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

// PgPublicationNamespaceName is a constant to the pg_publication_namespace name.
const PgPublicationNamespaceName = "pg_publication_namespace"

// InitPgPublicationNamespace handles registration of the pg_publication_namespace handler.
func InitPgPublicationNamespace() {
	tables.AddHandler(PgCatalogName, PgPublicationNamespaceName, PgPublicationNamespaceHandler{})
}

// PgPublicationNamespaceHandler is the handler for the pg_publication_namespace table.
type PgPublicationNamespaceHandler struct{}

var _ tables.Handler = PgPublicationNamespaceHandler{}

// Name implements the interface tables.Handler.
func (p PgPublicationNamespaceHandler) Name() string {
	return PgPublicationNamespaceName
}

// RowIter implements the interface tables.Handler.
func (p PgPublicationNamespaceHandler) RowIter(ctx *sql.Context) (sql.RowIter, error) {
	// TODO: Implement pg_publication_namespace row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgPublicationNamespaceHandler) Schema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgPublicationNamespaceSchema,
		PkOrdinals: nil,
	}
}

// pgPublicationNamespaceSchema is the schema for pg_publication_namespace.
var pgPublicationNamespaceSchema = sql.Schema{
	{Name: "oid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgPublicationNamespaceName},
	{Name: "pnpubid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgPublicationNamespaceName},
	{Name: "pnnspid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgPublicationNamespaceName},
}

// pgPublicationNamespaceRowIter is the sql.RowIter for the pg_publication_namespace table.
type pgPublicationNamespaceRowIter struct {
}

var _ sql.RowIter = (*pgPublicationNamespaceRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgPublicationNamespaceRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgPublicationNamespaceRowIter) Close(ctx *sql.Context) error {
	return nil
}
