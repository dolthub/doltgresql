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

	"github.com/dolthub/doltgresql/server/tables"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// PgStatioAllIndexesName is a constant to the pg_statio_all_indexes name.
const PgStatioAllIndexesName = "pg_statio_all_indexes"

// InitPgStatioAllIndexes handles registration of the pg_statio_all_indexes handler.
func InitPgStatioAllIndexes() {
	tables.AddHandler(PgCatalogName, PgStatioAllIndexesName, PgStatioAllIndexesHandler{})
}

// PgStatioAllIndexesHandler is the handler for the pg_statio_all_indexes table.
type PgStatioAllIndexesHandler struct{}

var _ tables.Handler = PgStatioAllIndexesHandler{}

// Name implements the interface tables.Handler.
func (p PgStatioAllIndexesHandler) Name() string {
	return PgStatioAllIndexesName
}

// RowIter implements the interface tables.Handler.
func (p PgStatioAllIndexesHandler) RowIter(ctx *sql.Context, partition sql.Partition) (sql.RowIter, error) {
	entries, err := getStatIndexEntries(ctx, statIndexesAll)
	if err != nil {
		return nil, err
	}
	return &pgStatioIndexesRowIter{entries: entries}, nil
}

// PkSchema implements the interface tables.Handler.
func (p PgStatioAllIndexesHandler) PkSchema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgStatioAllIndexesSchema,
		PkOrdinals: nil,
	}
}

// pgStatioAllIndexesSchema is the schema for pg_statio_all_indexes.
var pgStatioAllIndexesSchema = sql.Schema{
	{Name: "relid", Type: pgtypes.Oid, Default: nil, Nullable: true, Source: PgStatioAllIndexesName},
	{Name: "indexrelid", Type: pgtypes.Oid, Default: nil, Nullable: true, Source: PgStatioAllIndexesName},
	{Name: "schemaname", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgStatioAllIndexesName},
	{Name: "relname", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgStatioAllIndexesName},
	{Name: "indexrelname", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgStatioAllIndexesName},
	{Name: "idx_blks_read", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatioAllIndexesName},
	{Name: "idx_blks_hit", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatioAllIndexesName},
}

// pgStatioIndexesRowIter is the sql.RowIter for the pg_statio_all_indexes, pg_statio_sys_indexes,
// and pg_statio_user_indexes tables. All I/O counters are zero, since Doltgres does not track
// index block I/O statistics. This matches what a freshly-started Postgres server reports.
// TODO: fill in the statistics columns when index I/O statistics are tracked
type pgStatioIndexesRowIter struct {
	entries []statIndexEntry
	idx     int
}

var _ sql.RowIter = (*pgStatioIndexesRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgStatioIndexesRowIter) Next(ctx *sql.Context) (sql.Row, error) {
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
		int64(0),         // idx_blks_read
		int64(0),         // idx_blks_hit
	}, nil
}

// Close implements the interface sql.RowIter.
func (iter *pgStatioIndexesRowIter) Close(ctx *sql.Context) error {
	return nil
}
