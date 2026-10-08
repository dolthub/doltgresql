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

// PgAuthMembersName is a constant to the pg_auth_members name.
const PgAuthMembersName = "pg_auth_members"

// InitPgAuthMembers handles registration of the pg_auth_members handler.
func InitPgAuthMembers() {
	tables.AddHandler(PgCatalogName, PgAuthMembersName, PgAuthMembersHandler{})
}

// PgAuthMembersHandler is the handler for the pg_auth_members table.
type PgAuthMembersHandler struct{}

var _ tables.Handler = PgAuthMembersHandler{}

// Name implements the interface tables.Handler.
func (p PgAuthMembersHandler) Name() string {
	return PgAuthMembersName
}

// pgAuthMember represents a row in the pg_auth_members table.
type pgAuthMember struct {
	roleid      id.Id
	member      id.Id
	grantor     id.Id
	adminOption bool
}

// RowIter implements the interface tables.Handler.
func (p PgAuthMembersHandler) RowIter(ctx *sql.Context, partition sql.Partition) (sql.RowIter, error) {
	var roles []auth.Role
	var memberships []auth.RoleMembershipValue
	auth.LockRead(func() {
		roles = auth.AllRoles()
		memberships = auth.AllRoleMemberships()
	})
	namesByID := make(map[auth.RoleID]string)
	for _, role := range roles {
		namesByID[role.ID()] = role.Name
	}
	// roleIDToOid returns the OID for the given RoleID, or id.Null if the role no longer exists.
	roleIDToOid := func(roleID auth.RoleID) id.Id {
		if name, ok := namesByID[roleID]; ok {
			return roleOid(name)
		}
		return id.Null
	}
	members := make([]pgAuthMember, 0, len(memberships))
	for _, membership := range memberships {
		if _, ok := namesByID[membership.Member]; !ok {
			continue
		}
		if _, ok := namesByID[membership.Group]; !ok {
			continue
		}
		members = append(members, pgAuthMember{
			roleid:      roleIDToOid(membership.Group),
			member:      roleIDToOid(membership.Member),
			grantor:     roleIDToOid(membership.GrantedBy),
			adminOption: membership.WithAdminOption,
		})
	}
	return &pgAuthMembersRowIter{
		members: members,
		idx:     0,
	}, nil
}

// PkSchema implements the interface tables.Handler.
func (p PgAuthMembersHandler) PkSchema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgAuthMembersSchema,
		PkOrdinals: nil,
	}
}

// pgAuthMembersSchema is the schema for pg_auth_members.
var pgAuthMembersSchema = sql.Schema{
	{Name: "roleid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgAuthMembersName},
	{Name: "member", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgAuthMembersName},
	{Name: "grantor", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgAuthMembersName},
	{Name: "admin_option", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgAuthMembersName},
}

// pgAuthMembersRowIter is the sql.RowIter for the pg_auth_members table.
type pgAuthMembersRowIter struct {
	members []pgAuthMember
	idx     int
}

var _ sql.RowIter = (*pgAuthMembersRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgAuthMembersRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	if iter.idx >= len(iter.members) {
		return nil, io.EOF
	}
	iter.idx++
	member := iter.members[iter.idx-1]
	return sql.Row{
		member.roleid,      // roleid
		member.member,      // member
		member.grantor,     // grantor
		member.adminOption, // admin_option
	}, nil
}

// Close implements the interface sql.RowIter.
func (iter *pgAuthMembersRowIter) Close(ctx *sql.Context) error {
	return nil
}
