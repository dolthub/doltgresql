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
	"context"
	"runtime"
	"strings"
	"testing"

	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/dsess"
	denginetest "github.com/dolthub/dolt/go/libraries/doltcore/sqle/enginetest"
	gms "github.com/dolthub/go-mysql-server"
	"github.com/dolthub/go-mysql-server/enginetest"
	"github.com/dolthub/go-mysql-server/enginetest/scriptgen/setup"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/analyzer"
	"github.com/dolthub/vitess/go/vt/proto/query"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"
)

type DoltgresHarness struct {
	t                   *testing.T
	setupData           []setup.SetupScript
	resetData           []setup.SetupScript
	skippedQueries      []string
	parallelism int
	setupDbs            map[string]struct{}
	skipSetupCommit     bool
	configureStats      bool
	useLocalFilesystem  bool
	setupTestProcedures bool
}

func (d *DoltgresHarness) ValidateEngine(ctx *sql.Context, e *gms.Engine) error {
	panic("implement me")
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

func newDoltEnginetestHarness(t *testing.T) denginetest.DoltEnginetestHarness {
	return newDoltgresServerHarness(t)
}

var defaultSkippedQueries = []string{
	"show variables",             // we set extra variables
	"show create table fk_tbl",   // we create an extra key for the FK that vanilla gms does not
	"show indexes from",          // we create / expose extra indexes (for foreign keys)
	"show global variables like", // we set extra variables
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
	return &DoltgresQueryEngine{harness: d}, nil
}

func filterStatsOnlyQueries(scripts []setup.SetupScript) []setup.SetupScript {
	var ret []string
	for i := range scripts {
		for _, s := range scripts[i] {
			if strings.HasPrefix(s, "analyze table") {
				ret = append(ret, s)
			}
		}
	}
	return []setup.SetupScript{ret}
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
	for _, skipped := range d.skippedQueries {
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
	panic("implement me")
}

func (d *DoltgresHarness) NewContextWithClient(client sql.Client) *sql.Context {
	return sql.NewContext(context.Background(), sql.WithSession(d.newSessionWithClient(client)))
}

func (d *DoltgresHarness) NewSession() *sql.Context {
	panic("implement	me")
}

func (d *DoltgresHarness) newSessionWithClient(client sql.Client) *dsess.DoltSession {
	panic("implement me")
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
	harness *DoltgresHarness
}

func (d DoltgresQueryEngine) PrepareQuery(s *sql.Context, s2 string) (sql.Node, error) {
	panic("implement me")
}

func (d DoltgresQueryEngine) AnalyzeQuery(s *sql.Context, s2 string) (sql.Node, error) {
	panic("implement me")
}

func (d DoltgresQueryEngine) Query(ctx *sql.Context, query string) (sql.Schema, sql.RowIter, error) {
	panic("implement me")
}

func (d DoltgresQueryEngine) EngineAnalyzer() *analyzer.Analyzer {
	panic("implement me")
}

func (d DoltgresQueryEngine) EnginePreparedDataCache() *gms.PreparedDataCache {
	panic("implement me")
}

func (d DoltgresQueryEngine) QueryWithBindings(ctx *sql.Context, query string, parsed vitess.Statement, bindings map[string]*query.BindVariable) (sql.Schema, sql.RowIter, error) {
	panic("implement me")
}

func (d DoltgresQueryEngine) CloseSession(connID uint32) {
	panic("implement me")
}

func (d DoltgresQueryEngine) Close() error {
	panic("implement me")
}

var _ enginetest.QueryEngine = &DoltgresQueryEngine{}
