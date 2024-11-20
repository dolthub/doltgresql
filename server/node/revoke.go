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

package node

import (
	"errors"
	"fmt"

	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/plan"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/server/auth"
)

// Revoke handles all of the REVOKE statements.
type Revoke struct {
	RevokeTable    *RevokeTable
	RevokeSchema   *RevokeSchema
	RevokeDatabase *RevokeDatabase
	RevokeRole     *RevokeRole
	FromRoles      []string
	GrantedBy      string
	GrantOptionFor bool // This is "ADMIN OPTION FOR" for RevokeRole only
	Cascade        bool // When false, represents RESTRICT
}

// RevokeTable specifically handles the REVOKE ... ON TABLE statement.
type RevokeTable struct {
	Privileges []auth.Privilege
	Tables     []doltdb.TableName
}

// RevokeSchema specifically handles the REVOKE ... ON SCHEMA statement.
type RevokeSchema struct {
	Privileges []auth.Privilege
	Schemas    []string
}

// RevokeDatabase specifically handles the REVOKE ... ON DATABASE statement.
type RevokeDatabase struct {
	Privileges []auth.Privilege
	Databases  []string
}

// RevokeRole specifically handles the REVOKE <roles> FROM <roles> statement.
type RevokeRole struct {
	Groups []string
}

var _ sql.ExecSourceRel = (*Revoke)(nil)
var _ vitess.Injectable = (*Revoke)(nil)

// Children implements the interface sql.ExecSourceRel.
func (r *Revoke) Children() []sql.Node {
	return nil
}

// IsReadOnly implements the interface sql.ExecSourceRel.
func (r *Revoke) IsReadOnly() bool {
	return false
}

// Resolved implements the interface sql.ExecSourceRel.
func (r *Revoke) Resolved() bool {
	return true
}

// RowIter implements the interface sql.ExecSourceRel.
func (r *Revoke) RowIter(ctx *sql.Context, _ sql.Row) (sql.RowIter, error) {
	if r.Cascade {
		return nil, errors.New("REVOKE does not yet support CASCADE")
	}

	var err error
	auth.LockWrite(func() {
		switch {
		case r.RevokeTable != nil:
			if err = r.revokeTable(ctx); err != nil {
				return
			}
		case r.RevokeSchema != nil:
			if err = r.revokeSchema(ctx); err != nil {
				return
			}
		case r.RevokeDatabase != nil:
			if err = r.revokeDatabase(ctx); err != nil {
				return
			}
		case r.RevokeRole != nil:
			if err = r.revokeRole(ctx); err != nil {
				return
			}
		default:
			err = fmt.Errorf("REVOKE statement is not yet supported")
			return
		}
		err = auth.PersistChanges()
	})
	if err != nil {
		return nil, err
	}
	return sql.RowsToRowIter(), nil
}

// Schema implements the interface sql.ExecSourceRel.
func (r *Revoke) Schema() sql.Schema {
	return nil
}

// String implements the interface sql.ExecSourceRel.
func (r *Revoke) String() string {
	switch {
	case r.RevokeTable != nil:
		return "REVOKE TABLE"
	default:
		return "REVOKE"
	}
}

// WithChildren implements the interface sql.ExecSourceRel.
func (r *Revoke) WithChildren(children ...sql.Node) (sql.Node, error) {
	return plan.NillaryWithChildren(r, children...)
}

// WithResolvedChildren implements the interface vitess.Injectable.
func (r *Revoke) WithResolvedChildren(children []any) (any, error) {
	if len(children) != 0 {
		return nil, ErrVitessChildCount.New(0, len(children))
	}
	return r, nil
}

// common handles the initial logic for each REVOKE statement. `roles` are the `FromRoles`. `userRole` is the role of
// the session's selected user. `grantedByID` is the `GrantedBy` user if specified (or `userRole` if not).
func (r *Revoke) common(ctx *sql.Context) (roles []auth.Role, userRole auth.Role, grantedByID auth.RoleID, err error) {
	roles = make([]auth.Role, len(r.FromRoles))
	// First we'll verify that all of the roles exist
	for i, roleName := range r.FromRoles {
		roles[i] = auth.GetRole(roleName)
		if !roles[i].IsValid() {
			return nil, auth.Role{}, 0, fmt.Errorf(`role "%s" does not exist`, roleName)
		}
	}
	// Then we'll check that the role that is revoking the privileges exists
	userRole = auth.GetRole(ctx.Client().User)
	if !userRole.IsValid() {
		return nil, auth.Role{}, 0, fmt.Errorf(`role "%s" does not exist`, ctx.Client().User)
	}
	if len(r.GrantedBy) != 0 {
		grantedByRole := auth.GetRole(r.GrantedBy)
		if !grantedByRole.IsValid() {
			return nil, auth.Role{}, 0, fmt.Errorf(`role "%s" does not exist`, r.GrantedBy)
		}
		if groupID, _, _ := auth.IsRoleAMember(userRole.ID(), grantedByRole.ID()); !groupID.IsValid() {
			// TODO: grab the actual error message
			return nil, auth.Role{}, 0, fmt.Errorf(`role "%s" does not have permission to revoke this privilege`, userRole.Name)
		}
	} else {
		grantedByID = userRole.ID()
	}
	return roles, userRole, grantedByID, nil
}

