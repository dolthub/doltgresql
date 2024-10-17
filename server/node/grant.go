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
	"fmt"

	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/plan"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

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
func (c *Grant) CheckPrivileges(ctx *sql.Context, opChecker sql.PrivilegedOperationChecker) bool {
	return true
}

// Children implements the interface sql.ExecSourceRel.
func (c *Grant) Children() []sql.Node {
	return nil
}

// IsReadOnly implements the interface sql.ExecSourceRel.
func (c *Grant) IsReadOnly() bool {
	return false
}

// Resolved implements the interface sql.ExecSourceRel.
func (c *Grant) Resolved() bool {
	return true
}

// RowIter implements the interface sql.ExecSourceRel.
func (c *Grant) RowIter(ctx *sql.Context, r sql.Row) (sql.RowIter, error) {
	switch {
	case c.GrantTable != nil:
		if len(c.GrantTable.AllTablesInSchemas) > 0 {
			return nil, fmt.Errorf("granting privileges to all tables in the schema is not yet supported")
		}
		roles := make([]auth.Role, len(c.ToRoles))
		// First we'll verify that all of the roles exist
		for i, roleName := range c.ToRoles {
			roles[i] = auth.GetRole(roleName)
			if !roles[i].IsValid() {
				return nil, fmt.Errorf(`role "%s" does not exist`, roleName)
			}
		}
		// Then we'll check that the role that is granting the privileges exists
		var grantedByID auth.RoleID
		if len(c.GrantedBy) != 0 {
			// TODO: check the role chain to see if this session's user can assume this role
			grantedByRole := auth.GetRole(c.GrantedBy)
			if !grantedByRole.IsValid() {
				return nil, fmt.Errorf(`role "%s" does not exist`, c.GrantedBy)
			}
			grantedByID = grantedByRole.ID()
		} else {
			userRole := auth.GetRole(ctx.Client().User)
			if !userRole.IsValid() {
				return nil, fmt.Errorf(`role "%s" does not exist`, ctx.Client().User)
			}
			grantedByID = userRole.ID()
		}
		// TODO: check WITH GRANT OPTION, ownership, and superuser status before allowing this user to grant privileges
		// Next we'll assign all of the privileges to each role
		for _, role := range roles {
			for _, table := range c.GrantTable.Tables {
				key := auth.TablePrivilegeKey{
					Role:  role.ID(),
					Table: table,
				}
				for _, privilege := range c.GrantTable.Privileges {
					auth.AddTablePrivilege(key, auth.GrantedPrivilege{
						Privilege:       privilege,
						WithGrantOption: c.WithGrantOption,
						GrantedBy:       grantedByID,
					})
				}
			}
		}
	default:
		return nil, fmt.Errorf("GRANT statement is not yet supported")
	}
	return sql.RowsToRowIter(), nil
}

// Schema implements the interface sql.ExecSourceRel.
func (c *Grant) Schema() sql.Schema {
	return nil
}

// String implements the interface sql.ExecSourceRel.
func (c *Grant) String() string {
	switch {
	case c.GrantTable != nil:
		return "GRANT TABLE"
	default:
		return "GRANT"
	}
}

// WithChildren implements the interface sql.ExecSourceRel.
func (c *Grant) WithChildren(children ...sql.Node) (sql.Node, error) {
	return plan.NillaryWithChildren(c, children...)
}

// WithResolvedChildren implements the interface vitess.Injectable.
func (c *Grant) WithResolvedChildren(children []any) (any, error) {
	if len(children) != 0 {
		return nil, ErrVitessChildCount.New(0, len(children))
	}
	return c, nil
}
