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

// PgStatsName is a constant to the pg_stats name.
const PgStatsName = "pg_stats"

// InitPgStats handles registration of the pg_stats handler.
func InitPgStats() {
	tables.AddHandler(PgCatalogName, PgStatsName, PgStatsHandler{})
}

// PgStatsHandler is the handler for the pg_stats table.
type PgStatsHandler struct{}

var _ tables.Handler = PgStatsHandler{}

// Name implements the interface tables.Handler.
func (p PgStatsHandler) Name() string {
	return PgStatsName
}

// RowIter implements the interface tables.Handler.
func (p PgStatsHandler) RowIter(ctx *sql.Context, partition sql.Partition) (sql.RowIter, error) {
	// TODO: Implement pg_stats row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgStatsHandler) PkSchema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgStatsSchema,
		PkOrdinals: nil,
	}
}

// pgStatsSchema is the schema for pg_stats.
var pgStatsSchema = sql.Schema{
	{Name: "schemaname", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgStatsName},
	{Name: "tablename", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgStatsName},
	{Name: "attname", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgStatsName},
	{Name: "inherited", Type: pgtypes.Bool, Default: nil, Nullable: true, Source: PgStatsName},
	{Name: "null_frac", Type: pgtypes.Float32, Default: nil, Nullable: true, Source: PgStatsName},
	{Name: "avg_width", Type: pgtypes.Int32, Default: nil, Nullable: true, Source: PgStatsName},
	{Name: "n_distinct", Type: pgtypes.Float32, Default: nil, Nullable: true, Source: PgStatsName},
	{Name: "most_common_vals", Type: pgtypes.AnyArray, Default: nil, Nullable: true, Source: PgStatsName},
	{Name: "most_common_freqs", Type: pgtypes.Float32Array, Default: nil, Nullable: true, Source: PgStatsName},
	{Name: "histogram_bounds", Type: pgtypes.AnyArray, Default: nil, Nullable: true, Source: PgStatsName},
	{Name: "correlation", Type: pgtypes.Float32, Default: nil, Nullable: true, Source: PgStatsName},
	{Name: "most_common_elems", Type: pgtypes.AnyArray, Default: nil, Nullable: true, Source: PgStatsName},
	{Name: "most_common_elem_freqs", Type: pgtypes.Float32Array, Default: nil, Nullable: true, Source: PgStatsName},
	{Name: "elem_count_histogram", Type: pgtypes.Float32Array, Default: nil, Nullable: true, Source: PgStatsName},
}

// pgStatsRowIter is the sql.RowIter for the pg_stats table.
type pgStatsRowIter struct {
}

var _ sql.RowIter = (*pgStatsRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgStatsRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgStatsRowIter) Close(ctx *sql.Context) error {
	return nil
}
