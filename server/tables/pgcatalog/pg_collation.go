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

// PgCollationName is a constant to the pg_collation name.
const PgCollationName = "pg_collation"

// InitPgCollation handles registration of the pg_collation handler.
func InitPgCollation() {
	tables.AddHandler(PgCatalogName, PgCollationName, PgCollationHandler{})
}

// PgCollationHandler is the handler for the pg_collation table.
type PgCollationHandler struct{}

var _ tables.Handler = PgCollationHandler{}

// Name implements the interface tables.Handler.
func (p PgCollationHandler) Name() string {
	return PgCollationName
}

// RowIter implements the interface tables.Handler.
func (p PgCollationHandler) RowIter(ctx *sql.Context) (sql.RowIter, error) {
	// TODO: Implement pg_collation row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgCollationHandler) Schema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     PgCollationSchema,
		PkOrdinals: nil,
	}
}

// PgCollationSchema is the schema for pg_collation.
var PgCollationSchema = sql.Schema{
	{Name: "oid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgCollationName},
	{Name: "collname", Type: pgtypes.Name, Default: nil, Nullable: false, Source: PgCollationName},
	{Name: "collnamespace", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgCollationName},
	{Name: "collowner", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgCollationName},
	{Name: "collprovider", Type: pgtypes.BpChar, Default: nil, Nullable: false, Source: PgCollationName},
	{Name: "collisdeterministic", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgCollationName},
	{Name: "collencoding", Type: pgtypes.Int32, Default: nil, Nullable: false, Source: PgCollationName},
	{Name: "collcollate", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgCollationName},   // TODO: collation C
	{Name: "collctype", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgCollationName},     // TODO: collation C
	{Name: "colliculocale", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgCollationName}, // TODO: collation C
	{Name: "collversion", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgCollationName},   // TODO: collation C
}

// pgCollationRowIter is the sql.RowIter for the pg_collation table.
type pgCollationRowIter struct {
}

var _ sql.RowIter = (*pgCollationRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgCollationRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgCollationRowIter) Close(ctx *sql.Context) error {
	return nil
}
