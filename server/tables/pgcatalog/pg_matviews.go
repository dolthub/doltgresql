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

// PgMatviewsName is a constant to the pg_matviews name.
const PgMatviewsName = "pg_matviews"

// InitPgMatviews handles registration of the pg_matviews handler.
func InitPgMatviews() {
	tables.AddHandler(PgCatalogName, PgMatviewsName, PgMatviewsHandler{})
}

// PgMatviewsHandler is the handler for the pg_matviews table.
type PgMatviewsHandler struct{}

var _ tables.Handler = PgMatviewsHandler{}

// Name implements the interface tables.Handler.
func (p PgMatviewsHandler) Name() string {
	return PgMatviewsName
}

// RowIter implements the interface tables.Handler.
func (p PgMatviewsHandler) RowIter(ctx *sql.Context) (sql.RowIter, error) {
	// TODO: Implement pg_matviews row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgMatviewsHandler) Schema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgMatviewsSchema,
		PkOrdinals: nil,
	}
}

// pgMatviewsSchema is the schema for pg_matviews.
var pgMatviewsSchema = sql.Schema{
	{Name: "schemaname", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgMatviewsName},
	{Name: "matviewname", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgMatviewsName},
	{Name: "matviewowner", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgMatviewsName},
	{Name: "tablespace", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgMatviewsName},
	{Name: "hasindexes", Type: pgtypes.Bool, Default: nil, Nullable: true, Source: PgMatviewsName},
	{Name: "ispopulated", Type: pgtypes.Bool, Default: nil, Nullable: true, Source: PgMatviewsName},
	{Name: "definition", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgMatviewsName},
}

// pgMatviewsRowIter is the sql.RowIter for the pg_matviews table.
type pgMatviewsRowIter struct {
}

var _ sql.RowIter = (*pgMatviewsRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgMatviewsRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgMatviewsRowIter) Close(ctx *sql.Context) error {
	return nil
}
