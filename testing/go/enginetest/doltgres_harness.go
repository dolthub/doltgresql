// Copyright 2020 Dolthub, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package enginetest

import (
	gosql "database/sql"
	"fmt"
	"io"
	"regexp"
	"runtime"
	"strings"
	"testing"

	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/dsess"
	denginetest "github.com/dolthub/dolt/go/libraries/doltcore/sqle/enginetest"
	"github.com/dolthub/dolt/go/libraries/utils/svcs"
	gms "github.com/dolthub/go-mysql-server"
	"github.com/dolthub/go-mysql-server/enginetest"
	"github.com/dolthub/go-mysql-server/enginetest/scriptgen/setup"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/analyzer"
	gmstypes "github.com/dolthub/go-mysql-server/sql/types"
	"github.com/dolthub/vitess/go/vt/proto/query"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/stretchr/testify/require"

	"github.com/dolthub/doltgresql/server"
	"github.com/dolthub/doltgresql/servercfg"
)

type DoltgresHarness struct {
	t                  *testing.T
	setupData          []setup.SetupScript
	skippedQueries     []string
	parallelism        int
	skipSetupCommit    bool
	configureStats     bool
	useLocalFilesystem bool
}

func (d *DoltgresHarness) ValidateEngine(ctx *sql.Context, e *gms.Engine) error {
	// TODO
	return nil
}

func (d *DoltgresHarness) UseLocalFileSystem() {
	d.useLocalFilesystem = true
}

func (d *DoltgresHarness) Session() *dsess.DoltSession {
	panic("implement me")
}

func (d *DoltgresHarness) WithConfigureStats(configureStats bool) denginetest.DoltEnginetestHarness {
	nd := *d
	nd.configureStats = configureStats
	return &nd
}

func (d *DoltgresHarness) NewHarness(t *testing.T) denginetest.DoltEnginetestHarness {
	return newDoltgresServerHarness(t)
}

var _ denginetest.DoltEnginetestHarness = &DoltgresHarness{}

// newDoltgresServerHarness creates a new harness for testing Dolt, using an in-memory filesystem and an in-memory blob store.
func newDoltgresServerHarness(t *testing.T) denginetest.DoltEnginetestHarness {
	dh := &DoltgresHarness{
		t:              t,
		skippedQueries: defaultSkippedQueries,
	}

	return dh
}

var defaultSkippedQueries = []string{
	"show variables",             // we set extra variables
	"show create table fk_tbl",   // we create an extra key for the FK that vanilla gms does not
	"show indexes from",          // we create / expose extra indexes (for foreign keys)
	"show global variables like", // we set extra variables
	// unsupported doltgres syntax
	"WITH",
	"OVER",
	// subqueries are broken, breaks with an index out of bounds error
	"(SELECT",
	// string functions are broken due to incompatible types
	"HEX(",
	"TO_BASE64(",
}

// Setup sets the setup scripts for this DoltHarness's engine
func (d *DoltgresHarness) Setup(setupData ...[]setup.SetupScript) {
	d.setupData = nil
	for i := range setupData {
		d.setupData = append(d.setupData, setupData[i]...)
	}
}

func (d *DoltgresHarness) SkipSetupCommit() {
	d.skipSetupCommit = true
}

// NewEngine creates a new *gms.Engine or calls reset and clear scripts on the existing
// engine for reuse.
func (d *DoltgresHarness) NewEngine(t *testing.T) (enginetest.QueryEngine, error) {
	queryEngine := NewDoltgresQueryEngine(t, d)

	ctx := d.NewContext()
	for _, setupScript := range d.setupData {
		for _, s := range setupScript {
			runQuery, sanitized := sanitizeQuery(s)
			if !runQuery {
				t.Log("Skipping setup query: ", s)
				continue
			} else {
				t.Log("Running setup query: ", s)
			}
			_, rowIter, err := queryEngine.Query(ctx, sanitized)
			if err != nil {
				return nil, err
			}
			err = drainIter(ctx, rowIter)
			if err != nil {
				return nil, err
			}
		}
	}

	return queryEngine, nil
}

