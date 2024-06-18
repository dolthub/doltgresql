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

// PgNamespaceName is a constant to the pg_namespace name.
const PgNamespaceName = "pg_namespace"

// InitPgNamespace handles registration of the pg_namespace handler.
func InitPgNamespace() {
	tables.AddHandler(PgCatalogName, PgNamespaceName, PgNamespaceHandler{})
}

// PgNamespaceHandler is the handler for the pg_namespace table.
type PgNamespaceHandler struct{}

var _ tables.Handler = PgNamespaceHandler{}

// Name implements the interface tables.Handler.
func (p PgNamespaceHandler) Name() string {
	return PgNamespaceName
}

// RowIter implements the interface tables.Handler.
func (p PgNamespaceHandler) RowIter(ctx *sql.Context) (sql.RowIter, error) {
	// TODO: Implement pg_namespace row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgNamespaceHandler) Schema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgNamespaceSchema,
		PkOrdinals: nil,
	}
}

// pgNamespaceSchema is the schema for pg_namespace.
var pgNamespaceSchema = sql.Schema{
	{Name: "oid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgNamespaceName},
	{Name: "nspname", Type: pgtypes.Name, Default: nil, Nullable: false, Source: PgNamespaceName},
	{Name: "nspowner", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgNamespaceName},
	{Name: "nspacl", Type: pgtypes.TextArray, Default: nil, Nullable: true, Source: PgNamespaceName}, // TODO: type aclitem[]         // TODO: type pg_node_tree, collation C
}

// pgNamespaceRowIter is the sql.RowIter for the pg_namespace table.
type pgNamespaceRowIter struct {
	idx int
}

var _ sql.RowIter = (*pgNamespaceRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgNamespaceRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgNamespaceRowIter) Close(ctx *sql.Context) error {
	return nil
}
