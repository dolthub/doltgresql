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
	FromRoles      []string
	GrantedBy      string
	GrantOptionFor bool
	Cascade        bool // When false, represents RESTRICT
}

// RevokeTable specifically handles the REVOKE ... ON TABLE statement.
type RevokeTable struct {
	Privileges         []auth.Privilege
	Tables             []doltdb.TableName
	AllTablesInSchemas []string
}

var _ sql.ExecSourceRel = (*Revoke)(nil)
var _ vitess.Injectable = (*Revoke)(nil)

// CheckPrivileges implements the interface sql.ExecSourceRel.
func (r *Revoke) CheckPrivileges(ctx *sql.Context, opChecker sql.PrivilegedOperationChecker) bool {
	return true
}

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
			if len(r.RevokeTable.AllTablesInSchemas) > 0 {
				err = fmt.Errorf("revoking privileges to all tables in the schema is not yet supported")
				return
			}
			roles := make([]auth.Role, len(r.FromRoles))
			// First we'll verify that all of the roles exist
			for i, roleName := range r.FromRoles {
				roles[i] = auth.GetRole(roleName)
				if !roles[i].IsValid() {
					err = fmt.Errorf(`role "%s" does not exist`, roleName)
					return
				}
			}
			// Then we'll check that the role that is revoking the privileges exists
			userRole := auth.GetRole(ctx.Client().User)
			if !userRole.IsValid() {
				err = fmt.Errorf(`role "%s" does not exist`, ctx.Client().User)
				return
			}
			var grantedByID auth.RoleID
			if len(r.GrantedBy) != 0 {
				grantedByRole := auth.GetRole(r.GrantedBy)
				if !grantedByRole.IsValid() {
					err = fmt.Errorf(`role "%s" does not exist`, r.GrantedBy)
					return
				}
				grantedByID = grantedByRole.ID()
			}
			// Next we'll remove the privileges
			for _, role := range roles {
				for _, table := range r.RevokeTable.Tables {
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
					for _, privilege := range r.RevokeTable.Privileges {
						// TODO: we don't have to exactly match the GRANTED BY ID, we can also check if it's in the access chain
						if !userRole.IsSuperUser && !isOwner && userRole.ID() != grantedByID {
							// TODO: grab the actual error message
							err = fmt.Errorf(`role "%s" does not have permission to revoke this privilege`, userRole.Name)
							return
						}
						auth.RemoveTablePrivilege(key, auth.GrantedPrivilege{
							Privilege: privilege,
							GrantedBy: grantedByID,
						}, r.GrantOptionFor)
					}
				}
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
