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

	"github.com/dolthub/doltgresql/server/auth"
	"github.com/dolthub/doltgresql/server/tables"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// PgGroupName is a constant to the pg_group name.
const PgGroupName = "pg_group"

// InitPgGroup handles registration of the pg_group handler.
func InitPgGroup() {
	tables.AddHandler(PgCatalogName, PgGroupName, PgGroupHandler{})
}

// PgGroupHandler is the handler for the pg_group table.
type PgGroupHandler struct{}

var _ tables.Handler = PgGroupHandler{}

// Name implements the interface tables.Handler.
func (p PgGroupHandler) Name() string {
	return PgGroupName
}

// pgGroup represents a row in the pg_group table.
type pgGroup struct {
	name    string
	members []any
}

// RowIter implements the interface tables.Handler.
func (p PgGroupHandler) RowIter(ctx *sql.Context, partition sql.Partition) (sql.RowIter, error) {
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
	membersByGroup := make(map[auth.RoleID][]any)
	for _, membership := range memberships {
		memberName, ok := namesByID[membership.Member]
		if !ok {
			continue
		}
		membersByGroup[membership.Group] = append(membersByGroup[membership.Group], roleOid(memberName))
	}
	// Roles that cannot log in are considered groups
	var groups []pgGroup
	for _, role := range roles {
		if role.CanLogin {
			continue
		}
		members := membersByGroup[role.ID()]
		if members == nil {
			members = []any{}
		}
		groups = append(groups, pgGroup{
			name:    role.Name,
			members: members,
		})
	}
	return &pgGroupRowIter{
		groups: groups,
		idx:    0,
	}, nil
}

// PkSchema implements the interface tables.Handler.
func (p PgGroupHandler) PkSchema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgGroupSchema,
		PkOrdinals: nil,
	}
}

// pgGroupSchema is the schema for pg_group.
var pgGroupSchema = sql.Schema{
	{Name: "groname", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgGroupName},
	{Name: "grosysid", Type: pgtypes.Oid, Default: nil, Nullable: true, Source: PgGroupName},
	{Name: "grolist", Type: pgtypes.OidArray, Default: nil, Nullable: true, Source: PgGroupName},
}

// pgGroupRowIter is the sql.RowIter for the pg_group table.
type pgGroupRowIter struct {
	groups []pgGroup
	idx    int
}

var _ sql.RowIter = (*pgGroupRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgGroupRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	if iter.idx >= len(iter.groups) {
		return nil, io.EOF
	}
	iter.idx++
	group := iter.groups[iter.idx-1]
	return sql.Row{
		group.name,          // groname
		roleOid(group.name), // grosysid
		group.members,       // grolist
	}, nil
}

// Close implements the interface sql.RowIter.
func (iter *pgGroupRowIter) Close(ctx *sql.Context) error {
	return nil
}
