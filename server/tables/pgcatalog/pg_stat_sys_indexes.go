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

// PgStatSysIndexesName is a constant to the pg_stat_sys_indexes name.
const PgStatSysIndexesName = "pg_stat_sys_indexes"

// InitPgStatSysIndexes handles registration of the pg_stat_sys_indexes handler.
func InitPgStatSysIndexes() {
	tables.AddHandler(PgCatalogName, PgStatSysIndexesName, PgStatSysIndexesHandler{})
}

// PgStatSysIndexesHandler is the handler for the pg_stat_sys_indexes table.
type PgStatSysIndexesHandler struct{}

var _ tables.Handler = PgStatSysIndexesHandler{}

// Name implements the interface tables.Handler.
func (p PgStatSysIndexesHandler) Name() string {
	return PgStatSysIndexesName
}

// RowIter implements the interface tables.Handler.
func (p PgStatSysIndexesHandler) RowIter(ctx *sql.Context, partition sql.Partition) (sql.RowIter, error) {
	// TODO: Implement pg_stat_sys_indexes row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgStatSysIndexesHandler) PkSchema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgStatSysIndexesSchema,
		PkOrdinals: nil,
	}
}

// pgStatSysIndexesSchema is the schema for pg_stat_sys_indexes.
var pgStatSysIndexesSchema = sql.Schema{
	{Name: "relid", Type: pgtypes.Oid, Default: nil, Nullable: true, Source: PgStatSysIndexesName},
	{Name: "indexrelid", Type: pgtypes.Oid, Default: nil, Nullable: true, Source: PgStatSysIndexesName},
	{Name: "schemaname", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgStatSysIndexesName},
	{Name: "relname", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgStatSysIndexesName},
	{Name: "indexrelname", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgStatSysIndexesName},
	{Name: "idx_scan", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatSysIndexesName},
	{Name: "last_idx_scan", Type: pgtypes.TimestampTZ, Default: nil, Nullable: true, Source: PgStatSysIndexesName},
	{Name: "idx_tup_read", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatSysIndexesName},
	{Name: "idx_tup_fetch", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatSysIndexesName},
}

// pgStatSysIndexesRowIter is the sql.RowIter for the pg_stat_sys_indexes table.
type pgStatSysIndexesRowIter struct {
}

var _ sql.RowIter = (*pgStatSysIndexesRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgStatSysIndexesRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgStatSysIndexesRowIter) Close(ctx *sql.Context) error {
	return nil
}
