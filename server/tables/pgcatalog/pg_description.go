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

// PgDescriptionName is a constant to the pg_description name.
const PgDescriptionName = "pg_description"

// InitPgDescription handles registration of the pg_description handler.
func InitPgDescription() {
	tables.AddHandler(PgCatalogName, PgDescriptionName, PgDescriptionHandler{})
}

// PgDescriptionHandler is the handler for the pg_description table.
type PgDescriptionHandler struct{}

var _ tables.Handler = PgDescriptionHandler{}

// Name implements the interface tables.Handler.
func (p PgDescriptionHandler) Name() string {
	return PgDescriptionName
}

// RowIter implements the interface tables.Handler.
func (p PgDescriptionHandler) RowIter(ctx *sql.Context) (sql.RowIter, error) {
	// TODO: Implement pg_description row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgDescriptionHandler) Schema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgDescriptionSchema,
		PkOrdinals: nil,
	}
}

// pgDescriptionSchema is the schema for pg_description.
var pgDescriptionSchema = sql.Schema{
	{Name: "objoid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgDescriptionName},
	{Name: "classoid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgDescriptionName},
	{Name: "objsubid", Type: pgtypes.Int32, Default: nil, Nullable: false, Source: PgDescriptionName},
	{Name: "description", Type: pgtypes.Text, Default: nil, Nullable: false, Source: PgDescriptionName}, // TODO: collation C
}

// pgDescriptionRowIter is the sql.RowIter for the pg_description table.
type pgDescriptionRowIter struct {
	idx int
}

var _ sql.RowIter = (*pgDescriptionRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgDescriptionRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgDescriptionRowIter) Close(ctx *sql.Context) error {
	return nil
}
