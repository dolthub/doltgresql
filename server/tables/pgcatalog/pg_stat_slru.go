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

// PgStatSlruName is a constant to the pg_stat_slru name.
const PgStatSlruName = "pg_stat_slru"

// InitPgStatSlru handles registration of the pg_stat_slru handler.
func InitPgStatSlru() {
	tables.AddHandler(PgCatalogName, PgStatSlruName, PgStatSlruHandler{})
}

// PgStatSlruHandler is the handler for the pg_stat_slru table.
type PgStatSlruHandler struct{}

var _ tables.Handler = PgStatSlruHandler{}

// Name implements the interface tables.Handler.
func (p PgStatSlruHandler) Name() string {
	return PgStatSlruName
}

// RowIter implements the interface tables.Handler.
func (p PgStatSlruHandler) RowIter(ctx *sql.Context) (sql.RowIter, error) {
	// TODO: Implement pg_stat_slru row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgStatSlruHandler) Schema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgStatSlruSchema,
		PkOrdinals: nil,
	}
}

// pgStatSlruSchema is the schema for pg_stat_slru.
var pgStatSlruSchema = sql.Schema{
	{Name: "name", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgStatSlruName},
	{Name: "blks_zeroed", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatSlruName},
	{Name: "blks_hit", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatSlruName},
	{Name: "blks_read", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatSlruName},
	{Name: "blks_written", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatSlruName},
	{Name: "blks_exists", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatSlruName},
	{Name: "flushes", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatSlruName},
	{Name: "truncates", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatSlruName},
	{Name: "stats_reset", Type: pgtypes.TimestampTZ, Default: nil, Nullable: true, Source: PgStatSlruName},
}

// pgStatSlruRowIter is the sql.RowIter for the pg_stat_slru table.
type pgStatSlruRowIter struct {
}

var _ sql.RowIter = (*pgStatSlruRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgStatSlruRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgStatSlruRowIter) Close(ctx *sql.Context) error {
	return nil
}
