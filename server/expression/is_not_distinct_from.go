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

package expression

import (
	"context"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/expression"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// IsNotDistinctFrom represents IS NOT DISTINCT FROM expression.
type IsNotDistinctFrom struct {
	leftExpr           sql.Expression
	rightExpr          sql.Expression
	staticLeftLiteral  *expression.Literal
	staticRightLiteral *expression.Literal
	equalFunc          *framework.CompiledFunction
}

var _ vitess.Injectable = (*IsNotDistinctFrom)(nil)
var _ sql.Expression = (*IsNotDistinctFrom)(nil)

// NewIsNotDistinctFrom returns a new *IsNotDistinctFrom.
func NewIsNotDistinctFrom() *IsNotDistinctFrom {
	return &IsNotDistinctFrom{
		leftExpr:  nil,
		rightExpr: nil,
	}
}

// Children implements the sql.Expression interface.
func (n *IsNotDistinctFrom) Children() []sql.Expression {
	return []sql.Expression{n.leftExpr, n.rightExpr}
}

// Eval implements the sql.Expression interface.
func (n *IsNotDistinctFrom) Eval(ctx *sql.Context, row sql.Row) (any, error) {
	left, err := n.leftExpr.Eval(ctx, row)
	if err != nil {
		return nil, err
	}
	right, err := n.rightExpr.Eval(ctx, row)
	if err != nil {
		return nil, err
	}

	if left == nil && right == nil {
		return true, nil
	} else if left == nil || right == nil {
		return false, nil
	}

	n.staticLeftLiteral.Val = left
	n.staticRightLiteral.Val = right

	if n.equalFunc == nil {
		return nil, errors.Errorf("input types do not match: %s %s", n.leftExpr.Type(ctx).String(), n.rightExpr.Type(ctx).String())
	}
	return n.equalFunc.Eval(ctx, row)
}

// IsNullable implements the sql.Expression interface.
func (n *IsNotDistinctFrom) IsNullable(ctx *sql.Context) bool {
	return true
}

// Resolved implements the sql.Expression interface.
func (n *IsNotDistinctFrom) Resolved() bool {
	if n.leftExpr == nil || n.rightExpr == nil {
		return false
	}
	return n.leftExpr.Resolved() && n.rightExpr.Resolved()
}

// String implements the sql.Expression interface.
func (n *IsNotDistinctFrom) String() string {
	return n.leftExpr.String() + " IS NOT DISTINCT FROM " + n.rightExpr.String()
}

// Type implements the sql.Expression interface.
func (n *IsNotDistinctFrom) Type(ctx *sql.Context) sql.Type {
	return pgtypes.Bool
}

// WithChildren implements the sql.Expression interface.
func (n *IsNotDistinctFrom) WithChildren(ctx *sql.Context, children ...sql.Expression) (sql.Expression, error) {
	if len(children) != 2 {
		return nil, sql.ErrInvalidChildrenNumber.New(n, len(children), 2)
	}

	// This allows evaluating the arguments separate from function.Eval() in order to resolve NULL values.
	// This follows the same logic as InTuple expression.
	allAreWell := true
	leftType, ok := children[0].Type(ctx).(*pgtypes.DoltgresType)
	if !ok {
		allAreWell = false
	}
	rightType, ok := children[1].Type(ctx).(*pgtypes.DoltgresType)
	if !ok {
		allAreWell = false
	}
	staticLeftLiteral := expression.NewLiteral(nil, leftType)
	staticRightLiteral := expression.NewLiteral(nil, rightType)

	if allAreWell {
		cf := framework.GetBinaryFunction(framework.Operator_BinaryEqual).Compile(ctx, "internal_binary_operator_func_=", staticLeftLiteral, staticRightLiteral)
		return &IsNotDistinctFrom{
			leftExpr:           children[0],
			rightExpr:          children[1],
			staticLeftLiteral:  staticLeftLiteral,
			staticRightLiteral: staticRightLiteral,
			equalFunc:          cf,
		}, nil
	}

	return &IsNotDistinctFrom{
		leftExpr:  children[0],
		rightExpr: children[1],
	}, nil
}

// WithResolvedChildren implements the vitess.InjectableExpression interface.
func (n *IsNotDistinctFrom) WithResolvedChildren(ctx context.Context, children []any) (any, error) {
	if len(children) != 2 {
		return nil, errors.Errorf("invalid vitess child count, expected `2` but got `%d`", len(children))
	}
	left, ok := children[0].(sql.Expression)
	if !ok {
		return nil, errors.Errorf("expected vitess child to be an expression but has type `%T`", children[0])
	}
	right, ok := children[1].(sql.Expression)
	if !ok {
		return nil, errors.Errorf("expected vitess child to be an expression but has type `%T`", children[1])
	}

	return n.WithChildren(ctx.(*sql.Context), left, right)
}