var skippedSetupWords = []string{
	"auto_increment",
	"create index",
	"bigtable",           // "ERROR: blob/text column 't' used in key specification without a key length"
	"typestable",         // lots of work to do
	"datetime_table",     // invalid timestamp format
	"specialtable",       // invalid quoting
	"people",             // ERROR: blob/text column 'first_name' used in key specification without a key length
	"reservedWordsTable", // ERROR: blob/text column 'Timestamp' used in key specification without a key length
	"foo.othertable",     // ERROR: database schema not found: foo (errno 1105)
	"bus_routes",         // ERROR: blob/text column 'origin' used in key specification without a key length
	"parts",              // ERROR: blob/text column 'part' used in key specification without a key length
	"xy_hasnull_idx",     // needs an index during creation
	"xy ",                // needs an index during creation
	"rs ",                // needs an index during creation
	"analyze table",      // unsupported syntax
}

var commentClause = regexp.MustCompile(`(?i)comment '.*?'`)
var createIndexStatement = regexp.MustCompile(`(?i)create.*?index`)
var alterTableStatement = regexp.MustCompile(`(?i)alter table`)
var createTableStatement = regexp.MustCompile(`(?i)create table`)
var floatKeyword = regexp.MustCompile(`(?i)\bfloat\b`)
var doubleKeyword = regexp.MustCompile(`(?i)\bdouble\b`)
var datetimeKeyword = regexp.MustCompile(`(?i)\bdatetime\b`)
var mediumIntKeyword = regexp.MustCompile(`(?i)\bmediumint\b`)
var tinyIntKeyword = regexp.MustCompile(`(?i)\btinyint\b`)
var backtick = "`"

// sanitizeQuery strips the query string given of any unsupported constructs without attempting to actually convert
// to Postgres syntax.
func sanitizeQuery(s string) (bool, string) {
	for _, word := range skippedSetupWords {
		if strings.Contains(s, word) {
			return false, ""
		}
	}

	if createIndexStatement.MatchString(s) {
		return false, ""
	}
	if alterTableStatement.MatchString(s) {
		return false, ""
	}

	if createTableStatement.MatchString(s) {
		s = replaceTypes(s)
	}

	s = commentClause.ReplaceAllString(s, "")
	s = strings.ReplaceAll(s, backtick, `"`)

	return true, s
}

func replaceTypes(s string) string {
	s = floatKeyword.ReplaceAllString(s, "real")
	s = doubleKeyword.ReplaceAllString(s, "double precision")
	s = datetimeKeyword.ReplaceAllString(s, "timestamp")
	s = mediumIntKeyword.ReplaceAllString(s, "integer")
	s = tinyIntKeyword.ReplaceAllString(s, "smallint")
	return s
}

func drainIter(ctx *sql.Context, rowIter sql.RowIter) error {
	for {
		_, err := rowIter.Next(ctx)
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}
	}
	return rowIter.Close(ctx)
}

// WithParallelism returns a copy of the harness with parallelism set to the given number of threads. A value of 0 or
// less means to use the system parallelism settings.
func (d *DoltgresHarness) WithParallelism(parallelism int) denginetest.DoltEnginetestHarness {
	nd := *d
	nd.parallelism = parallelism
	return &nd
}

// WithSkippedQueries returns a copy of the harness with the given queries skipped
func (d *DoltgresHarness) WithSkippedQueries(queries []string) denginetest.DoltEnginetestHarness {
	nd := *d
	nd.skippedQueries = append(d.skippedQueries, queries...)
	return &nd
}

func (d *DoltgresHarness) Engine() *gms.Engine {
	panic("implement me")
}

// SkipQueryTest returns whether to skip a query
func (d *DoltgresHarness) SkipQueryTest(query string) bool {
	lowerQuery := strings.ToLower(query)
	for _, skipped := range append(d.skippedQueries, skippedSetupWords...) {
		if strings.Contains(lowerQuery, strings.ToLower(skipped)) {
			return true
		}
	}

	return false
}

