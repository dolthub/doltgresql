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
	"context"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/plan"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/server/auth"
)

// AlterDefaultPrivileges handles the ALTER DEFAULT PRIVILEGES statement.
type AlterDefaultPrivileges struct {
	OwnerRole   string
	Schemas     []string
	ObjectType  auth.PrivilegeObject
	Privileges  []auth.Privilege
	Grantees    []string
	Grant       bool // false = REVOKE
	GrantOption bool
	Cascade     bool
}

var _ sql.ExecSourceRel = (*AlterDefaultPrivileges)(nil)
var _ vitess.Injectable = (*AlterDefaultPrivileges)(nil)

// Children implements the interface sql.ExecSourceRel.
func (n *AlterDefaultPrivileges) Children() []sql.Node {
	return nil
}

// IsReadOnly implements the interface sql.ExecSourceRel.
func (n *AlterDefaultPrivileges) IsReadOnly() bool {
	return false
}

// Resolved implements the interface sql.ExecSourceRel.
func (n *AlterDefaultPrivileges) Resolved() bool {
	return true
}

// RowIter implements the interface sql.ExecSourceRel.
func (n *AlterDefaultPrivileges) RowIter(ctx *sql.Context, _ sql.Row) (sql.RowIter, error) {
	if n.Cascade {
		return nil, errors.New("ALTER DEFAULT PRIVILEGES does not yet support CASCADE")
	}
	var err error
	auth.LockWrite(func() {
		err = n.execute(ctx)
		if err != nil {
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
func (n *AlterDefaultPrivileges) Schema(ctx *sql.Context) sql.Schema {
	return nil
}

// String implements the interface sql.ExecSourceRel.
func (n *AlterDefaultPrivileges) String() string {
	return "ALTER DEFAULT PRIVILEGES"
}

// WithChildren implements the interface sql.ExecSourceRel.
func (n *AlterDefaultPrivileges) WithChildren(ctx *sql.Context, children ...sql.Node) (sql.Node, error) {
	return plan.NillaryWithChildren(n, children...)
}

// WithResolvedChildren implements the interface vitess.Injectable.
func (n *AlterDefaultPrivileges) WithResolvedChildren(ctx context.Context, children []any) (any, error) {
	if len(children) != 0 {
		return nil, ErrVitessChildCount.New(0, len(children))
	}
	return n, nil
}

// execute performs the actual default privilege changes.
func (n *AlterDefaultPrivileges) execute(ctx *sql.Context) error {
	ownerRole, err := n.resolveOwnerRole(ctx)
	if err != nil {
		return err
	}

	granteeRoles := make([]auth.Role, len(n.Grantees))
	for i, name := range n.Grantees {
		role := auth.GetRole(name)
		if !role.IsValid() {
			return errors.Errorf(`role "%s" does not exist`, name)
		}
		granteeRoles[i] = role
	}

	schemas := n.Schemas
	// empty list means all schemas
	if len(schemas) == 0 {
		// TODO: get all schemas
		schemas = []string{""}
	}
	for _, schema := range schemas {
		key := auth.DefaultPrivilegeKey{
			OwnerRole:  ownerRole.ID(),
			Schema:     schema,
			ObjectType: n.ObjectType,
		}
		for _, granteeRole := range granteeRoles {
			for _, priv := range n.Privileges {
				grantedPrivilege := auth.GrantedPrivilege{
					Privilege: priv,
					GrantedBy: ownerRole.ID(),
				}
				if n.Grant {
					auth.AddDefaultPrivilege(key, granteeRole.ID(), grantedPrivilege, n.GrantOption)
				} else {
					auth.RemoveDefaultPrivilege(key, granteeRole.ID(), grantedPrivilege, n.GrantOption)
				}
			}
		}
	}
	return nil
}

// resolveOwnerRoles returns the roles that own the default privileges being modified.
// When no roles are explicitly specified, the current session user is used.
func (n *AlterDefaultPrivileges) resolveOwnerRole(ctx *sql.Context) (auth.Role, error) {
	// empty means current user
	if n.OwnerRole == "" {
		userRole := auth.GetRole(ctx.Client().User)
		if !userRole.IsValid() {
			return auth.Role{}, errors.Errorf(`role "%s" does not exist`, ctx.Client().User)
		}
		return userRole, nil
	} else {
		role := auth.GetRole(n.OwnerRole)
		if !role.IsValid() {
			return auth.Role{}, errors.Errorf(`role "%s" does not exist`, n.OwnerRole)
		}
		return role, nil
	}
}
