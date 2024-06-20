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

// PgAttrdefName is a constant to the pg_attrdef name.
const PgAttrdefName = "pg_attrdef"

// InitPgAttrdef handles registration of the pg_attrdef handler.
func InitPgAttrdef() {
	tables.AddHandler(PgCatalogName, PgAttrdefName, PgAttrdefHandler{})
}

// PgAttrdefHandler is the handler for the pg_attrdef table.
type PgAttrdefHandler struct{}

var _ tables.Handler = PgAttrdefHandler{}

// Name implements the interface tables.Handler.
func (p PgAttrdefHandler) Name() string {
	return PgAttrdefName
}

// RowIter implements the interface tables.Handler.
func (p PgAttrdefHandler) RowIter(ctx *sql.Context) (sql.RowIter, error) {
	// TODO: Implement pg_attrdef row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgAttrdefHandler) Schema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgAttrdefSchema,
		PkOrdinals: nil,
	}
}

// pgAttrdefSchema is the schema for pg_attrdef.
var pgAttrdefSchema = sql.Schema{
	{Name: "oid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgAttributeName},
	{Name: "adrelid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgAttributeName},
	{Name: "adnum", Type: pgtypes.Int16, Default: nil, Nullable: false, Source: PgAttributeName},
	{Name: "adbin", Type: pgtypes.Text, Default: nil, Nullable: false, Source: PgAttributeName}, // TODO: collation C, type pg_node_tree
}

// pgAttrdefRowIter is the sql.RowIter for the pg_attrdef table.
type pgAttrdefRowIter struct {
}

var _ sql.RowIter = (*pgAttrdefRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgAttrdefRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgAttrdefRowIter) Close(ctx *sql.Context) error {
	return nil
}
