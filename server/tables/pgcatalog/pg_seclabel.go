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

// PgSeclabelName is a constant to the pg_seclabel name.
const PgSeclabelName = "pg_seclabel"

// InitPgSeclabel handles registration of the pg_seclabel handler.
func InitPgSeclabel() {
	tables.AddHandler(PgCatalogName, PgSeclabelName, PgSeclabelHandler{})
}

// PgSeclabelHandler is the handler for the pg_seclabel table.
type PgSeclabelHandler struct{}

var _ tables.Handler = PgSeclabelHandler{}

// Name implements the interface tables.Handler.
func (p PgSeclabelHandler) Name() string {
	return PgSeclabelName
}

// RowIter implements the interface tables.Handler.
func (p PgSeclabelHandler) RowIter(ctx *sql.Context, partition sql.Partition) (sql.RowIter, error) {
	// TODO: Implement pg_seclabel row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgSeclabelHandler) Schema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgSeclabelSchema,
		PkOrdinals: nil,
	}
}

// pgSeclabelSchema is the schema for pg_seclabel.
var pgSeclabelSchema = sql.Schema{
	{Name: "objoid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgSeclabelName},
	{Name: "classoid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgSeclabelName},
	{Name: "objsubid", Type: pgtypes.Int32, Default: nil, Nullable: false, Source: PgSeclabelName},
	{Name: "provider", Type: pgtypes.Text, Default: nil, Nullable: false, Source: PgSeclabelName}, // TODO: collation C
	{Name: "label", Type: pgtypes.Text, Default: nil, Nullable: false, Source: PgSeclabelName},    // TODO: collation C
}

// pgSeclabelRowIter is the sql.RowIter for the pg_seclabel table.
type pgSeclabelRowIter struct {
}

var _ sql.RowIter = (*pgSeclabelRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgSeclabelRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgSeclabelRowIter) Close(ctx *sql.Context) error {
	return nil
}
