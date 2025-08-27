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

// PgShseclabelName is a constant to the pg_shseclabel name.
const PgShseclabelName = "pg_shseclabel"

// InitPgShseclabel handles registration of the pg_shseclabel handler.
func InitPgShseclabel() {
	tables.AddHandler(PgCatalogName, PgShseclabelName, PgShseclabelHandler{})
}

// PgShseclabelHandler is the handler for the pg_shseclabel table.
type PgShseclabelHandler struct{}

var _ tables.Handler = PgShseclabelHandler{}

// Name implements the interface tables.Handler.
func (p PgShseclabelHandler) Name() string {
	return PgShseclabelName
}

// RowIter implements the interface tables.Handler.
func (p PgShseclabelHandler) RowIter(ctx *sql.Context, partition sql.Partition) (sql.RowIter, error) {
	// TODO: Implement pg_shseclabel row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgShseclabelHandler) Schema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgShseclabelSchema,
		PkOrdinals: nil,
	}
}

// pgShseclabelSchema is the schema for pg_shseclabel.
var pgShseclabelSchema = sql.Schema{
	{Name: "objoid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgShseclabelName},
	{Name: "classoid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgShseclabelName},
	{Name: "provider", Type: pgtypes.Text, Default: nil, Nullable: false, Source: PgShseclabelName}, // TODO: collation C
	{Name: "label", Type: pgtypes.Text, Default: nil, Nullable: false, Source: PgShseclabelName},    // TODO: collation C
}

// pgShseclabelRowIter is the sql.RowIter for the pg_shseclabel table.
type pgShseclabelRowIter struct {
}

var _ sql.RowIter = (*pgShseclabelRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgShseclabelRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgShseclabelRowIter) Close(ctx *sql.Context) error {
	return nil
}
