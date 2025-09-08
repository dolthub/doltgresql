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

// PgRolesName is a constant to the pg_roles name.
const PgRolesName = "pg_roles"

// InitPgRoles handles registration of the pg_roles handler.
func InitPgRoles() {
	tables.AddHandler(PgCatalogName, PgRolesName, PgRolesHandler{})
}

// PgRolesHandler is the handler for the pg_roles table.
type PgRolesHandler struct{}

var _ tables.Handler = PgRolesHandler{}

// Name implements the interface tables.Handler.
func (p PgRolesHandler) Name() string {
	return PgRolesName
}

// RowIter implements the interface tables.Handler.
func (p PgRolesHandler) RowIter(ctx *sql.Context, partition sql.Partition) (sql.RowIter, error) {
	// TODO: Implement pg_roles row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgRolesHandler) PkSchema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgRolesSchema,
		PkOrdinals: nil,
	}
}

// pgRolesSchema is the schema for pg_roles.
var pgRolesSchema = sql.Schema{
	{Name: "rolname", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgRolesName},
	{Name: "rolsuper", Type: pgtypes.Bool, Default: nil, Nullable: true, Source: PgRolesName},
	{Name: "rolinherit", Type: pgtypes.Bool, Default: nil, Nullable: true, Source: PgRolesName},
	{Name: "rolcreaterole", Type: pgtypes.Bool, Default: nil, Nullable: true, Source: PgRolesName},
	{Name: "rolcreatedb", Type: pgtypes.Bool, Default: nil, Nullable: true, Source: PgRolesName},
	{Name: "rolcanlogin", Type: pgtypes.Bool, Default: nil, Nullable: true, Source: PgRolesName},
	{Name: "rolreplication", Type: pgtypes.Bool, Default: nil, Nullable: true, Source: PgRolesName},
	{Name: "rolconnlimit", Type: pgtypes.Int32, Default: nil, Nullable: true, Source: PgRolesName},
	{Name: "rolpassword", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgRolesName},
	{Name: "rolvaliduntil", Type: pgtypes.TimestampTZ, Default: nil, Nullable: true, Source: PgRolesName},
	{Name: "rolbypassrls", Type: pgtypes.Bool, Default: nil, Nullable: true, Source: PgRolesName},
	{Name: "rolconfig", Type: pgtypes.TextArray, Default: nil, Nullable: true, Source: PgRolesName}, // TODO: collation C
	{Name: "oid", Type: pgtypes.Oid, Default: nil, Nullable: true, Source: PgRolesName},
}

// pgRolesRowIter is the sql.RowIter for the pg_roles table.
type pgRolesRowIter struct {
}

var _ sql.RowIter = (*pgRolesRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgRolesRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgRolesRowIter) Close(ctx *sql.Context) error {
	return nil
}
