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

// PgStatUserTablesName is a constant to the pg_stat_user_tables name.
const PgStatUserTablesName = "pg_stat_user_tables"

// InitPgStatUserTables handles registration of the pg_stat_user_tables handler.
func InitPgStatUserTables() {
	tables.AddHandler(PgCatalogName, PgStatUserTablesName, PgStatUserTablesHandler{})
}

// PgStatUserTablesHandler is the handler for the pg_stat_user_tables table.
type PgStatUserTablesHandler struct{}

var _ tables.Handler = PgStatUserTablesHandler{}

// Name implements the interface tables.Handler.
func (p PgStatUserTablesHandler) Name() string {
	return PgStatUserTablesName
}

// RowIter implements the interface tables.Handler.
func (p PgStatUserTablesHandler) RowIter(ctx *sql.Context, partition sql.Partition) (sql.RowIter, error) {
	// TODO: Implement pg_stat_user_tables row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgStatUserTablesHandler) PkSchema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgStatUserTablesSchema,
		PkOrdinals: nil,
	}
}

// pgStatUserTablesSchema is the schema for pg_stat_user_tables.
var pgStatUserTablesSchema = sql.Schema{
	{Name: "relid", Type: pgtypes.Oid, Default: nil, Nullable: true, Source: PgStatUserTablesName},
	{Name: "schemaname", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgStatUserTablesName},
	{Name: "relname", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgStatUserTablesName},
	{Name: "seq_scan", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatUserTablesName},
	{Name: "last_seq_scan", Type: pgtypes.TimestampTZ, Default: nil, Nullable: true, Source: PgStatUserTablesName},
	{Name: "seq_tup_read", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatUserTablesName},
	{Name: "idx_scan", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatUserTablesName},
	{Name: "last_idx_scan", Type: pgtypes.TimestampTZ, Default: nil, Nullable: true, Source: PgStatUserTablesName},
	{Name: "idx_tup_fetch", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatUserTablesName},
	{Name: "n_tup_ins", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatUserTablesName},
	{Name: "n_tup_upd", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatUserTablesName},
	{Name: "n_tup_del", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatUserTablesName},
	{Name: "n_tup_hot_upd", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatUserTablesName},
	{Name: "n_tup_newpage_upd", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatUserTablesName},
	{Name: "n_live_tup", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatUserTablesName},
	{Name: "n_dead_tup", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatUserTablesName},
	{Name: "n_mod_since_analyze", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatUserTablesName},
	{Name: "n_ins_since_vacuum", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatUserTablesName},
	{Name: "last_vacuum", Type: pgtypes.TimestampTZ, Default: nil, Nullable: true, Source: PgStatUserTablesName},
	{Name: "last_autovacuum", Type: pgtypes.TimestampTZ, Default: nil, Nullable: true, Source: PgStatUserTablesName},
	{Name: "last_analyze", Type: pgtypes.TimestampTZ, Default: nil, Nullable: true, Source: PgStatUserTablesName},
	{Name: "last_autoanalyze", Type: pgtypes.TimestampTZ, Default: nil, Nullable: true, Source: PgStatUserTablesName},
	{Name: "vacuum_count", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatUserTablesName},
	{Name: "autovacuum_count", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatUserTablesName},
	{Name: "analyze_count", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatUserTablesName},
	{Name: "autoanalyze_count", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatUserTablesName},
}

// pgStatUserTablesRowIter is the sql.RowIter for the pg_stat_user_tables table.
type pgStatUserTablesRowIter struct {
}

var _ sql.RowIter = (*pgStatUserTablesRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgStatUserTablesRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgStatUserTablesRowIter) Close(ctx *sql.Context) error {
	return nil
}
