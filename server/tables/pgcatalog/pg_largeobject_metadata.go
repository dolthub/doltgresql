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

// PgLargeobjectMetadataName is a constant to the pg_largeobject_metadata name.
const PgLargeobjectMetadataName = "pg_largeobject_metadata"

// InitPgLargeobjectMetadata handles registration of the pg_largeobject_metadata handler.
func InitPgLargeobjectMetadata() {
	tables.AddHandler(PgCatalogName, PgLargeobjectMetadataName, PgLargeobjectMetadataHandler{})
}

// PgLargeobjectMetadataHandler is the handler for the pg_largeobject_metadata table.
type PgLargeobjectMetadataHandler struct{}

var _ tables.Handler = PgLargeobjectMetadataHandler{}

// Name implements the interface tables.Handler.
func (p PgLargeobjectMetadataHandler) Name() string {
	return PgLargeobjectMetadataName
}

// RowIter implements the interface tables.Handler.
func (p PgLargeobjectMetadataHandler) RowIter(ctx *sql.Context) (sql.RowIter, error) {
	// TODO: Implement pg_largeobject_metadata row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgLargeobjectMetadataHandler) Schema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgLargeobjectMetadataSchema,
		PkOrdinals: nil,
	}
}

// pgLargeobjectMetadataSchema is the schema for pg_largeobject_metadata.
var pgLargeobjectMetadataSchema = sql.Schema{
	{Name: "oid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgLargeobjectMetadataName},
	{Name: "lomowner", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgLargeobjectMetadataName},
	{Name: "lomacl", Type: pgtypes.TextArray, Default: nil, Nullable: true, Source: PgLargeobjectMetadataName}, // TODO: aclitem[] type
}

// pgLargeobjectMetadataRowIter is the sql.RowIter for the pg_largeobject_metadata table.
type pgLargeobjectMetadataRowIter struct {
}

var _ sql.RowIter = (*pgLargeobjectMetadataRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgLargeobjectMetadataRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgLargeobjectMetadataRowIter) Close(ctx *sql.Context) error {
	return nil
}
