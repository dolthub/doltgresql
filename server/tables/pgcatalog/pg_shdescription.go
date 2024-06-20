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

// PgShdescriptionName is a constant to the pg_shdescription name.
const PgShdescriptionName = "pg_shdescription"

// InitPgShdescription handles registration of the pg_shdescription handler.
func InitPgShdescription() {
	tables.AddHandler(PgCatalogName, PgShdescriptionName, PgShdescriptionHandler{})
}

// PgShdescriptionHandler is the handler for the pg_shdescription table.
type PgShdescriptionHandler struct{}

var _ tables.Handler = PgShdescriptionHandler{}

// Name implements the interface tables.Handler.
func (p PgShdescriptionHandler) Name() string {
	return PgShdescriptionName
}

// RowIter implements the interface tables.Handler.
func (p PgShdescriptionHandler) RowIter(ctx *sql.Context) (sql.RowIter, error) {
	// TODO: Implement pg_shdescription row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgShdescriptionHandler) Schema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgShdescriptionSchema,
		PkOrdinals: nil,
	}
}

// pgShdescriptionSchema is the schema for pg_shdescription.
var pgShdescriptionSchema = sql.Schema{
	{Name: "objoid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgShdescriptionName},
	{Name: "classoid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgShdescriptionName},
	{Name: "description", Type: pgtypes.Text, Default: nil, Nullable: false, Source: PgShdescriptionName}, // TODO: collation C
}

// pgShdescriptionRowIter is the sql.RowIter for the pg_shdescription table.
type pgShdescriptionRowIter struct {
}

var _ sql.RowIter = (*pgShdescriptionRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgShdescriptionRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgShdescriptionRowIter) Close(ctx *sql.Context) error {
	return nil
}
