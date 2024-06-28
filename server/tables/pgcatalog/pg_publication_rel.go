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

// PgPublicationRelName is a constant to the pg_publication_rel name.
const PgPublicationRelName = "pg_publication_rel"

// InitPgPublicationRel handles registration of the pg_publication_rel handler.
func InitPgPublicationRel() {
	tables.AddHandler(PgCatalogName, PgPublicationRelName, PgPublicationRelHandler{})
}

// PgPublicationRelHandler is the handler for the pg_publication_rel table.
type PgPublicationRelHandler struct{}

var _ tables.Handler = PgPublicationRelHandler{}

// Name implements the interface tables.Handler.
func (p PgPublicationRelHandler) Name() string {
	return PgPublicationRelName
}

// RowIter implements the interface tables.Handler.
func (p PgPublicationRelHandler) RowIter(ctx *sql.Context) (sql.RowIter, error) {
	// TODO: Implement pg_publication_rel row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgPublicationRelHandler) Schema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgPublicationRelSchema,
		PkOrdinals: nil,
	}
}

// pgPublicationRelSchema is the schema for pg_publication_rel.
var pgPublicationRelSchema = sql.Schema{
	{Name: "oid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgPublicationRelName},
	{Name: "prpubid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgPublicationRelName},
	{Name: "prrelid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgPublicationRelName},
	{Name: "prqual", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgPublicationRelName},        // TODO: pg_node_tree type, collation C
	{Name: "prattrs", Type: pgtypes.Int16Array, Default: nil, Nullable: true, Source: PgPublicationRelName}, // TODO: int2vector type
}

// pgPublicationRelRowIter is the sql.RowIter for the pg_publication_rel table.
type pgPublicationRelRowIter struct {
}

var _ sql.RowIter = (*pgPublicationRelRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgPublicationRelRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgPublicationRelRowIter) Close(ctx *sql.Context) error {
	return nil
}
