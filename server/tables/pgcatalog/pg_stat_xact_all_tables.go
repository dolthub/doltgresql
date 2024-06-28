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

// PgStatXactAllTablesName is a constant to the pg_stat_xact_all_tables name.
const PgStatXactAllTablesName = "pg_stat_xact_all_tables"

// InitPgStatXactAllTables handles registration of the pg_stat_xact_all_tables handler.
func InitPgStatXactAllTables() {
	tables.AddHandler(PgCatalogName, PgStatXactAllTablesName, PgStatXactAllTablesHandler{})
}

// PgStatXactAllTablesHandler is the handler for the pg_stat_xact_all_tables table.
type PgStatXactAllTablesHandler struct{}

var _ tables.Handler = PgStatXactAllTablesHandler{}

// Name implements the interface tables.Handler.
func (p PgStatXactAllTablesHandler) Name() string {
	return PgStatXactAllTablesName
}

// RowIter implements the interface tables.Handler.
func (p PgStatXactAllTablesHandler) RowIter(ctx *sql.Context) (sql.RowIter, error) {
	// TODO: Implement pg_stat_xact_all_tables row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgStatXactAllTablesHandler) Schema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgStatXactAllTablesSchema,
		PkOrdinals: nil,
	}
}

// pgStatXactAllTablesSchema is the schema for pg_stat_xact_all_tables.
var pgStatXactAllTablesSchema = sql.Schema{
	{Name: "relid", Type: pgtypes.Oid, Default: nil, Nullable: true, Source: PgStatXactAllTablesName},
	{Name: "schemaname", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgStatXactAllTablesName},
	{Name: "relname", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgStatXactAllTablesName},
	{Name: "seq_scan", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatXactAllTablesName},
	{Name: "seq_tup_read", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatXactAllTablesName},
	{Name: "idx_scan", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatXactAllTablesName},
	{Name: "idx_tup_fetch", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatXactAllTablesName},
	{Name: "n_tup_ins", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatXactAllTablesName},
	{Name: "n_tup_upd", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatXactAllTablesName},
	{Name: "n_tup_del", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatXactAllTablesName},
	{Name: "n_tup_hot_upd", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatXactAllTablesName},
	{Name: "n_tup_newpage_upd", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatXactAllTablesName},
}

// pgStatXactAllTablesRowIter is the sql.RowIter for the pg_stat_xact_all_tables table.
type pgStatXactAllTablesRowIter struct {
}

var _ sql.RowIter = (*pgStatXactAllTablesRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgStatXactAllTablesRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgStatXactAllTablesRowIter) Close(ctx *sql.Context) error {
	return nil
}
