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

// PgAttributeName is a constant to the pg_attribute name.
const PgAttributeName = "pg_attribute"

// InitPgAttribute handles registration of the pg_attribute handler.
func InitPgAttribute() {
	tables.AddHandler(PgCatalogName, PgAttributeName, PgAttributeHandler{})
}

// PgAttributeHandler is the handler for the pg_attribute table.
type PgAttributeHandler struct{}

var _ tables.Handler = PgAttributeHandler{}

// Name implements the interface tables.Handler.
func (p PgAttributeHandler) Name() string {
	return PgAttributeName
}

// emptyRowIter implements the sql.RowIter for empty table.
func emptyRowIter() (sql.RowIter, error) {
	return sql.RowsToRowIter(), nil
}

// RowIter implements the interface tables.Handler.
func (p PgAttributeHandler) RowIter(ctx *sql.Context) (sql.RowIter, error) {
	// TODO: Implement pg_attribute row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgAttributeHandler) Schema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgAttributeSchema,
		PkOrdinals: nil,
	}
}

// pgAttributeSchema is the schema for pg_attribute.
var pgAttributeSchema = sql.Schema{
	{Name: "attrelid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgAttributeName},
	{Name: "attname", Type: pgtypes.Name, Default: nil, Nullable: false, Source: PgAttributeName},
	{Name: "atttypid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgAttributeName},
	{Name: "attlen", Type: pgtypes.Int16, Default: nil, Nullable: false, Source: PgAttributeName},
	{Name: "attnum", Type: pgtypes.Int16, Default: nil, Nullable: false, Source: PgAttributeName},
	{Name: "attcacheoff", Type: pgtypes.Int32, Default: nil, Nullable: false, Source: PgAttributeName},
	{Name: "atttypmod", Type: pgtypes.Int32, Default: nil, Nullable: false, Source: PgAttributeName},
	{Name: "attndims", Type: pgtypes.Int16, Default: nil, Nullable: false, Source: PgAttributeName},
	{Name: "attbyval", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgAttributeName},
	{Name: "attalign", Type: pgtypes.BpChar, Default: nil, Nullable: false, Source: PgAttributeName},
	{Name: "attstorage", Type: pgtypes.BpChar, Default: nil, Nullable: false, Source: PgAttributeName},
	{Name: "attcompression", Type: pgtypes.BpChar, Default: nil, Nullable: false, Source: PgAttributeName},
	{Name: "attnotnull", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgAttributeName},
	{Name: "atthasdef", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgAttributeName},
	{Name: "atthasmissing", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgAttributeName},
	{Name: "attidentity", Type: pgtypes.BpChar, Default: nil, Nullable: false, Source: PgAttributeName},
	{Name: "attgenerated", Type: pgtypes.BpChar, Default: nil, Nullable: false, Source: PgAttributeName},
	{Name: "attisdropped", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgAttributeName},
	{Name: "attislocal", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgAttributeName},
	{Name: "attinhcount", Type: pgtypes.Int16, Default: nil, Nullable: false, Source: PgAttributeName},
	{Name: "attstattarget", Type: pgtypes.Int16, Default: nil, Nullable: false, Source: PgAttributeName},
	{Name: "attcollation", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgAttributeName},
	{Name: "attacl", Type: pgtypes.TextArray, Default: nil, Nullable: true, Source: PgAttributeName},        // TODO: type aclitem[]
	{Name: "attoptions", Type: pgtypes.TextArray, Default: nil, Nullable: true, Source: PgAttributeName},    // TODO: collation C
	{Name: "attfdwoptions", Type: pgtypes.TextArray, Default: nil, Nullable: true, Source: PgAttributeName}, // TODO: collation C
	{Name: "attmissingval", Type: pgtypes.AnyArray, Default: nil, Nullable: true, Source: PgAttributeName},
}

// pgAttributeRowIter is the sql.RowIter for the pg_attribute table.
type pgAttributeRowIter struct {
	idx int
}

var _ sql.RowIter = (*pgAttributeRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgAttributeRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgAttributeRowIter) Close(ctx *sql.Context) error {
	return nil
}
