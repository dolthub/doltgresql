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

// PgTablespaceName is a constant to the pg_tablespace name.
const PgTablespaceName = "pg_tablespace"

// InitPgTablespace handles registration of the pg_tablespace handler.
func InitPgTablespace() {
	tables.AddHandler(PgCatalogName, PgTablespaceName, PgTablespaceHandler{})
}

// PgTablespaceHandler is the handler for the pg_tablespace table.
type PgTablespaceHandler struct{}

var _ tables.Handler = PgTablespaceHandler{}

// Name implements the interface tables.Handler.
func (p PgTablespaceHandler) Name() string {
	return PgTablespaceName
}

// RowIter implements the interface tables.Handler.
func (p PgTablespaceHandler) RowIter(ctx *sql.Context, partition sql.Partition) (sql.RowIter, error) {
	// TODO: Implement pg_tablespace row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgTablespaceHandler) PkSchema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgTablespaceSchema,
		PkOrdinals: nil,
	}
}

// pgTablespaceSchema is the schema for pg_tablespace.
var pgTablespaceSchema = sql.Schema{
	{Name: "oid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgTablespaceName},
	{Name: "spcname", Type: pgtypes.Name, Default: nil, Nullable: false, Source: PgTablespaceName},
	{Name: "spcowner", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgTablespaceName},
	{Name: "spcacl", Type: pgtypes.TextArray, Default: nil, Nullable: true, Source: PgTablespaceName},     // TODO: aclitem[] type
	{Name: "spcoptions", Type: pgtypes.TextArray, Default: nil, Nullable: true, Source: PgTablespaceName}, // TODO: collation C
}

// pgTablespaceRowIter is the sql.RowIter for the pg_tablespace table.
type pgTablespaceRowIter struct {
}

var _ sql.RowIter = (*pgTablespaceRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgTablespaceRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgTablespaceRowIter) Close(ctx *sql.Context) error {
	return nil
}
