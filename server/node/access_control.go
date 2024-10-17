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

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/plan"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/server/auth"
	pgtransform "github.com/dolthub/doltgresql/server/transform"
)

// AccessControl handles privilege checking, ownership, etc. to the child node.
type AccessControl struct {
	Child        sql.Node
	Checks       []AccessControlPrivilegeCheck
	cachedTables *[]sql.Table
}

// AccessControlPrivilegeCheck contains the privilege and object that needs to be checked. AccessControl will find the
// appropriate objects to test by searching the child.
type AccessControlPrivilegeCheck struct {
	auth.Privilege
	auth.PrivilegeObject
}

var _ sql.ExecSourceRel = (*AccessControl)(nil)
var _ vitess.Injectable = (*AccessControl)(nil)

//var _ vitess.ExprWrapper = (*AccessControl)(nil) // TODO: uncomment me

// CheckAccess checks that the user has access to the underlying node.
func (ac *AccessControl) CheckAccess(ctx *sql.Context) error {
	for _, check := range ac.Checks {
		// TODO: the rest of the privileges
		switch check.Privilege {
		case auth.Privilege_SELECT:
			if err := ac.privilegeSelect(ctx, check); err != nil {
				return err
			}
		}
	}
	return nil
}

// CheckPrivileges implements the interface sql.ExecSourceRel. This is only used by GMS as it is built specifically for
// MySQL, and due to Go limitations, it cannot contain a dual implementation (by replacing the function pointer for
// example). It also cannot handle concepts such as Row-Level Security, as such a thing cannot be done without the use
// of another function or an analyzer rule.
func (ac *AccessControl) CheckPrivileges(ctx *sql.Context, opChecker sql.PrivilegedOperationChecker) bool {
	return true
}

// Children implements the interface sql.ExecSourceRel.
func (ac *AccessControl) Children() []sql.Node {
	return []sql.Node{ac.Child}
}

// IsReadOnly implements the interface sql.ExecSourceRel.
func (ac *AccessControl) IsReadOnly() bool {
	return ac.Child.IsReadOnly()
}

// Resolved implements the interface sql.ExecSourceRel.
func (ac *AccessControl) Resolved() bool {
	return ac.Child.Resolved()
}

// RowIter implements the interface sql.ExecSourceRel.
func (ac *AccessControl) RowIter(ctx *sql.Context, r sql.Row) (sql.RowIter, error) {
	// TODO: implement Row-Level Security -> https://www.postgresql.org/docs/15/ddl-rowsecurity.html
	if execSourceRel, ok := ac.Child.(sql.ExecSourceRel); ok {
		return execSourceRel.RowIter(ctx, r)
	}
	return sql.RowsToRowIter(), nil
}

// Schema implements the interface sql.ExecSourceRel.
func (ac *AccessControl) Schema() sql.Schema {
	return ac.Child.Schema()
}

// String implements the interface sql.ExecSourceRel.
func (ac *AccessControl) String() string {
	return ac.Child.String()
}

// WithChildren implements the interface sql.ExecSourceRel.
func (ac *AccessControl) WithChildren(children ...sql.Node) (sql.Node, error) {
	if len(children) != 1 {
		return nil, sql.ErrInvalidChildrenNumber.New(ac, len(children), 1)
	}
	nc := *ac
	nc.Child = children[0]
	return &nc, nil
}

// WithResolvedChildren implements the interface vitess.Injectable.
func (ac *AccessControl) WithResolvedChildren(children []any) (any, error) {
	if len(children) != 0 {
		return nil, ErrVitessChildCount.New(0, len(children))
	}
	return ac, nil
}

// findTables returns all tables that can be found in the child.
func (ac *AccessControl) findTables() {
	if ac.cachedTables != nil {
		return
	}
	var tables []sql.Table
	pgtransform.InspectNode(ac.Child, func(node sql.Node) bool {
		switch node := node.(type) {
		case *plan.ResolvedTable:
			tables = append(tables, node.Table)
		case sql.TableNode:
			tables = append(tables, node.UnderlyingTable())
		case sql.TableWrapper:
			tables = append(tables, node.Underlying())
		}
		return false
	})
	ac.cachedTables = &tables
}

// privilegeSelect handles the SELECT privilege.
func (ac *AccessControl) privilegeSelect(ctx *sql.Context, check AccessControlPrivilegeCheck) error {
	switch check.PrivilegeObject {
	case auth.PrivilegeObject_LARGE_OBJECT:
		// TODO: implement me
	case auth.PrivilegeObject_SEQUENCE:
		// TODO: implement me
	case auth.PrivilegeObject_TABLE:
		ac.findTables()
		if len(*ac.cachedTables) == 0 {
			return errors.New("cannot find tables to execute SELECT privilege check")
		}
		for _, table := range *ac.cachedTables {
			// TODO: check that the table is either owned by the user, or the user has a SELECT privilege
		}
	case auth.PrivilegeObject_TABLE_COLUMN:
		// TODO: implement me
	}
	return nil
}
