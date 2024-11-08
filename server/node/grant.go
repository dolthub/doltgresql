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

// Grant handles all of the GRANT statements.
type Grant struct {
	GrantTable      *GrantTable
	ToRoles         []string
	WithGrantOption bool // Does not apply to the GRANT <roles> TO <roles> statement
	GrantedBy       string
}

// GrantTable specifically handles the GRANT ... ON TABLE statement.
type GrantTable struct {
	Privileges         []auth.Privilege
	Tables             []doltdb.TableName
	AllTablesInSchemas []string
}

var _ sql.ExecSourceRel = (*Grant)(nil)
var _ vitess.Injectable = (*Grant)(nil)

// CheckPrivileges implements the interface sql.ExecSourceRel.
func (g *Grant) CheckPrivileges(ctx *sql.Context, opChecker sql.PrivilegedOperationChecker) bool {
	return true
}

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
			if len(g.GrantTable.AllTablesInSchemas) > 0 {
				err = fmt.Errorf("granting privileges to all tables in the schema is not yet supported")
				return
			}
			roles := make([]auth.Role, len(g.ToRoles))
			// First we'll verify that all of the roles exist
			for i, roleName := range g.ToRoles {
				roles[i] = auth.GetRole(roleName)
				if !roles[i].IsValid() {
					err = fmt.Errorf(`role "%s" does not exist`, roleName)
					return
				}
			}
			// Then we'll check that the role that is granting the privileges exists
			userRole := auth.GetRole(ctx.Client().User)
			if !userRole.IsValid() {
				err = fmt.Errorf(`role "%s" does not exist`, ctx.Client().User)
				return
			}
			var grantedByID auth.RoleID
			if len(g.GrantedBy) != 0 {
				// TODO: check the role chain to see if this session's user can assume this role
				grantedByRole := auth.GetRole(g.GrantedBy)
				if !grantedByRole.IsValid() {
					err = fmt.Errorf(`role "%s" does not exist`, g.GrantedBy)
					return
				}
				grantedByID = grantedByRole.ID()
				// TODO: check if owners may arbitrarily set the GRANTED BY
				if !userRole.IsSuperUser {
					err = errors.New("REVOKE currently only allows superusers to set GRANTED BY")
					return
				}
			} else {
				grantedByID = userRole.ID()
			}
			// Next we'll assign all of the privileges to each role
			for _, role := range roles {
				for _, table := range g.GrantTable.Tables {
					var schemaName string
					schemaName, err = core.GetSchemaName(ctx, nil, table.Schema)
					if err != nil {
						return
					}
					key := auth.TablePrivilegeKey{
						Role:  role.ID(),
						Table: doltdb.TableName{Name: table.Name, Schema: schemaName},
					}
					isOwner := auth.IsOwner(auth.OwnershipKey{
						PrivilegeObject: auth.PrivilegeObject_TABLE,
						Schema:          schemaName,
						Name:            table.Name,
					}, userRole.ID())
					for _, privilege := range g.GrantTable.Privileges {
						if !userRole.IsSuperUser && !isOwner && !auth.HasTablePrivilegeGrantOption(key, privilege) {
							// TODO: grab the actual error message
							err = fmt.Errorf(`role "%s" does not have permission to grant this privilege`, userRole.Name)
							return
						}
						auth.AddTablePrivilege(key, auth.GrantedPrivilege{
							Privilege: privilege,
							GrantedBy: grantedByID,
						}, g.WithGrantOption)
					}
				}
			}
		default:
			err = fmt.Errorf("GRANT statement is not yet supported")
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
