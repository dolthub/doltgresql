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

// PgConstraintName is a constant to the pg_constraint name.
const PgConstraintName = "pg_constraint"

// InitPgConstraint handles registration of the pg_constraint handler.
func InitPgConstraint() {
	tables.AddHandler(PgCatalogName, PgConstraintName, PgConstraintHandler{})
}

// PgConstraintHandler is the handler for the pg_constraint table.
type PgConstraintHandler struct{}

var _ tables.Handler = PgConstraintHandler{}

// Name implements the interface tables.Handler.
func (p PgConstraintHandler) Name() string {
	return PgConstraintName
}

// RowIter implements the interface tables.Handler.
func (p PgConstraintHandler) RowIter(ctx *sql.Context) (sql.RowIter, error) {
	// TODO: Implement pg_constraint row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgConstraintHandler) Schema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     PgConstraintSchema,
		PkOrdinals: nil,
	}
}

// PgConstraintSchema is the schema for pg_constraint.
var PgConstraintSchema = sql.Schema{
	{Name: "oid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgConstraintName},
	{Name: "conname", Type: pgtypes.Name, Default: nil, Nullable: false, Source: PgConstraintName},
	{Name: "connamespace", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgConstraintName},
	{Name: "contype", Type: pgtypes.BpChar, Default: nil, Nullable: false, Source: PgConstraintName},
	{Name: "condeferrable", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgConstraintName},
	{Name: "condeferred", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgConstraintName},
	{Name: "convalidated", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgConstraintName},
	{Name: "conrelid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgConstraintName},
	{Name: "contypid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgConstraintName},
	{Name: "conindid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgConstraintName},
	{Name: "conparentid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgConstraintName},
	{Name: "confrelid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgConstraintName},
	{Name: "confupdtype", Type: pgtypes.BpChar, Default: nil, Nullable: false, Source: PgConstraintName},
	{Name: "confdeltype", Type: pgtypes.BpChar, Default: nil, Nullable: false, Source: PgConstraintName},
	{Name: "confmatchtype", Type: pgtypes.BpChar, Default: nil, Nullable: false, Source: PgConstraintName},
	{Name: "conislocal", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgConstraintName},
	{Name: "coninhcount", Type: pgtypes.Int16, Default: nil, Nullable: false, Source: PgConstraintName},
	{Name: "connoinherit", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgConstraintName},
	{Name: "conkey", Type: pgtypes.Int16Array, Default: nil, Nullable: true, Source: PgConstraintName},
	{Name: "confkey", Type: pgtypes.Int16Array, Default: nil, Nullable: true, Source: PgConstraintName},
	{Name: "conpfeqop", Type: pgtypes.OidArray, Default: nil, Nullable: true, Source: PgConstraintName},
	{Name: "conppeqop", Type: pgtypes.OidArray, Default: nil, Nullable: true, Source: PgConstraintName},
	{Name: "conffeqop", Type: pgtypes.OidArray, Default: nil, Nullable: true, Source: PgConstraintName},
	{Name: "confdelsetcols", Type: pgtypes.Int16Array, Default: nil, Nullable: true, Source: PgConstraintName},
	{Name: "conexclop", Type: pgtypes.OidArray, Default: nil, Nullable: true, Source: PgConstraintName},
	{Name: "conbin", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgConstraintName}, // TODO: type pg_node_tree, collation C
}

// pgConstraintRowIter is the sql.RowIter for the pg_constraint table.
type pgConstraintRowIter struct {
}

var _ sql.RowIter = (*pgConstraintRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgConstraintRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgConstraintRowIter) Close(ctx *sql.Context) error {
	return nil
}
