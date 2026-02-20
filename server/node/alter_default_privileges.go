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

package node

import (
	"github.com/cockroachdb/errors"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/plan"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/server/auth"
)

// AlterDefaultPrivileges handles the ALTER DEFAULT PRIVILEGES statement.
type AlterDefaultPrivileges struct {
	// ForRoles is the list of roles for whose future object creations these defaults apply.
	// If empty, defaults to the current session user.
	ForRoles []string
	// Schemas restricts the default privileges to specific schemas. Empty means all schemas.
	Schemas []string
	// ObjectType is the type of object (TABLE, SEQUENCE, FUNCTION, TYPE, SCHEMA).
	ObjectType auth.PrivilegeObject
	// Grantees are the roles that receive (GRANT) or lose (REVOKE) the privileges.
	Grantees []string
	// Privileges is the list of privileges to grant or revoke.
	Privileges []auth.Privilege
	// WithGrantOption indicates WITH GRANT OPTION (for GRANT) or GRANT OPTION FOR (for REVOKE).
	WithGrantOption bool
	// IsGrant is true for GRANT, false for REVOKE.
	IsGrant bool
	// Cascade is true when CASCADE is specified on REVOKE.
	Cascade bool
}

var _ sql.ExecSourceRel = (*AlterDefaultPrivileges)(nil)
var _ vitess.Injectable = (*AlterDefaultPrivileges)(nil)

// Children implements the interface sql.ExecSourceRel.
func (a *AlterDefaultPrivileges) Children() []sql.Node {
	return nil
}

// IsReadOnly implements the interface sql.ExecSourceRel.
func (a *AlterDefaultPrivileges) IsReadOnly() bool {
	return false
}

// Resolved implements the interface sql.ExecSourceRel.
func (a *AlterDefaultPrivileges) Resolved() bool {
	return true
}

// RowIter implements the interface sql.ExecSourceRel.
func (a *AlterDefaultPrivileges) RowIter(ctx *sql.Context, r sql.Row) (sql.RowIter, error) {
	if a.Cascade {
		return nil, errors.New("ALTER DEFAULT PRIVILEGES does not yet support CASCADE")
	}

	var err error
	auth.LockWrite(func() {
		// Resolve the ForRoles: if none specified, use the current session user.
		forRoles := make([]auth.Role, 0, len(a.ForRoles))
		if len(a.ForRoles) == 0 {
			userRole := auth.GetRole(ctx.Client().User)
			if !userRole.IsValid() {
				err = errors.Errorf(`role "%s" does not exist`, ctx.Client().User)
				return
			}
			forRoles = append(forRoles, userRole)
		} else {
			for _, name := range a.ForRoles {
				role := auth.GetRole(name)
				if !role.IsValid() {
					err = errors.Errorf(`role "%s" does not exist`, name)
					return
				}
				forRoles = append(forRoles, role)
			}
		}

		// Validate grantees.
		grantees := make([]auth.Role, len(a.Grantees))
		for i, name := range a.Grantees {
			grantees[i] = auth.GetRole(name)
			if !grantees[i].IsValid() {
				err = errors.Errorf(`role "%s" does not exist`, name)
				return
			}
		}

		// Resolve schemas: empty slice means all schemas (represented as empty string key).
		schemas := a.Schemas
		if len(schemas) == 0 {
			schemas = []string{""}
		}

		if a.IsGrant {
			for _, forRole := range forRoles {
				for _, schema := range schemas {
					for _, grantee := range grantees {
						key := auth.DefaultPrivilegeKey{
							ForRole:    forRole.ID(),
							Schema:     schema,
							ObjectType: a.ObjectType,
							Grantee:    grantee.ID(),
						}
						for _, priv := range a.Privileges {
							auth.AddDefaultPrivilege(key, priv, a.WithGrantOption)
						}
					}
				}
			}
		} else {
			for _, forRole := range forRoles {
				for _, schema := range schemas {
					for _, grantee := range grantees {
						key := auth.DefaultPrivilegeKey{
							ForRole:    forRole.ID(),
							Schema:     schema,
							ObjectType: a.ObjectType,
							Grantee:    grantee.ID(),
						}
						for _, priv := range a.Privileges {
							auth.RemoveDefaultPrivilege(key, priv, a.WithGrantOption)
						}
					}
				}
			}
		}

		err = auth.PersistChanges()
	})
	if err != nil {
		return nil, err
	}
	return sql.RowsToRowIter(), nil
}

// Schema implements the interface sql.ExecSourceRel.
func (a *AlterDefaultPrivileges) Schema() sql.Schema {
	return nil
}

// String implements the interface sql.ExecSourceRel.
func (a *AlterDefaultPrivileges) String() string {
	return "ALTER DEFAULT PRIVILEGES"
}

// WithChildren implements the interface sql.ExecSourceRel.
func (a *AlterDefaultPrivileges) WithChildren(children ...sql.Node) (sql.Node, error) {
	return plan.NillaryWithChildren(a, children...)
}

// WithResolvedChildren implements the interface vitess.Injectable.
func (a *AlterDefaultPrivileges) WithResolvedChildren(children []any) (any, error) {
	if len(children) != 0 {
		return nil, ErrVitessChildCount.New(0, len(children))
	}
	return a, nil
}
