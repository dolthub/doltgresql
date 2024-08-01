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

// PgDependName is a constant to the pg_depend name.
const PgDependName = "pg_depend"

// InitPgDepend handles registration of the pg_depend handler.
func InitPgDepend() {
	tables.AddHandler(PgCatalogName, PgDependName, PgDependHandler{})
}

// PgDependHandler is the handler for the pg_depend table.
type PgDependHandler struct{}

var _ tables.Handler = PgDependHandler{}

// Name implements the interface tables.Handler.
func (p PgDependHandler) Name() string {
	return PgDependName
}

// RowIter implements the interface tables.Handler.
func (p PgDependHandler) RowIter(ctx *sql.Context) (sql.RowIter, error) {
	// TODO: Implement pg_depend row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgDependHandler) Schema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgDependSchema,
		PkOrdinals: nil,
	}
}

// pgDependSchema is the schema for pg_depend.
var pgDependSchema = sql.Schema{
	{Name: "classid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgDependName},
	{Name: "objid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgDependName},
	{Name: "objsubid", Type: pgtypes.Int32, Default: nil, Nullable: false, Source: PgDependName},
	{Name: "refclassid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgDependName},
	{Name: "refobjid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgDependName},
	{Name: "refobjsubid", Type: pgtypes.Int32, Default: nil, Nullable: false, Source: PgDependName},
	{Name: "deptype", Type: pgtypes.InternalChar, Default: nil, Nullable: false, Source: PgDependName},
}

// pgDependRowIter is the sql.RowIter for the pg_depend table.
type pgDependRowIter struct {
}

var _ sql.RowIter = (*pgDependRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgDependRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgDependRowIter) Close(ctx *sql.Context) error {
	return nil
}
