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

// PgStatsExtName is a constant to the pg_stats_ext name.
const PgStatsExtName = "pg_stats_ext"

// InitPgStatsExt handles registration of the pg_stats_ext handler.
func InitPgStatsExt() {
	tables.AddHandler(PgCatalogName, PgStatsExtName, PgStatsExtHandler{})
}

// PgStatsExtHandler is the handler for the pg_stats_ext table.
type PgStatsExtHandler struct{}

var _ tables.Handler = PgStatsExtHandler{}

// Name implements the interface tables.Handler.
func (p PgStatsExtHandler) Name() string {
	return PgStatsExtName
}

// RowIter implements the interface tables.Handler.
func (p PgStatsExtHandler) RowIter(ctx *sql.Context) (sql.RowIter, error) {
	// TODO: Implement pg_stats_ext row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgStatsExtHandler) Schema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgStatsExtSchema,
		PkOrdinals: nil,
	}
}

// pgStatsExtSchema is the schema for pg_stats_ext.
var pgStatsExtSchema = sql.Schema{
	{Name: "schemaname", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgStatsExtName},
	{Name: "tablename", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgStatsExtName},
	{Name: "statistics_schemaname", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgStatsExtName},
	{Name: "statistics_name", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgStatsExtName},
	{Name: "statistics_owner", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgStatsExtName},
	{Name: "attnames", Type: pgtypes.NameArray, Default: nil, Nullable: true, Source: PgStatsExtName},
	{Name: "exprs", Type: pgtypes.TextArray, Default: nil, Nullable: true, Source: PgStatsExtName},
	{Name: "kinds", Type: pgtypes.InternalCharArray, Default: nil, Nullable: true, Source: PgStatsExtName},
	{Name: "inherited", Type: pgtypes.Bool, Default: nil, Nullable: true, Source: PgStatsExtName},
	{Name: "n_distinct", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgStatsExtName},   // TODO: pg_ndistinct type AND collation C
	{Name: "dependencies", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgStatsExtName}, // TODO: pg_dependencies type AND collation C
	{Name: "most_common_vals", Type: pgtypes.TextArray, Default: nil, Nullable: true, Source: PgStatsExtName},
	{Name: "most_common_val_nulls", Type: pgtypes.BoolArray, Default: nil, Nullable: true, Source: PgStatsExtName},
	{Name: "most_common_freqs", Type: pgtypes.Float64Array, Default: nil, Nullable: true, Source: PgStatsExtName},
	{Name: "most_common_base_freqs", Type: pgtypes.Float64Array, Default: nil, Nullable: true, Source: PgStatsExtName},
}

// pgStatsExtRowIter is the sql.RowIter for the pg_stats_ext table.
type pgStatsExtRowIter struct {
}

var _ sql.RowIter = (*pgStatsExtRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgStatsExtRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgStatsExtRowIter) Close(ctx *sql.Context) error {
	return nil
}
