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
	"github.com/cockroachdb/errors"

	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/plan"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/server/auth"
)

// Grant handles all of the GRANT statements.
type Grant struct {
	GrantTable      *GrantTable
	GrantSchema     *GrantSchema
	GrantDatabase   *GrantDatabase
	GrantRole       *GrantRole
	ToRoles         []string
	WithGrantOption bool // This is "WITH ADMIN OPTION" for GrantRole only
	GrantedBy       string
}

// GrantTable specifically handles the GRANT ... ON TABLE statement.
type GrantTable struct {
	Privileges []auth.Privilege
	Tables     []doltdb.TableName
}

// GrantSchema specifically handles the GRANT ... ON SCHEMA statement.
type GrantSchema struct {
	Privileges []auth.Privilege
	Schemas    []string
}

// GrantDatabase specifically handles the GRANT ... ON DATABASE statement.
type GrantDatabase struct {
	Privileges []auth.Privilege
	Databases  []string
}

// GrantRole specifically handles the GRANT <roles> TO <roles> statement.
type GrantRole struct {
	Groups []string
}

var _ sql.ExecSourceRel = (*Grant)(nil)
var _ vitess.Injectable = (*Grant)(nil)

// Children implements the interface sql.ExecSourceRel.
func (g *Grant) Children() []sql.Node {
	return nil
}

// IsReadOnly implements the interface sql.ExecSourceRel.
func (g *Grant) IsReadOnly() bool {
	return false
}

// Resolved implements the interface sql.ExecSourceRel.
func (g *Grant) Resolved() bool {
	return true
}

