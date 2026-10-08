// Copyright 2024 Dolthub, Inc.
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

package pgcatalog

import (
	"io"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/server/functions"
	"github.com/dolthub/doltgresql/server/tables"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// PgStatAllTablesName is a constant to the pg_stat_all_tables name.
const PgStatAllTablesName = "pg_stat_all_tables"

// InitPgStatAllTables handles registration of the pg_stat_all_tables handler.
func InitPgStatAllTables() {
	tables.AddHandler(PgCatalogName, PgStatAllTablesName, PgStatAllTablesHandler{})
}

// PgStatAllTablesHandler is the handler for the pg_stat_all_tables table.
type PgStatAllTablesHandler struct{}

var _ tables.Handler = PgStatAllTablesHandler{}

// Name implements the interface tables.Handler.
func (p PgStatAllTablesHandler) Name() string {
	return PgStatAllTablesName
}

// RowIter implements the interface tables.Handler.
func (p PgStatAllTablesHandler) RowIter(ctx *sql.Context, partition sql.Partition) (sql.RowIter, error) {
	entries, err := getStatTableEntries(ctx, statSchemaAll)
	if err != nil {
		return nil, err
	}
	return &pgStatTablesRowIter{entries: entries}, nil
}

// PkSchema implements the interface tables.Handler.
func (p PgStatAllTablesHandler) PkSchema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgStatAllTablesSchema,
		PkOrdinals: nil,
	}
}

// pgStatAllTablesSchema is the schema for pg_stat_all_tables.
var pgStatAllTablesSchema = sql.Schema{
	{Name: "relid", Type: pgtypes.Oid, Default: nil, Nullable: true, Source: PgStatAllTablesName},
	{Name: "schemaname", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgStatAllTablesName},
	{Name: "relname", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgStatAllTablesName},
	{Name: "seq_scan", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatAllTablesName},
	{Name: "last_seq_scan", Type: pgtypes.TimestampTZ, Default: nil, Nullable: true, Source: PgStatAllTablesName},
	{Name: "seq_tup_read", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatAllTablesName},
	{Name: "idx_scan", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatAllTablesName},
	{Name: "last_idx_scan", Type: pgtypes.TimestampTZ, Default: nil, Nullable: true, Source: PgStatAllTablesName},
	{Name: "idx_tup_fetch", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatAllTablesName},
	{Name: "n_tup_ins", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatAllTablesName},
	{Name: "n_tup_upd", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatAllTablesName},
	{Name: "n_tup_del", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatAllTablesName},
	{Name: "n_tup_hot_upd", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatAllTablesName},
	{Name: "n_tup_newpage_upd", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatAllTablesName},
	{Name: "n_live_tup", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatAllTablesName},
	{Name: "n_dead_tup", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatAllTablesName},
	{Name: "n_mod_since_analyze", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatAllTablesName},
	{Name: "n_ins_since_vacuum", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatAllTablesName},
	{Name: "last_vacuum", Type: pgtypes.TimestampTZ, Default: nil, Nullable: true, Source: PgStatAllTablesName},
	{Name: "last_autovacuum", Type: pgtypes.TimestampTZ, Default: nil, Nullable: true, Source: PgStatAllTablesName},
	{Name: "last_analyze", Type: pgtypes.TimestampTZ, Default: nil, Nullable: true, Source: PgStatAllTablesName},
	{Name: "last_autoanalyze", Type: pgtypes.TimestampTZ, Default: nil, Nullable: true, Source: PgStatAllTablesName},
	{Name: "vacuum_count", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatAllTablesName},
	{Name: "autovacuum_count", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatAllTablesName},
	{Name: "analyze_count", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatAllTablesName},
	{Name: "autoanalyze_count", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatAllTablesName},
}

// statTableEntry identifies a table row in the pg_stat_*_tables and pg_stat_xact_*_tables tables.
type statTableEntry struct {
	oid        id.Id
	schemaName string
	tableName  string
	isSystem   bool
}

// statSchemaFilter determines which entries are included in a pg_stat_* table, based on whether the
// schema they belong to is a system schema. It is shared by the table, index, and sequence stat views.
type statSchemaFilter func(isSystem bool) bool

// statSchemaAll includes entries from all schemas.
func statSchemaAll(isSystem bool) bool {
	return true
}

// statSchemaSys includes only entries from system schemas.
func statSchemaSys(isSystem bool) bool {
	return isSystem
}

// statSchemaUser includes only entries from non-system schemas.
func statSchemaUser(isSystem bool) bool {
	return !isSystem
}

// getStatTableEntries returns a statTableEntry for each table in the current database whose schema
// matches the given filter. The unfiltered entry list is cached in the session's pg_catalog cache,
// since iterating all schema elements is expensive.
func getStatTableEntries(ctx *sql.Context, filter statSchemaFilter) ([]statTableEntry, error) {
	pgCatalogCache, err := getPgCatalogCache(ctx)
	if err != nil {
		return nil, err
	}

	if pgCatalogCache.statTableEntries == nil {
		var entries []statTableEntry
		err := functions.IterateCurrentDatabase(ctx, functions.Callbacks{
			Table: func(ctx *sql.Context, schema functions.ItemSchema, table functions.ItemTable) (cont bool, err error) {
				entries = append(entries, statTableEntry{
					oid:        table.OID.AsId(),
					schemaName: schema.Item.SchemaName(),
					tableName:  table.Item.Name(),
					isSystem:   schema.IsSystemSchema(),
				})
				return true, nil
			},
		})
		if err != nil {
			return nil, err
		}
		pgCatalogCache.statTableEntries = entries
	}

	var filtered []statTableEntry
	for _, entry := range pgCatalogCache.statTableEntries {
		if filter(entry.isSystem) {
			filtered = append(filtered, entry)
		}
	}
	return filtered, nil
}

// pgStatTablesRowIter is the sql.RowIter for the pg_stat_all_tables, pg_stat_sys_tables, and
// pg_stat_user_tables tables. All statistics counters are zero and all timestamps are NULL, since
// Doltgres does not track table access statistics. This matches what a freshly-started Postgres
// server reports.
// TODO: fill in the statistics columns when table access statistics are tracked
type pgStatTablesRowIter struct {
	entries []statTableEntry
	idx     int
}

var _ sql.RowIter = (*pgStatTablesRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgStatTablesRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	if iter.idx >= len(iter.entries) {
		return nil, io.EOF
	}
	iter.idx++
	entry := iter.entries[iter.idx-1]
	return sql.Row{
		entry.oid,        // relid
		entry.schemaName, // schemaname
		entry.tableName,  // relname
		int64(0),         // seq_scan
		nil,              // last_seq_scan
		int64(0),         // seq_tup_read
		int64(0),         // idx_scan
		nil,              // last_idx_scan
		int64(0),         // idx_tup_fetch
		int64(0),         // n_tup_ins
		int64(0),         // n_tup_upd
		int64(0),         // n_tup_del
		int64(0),         // n_tup_hot_upd
		int64(0),         // n_tup_newpage_upd
		int64(0),         // n_live_tup
		int64(0),         // n_dead_tup
		int64(0),         // n_mod_since_analyze
		int64(0),         // n_ins_since_vacuum
		nil,              // last_vacuum
		nil,              // last_autovacuum
		nil,              // last_analyze
		nil,              // last_autoanalyze
		int64(0),         // vacuum_count
		int64(0),         // autovacuum_count
		int64(0),         // analyze_count
		int64(0),         // autoanalyze_count
	}, nil
}

// Close implements the interface sql.RowIter.
func (iter *pgStatTablesRowIter) Close(ctx *sql.Context) error {
	return nil
}
