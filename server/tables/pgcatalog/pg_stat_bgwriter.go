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

// PgStatBgwriterName is a constant to the pg_stat_bgwriter name.
const PgStatBgwriterName = "pg_stat_bgwriter"

// InitPgStatBgwriter handles registration of the pg_stat_bgwriter handler.
func InitPgStatBgwriter() {
	tables.AddHandler(PgCatalogName, PgStatBgwriterName, PgStatBgwriterHandler{})
}

// PgStatBgwriterHandler is the handler for the pg_stat_bgwriter table.
type PgStatBgwriterHandler struct{}

var _ tables.Handler = PgStatBgwriterHandler{}

// Name implements the interface tables.Handler.
func (p PgStatBgwriterHandler) Name() string {
	return PgStatBgwriterName
}

// RowIter implements the interface tables.Handler.
func (p PgStatBgwriterHandler) RowIter(ctx *sql.Context, partition sql.Partition) (sql.RowIter, error) {
	// TODO: Implement pg_stat_bgwriter row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgStatBgwriterHandler) PkSchema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgStatBgwriterSchema,
		PkOrdinals: nil,
	}
}

// pgStatBgwriterSchema is the schema for pg_stat_bgwriter.
var pgStatBgwriterSchema = sql.Schema{
	{Name: "checkpoints_timed", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatBgwriterName},
	{Name: "checkpoints_req", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatBgwriterName},
	{Name: "checkpoint_write_time", Type: pgtypes.Float64, Default: nil, Nullable: true, Source: PgStatBgwriterName},
	{Name: "checkpoint_sync_time", Type: pgtypes.Float64, Default: nil, Nullable: true, Source: PgStatBgwriterName},
	{Name: "buffers_checkpoint", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatBgwriterName},
	{Name: "buffers_clean", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatBgwriterName},
	{Name: "maxwritten_clean", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatBgwriterName},
	{Name: "buffers_backend", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatBgwriterName},
	{Name: "buffers_backend_fsync", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatBgwriterName},
	{Name: "buffers_alloc", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatBgwriterName},
	{Name: "stats_reset", Type: pgtypes.TimestampTZ, Default: nil, Nullable: true, Source: PgStatBgwriterName},
}

// pgStatBgwriterRowIter is the sql.RowIter for the pg_stat_bgwriter table.
type pgStatBgwriterRowIter struct {
}

var _ sql.RowIter = (*pgStatBgwriterRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgStatBgwriterRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgStatBgwriterRowIter) Close(ctx *sql.Context) error {
	return nil
}
