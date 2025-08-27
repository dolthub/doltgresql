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
	// TODO: Implement pg_stat_all_tables row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgStatAllTablesHandler) Schema() sql.PrimaryKeySchema {
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

// pgStatAllTablesRowIter is the sql.RowIter for the pg_stat_all_tables table.
type pgStatAllTablesRowIter struct {
}

var _ sql.RowIter = (*pgStatAllTablesRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgStatAllTablesRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgStatAllTablesRowIter) Close(ctx *sql.Context) error {
	return nil
}
