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

// PgAmName is a constant to the pg_am name.
const PgAmName = "pg_am"

// InitPgAm handles registration of the pg_am handler.
func InitPgAm() {
	tables.AddHandler(PgCatalogName, PgAmName, PgAmHandler{})
}

// PgAmHandler is the handler for the pg_am table.
type PgAmHandler struct{}

var _ tables.Handler = PgAmHandler{}

// Name implements the interface tables.Handler.
func (p PgAmHandler) Name() string {
	return PgAmName
}

// RowIter implements the interface tables.Handler.
func (p PgAmHandler) RowIter(ctx *sql.Context) (sql.RowIter, error) {
	// TODO: Implement pg_am row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgAmHandler) Schema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgAmSchema,
		PkOrdinals: nil,
	}
}

// pgAmSchema is the schema for pg_am.
var pgAmSchema = sql.Schema{
	{Name: "oid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgAmName},
	{Name: "amname", Type: pgtypes.Name, Default: nil, Nullable: false, Source: PgAmName},
	{Name: "amhandler", Type: pgtypes.Text, Default: nil, Nullable: false, Source: PgAmName}, // TODO: type regproc
	{Name: "amtype", Type: pgtypes.BpChar, Default: nil, Nullable: false, Source: PgAmName},
}

// pgAmRowIter is the sql.RowIter for the pg_am table.
type pgAmRowIter struct {
}

var _ sql.RowIter = (*pgAmRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgAmRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgAmRowIter) Close(ctx *sql.Context) error {
	return nil
}
