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

// PgShdependName is a constant to the pg_shdepend name.
const PgShdependName = "pg_shdepend"

// InitPgShdepend handles registration of the pg_shdepend handler.
func InitPgShdepend() {
	tables.AddHandler(PgCatalogName, PgShdependName, PgShdependHandler{})
}

// PgShdependHandler is the handler for the pg_shdepend table.
type PgShdependHandler struct{}

var _ tables.Handler = PgShdependHandler{}

// Name implements the interface tables.Handler.
func (p PgShdependHandler) Name() string {
	return PgShdependName
}

// RowIter implements the interface tables.Handler.
func (p PgShdependHandler) RowIter(ctx *sql.Context, partition sql.Partition) (sql.RowIter, error) {
	// TODO: Implement pg_shdepend row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgShdependHandler) Schema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgShdependSchema,
		PkOrdinals: nil,
	}
}

// pgShdependSchema is the schema for pg_shdepend.
var pgShdependSchema = sql.Schema{
	{Name: "dbid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgShdependName},
	{Name: "classid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgShdependName},
	{Name: "objid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgShdependName},
	{Name: "objsubid", Type: pgtypes.Int32, Default: nil, Nullable: false, Source: PgShdependName},
	{Name: "refclassid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgShdependName},
	{Name: "refobjid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgShdependName},
	{Name: "deptype", Type: pgtypes.InternalChar, Default: nil, Nullable: false, Source: PgShdependName},
}

// pgShdependRowIter is the sql.RowIter for the pg_shdepend table.
type pgShdependRowIter struct {
}

var _ sql.RowIter = (*pgShdependRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgShdependRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgShdependRowIter) Close(ctx *sql.Context) error {
	return nil
}