func (d *DoltgresHarness) Parallelism() int {
	if d.parallelism <= 0 {

		// always test with some parallelism
		parallelism := runtime.NumCPU()

		if parallelism <= 1 {
			parallelism = 2
		}

		return parallelism
	}

	return d.parallelism
}

func (d *DoltgresHarness) NewContext() *sql.Context {
	return sql.NewEmptyContext()
}

func (d *DoltgresHarness) NewContextWithClient(client sql.Client) *sql.Context {
	// unused for now, linter is complaining
	// return sql.NewContext(context.Background(), sql.WithSession(d.newSessionWithClient(client)))
	panic("implement me")
}

func (d *DoltgresHarness) NewSession() *sql.Context {
	panic("implement	me")
}

func (d *DoltgresHarness) SupportsNativeIndexCreation() bool {
	return true
}

func (d *DoltgresHarness) SupportsForeignKeys() bool {
	return true
}

func (d *DoltgresHarness) SupportsKeylessTables() bool {
	return true
}

func (d *DoltgresHarness) NewDatabases(names ...string) []sql.Database {
	panic("implement me")
}

func (d *DoltgresHarness) NewReadOnlyEngine(provider sql.DatabaseProvider) (enginetest.QueryEngine, error) {
	panic("implement me")
}

func (d *DoltgresHarness) NewDatabaseProvider() sql.MutableDatabaseProvider {
	panic("implement me")
}

func (d *DoltgresHarness) Close() {
}

// NewTableAsOf implements enginetest.VersionedHarness
// Dolt doesn't version tables per se, just the entire database. So ignore the name and schema and just create a new
// branch with the given name.
func (d *DoltgresHarness) NewTableAsOf(db sql.VersionedDatabase, name string, schema sql.PrimaryKeySchema, asOf interface{}) sql.Table {
	panic("implement me")
}

// SnapshotTable implements enginetest.VersionedHarness
// Dolt doesn't version tables per se, just the entire database. So ignore the name and schema and just create a new
// branch with the given name.
func (d *DoltgresHarness) SnapshotTable(db sql.VersionedDatabase, tableName string, asOf interface{}) error {
	panic("implement me")
}

type DoltgresQueryEngine struct {
	harness    *DoltgresHarness
	controller *svcs.Controller
}

// Ptr is a helper function that returns a pointer to the value passed in. This is necessary to e.g. get a pointer to
// a const value without assigning to an intermediate variable.
func Ptr[T any](v T) *T {
	return &v
}

func NewDoltgresQueryEngine(t *testing.T, harness *DoltgresHarness) *DoltgresQueryEngine {
	ctrl, err := server.RunInMemory(&servercfg.DoltgresConfig{
		LogLevelStr: Ptr("debug"),
	})
	require.NoError(t, err)
	return &DoltgresQueryEngine{
		harness:    harness,
		controller: ctrl,
	}
}

func (d DoltgresQueryEngine) PrepareQuery(s *sql.Context, s2 string) (sql.Node, error) {
	panic("implement me")
}

func (d DoltgresQueryEngine) AnalyzeQuery(s *sql.Context, s2 string) (sql.Node, error) {
	panic("implement me")
}

// TODO: random port
var doltgresNoDbDsn = "postgresql://doltgres:password@127.0.0.1:5432/?sslmode=disable"

func (d DoltgresQueryEngine) Query(ctx *sql.Context, query string) (sql.Schema, sql.RowIter, error) {
	db, err := gosql.Open("pgx", doltgresNoDbDsn)
	if err != nil {
		return nil, nil, err
	}

	rows, err := db.Query(query)
	if rows != nil {
		defer rows.Close()
	}

	if err != nil {
		return nil, nil, err
	}

	schema, columns, err := columns(rows)
	if err != nil {
		return nil, nil, err
	}

	results := make([]sql.Row, 0)
	for rows.Next() {
		err := rows.Scan(columns...)
		if err != nil {
			return nil, nil, err
		}

		row, err := toRow(schema, columns)
		if err != nil {
			return nil, nil, err
		}

		results = append(results, row)
	}

	if rows.Err() != nil {
		return nil, nil, rows.Err()
	}

	return schema, sql.RowsToRowIter(results...), nil
}

