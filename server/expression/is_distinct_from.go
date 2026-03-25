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
	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/server/functions/framework"

	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// IsDistinctFrom represents IS DISTINCT FROM expression.
type IsDistinctFrom struct {
	leftExpr     sql.Expression
	rightExpr    sql.Expression
	notEqualFunc framework.Function
}

var _ vitess.Injectable = (*IsDistinctFrom)(nil)
var _ sql.Expression = (*IsDistinctFrom)(nil)

// NewIsDistinctFrom returns a new *IsDistinctFrom.
func NewIsDistinctFrom() *IsDistinctFrom {
	return &IsDistinctFrom{
		leftExpr:  nil,
		rightExpr: nil,
	}
}

// Children implements the sql.Expression interface.
func (n *IsDistinctFrom) Children() []sql.Expression {
	return []sql.Expression{n.leftExpr, n.rightExpr}
}

// Eval implements the sql.Expression interface.
func (n *IsDistinctFrom) Eval(ctx *sql.Context, row sql.Row) (any, error) {
	val, err := n.notEqualFunc.Eval(ctx, row)
	if err != nil {
		return nil, err
	}
	if val == nil {
		left, err := n.leftExpr.Eval(ctx, row)
		if err != nil {
			return nil, err
		}
		right, err := n.rightExpr.Eval(ctx, row)
		if err != nil {
			return nil, err
		}
		if left == nil && right == nil {
			return false, nil
		}
		return true, nil
	}
	return val, nil
	//// Both values must have the same type
	//res, err := n.leftExpr.Type().(*pgtypes.DoltgresType).Compare(ctx, left, right)
	//if err != nil {
	//	return nil, err
	//}
	//return res != 0, nil
}

// IsNullable implements the sql.Expression interface.
func (n *IsDistinctFrom) IsNullable() bool {
	return true
}

// Resolved implements the sql.Expression interface.
func (n *IsDistinctFrom) Resolved() bool {
	if n.leftExpr == nil || n.rightExpr == nil {
		return false
	}
	return n.leftExpr.Resolved() && n.rightExpr.Resolved()
}

// String implements the sql.Expression interface.
func (n *IsDistinctFrom) String() string {
	return n.leftExpr.String() + " IS DISTINCT FROM " + n.rightExpr.String()
}

// Type implements the sql.Expression interface.
func (n *IsDistinctFrom) Type() sql.Type {
	return pgtypes.Bool
}

// WithChildren implements the sql.Expression interface.
func (n *IsDistinctFrom) WithChildren(children ...sql.Expression) (sql.Expression, error) {
	if len(children) != 2 {
		return nil, sql.ErrInvalidChildrenNumber.New(n, len(children), 2)
	}
	if n.notEqualFunc != nil {
		compiledFunc, err := n.notEqualFunc.WithChildren(children...)
		if err != nil {
			return nil, err
		}
		return &IsDistinctFrom{
			leftExpr:     children[0],
			rightExpr:    children[1],
			notEqualFunc: compiledFunc.(framework.Function),
		}, nil
	} else {
		i, err := n.WithResolvedChildren([]any{children[0], children[1]})
		if err != nil {
			return nil, err
		}
		return i.(sql.Expression), nil
	}
}

// WithResolvedChildren implements the vitess.InjectableExpression interface.
func (n *IsDistinctFrom) WithResolvedChildren(children []any) (any, error) {
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

	compiledFunc := framework.GetBinaryFunction(framework.Operator_BinaryNotEqual).Compile("internal_binary_operator_func_<>", left, right)
	if compiledFunc == nil {
		return nil, errors.Errorf("input types do not match: %s %s", left.Type().String(), right.Type().String())
	}
	return &IsDistinctFrom{
		leftExpr:     left,
		rightExpr:    right,
		notEqualFunc: compiledFunc,
	}, nil
}
