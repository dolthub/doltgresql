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

// PgStatIoName is a constant to the pg_stat_io name.
const PgStatIoName = "pg_stat_io"

// InitPgStatIo handles registration of the pg_stat_io handler.
func InitPgStatIo() {
	tables.AddHandler(PgCatalogName, PgStatIoName, PgStatIoHandler{})
}

// PgStatIoHandler is the handler for the pg_stat_io table.
type PgStatIoHandler struct{}

var _ tables.Handler = PgStatIoHandler{}

// Name implements the interface tables.Handler.
func (p PgStatIoHandler) Name() string {
	return PgStatIoName
}

// RowIter implements the interface tables.Handler.
func (p PgStatIoHandler) RowIter(ctx *sql.Context) (sql.RowIter, error) {
	// TODO: Implement pg_stat_io row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgStatIoHandler) Schema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgStatIoSchema,
		PkOrdinals: nil,
	}
}

// pgStatIoSchema is the schema for pg_stat_io.
var pgStatIoSchema = sql.Schema{
	{Name: "backend_type", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgStatIoName},
	{Name: "object", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgStatIoName},
	{Name: "context", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgStatIoName},
	{Name: "reads", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatIoName},
	{Name: "read_time", Type: pgtypes.Float64, Default: nil, Nullable: true, Source: PgStatIoName},
	{Name: "writes", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatIoName},
	{Name: "write_time", Type: pgtypes.Float64, Default: nil, Nullable: true, Source: PgStatIoName},
	{Name: "writebacks", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatIoName},
	{Name: "writeback_time", Type: pgtypes.Float64, Default: nil, Nullable: true, Source: PgStatIoName},
	{Name: "extends", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatIoName},
	{Name: "extend_time", Type: pgtypes.Float64, Default: nil, Nullable: true, Source: PgStatIoName},
	{Name: "op_bytes", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatIoName},
	{Name: "hits", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatIoName},
	{Name: "evictions", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatIoName},
	{Name: "reuses", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatIoName},
	{Name: "fsyncs", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatIoName},
	{Name: "fsync_time", Type: pgtypes.Float64, Default: nil, Nullable: true, Source: PgStatIoName},
	{Name: "stats_reset", Type: pgtypes.TimestampTZ, Default: nil, Nullable: true, Source: PgStatIoName},
}

// pgStatIoRowIter is the sql.RowIter for the pg_stat_io table.
type pgStatIoRowIter struct {
}

var _ sql.RowIter = (*pgStatIoRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgStatIoRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgStatIoRowIter) Close(ctx *sql.Context) error {
	return nil
}
