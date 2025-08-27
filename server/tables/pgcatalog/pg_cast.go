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

// PgCastName is a constant to the pg_cast name.
const PgCastName = "pg_cast"

// InitPgCast handles registration of the pg_cast handler.
func InitPgCast() {
	tables.AddHandler(PgCatalogName, PgCastName, PgCastHandler{})
}

// PgCastHandler is the handler for the pg_cast table.
type PgCastHandler struct{}

var _ tables.Handler = PgCastHandler{}

// Name implements the interface tables.Handler.
func (p PgCastHandler) Name() string {
	return PgCastName
}

// RowIter implements the interface tables.Handler.
func (p PgCastHandler) RowIter(ctx *sql.Context, partition sql.Partition) (sql.RowIter, error) {
	// TODO: Implement pg_cast row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgCastHandler) Schema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgCastSchema,
		PkOrdinals: nil,
	}
}

// pgCastSchema is the schema for pg_cast.
var pgCastSchema = sql.Schema{
	{Name: "oid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgCastName},
	{Name: "castsource", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgCastName},
	{Name: "casttarget", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgCastName},
	{Name: "castfunc", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgCastName},
	{Name: "castcontext", Type: pgtypes.InternalChar, Default: nil, Nullable: false, Source: PgCastName},
	{Name: "castmethod", Type: pgtypes.InternalChar, Default: nil, Nullable: false, Source: PgCastName},
}

// pgCastRowIter is the sql.RowIter for the pg_cast table.
type pgCastRowIter struct {
}

var _ sql.RowIter = (*pgCastRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgCastRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgCastRowIter) Close(ctx *sql.Context) error {
	return nil
}
