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

// PgStatioSysTablesName is a constant to the pg_statio_sys_tables name.
const PgStatioSysTablesName = "pg_statio_sys_tables"

// InitPgStatioSysTables handles registration of the pg_statio_sys_tables handler.
func InitPgStatioSysTables() {
	tables.AddHandler(PgCatalogName, PgStatioSysTablesName, PgStatioSysTablesHandler{})
}

// PgStatioSysTablesHandler is the handler for the pg_statio_sys_tables table.
type PgStatioSysTablesHandler struct{}

var _ tables.Handler = PgStatioSysTablesHandler{}

// Name implements the interface tables.Handler.
func (p PgStatioSysTablesHandler) Name() string {
	return PgStatioSysTablesName
}

// RowIter implements the interface tables.Handler.
func (p PgStatioSysTablesHandler) RowIter(ctx *sql.Context) (sql.RowIter, error) {
	// TODO: Implement pg_statio_sys_tables row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgStatioSysTablesHandler) Schema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgStatioSysTablesSchema,
		PkOrdinals: nil,
	}
}

// pgStatioSysTablesSchema is the schema for pg_statio_sys_tables.
var pgStatioSysTablesSchema = sql.Schema{
	{Name: "relid", Type: pgtypes.Oid, Default: nil, Nullable: true, Source: PgStatioSysTablesName},
	{Name: "schemaname", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgStatioSysTablesName},
	{Name: "relname", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgStatioSysTablesName},
	{Name: "heap_blks_read", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatioSysTablesName},
	{Name: "heap_blks_hit", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatioSysTablesName},
	{Name: "idx_blks_read", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatioSysTablesName},
	{Name: "idx_blks_hit", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatioSysTablesName},
	{Name: "toast_blks_read", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatioSysTablesName},
	{Name: "toast_blks_hit", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatioSysTablesName},
	{Name: "tidx_blks_read", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatioSysTablesName},
	{Name: "tidx_blks_hit", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatioSysTablesName},
}

// pgStatioSysTablesRowIter is the sql.RowIter for the pg_statio_sys_tables table.
type pgStatioSysTablesRowIter struct {
}

var _ sql.RowIter = (*pgStatioSysTablesRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgStatioSysTablesRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgStatioSysTablesRowIter) Close(ctx *sql.Context) error {
	return nil
}
