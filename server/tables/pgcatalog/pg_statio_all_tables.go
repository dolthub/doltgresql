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

// PgStatioAllTablesName is a constant to the pg_statio_all_tables name.
const PgStatioAllTablesName = "pg_statio_all_tables"

// InitPgStatioAllTables handles registration of the pg_statio_all_tables handler.
func InitPgStatioAllTables() {
	tables.AddHandler(PgCatalogName, PgStatioAllTablesName, PgStatioAllTablesHandler{})
}

// PgStatioAllTablesHandler is the handler for the pg_statio_all_tables table.
type PgStatioAllTablesHandler struct{}

var _ tables.Handler = PgStatioAllTablesHandler{}

// Name implements the interface tables.Handler.
func (p PgStatioAllTablesHandler) Name() string {
	return PgStatioAllTablesName
}

// RowIter implements the interface tables.Handler.
func (p PgStatioAllTablesHandler) RowIter(ctx *sql.Context, partition sql.Partition) (sql.RowIter, error) {
	// TODO: Implement pg_statio_all_tables row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgStatioAllTablesHandler) Schema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgStatioAllTablesSchema,
		PkOrdinals: nil,
	}
}

// pgStatioAllTablesSchema is the schema for pg_statio_all_tables.
var pgStatioAllTablesSchema = sql.Schema{
	{Name: "relid", Type: pgtypes.Oid, Default: nil, Nullable: true, Source: PgStatioAllTablesName},
	{Name: "schemaname", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgStatioAllTablesName},
	{Name: "relname", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgStatioAllTablesName},
	{Name: "heap_blks_read", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatioAllTablesName},
	{Name: "heap_blks_hit", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatioAllTablesName},
	{Name: "idx_blks_read", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatioAllTablesName},
	{Name: "idx_blks_hit", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatioAllTablesName},
	{Name: "toast_blks_read", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatioAllTablesName},
	{Name: "toast_blks_hit", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatioAllTablesName},
	{Name: "tidx_blks_read", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatioAllTablesName},
	{Name: "tidx_blks_hit", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatioAllTablesName},
}

// pgStatioAllTablesRowIter is the sql.RowIter for the pg_statio_all_tables table.
type pgStatioAllTablesRowIter struct {
}

var _ sql.RowIter = (*pgStatioAllTablesRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgStatioAllTablesRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgStatioAllTablesRowIter) Close(ctx *sql.Context) error {
	return nil
}
