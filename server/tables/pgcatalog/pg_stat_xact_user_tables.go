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

// PgStatXactUserTablesName is a constant to the pg_stat_xact_user_tables name.
const PgStatXactUserTablesName = "pg_stat_xact_user_tables"

// InitPgStatXactUserTables handles registration of the pg_stat_xact_user_tables handler.
func InitPgStatXactUserTables() {
	tables.AddHandler(PgCatalogName, PgStatXactUserTablesName, PgStatXactUserTablesHandler{})
}

// PgStatXactUserTablesHandler is the handler for the pg_stat_xact_user_tables table.
type PgStatXactUserTablesHandler struct{}

var _ tables.Handler = PgStatXactUserTablesHandler{}

// Name implements the interface tables.Handler.
func (p PgStatXactUserTablesHandler) Name() string {
	return PgStatXactUserTablesName
}

// RowIter implements the interface tables.Handler.
func (p PgStatXactUserTablesHandler) RowIter(ctx *sql.Context, partition sql.Partition) (sql.RowIter, error) {
	// TODO: Implement pg_stat_xact_user_tables row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgStatXactUserTablesHandler) PkSchema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgStatXactUserTablesSchema,
		PkOrdinals: nil,
	}
}

// pgStatXactUserTablesSchema is the schema for pg_stat_xact_user_tables.
var pgStatXactUserTablesSchema = sql.Schema{
	{Name: "relid", Type: pgtypes.Oid, Default: nil, Nullable: true, Source: PgStatXactUserTablesName},
	{Name: "schemaname", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgStatXactUserTablesName},
	{Name: "relname", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgStatXactUserTablesName},
	{Name: "seq_scan", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatXactUserTablesName},
	{Name: "seq_tup_read", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatXactUserTablesName},
	{Name: "idx_scan", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatXactUserTablesName},
	{Name: "idx_tup_fetch", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatXactUserTablesName},
	{Name: "n_tup_ins", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatXactUserTablesName},
	{Name: "n_tup_upd", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatXactUserTablesName},
	{Name: "n_tup_del", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatXactUserTablesName},
	{Name: "n_tup_hot_upd", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatXactUserTablesName},
	{Name: "n_tup_newpage_upd", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatXactUserTablesName},
}

// pgStatXactUserTablesRowIter is the sql.RowIter for the pg_stat_xact_user_tables table.
type pgStatXactUserTablesRowIter struct {
}

var _ sql.RowIter = (*pgStatXactUserTablesRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgStatXactUserTablesRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgStatXactUserTablesRowIter) Close(ctx *sql.Context) error {
	return nil
}
