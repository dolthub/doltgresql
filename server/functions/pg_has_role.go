// Copyright 2026 Dolthub, Inc.
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

package functions

import (
	"strings"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
	"github.com/dolthub/doltgresql/server/auth"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initPgHasRole registers PostgreSQL's role privilege inquiry overloads.
func initPgHasRole() {
	framework.RegisterFunction(pg_has_role_name_name_text)
	framework.RegisterFunction(pg_has_role_name_oid_text)
	framework.RegisterFunction(pg_has_role_oid_name_text)
	framework.RegisterFunction(pg_has_role_oid_oid_text)
	framework.RegisterFunction(pg_has_role_name_text)
	framework.RegisterFunction(pg_has_role_oid_text)
}

// pg_has_role_name_name_text represents the PostgreSQL function of the same name.
var pg_has_role_name_name_text = framework.Function3{
	Name:               "pg_has_role",
	Return:             pgtypes.Bool,
	Parameters:         [3]*pgtypes.DoltgresType{pgtypes.Name, pgtypes.Name, pgtypes.Text},
	Strict:             true,
	IsNonDeterministic: true,
	Callable: func(ctx *sql.Context, _ [4]*pgtypes.DoltgresType, user, role, privilege any) (any, error) {
		return pgHasRole(ctx, user, role, privilege)
	},
}

// pg_has_role_name_oid_text represents the PostgreSQL function of the same name.
var pg_has_role_name_oid_text = framework.Function3{
	Name:               "pg_has_role",
	Return:             pgtypes.Bool,
	Parameters:         [3]*pgtypes.DoltgresType{pgtypes.Name, pgtypes.Oid, pgtypes.Text},
	Strict:             true,
	IsNonDeterministic: true,
	Callable: func(ctx *sql.Context, _ [4]*pgtypes.DoltgresType, user, role, privilege any) (any, error) {
		return pgHasRole(ctx, user, role, privilege)
	},
}

// pg_has_role_oid_name_text represents the PostgreSQL function of the same name.
var pg_has_role_oid_name_text = framework.Function3{
	Name:               "pg_has_role",
	Return:             pgtypes.Bool,
	Parameters:         [3]*pgtypes.DoltgresType{pgtypes.Oid, pgtypes.Name, pgtypes.Text},
	Strict:             true,
	IsNonDeterministic: true,
	Callable: func(ctx *sql.Context, _ [4]*pgtypes.DoltgresType, user, role, privilege any) (any, error) {
		return pgHasRole(ctx, user, role, privilege)
	},
}

// pg_has_role_oid_oid_text represents the PostgreSQL function of the same name.
var pg_has_role_oid_oid_text = framework.Function3{
	Name:               "pg_has_role",
	Return:             pgtypes.Bool,
	Parameters:         [3]*pgtypes.DoltgresType{pgtypes.Oid, pgtypes.Oid, pgtypes.Text},
	Strict:             true,
	IsNonDeterministic: true,
	Callable: func(ctx *sql.Context, _ [4]*pgtypes.DoltgresType, user, role, privilege any) (any, error) {
		return pgHasRole(ctx, user, role, privilege)
	},
}

// pg_has_role_name_text represents the PostgreSQL function of the same name.
var pg_has_role_name_text = framework.Function2{
	Name:               "pg_has_role",
	Return:             pgtypes.Bool,
	Parameters:         [2]*pgtypes.DoltgresType{pgtypes.Name, pgtypes.Text},
	Strict:             true,
	IsNonDeterministic: true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, role, privilege any) (any, error) {
		return pgHasRole(ctx, nil, role, privilege)
	},
}

// pg_has_role_oid_text represents the PostgreSQL function of the same name.
var pg_has_role_oid_text = framework.Function2{
	Name:               "pg_has_role",
	Return:             pgtypes.Bool,
	Parameters:         [2]*pgtypes.DoltgresType{pgtypes.Oid, pgtypes.Text},
	Strict:             true,
	IsNonDeterministic: true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, role, privilege any) (any, error) {
		return pgHasRole(ctx, nil, role, privilege)
	},
}