func toRow(schema sql.Schema, r []interface{}) (sql.Row, error) {
	row := make(sql.Row, len(schema))
	for i, col := range schema {
		val, err := unwrapResultColumn(r[i])
		if err != nil {
			return nil, err
		}

		row[i], _, err = col.Type.Convert(val)
		if err != nil {
			return nil, err
		}
	}
	return row, nil
}

func unwrapResultColumn(v any) (any, error) {
	switch v := v.(type) {
	case *gosql.NullBool:
		if v.Valid {
			return v.Bool, nil
		}
		return nil, nil
	case *gosql.NullString:
		if v.Valid {
			return v.String, nil
		}
		return nil, nil
	case *gosql.NullFloat64:
		if v.Valid {
			return v.Float64, nil
		}
		return nil, nil
	case *gosql.NullInt64:
		if v.Valid {
			return v.Int64, nil
		}
		return nil, nil
	case *gosql.NullTime:
		if v.Valid {
			return v.Time, nil
		}
		return nil, nil
	case *gosql.NullInt32:
		if v.Valid {
			return v.Int32, nil
		}
		return nil, nil
	case *gosql.NullInt16:
		if v.Valid {
			return v.Int16, nil
		}
		return nil, nil
	default:
		return nil, fmt.Errorf("unsupported type %T", v)
	}
}

func (d DoltgresQueryEngine) EngineAnalyzer() *analyzer.Analyzer {
	panic("implement me")
}

func (d DoltgresQueryEngine) EnginePreparedDataCache() *gms.PreparedDataCache {
	panic("implement me")
}

func (d DoltgresQueryEngine) QueryWithBindings(ctx *sql.Context, query string, parsed vitess.Statement, bindings map[string]*query.BindVariable) (sql.Schema, sql.RowIter, error) {
	if len(bindings) > 0 {
		return nil, nil, fmt.Errorf("bindings not supported")
	}

	return d.Query(ctx, query)
}

func (d DoltgresQueryEngine) CloseSession(connID uint32) {
	panic("implement me")
}

func (d DoltgresQueryEngine) Close() error {
	d.controller.Stop()
	return d.controller.WaitForStop()
}

var _ enginetest.QueryEngine = &DoltgresQueryEngine{}

func columns(rows *gosql.Rows) (sql.Schema, []interface{}, error) {
	types, err := rows.ColumnTypes()
	if err != nil {
		return nil, nil, err
	}

	schema := make(sql.Schema, 0, len(types))
	columnVals := make([]interface{}, 0, len(types))

	for _, columnType := range types {
		switch columnType.DatabaseTypeName() {
		case "BIT":
			colVal := gosql.NullBool{}
			columnVals = append(columnVals, &colVal)
			schema = append(schema, &sql.Column{Name: columnType.Name(), Type: gmstypes.Int8, Nullable: true})
		case "TEXT", "VARCHAR", "MEDIUMTEXT", "CHAR", "TINYTEXT", "NAME", "BYTEA":
			colVal := gosql.NullString{}
			columnVals = append(columnVals, &colVal)
			schema = append(schema, &sql.Column{Name: columnType.Name(), Type: gmstypes.LongText, Nullable: true})
		case "DECIMAL", "DOUBLE", "FLOAT", "FLOAT4", "FLOAT8", "NUMERIC":
			colVal := gosql.NullFloat64{}
			columnVals = append(columnVals, &colVal)
			schema = append(schema, &sql.Column{Name: columnType.Name(), Type: gmstypes.Float64, Nullable: true})
		case "MEDIUMINT", "INT", "BIGINT", "TINYINT", "SMALLINT", "INT2", "INT4", "INT8":
			colVal := gosql.NullInt64{}
			columnVals = append(columnVals, &colVal)
			schema = append(schema, &sql.Column{Name: columnType.Name(), Type: gmstypes.Int64, Nullable: true})
		default:
			return nil, nil, fmt.Errorf("Unhandled type %s", columnType.DatabaseTypeName())
		}
	}

	return schema, columnVals, nil
}
