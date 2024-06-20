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

// PgCursorsName is a constant to the pg_cursors name.
const PgCursorsName = "pg_cursors"

// InitPgCursors handles registration of the pg_cursors handler.
func InitPgCursors() {
	tables.AddHandler(PgCatalogName, PgCursorsName, PgCursorsHandler{})
}

// PgCursorsHandler is the handler for the pg_cursors table.
type PgCursorsHandler struct{}

var _ tables.Handler = PgCursorsHandler{}

// Name implements the interface tables.Handler.
func (p PgCursorsHandler) Name() string {
	return PgCursorsName
}

// RowIter implements the interface tables.Handler.
func (p PgCursorsHandler) RowIter(ctx *sql.Context) (sql.RowIter, error) {
	// TODO: Implement pg_cursors row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgCursorsHandler) Schema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgCursorsSchema,
		PkOrdinals: nil,
	}
}

// pgCursorsSchema is the schema for pg_cursors.
var pgCursorsSchema = sql.Schema{
	{Name: "name", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgCursorsName},
	{Name: "statement", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgCursorsName},
	{Name: "is_holdable", Type: pgtypes.Bool, Default: nil, Nullable: true, Source: PgCursorsName},
	{Name: "is_binary", Type: pgtypes.Bool, Default: nil, Nullable: true, Source: PgCursorsName},
	{Name: "is_scrollable", Type: pgtypes.Bool, Default: nil, Nullable: true, Source: PgCursorsName},
	{Name: "creation_time", Type: pgtypes.TimestampTZ, Default: nil, Nullable: true, Source: PgCursorsName},
}

// pgCursorsRowIter is the sql.RowIter for the pg_cursors table.
type pgCursorsRowIter struct {
}

var _ sql.RowIter = (*pgCursorsRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgCursorsRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgCursorsRowIter) Close(ctx *sql.Context) error {
	return nil
}
