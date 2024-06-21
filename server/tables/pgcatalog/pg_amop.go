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

// PgAmopName is a constant to the pg_amop name.
const PgAmopName = "pg_amop"

// InitPgAmop handles registration of the pg_amop handler.
func InitPgAmop() {
	tables.AddHandler(PgCatalogName, PgAmopName, PgAmopHandler{})
}

// PgAmopHandler is the handler for the pg_amop table.
type PgAmopHandler struct{}

var _ tables.Handler = PgAmopHandler{}

// Name implements the interface tables.Handler.
func (p PgAmopHandler) Name() string {
	return PgAmopName
}

// RowIter implements the interface tables.Handler.
func (p PgAmopHandler) RowIter(ctx *sql.Context) (sql.RowIter, error) {
	// TODO: Implement pg_amop row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgAmopHandler) Schema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgAmopSchema,
		PkOrdinals: nil,
	}
}

// pgAmopSchema is the schema for pg_amop.
var pgAmopSchema = sql.Schema{
	{Name: "oid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgAmopName},
	{Name: "amopfamily", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgAmopName},
	{Name: "amoplefttype", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgAmopName},
	{Name: "amoprighttype", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgAmopName},
	{Name: "amopstrategy", Type: pgtypes.Int16, Default: nil, Nullable: false, Source: PgAmopName},
	{Name: "amoppurpose", Type: pgtypes.BpChar, Default: nil, Nullable: false, Source: PgAmopName},
	{Name: "amopopr", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgAmopName},
	{Name: "amopmethod", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgAmopName},
	{Name: "amopsortfamily", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgAmopName},
}

// pgAmopRowIter is the sql.RowIter for the pg_amop table.
type pgAmopRowIter struct {
}

var _ sql.RowIter = (*pgAmopRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgAmopRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgAmopRowIter) Close(ctx *sql.Context) error {
	return nil
}