// pgHasRole evaluates against a consistent snapshot of the auth graph. A nil
// user selects the current effective role; SQL NULLs are handled by Strict.
func pgHasRole(ctx *sql.Context, user, target, privilege any) (any, error) {
	var result bool
	var err error
	auth.LockRead(func() {
		var actor, role auth.Role
		var actorOID, roleOID id.Id
		if user == nil {
			actor, err = auth.CurrentRoleLocked(ctx)
			actorOID = id.NewId(id.Section_User, actor.Name)
		} else {
			actor, actorOID, err = pgHasRoleLookup(ctx, user)
		}
		if err != nil {
			return
		}
		role, roleOID, err = pgHasRoleLookup(ctx, target)
		if err != nil {
			return
		}
		var privileges string
		privileges, err = framework.UnwrapString(ctx, privilege)
		if err != nil {
			return
		}
		// Validate the whole list before evaluating its OR semantics.
		var member, usage, admin bool
		for _, item := range strings.Split(privileges, ",") {
			switch strings.ToUpper(strings.TrimSpace(item)) {
			case "MEMBER":
				member = true
			case "USAGE":
				usage = true
			case "MEMBER WITH ADMIN OPTION", "MEMBER WITH GRANT OPTION", "USAGE WITH ADMIN OPTION", "USAGE WITH GRANT OPTION":
				admin = true
			default:
				err = pgerror.Newf(pgcode.InvalidParameterValue, "unrecognized privilege type: %q", strings.TrimSpace(item))
				return
			}
		}
		// Like PostgreSQL, OID overloads do not require the OIDs to exist:
		// a superuser has all privileges, and an OID has its own MEMBER/USAGE.
		if actor.IsSuperUser {
			result = true
			return
		}
		if actorOID == roleOID {
			result = member || usage
			return
		}
		if !actor.IsValid() || !role.IsValid() {
			return
		}
		result = (member && auth.HasRoleMembership(actor.ID(), role.ID())) ||
			(usage && auth.InheritsPrivileges(actor.ID(), role.ID())) ||
			(admin && pgHasRoleAdmin(actor.ID(), role.ID()))
	})
	return result, err
}

// pgHasRoleLookup resolves catalog OIDs separately from auth's internal RoleIDs.
// The caller must hold the auth lock. Missing names error; missing OIDs do not.
func pgHasRoleLookup(ctx *sql.Context, value any) (auth.Role, id.Id, error) {
	if oid, ok := value.(id.Id); ok {
		if oid.Section() == id.Section_User && oid.SegmentCount() == 1 {
			role, _ := auth.LookupRole(oid.Segment(0))
			return role, oid, nil
		}
		return auth.Role{}, oid, nil
	}
	name, err := framework.UnwrapString(ctx, value)
	if err != nil {
		return auth.Role{}, id.Null, err
	}
	role, ok := auth.LookupRole(name)
	// PUBLIC is stored internally for ACLs, but is not a PostgreSQL role.
	if !ok || name == "public" {
		return auth.Role{}, id.Null, pgerror.Newf(pgcode.UndefinedObject, "role %q does not exist", name)
	}
	return role, id.NewId(id.Section_User, name), nil
}

// pgHasRoleAdmin follows membership regardless of INHERIT, as PostgreSQL 15's
// inquiry does. CanAdministerRole instead checks immediately available authority.
// Only the final grant to the target needs ADMIN OPTION. Caller holds auth lock.
func pgHasRoleAdmin(member, target auth.RoleID) bool {
	seen := make(map[auth.RoleID]bool)
	pending := []auth.RoleID{member}
	for len(pending) > 0 {
		current := pending[len(pending)-1]
		pending = pending[:len(pending)-1]
		if seen[current] {
			continue
		}
		seen[current] = true
		for _, grant := range auth.MembershipGrants(current) {
			if grant.GrantedRole == target && grant.AdminOption {
				return true
			}
			pending = append(pending, grant.GrantedRole)
		}
	}
	return false
}
