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

// PgStatXactSysTablesName is a constant to the pg_stat_xact_sys_tables name.
const PgStatXactSysTablesName = "pg_stat_xact_sys_tables"

// InitPgStatXactSysTables handles registration of the pg_stat_xact_sys_tables handler.
func InitPgStatXactSysTables() {
	tables.AddHandler(PgCatalogName, PgStatXactSysTablesName, PgStatXactSysTablesHandler{})
}

// PgStatXactSysTablesHandler is the handler for the pg_stat_xact_sys_tables table.
type PgStatXactSysTablesHandler struct{}

var _ tables.Handler = PgStatXactSysTablesHandler{}

// Name implements the interface tables.Handler.
func (p PgStatXactSysTablesHandler) Name() string {
	return PgStatXactSysTablesName
}

// RowIter implements the interface tables.Handler.
func (p PgStatXactSysTablesHandler) RowIter(ctx *sql.Context, partition sql.Partition) (sql.RowIter, error) {
	// TODO: Implement pg_stat_xact_sys_tables row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgStatXactSysTablesHandler) Schema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgStatXactSysTablesSchema,
		PkOrdinals: nil,
	}
}

// pgStatXactSysTablesSchema is the schema for pg_stat_xact_sys_tables.
var pgStatXactSysTablesSchema = sql.Schema{
	{Name: "relid", Type: pgtypes.Oid, Default: nil, Nullable: true, Source: PgStatXactSysTablesName},
	{Name: "schemaname", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgStatXactSysTablesName},
	{Name: "relname", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgStatXactSysTablesName},
	{Name: "seq_scan", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatXactSysTablesName},
	{Name: "seq_tup_read", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatXactSysTablesName},
	{Name: "idx_scan", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatXactSysTablesName},
	{Name: "idx_tup_fetch", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatXactSysTablesName},
	{Name: "n_tup_ins", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatXactSysTablesName},
	{Name: "n_tup_upd", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatXactSysTablesName},
	{Name: "n_tup_del", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatXactSysTablesName},
	{Name: "n_tup_hot_upd", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatXactSysTablesName},
	{Name: "n_tup_newpage_upd", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatXactSysTablesName},
}

// pgStatXactSysTablesRowIter is the sql.RowIter for the pg_stat_xact_sys_tables table.
type pgStatXactSysTablesRowIter struct {
}

var _ sql.RowIter = (*pgStatXactSysTablesRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgStatXactSysTablesRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgStatXactSysTablesRowIter) Close(ctx *sql.Context) error {
	return nil
}
