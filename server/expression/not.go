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

package expression

import (
	"fmt"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/expression"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// Not represents a NOT expression.
type Not struct {
	child sql.Expression
}

var _ vitess.Injectable = (*BinaryOperator)(nil)
var _ sql.Expression = (*BinaryOperator)(nil)
var _ expression.BinaryExpression = (*BinaryOperator)(nil)

// NewNot returns a new *Not.
func NewNot() *Not {
	return &Not{
		child: nil,
	}
}

// Children implements the sql.Expression interface.
func (n *Not) Children() []sql.Expression {
	return []sql.Expression{n.child}
}

// Eval implements the sql.Expression interface.
func (n *Not) Eval(ctx *sql.Context, row sql.Row) (any, error) {
	val, err := n.child.Eval(ctx, row)
	if err != nil {
		return nil, err
	}
	if val == nil {
		return nil, nil
	}
	boolVal, ok := val.(bool)
	if !ok {
		return nil, fmt.Errorf("NOT only applies to boolean values")
	}
	return !boolVal, nil
}

// IsNullable implements the sql.Expression interface.
func (n *Not) IsNullable() bool {
	return true
}

// Resolved implements the sql.Expression interface.
func (n *Not) Resolved() bool {
	return n.child != nil && n.child.Resolved()
}

// String implements the sql.Expression interface.
func (n *Not) String() string {
	if n.child == nil {
		return "NOT ?"
	}
	return "NOT " + n.child.String()
}

// Type implements the sql.Expression interface.
func (n *Not) Type() sql.Type {
	return pgtypes.Bool
}

// WithChildren implements the sql.Expression interface.
func (n *Not) WithChildren(children ...sql.Expression) (sql.Expression, error) {
	if len(children) != 1 {
		return nil, sql.ErrInvalidChildrenNumber.New(n, len(children), 1)
	}
	return &Not{
		child: children[0],
	}, nil
}

// WithResolvedChildren implements the vitess.InjectableExpression interface.
func (n *Not) WithResolvedChildren(children []any) (any, error) {
	if len(children) != 1 {
		return nil, fmt.Errorf("invalid vitess child count, expected `1` but got `%d`", len(children))
	}
	child, ok := children[0].(sql.Expression)
	if !ok {
		return nil, fmt.Errorf("expected vitess child to be an expression but has type `%T`", children[0])
	}
	return &Not{
		child: child,
	}, nil
}
