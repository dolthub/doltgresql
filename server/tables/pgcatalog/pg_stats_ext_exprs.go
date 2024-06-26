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

// PgStatsExtExprsName is a constant to the pg_stats_ext_exprs name.
const PgStatsExtExprsName = "pg_stats_ext_exprs"

// InitPgStatsExtExprs handles registration of the pg_stats_ext_exprs handler.
func InitPgStatsExtExprs() {
	tables.AddHandler(PgCatalogName, PgStatsExtExprsName, PgStatsExtExprsHandler{})
}

// PgStatsExtExprsHandler is the handler for the pg_stats_ext_exprs table.
type PgStatsExtExprsHandler struct{}

var _ tables.Handler = PgStatsExtExprsHandler{}

// Name implements the interface tables.Handler.
func (p PgStatsExtExprsHandler) Name() string {
	return PgStatsExtExprsName
}

// RowIter implements the interface tables.Handler.
func (p PgStatsExtExprsHandler) RowIter(ctx *sql.Context) (sql.RowIter, error) {
	// TODO: Implement pg_stats_ext_exprs row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgStatsExtExprsHandler) Schema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgStatsExtExprsSchema,
		PkOrdinals: nil,
	}
}

// pgStatsExtExprsSchema is the schema for pg_stats_ext_exprs.
var pgStatsExtExprsSchema = sql.Schema{
	{Name: "schemaname", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgStatsExtExprsName},
	{Name: "tablename", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgStatsExtExprsName},
	{Name: "statistics_schemaname", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgStatsExtExprsName},
	{Name: "statistics_name", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgStatsExtExprsName},
	{Name: "statistics_owner", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgStatsExtExprsName},
	{Name: "expr", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgStatsExtExprsName},
	{Name: "inherited", Type: pgtypes.Bool, Default: nil, Nullable: true, Source: PgStatsExtExprsName},
	{Name: "null_frac", Type: pgtypes.Float32, Default: nil, Nullable: true, Source: PgStatsExtExprsName},
	{Name: "avg_width", Type: pgtypes.Int32, Default: nil, Nullable: true, Source: PgStatsExtExprsName},
	{Name: "n_distinct", Type: pgtypes.Float32, Default: nil, Nullable: true, Source: PgStatsExtExprsName},
	{Name: "most_common_vals", Type: pgtypes.AnyArray, Default: nil, Nullable: true, Source: PgStatsExtExprsName},
	{Name: "most_common_freqs", Type: pgtypes.Float32Array, Default: nil, Nullable: true, Source: PgStatsExtExprsName},
	{Name: "histogram_bounds", Type: pgtypes.AnyArray, Default: nil, Nullable: true, Source: PgStatsExtExprsName},
	{Name: "correlation", Type: pgtypes.Float32, Default: nil, Nullable: true, Source: PgStatsExtExprsName},
	{Name: "most_common_elems", Type: pgtypes.AnyArray, Default: nil, Nullable: true, Source: PgStatsExtExprsName},
	{Name: "most_common_elem_freqs", Type: pgtypes.Float32Array, Default: nil, Nullable: true, Source: PgStatsExtExprsName},
	{Name: "elem_count_histogram", Type: pgtypes.Float32Array, Default: nil, Nullable: true, Source: PgStatsExtExprsName},
}

// pgStatsExtExprsRowIter is the sql.RowIter for the pg_stats_ext_exprs table.
type pgStatsExtExprsRowIter struct {
}

var _ sql.RowIter = (*pgStatsExtExprsRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgStatsExtExprsRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgStatsExtExprsRowIter) Close(ctx *sql.Context) error {
	return nil
}
