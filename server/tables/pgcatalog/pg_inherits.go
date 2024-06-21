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

// PgInheritsName is a constant to the pg_inherits name.
const PgInheritsName = "pg_inherits"

// InitPgInherits handles registration of the pg_inherits handler.
func InitPgInherits() {
	tables.AddHandler(PgCatalogName, PgInheritsName, PgInheritsHandler{})
}

// PgInheritsHandler is the handler for the pg_inherits table.
type PgInheritsHandler struct{}

var _ tables.Handler = PgInheritsHandler{}

// Name implements the interface tables.Handler.
func (p PgInheritsHandler) Name() string {
	return PgInheritsName
}

// RowIter implements the interface tables.Handler.
func (p PgInheritsHandler) RowIter(ctx *sql.Context) (sql.RowIter, error) {
	// TODO: Implement pg_inherits row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgInheritsHandler) Schema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgInheritsSchema,
		PkOrdinals: nil,
	}
}

// pgInheritsSchema is the schema for pg_inherits.
var pgInheritsSchema = sql.Schema{
	{Name: "inhrelid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgInheritsName},
	{Name: "inhparent", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgInheritsName},
	{Name: "inhseqno", Type: pgtypes.Int32, Default: nil, Nullable: false, Source: PgInheritsName},
	{Name: "inhdetachpending", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgInheritsName},
}

// pgInheritsRowIter is the sql.RowIter for the pg_inherits table.
type pgInheritsRowIter struct {
}

var _ sql.RowIter = (*pgInheritsRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgInheritsRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgInheritsRowIter) Close(ctx *sql.Context) error {
	return nil
}
