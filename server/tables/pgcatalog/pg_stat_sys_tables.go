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

// PgStatSysTablesName is a constant to the pg_stat_sys_tables name.
const PgStatSysTablesName = "pg_stat_sys_tables"

// InitPgStatSysTables handles registration of the pg_stat_sys_tables handler.
func InitPgStatSysTables() {
	tables.AddHandler(PgCatalogName, PgStatSysTablesName, PgStatSysTablesHandler{})
}

// PgStatSysTablesHandler is the handler for the pg_stat_sys_tables table.
type PgStatSysTablesHandler struct{}

var _ tables.Handler = PgStatSysTablesHandler{}

// Name implements the interface tables.Handler.
func (p PgStatSysTablesHandler) Name() string {
	return PgStatSysTablesName
}

// RowIter implements the interface tables.Handler.
func (p PgStatSysTablesHandler) RowIter(ctx *sql.Context, partition sql.Partition) (sql.RowIter, error) {
	// TODO: Implement pg_stat_sys_tables row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgStatSysTablesHandler) PkSchema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgStatSysTablesSchema,
		PkOrdinals: nil,
	}
}

// pgStatSysTablesSchema is the schema for pg_stat_sys_tables.
var pgStatSysTablesSchema = sql.Schema{
	{Name: "relid", Type: pgtypes.Oid, Default: nil, Nullable: true, Source: PgStatSysTablesName},
	{Name: "schemaname", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgStatSysTablesName},
	{Name: "relname", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgStatSysTablesName},
	{Name: "seq_scan", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatSysTablesName},
	{Name: "last_seq_scan", Type: pgtypes.TimestampTZ, Default: nil, Nullable: true, Source: PgStatSysTablesName},
	{Name: "seq_tup_read", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatSysTablesName},
	{Name: "idx_scan", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatSysTablesName},
	{Name: "last_idx_scan", Type: pgtypes.TimestampTZ, Default: nil, Nullable: true, Source: PgStatSysTablesName},
	{Name: "idx_tup_fetch", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatSysTablesName},
	{Name: "n_tup_ins", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatSysTablesName},
	{Name: "n_tup_upd", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatSysTablesName},
	{Name: "n_tup_del", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatSysTablesName},
	{Name: "n_tup_hot_upd", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatSysTablesName},
	{Name: "n_tup_newpage_upd", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatSysTablesName},
	{Name: "n_live_tup", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatSysTablesName},
	{Name: "n_dead_tup", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatSysTablesName},
	{Name: "n_mod_since_analyze", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatSysTablesName},
	{Name: "n_ins_since_vacuum", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatSysTablesName},
	{Name: "last_vacuum", Type: pgtypes.TimestampTZ, Default: nil, Nullable: true, Source: PgStatSysTablesName},
	{Name: "last_autovacuum", Type: pgtypes.TimestampTZ, Default: nil, Nullable: true, Source: PgStatSysTablesName},
	{Name: "last_analyze", Type: pgtypes.TimestampTZ, Default: nil, Nullable: true, Source: PgStatSysTablesName},
	{Name: "last_autoanalyze", Type: pgtypes.TimestampTZ, Default: nil, Nullable: true, Source: PgStatSysTablesName},
	{Name: "vacuum_count", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatSysTablesName},
	{Name: "autovacuum_count", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatSysTablesName},
	{Name: "analyze_count", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatSysTablesName},
	{Name: "autoanalyze_count", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatSysTablesName},
}

// pgStatSysTablesRowIter is the sql.RowIter for the pg_stat_sys_tables table.
type pgStatSysTablesRowIter struct {
}

var _ sql.RowIter = (*pgStatSysTablesRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgStatSysTablesRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgStatSysTablesRowIter) Close(ctx *sql.Context) error {
	return nil
}
