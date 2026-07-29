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

	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/server/auth"
	"github.com/dolthub/doltgresql/server/tables"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// PgAuthidName is a constant to the pg_authid name.
const PgAuthidName = "pg_authid"

// InitPgAuthid handles registration of the pg_authid handler.
func InitPgAuthid() {
	tables.AddHandler(PgCatalogName, PgAuthidName, PgAuthidHandler{})
}

// PgAuthidHandler is the handler for the pg_authid table.
type PgAuthidHandler struct{}

var _ tables.Handler = PgAuthidHandler{}

// Name implements the interface tables.Handler.
func (p PgAuthidHandler) Name() string {
	return PgAuthidName
}

// roleOid returns the id.Id used as the OID for the role with the given name. Roles use an auth.RoleID (a uint64
// counter) internally rather than an id.Id, so we construct a name-based ID here. This gives every role a stable,
// unique ID for the duration of the process, which the OID cache maps to a numeric OID.
// TODO: replace with a proper wrapper (e.g. id.InternalUser) once roles use the id system (see auth.RoleID).
func roleOid(roleName string) id.Id {
	return id.NewId(id.Section_User, roleName)
}

// allRoles returns every role from the global auth database, sorted by name. This handles locking internally.
func allRoles() []auth.Role {
	var roles []auth.Role
	auth.LockRead(func() {
		roles = auth.AllRoles()
	})
	return roles
}

// RowIter implements the interface tables.Handler.
func (p PgAuthidHandler) RowIter(ctx *sql.Context, partition sql.Partition) (sql.RowIter, error) {
	return &pgAuthidRowIter{
		roles: allRoles(),
		idx:   0,
	}, nil
}

// PkSchema implements the interface tables.Handler.
func (p PgAuthidHandler) PkSchema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgAuthidSchema,
		PkOrdinals: nil,
	}
}

// pgAuthidSchema is the schema for pg_authid.
var pgAuthidSchema = sql.Schema{
	{Name: "oid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgAuthidName},
	{Name: "rolname", Type: pgtypes.Name, Default: nil, Nullable: false, Source: PgAuthidName},
	{Name: "rolsuper", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgAuthidName},
	{Name: "rolinherit", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgAuthidName},
	{Name: "rolcreaterole", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgAuthidName},
	{Name: "rolcreatedb", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgAuthidName},
	{Name: "rolcanlogin", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgAuthidName},
	{Name: "rolreplication", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgAuthidName},
	{Name: "rolbypassrls", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgAuthidName},
	{Name: "rolconnlimit", Type: pgtypes.Int32, Default: nil, Nullable: false, Source: PgAuthidName},
	{Name: "rolpassword", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgAuthidName}, // TODO: collation C
	{Name: "rolvaliduntil", Type: pgtypes.TimestampTZ, Default: nil, Nullable: true, Source: PgAuthidName},
}

// pgAuthidRowIter is the sql.RowIter for the pg_authid table.
type pgAuthidRowIter struct {
	roles []auth.Role
	idx   int
}

var _ sql.RowIter = (*pgAuthidRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgAuthidRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	if iter.idx >= len(iter.roles) {
		return nil, io.EOF
	}
	iter.idx++
	role := iter.roles[iter.idx-1]
	var rolPassword any
	if role.Password != nil {
		rolPassword = role.Password.AsPasswordString()
	}
	var rolValidUntil any
	if role.ValidUntil != nil {
		rolValidUntil = *role.ValidUntil
	}
	return sql.Row{
		roleOid(role.Name),             // oid
		role.Name,                      // rolname
		role.IsSuperUser,               // rolsuper
		role.InheritPrivileges,         // rolinherit
		role.CanCreateRoles,            // rolcreaterole
		role.CanCreateDB,               // rolcreatedb
		role.CanLogin,                  // rolcanlogin
		role.IsReplicationRole,         // rolreplication
		role.CanBypassRowLevelSecurity, // rolbypassrls
		role.ConnectionLimit,           // rolconnlimit
		rolPassword,                    // rolpassword
		rolValidUntil,                  // rolvaliduntil
	}, nil
}

// Close implements the interface sql.RowIter.
func (iter *pgAuthidRowIter) Close(ctx *sql.Context) error {
	return nil
}
