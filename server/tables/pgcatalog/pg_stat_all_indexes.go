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

// PgStatAllIndexesName is a constant to the pg_stat_all_indexes name.
const PgStatAllIndexesName = "pg_stat_all_indexes"

// InitPgStatAllIndexes handles registration of the pg_stat_all_indexes handler.
func InitPgStatAllIndexes() {
	tables.AddHandler(PgCatalogName, PgStatAllIndexesName, PgStatAllIndexesHandler{})
}

// PgStatAllIndexesHandler is the handler for the pg_stat_all_indexes table.
type PgStatAllIndexesHandler struct{}

var _ tables.Handler = PgStatAllIndexesHandler{}

// Name implements the interface tables.Handler.
func (p PgStatAllIndexesHandler) Name() string {
	return PgStatAllIndexesName
}

// RowIter implements the interface tables.Handler.
func (p PgStatAllIndexesHandler) RowIter(ctx *sql.Context, partition sql.Partition) (sql.RowIter, error) {
	entries, err := getStatIndexEntries(ctx, statSchemaAll)
	if err != nil {
		return nil, err
	}
	return &pgStatIndexesRowIter{entries: entries}, nil
}

// PkSchema implements the interface tables.Handler.
func (p PgStatAllIndexesHandler) PkSchema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgStatAllIndexesSchema,
		PkOrdinals: nil,
	}
}

// pgStatAllIndexesSchema is the schema for pg_stat_all_indexes.
var pgStatAllIndexesSchema = sql.Schema{
	{Name: "relid", Type: pgtypes.Oid, Default: nil, Nullable: true, Source: PgStatAllIndexesName},
	{Name: "indexrelid", Type: pgtypes.Oid, Default: nil, Nullable: true, Source: PgStatAllIndexesName},
	{Name: "schemaname", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgStatAllIndexesName},
	{Name: "relname", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgStatAllIndexesName},
	{Name: "indexrelname", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgStatAllIndexesName},
	{Name: "idx_scan", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatAllIndexesName},
	{Name: "last_idx_scan", Type: pgtypes.TimestampTZ, Default: nil, Nullable: true, Source: PgStatAllIndexesName},
	{Name: "idx_tup_read", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatAllIndexesName},
	{Name: "idx_tup_fetch", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatAllIndexesName},
}

// statIndexEntry identifies an index row in the pg_stat_*_indexes and pg_statio_*_indexes tables.
type statIndexEntry struct {
	tableOid   id.Id
	indexOid   id.Id
	schemaName string
	tableName  string
	indexName  string
	isSystem   bool
}

// getStatIndexEntries returns a statIndexEntry for each index on a table in the current database
// whose schema matches the given filter. The unfiltered entry list is cached in the session's
// pg_catalog cache, since iterating all schema elements is expensive.
func getStatIndexEntries(ctx *sql.Context, filter statSchemaFilter) ([]statIndexEntry, error) {
	pgCatalogCache, err := getPgCatalogCache(ctx)
	if err != nil {
		return nil, err
	}

	if pgCatalogCache.statIndexEntries == nil {
		var entries []statIndexEntry
		err := functions.IterateCurrentDatabase(ctx, functions.Callbacks{
			Index: func(ctx *sql.Context, schema functions.ItemSchema, table functions.ItemTable, index functions.ItemIndex) (cont bool, err error) {
				entries = append(entries, statIndexEntry{
					tableOid:   table.OID.AsId(),
					indexOid:   index.OID.AsId(),
					schemaName: schema.Item.SchemaName(),
					tableName:  table.Item.Name(),
					indexName:  formatIndexName(index.Item),
					isSystem:   schema.IsSystemSchema(),
				})
				return true, nil
			},
		})
		if err != nil {
			return nil, err
		}
		pgCatalogCache.statIndexEntries = entries
	}

	var filtered []statIndexEntry
	for _, entry := range pgCatalogCache.statIndexEntries {
		if filter(entry.isSystem) {
			filtered = append(filtered, entry)
		}
	}
	return filtered, nil
}

// pgStatIndexesRowIter is the sql.RowIter for the pg_stat_all_indexes, pg_stat_sys_indexes, and
// pg_stat_user_indexes tables. All statistics counters are zero and all timestamps are NULL, since
// Doltgres does not track index access statistics. This matches what a freshly-started Postgres
// server reports.
// TODO: fill in the statistics columns when index access statistics are tracked
type pgStatIndexesRowIter struct {
	entries []statIndexEntry
	idx     int
}

var _ sql.RowIter = (*pgStatIndexesRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgStatIndexesRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	if iter.idx >= len(iter.entries) {
		return nil, io.EOF
	}
	iter.idx++
	entry := iter.entries[iter.idx-1]
	return sql.Row{
		entry.tableOid,   // relid
		entry.indexOid,   // indexrelid
		entry.schemaName, // schemaname
		entry.tableName,  // relname
		entry.indexName,  // indexrelname
		int64(0),         // idx_scan
		nil,              // last_idx_scan
		int64(0),         // idx_tup_read
		int64(0),         // idx_tup_fetch
	}, nil
}

// Close implements the interface sql.RowIter.
func (iter *pgStatIndexesRowIter) Close(ctx *sql.Context) error {
	return nil
}