// RowIter implements the interface sql.ExecSourceRel.
func (g *Grant) RowIter(ctx *sql.Context, r sql.Row) (sql.RowIter, error) {
	var err error
	auth.LockWrite(func() {
		switch {
		case g.GrantTable != nil:
			if err = g.grantTable(ctx); err != nil {
				return
			}
		case g.GrantSchema != nil:
			if err = g.grantSchema(ctx); err != nil {
				return
			}
		case g.GrantDatabase != nil:
			if err = g.grantDatabase(ctx); err != nil {
				return
			}
		case g.GrantRole != nil:
			if err = g.grantRole(ctx); err != nil {
				return
			}
		default:
			err = errors.Errorf("GRANT statement is not yet supported")
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
func (g *Grant) Schema() sql.Schema {
	return nil
}

// String implements the interface sql.ExecSourceRel.
func (g *Grant) String() string {
	switch {
	case g.GrantTable != nil:
		return "GRANT TABLE"
	default:
		return "GRANT"
	}
}

// WithChildren implements the interface sql.ExecSourceRel.
func (g *Grant) WithChildren(children ...sql.Node) (sql.Node, error) {
	return plan.NillaryWithChildren(g, children...)
}

// WithResolvedChildren implements the interface vitess.Injectable.
func (g *Grant) WithResolvedChildren(children []any) (any, error) {
	if len(children) != 0 {
		return nil, ErrVitessChildCount.New(0, len(children))
	}
	return g, nil
}

// common handles the initial logic for each GRANT statement. `roles` are the `ToRoles`. `userRole` is the role of the
// session's selected user.
func (g *Grant) common(ctx *sql.Context) (roles []auth.Role, userRole auth.Role, err error) {
	roles = make([]auth.Role, len(g.ToRoles))
	// First we'll verify that all of the roles exist
	for i, roleName := range g.ToRoles {
		roles[i] = auth.GetRole(roleName)
		if !roles[i].IsValid() {
			return nil, auth.Role{}, errors.Errorf(`role "%s" does not exist`, roleName)
		}
	}
	// Then we'll check that the role that is granting the privileges exists
	userRole = auth.GetRole(ctx.Client().User)
	if !userRole.IsValid() {
		return nil, auth.Role{}, errors.Errorf(`role "%s" does not exist`, ctx.Client().User)
	}
	if len(g.GrantedBy) != 0 {
		grantedByRole := auth.GetRole(g.GrantedBy)
		if !grantedByRole.IsValid() {
			return nil, auth.Role{}, errors.Errorf(`role "%s" does not exist`, g.GrantedBy)
		}
		if userRole.ID() != grantedByRole.ID() {
			// TODO: grab the actual error message
			return nil, auth.Role{}, errors.New("GRANTED BY may only be set to the calling user")
		}
	}
	return roles, userRole, nil
}

// grantTable handles *GrantTable from within RowIter.
func (g *Grant) grantTable(ctx *sql.Context) error {
	roles, userRole, err := g.common(ctx)
	if err != nil {
		return err
	}
	for _, role := range roles {
		for _, table := range g.GrantTable.Tables {
			schemaName, err := core.GetSchemaName(ctx, nil, table.Schema)
			if err != nil {
				return err
			}
			key := auth.TablePrivilegeKey{
				Role:  userRole.ID(),
				Table: doltdb.TableName{Name: table.Name, Schema: schemaName},
			}
			for _, privilege := range g.GrantTable.Privileges {
				grantedBy := auth.HasTablePrivilegeGrantOption(key, privilege)
				if !grantedBy.IsValid() {
					// TODO: grab the actual error message
					return errors.Errorf(`role "%s" does not have permission to grant this privilege`, userRole.Name)
				}
				auth.AddTablePrivilege(auth.TablePrivilegeKey{
					Role:  role.ID(),
					Table: doltdb.TableName{Name: table.Name, Schema: schemaName},
				}, auth.GrantedPrivilege{
					Privilege: privilege,
					GrantedBy: grantedBy,
				}, g.WithGrantOption)
			}
		}
	}
	return nil
}

// grantSchema handles *GrantSchema from within RowIter.
func (g *Grant) grantSchema(ctx *sql.Context) error {
	roles, userRole, err := g.common(ctx)
	if err != nil {
		return err
	}
	for _, role := range roles {
		for _, schema := range g.GrantSchema.Schemas {
			key := auth.SchemaPrivilegeKey{
				Role:   userRole.ID(),
				Schema: schema,
			}
			for _, privilege := range g.GrantSchema.Privileges {
				grantedBy := auth.HasSchemaPrivilegeGrantOption(key, privilege)
				if !grantedBy.IsValid() {
					// TODO: grab the actual error message
					return errors.Errorf(`role "%s" does not have permission to grant this privilege`, userRole.Name)
				}
				auth.AddSchemaPrivilege(auth.SchemaPrivilegeKey{
					Role:   role.ID(),
					Schema: schema,
				}, auth.GrantedPrivilege{
					Privilege: privilege,
					GrantedBy: grantedBy,
				}, g.WithGrantOption)
			}
		}
	}
	return nil
}

// grantDatabase handles *GrantDatabase from within RowIter.
func (g *Grant) grantDatabase(ctx *sql.Context) error {
	roles, userRole, err := g.common(ctx)
	if err != nil {
		return err
	}
	for _, role := range roles {
		for _, database := range g.GrantDatabase.Databases {
			key := auth.DatabasePrivilegeKey{
				Role: userRole.ID(),
				Name: database,
			}
			for _, privilege := range g.GrantDatabase.Privileges {
				grantedBy := auth.HasDatabasePrivilegeGrantOption(key, privilege)
				if !grantedBy.IsValid() {
					// TODO: grab the actual error message
					return errors.Errorf(`role "%s" does not have permission to grant this privilege`, userRole.Name)
				}
				auth.AddDatabasePrivilege(auth.DatabasePrivilegeKey{
					Role: role.ID(),
					Name: database,
				}, auth.GrantedPrivilege{
					Privilege: privilege,
					GrantedBy: grantedBy,
				}, g.WithGrantOption)
			}
		}
	}
	return nil
}

// grantRole handles *GrantRole from within RowIter.
func (g *Grant) grantRole(ctx *sql.Context) error {
	members, userRole, err := g.common(ctx)
	if err != nil {
		return err
	}
	groups := make([]auth.Role, len(g.GrantRole.Groups))
	for i, groupName := range g.GrantRole.Groups {
		groups[i] = auth.GetRole(groupName)
		if !groups[i].IsValid() {
			return errors.Errorf(`role "%s" does not exist`, groupName)
		}
	}
	for _, member := range members {
		for _, group := range groups {
			memberGroupID, _, withAdminOption := auth.IsRoleAMember(userRole.ID(), group.ID())
			if !memberGroupID.IsValid() || !withAdminOption {
				// TODO: grab the actual error message
				return errors.Errorf(`role "%s" does not have permission to grant role "%s"`, userRole.Name, group.Name)
			}
			auth.AddMemberToGroup(member.ID(), group.ID(), g.WithGrantOption, memberGroupID)
		}
	}
	return nil
}
