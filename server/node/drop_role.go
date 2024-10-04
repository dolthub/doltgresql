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

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/plan"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/server/auth"
)

// DropRole handles the DROP ROLE statement.
type DropRole struct {
	Names    []string
	IfExists bool
}

var _ sql.ExecSourceRel = (*DropRole)(nil)
var _ vitess.Injectable = (*DropRole)(nil)

// CheckPrivileges implements the interface sql.ExecSourceRel.
func (c *DropRole) CheckPrivileges(ctx *sql.Context, opChecker sql.PrivilegedOperationChecker) bool {
	return true
}

// Children implements the interface sql.ExecSourceRel.
func (c *DropRole) Children() []sql.Node {
	return nil
}

// IsReadOnly implements the interface sql.ExecSourceRel.
func (c *DropRole) IsReadOnly() bool {
	return false
}

// Resolved implements the interface sql.ExecSourceRel.
func (c *DropRole) Resolved() bool {
	return true
}

// RowIter implements the interface sql.ExecSourceRel.
func (c *DropRole) RowIter(ctx *sql.Context, r sql.Row) (sql.RowIter, error) {
	// TODO: handle concurrency
	// First we'll loop over all of the names to check that they all exist
	for _, roleName := range c.Names {
		if !auth.RoleExists(roleName) && !c.IfExists {
			return nil, fmt.Errorf(`role "%s" does not exist`, roleName)
		}
	}
	// Then we'll loop again, dropping all of the users
	for _, roleName := range c.Names {
		auth.DropRole(roleName)
	}
	return sql.RowsToRowIter(), nil
}

// Schema implements the interface sql.ExecSourceRel.
func (c *DropRole) Schema() sql.Schema {
	return nil
}

// String implements the interface sql.ExecSourceRel.
func (c *DropRole) String() string {
	return "DROP ROLE"
}

// WithChildren implements the interface sql.ExecSourceRel.
func (c *DropRole) WithChildren(children ...sql.Node) (sql.Node, error) {
	return plan.NillaryWithChildren(c, children...)
}

// WithResolvedChildren implements the interface vitess.Injectable.
func (c *DropRole) WithResolvedChildren(children []any) (any, error) {
	if len(children) != 0 {
		return nil, fmt.Errorf("invalid vitess child count, expected `0` but got `%d`", len(children))
	}
	return c, nil
}
