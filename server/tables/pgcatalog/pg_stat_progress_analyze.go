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

// PgStatProgressAnalyzeName is a constant to the pg_stat_progress_analyze name.
const PgStatProgressAnalyzeName = "pg_stat_progress_analyze"

// InitPgStatProgressAnalyze handles registration of the pg_stat_progress_analyze handler.
func InitPgStatProgressAnalyze() {
	tables.AddHandler(PgCatalogName, PgStatProgressAnalyzeName, PgStatProgressAnalyzeHandler{})
}

// PgStatProgressAnalyzeHandler is the handler for the pg_stat_progress_analyze table.
type PgStatProgressAnalyzeHandler struct{}

var _ tables.Handler = PgStatProgressAnalyzeHandler{}

// Name implements the interface tables.Handler.
func (p PgStatProgressAnalyzeHandler) Name() string {
	return PgStatProgressAnalyzeName
}

// RowIter implements the interface tables.Handler.
func (p PgStatProgressAnalyzeHandler) RowIter(ctx *sql.Context) (sql.RowIter, error) {
	// TODO: Implement pg_stat_progress_analyze row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgStatProgressAnalyzeHandler) Schema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgStatProgressAnalyzeSchema,
		PkOrdinals: nil,
	}
}

// pgStatProgressAnalyzeSchema is the schema for pg_stat_progress_analyze.
var pgStatProgressAnalyzeSchema = sql.Schema{
	{Name: "pid", Type: pgtypes.Int32, Default: nil, Nullable: true, Source: PgStatProgressAnalyzeName},
	{Name: "datid", Type: pgtypes.Oid, Default: nil, Nullable: true, Source: PgStatProgressAnalyzeName},
	{Name: "datname", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgStatProgressAnalyzeName},
	{Name: "relid", Type: pgtypes.Oid, Default: nil, Nullable: true, Source: PgStatProgressAnalyzeName},
	{Name: "phase", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgStatProgressAnalyzeName},
	{Name: "sample_blks_total", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatProgressAnalyzeName},
	{Name: "sample_blks_scanned", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatProgressAnalyzeName},
	{Name: "ext_stats_total", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatProgressAnalyzeName},
	{Name: "ext_stats_computed", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatProgressAnalyzeName},
	{Name: "child_tables_total", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatProgressAnalyzeName},
	{Name: "child_tables_done", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatProgressAnalyzeName},
	{Name: "current_child_table_relid", Type: pgtypes.Oid, Default: nil, Nullable: true, Source: PgStatProgressAnalyzeName},
}

// pgStatProgressAnalyzeRowIter is the sql.RowIter for the pg_stat_progress_analyze table.
type pgStatProgressAnalyzeRowIter struct {
}

var _ sql.RowIter = (*pgStatProgressAnalyzeRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgStatProgressAnalyzeRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgStatProgressAnalyzeRowIter) Close(ctx *sql.Context) error {
	return nil
}