// revokeTable handles *RevokeTable from within RowIter.
func (r *Revoke) revokeTable(ctx *sql.Context) error {
	roles, userRole, grantedByID, err := r.common(ctx)
	if err != nil {
		return err
	}
	for _, role := range roles {
		for _, table := range r.RevokeTable.Tables {
			schemaName, err := core.GetSchemaName(ctx, nil, table.Schema)
			if err != nil {
				return err
			}
			key := auth.TablePrivilegeKey{
				Role:  userRole.ID(),
				Table: doltdb.TableName{Name: table.Name, Schema: schemaName},
			}
			for _, privilege := range r.RevokeTable.Privileges {
				if id := auth.HasTablePrivilegeGrantOption(key, privilege); !id.IsValid() {
					// TODO: grab the actual error message
					return fmt.Errorf(`role "%s" does not have permission to revoke this privilege`, userRole.Name)
				}
				auth.RemoveTablePrivilege(auth.TablePrivilegeKey{
					Role:  role.ID(),
					Table: doltdb.TableName{Name: table.Name, Schema: schemaName},
				}, auth.GrantedPrivilege{
					Privilege: privilege,
					GrantedBy: grantedByID,
				}, r.GrantOptionFor)
			}
		}
	}
	return nil
}

// revokeSchema handles *RevokeSchema from within RowIter.
func (r *Revoke) revokeSchema(ctx *sql.Context) error {
	roles, userRole, grantedByID, err := r.common(ctx)
	if err != nil {
		return err
	}
	for _, role := range roles {
		for _, schema := range r.RevokeSchema.Schemas {
			key := auth.SchemaPrivilegeKey{
				Role:   userRole.ID(),
				Schema: schema,
			}
			for _, privilege := range r.RevokeTable.Privileges {
				if id := auth.HasSchemaPrivilegeGrantOption(key, privilege); !id.IsValid() {
					// TODO: grab the actual error message
					return fmt.Errorf(`role "%s" does not have permission to revoke this privilege`, userRole.Name)
				}
				auth.RemoveSchemaPrivilege(auth.SchemaPrivilegeKey{
					Role:   role.ID(),
					Schema: schema,
				}, auth.GrantedPrivilege{
					Privilege: privilege,
					GrantedBy: grantedByID,
				}, r.GrantOptionFor)
			}
		}
	}
	return nil
}

// revokeDatabase handles *RevokeDatabase from within RowIter.
func (r *Revoke) revokeDatabase(ctx *sql.Context) error {
	roles, userRole, grantedByID, err := r.common(ctx)
	if err != nil {
		return err
	}
	for _, role := range roles {
		for _, databases := range r.RevokeDatabase.Databases {
			key := auth.DatabasePrivilegeKey{
				Role: userRole.ID(),
				Name: databases,
			}
			for _, privilege := range r.RevokeDatabase.Privileges {
				if id := auth.HasDatabasePrivilegeGrantOption(key, privilege); !id.IsValid() {
					// TODO: grab the actual error message
					return fmt.Errorf(`role "%s" does not have permission to revoke this privilege`, userRole.Name)
				}
				auth.RemoveDatabasePrivilege(auth.DatabasePrivilegeKey{
					Role: role.ID(),
					Name: databases,
				}, auth.GrantedPrivilege{
					Privilege: privilege,
					GrantedBy: grantedByID,
				}, r.GrantOptionFor)
			}
		}
	}
	return nil
}

// revokeRole handles *RevokeRole from within RowIter.
func (r *Revoke) revokeRole(ctx *sql.Context) error {
	members, userRole, _, err := r.common(ctx)
	if err != nil {
		return err
	}
	groups := make([]auth.Role, len(r.RevokeRole.Groups))
	for i, groupName := range r.RevokeRole.Groups {
		groups[i] = auth.GetRole(groupName)
		if !groups[i].IsValid() {
			return fmt.Errorf(`role "%s" does not exist`, groupName)
		}
	}
	for _, member := range members {
		for _, group := range groups {
			memberGroupID, _, withAdminOption := auth.IsRoleAMember(userRole.ID(), group.ID())
			if !memberGroupID.IsValid() || !withAdminOption {
				// TODO: grab the actual error message
				return fmt.Errorf(`role "%s" does not have permission to revoke role "%s"`, userRole.Name, group.Name)
			}
			auth.RemoveMemberFromGroup(member.ID(), group.ID(), r.GrantOptionFor)
		}
	}
	return nil
}
