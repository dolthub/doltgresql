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

// PgConversionName is a constant to the pg_conversion name.
const PgConversionName = "pg_conversion"

// InitPgConversion handles registration of the pg_conversion handler.
func InitPgConversion() {
	tables.AddHandler(PgCatalogName, PgConversionName, PgConversionHandler{})
}

// PgConversionHandler is the handler for the pg_conversion table.
type PgConversionHandler struct{}

var _ tables.Handler = PgConversionHandler{}

// Name implements the interface tables.Handler.
func (p PgConversionHandler) Name() string {
	return PgConversionName
}

// RowIter implements the interface tables.Handler.
func (p PgConversionHandler) RowIter(ctx *sql.Context, partition sql.Partition) (sql.RowIter, error) {
	// TODO: Implement pg_conversion row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgConversionHandler) PkSchema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     PgConversionSchema,
		PkOrdinals: nil,
	}
}

// PgConversionSchema is the schema for pg_conversion.
var PgConversionSchema = sql.Schema{
	{Name: "oid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgConversionName},
	{Name: "conname", Type: pgtypes.Name, Default: nil, Nullable: false, Source: PgConversionName},
	{Name: "connamespace", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgConversionName},
	{Name: "conowner", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgConversionName},
	{Name: "conforencoding", Type: pgtypes.Int32, Default: nil, Nullable: false, Source: PgConversionName},
	{Name: "contoencoding", Type: pgtypes.Int32, Default: nil, Nullable: false, Source: PgConversionName},
	{Name: "conproc", Type: pgtypes.Text, Default: nil, Nullable: false, Source: PgConversionName}, // TODO: regproc type
	{Name: "condefault", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgConversionName},
}

// pgConversionRowIter is the sql.RowIter for the pg_conversion table.
type pgConversionRowIter struct {
}

var _ sql.RowIter = (*pgConversionRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgConversionRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgConversionRowIter) Close(ctx *sql.Context) error {
	return nil
}
