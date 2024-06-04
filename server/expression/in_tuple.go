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

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// InTuple represents a VALUE IN (<VALUES>) expression.
type InTuple struct {
	left  sql.Expression
	right expression.Tuple
}

var _ vitess.Injectable = (*BinaryOperator)(nil)
var _ sql.Expression = (*BinaryOperator)(nil)
var _ expression.BinaryExpression = (*BinaryOperator)(nil)

// NewInTuple returns a new *InTuple.
func NewInTuple() *InTuple {
	return &InTuple{
		left:  nil,
		right: nil,
	}
}

// Children implements the sql.Expression interface.
func (it *InTuple) Children() []sql.Expression {
	return []sql.Expression{it.left, it.right}
}

// Eval implements the sql.Expression interface.
func (it *InTuple) Eval(ctx *sql.Context, row sql.Row) (any, error) {
	left, err := it.left.Eval(ctx, row)
	if err != nil {
		return nil, err
	}
	rightInterface, err := it.right.Eval(ctx, row)
	if err != nil {
		return nil, err
	}
	rightValues, ok := rightInterface.([]any)
	if !ok {
		// Tuples will return the value directly if it has a length of one, so we'll check for that first
		if len(it.right) == 1 {
			rightValues = []any{rightInterface}
		} else {
			return nil, fmt.Errorf("%T: expected right child to return `%T` but returned `%T`", it, []any{}, rightInterface)
		}
	}
	leftType, ok := it.left.Type().(pgtypes.DoltgresType)
	if !ok {
		return nil, fmt.Errorf("%T: GMS type `%s` on left child", it, it.left.Type().String())
	}
	for i, rightValue := range rightValues {
		rightType, ok := it.right[i].Type().(pgtypes.DoltgresType)
		if !ok {
			return nil, fmt.Errorf("%T: GMS type `%s` within right child", it, it.right[i].Type().String())
		}
		// TODO: this should use the BinaryOperator expression, but since equality is not yet implemented, we implicitly cast
		if !leftType.Equals(rightType) {
			castFunc := framework.GetImplicitCast(rightType.BaseID(), leftType.BaseID())
			if castFunc == nil {
				return nil, fmt.Errorf("operator does not exist: %s = %s",
					leftType.String(), rightType.String())
			}
			rightValue, err = castFunc(ctx, rightValue, leftType)
			if err != nil {
				return nil, err
			}
		}
		if res, err := leftType.Compare(left, rightValue); err != nil {
			return nil, err
		} else if res == 0 {
			return true, nil
		}
	}
	return false, nil
}

// IsNullable implements the sql.Expression interface.
func (it *InTuple) IsNullable() bool {
	return true
}

// Resolved implements the sql.Expression interface.
func (it *InTuple) Resolved() bool {
	return it.left != nil && it.left.Resolved() && it.right != nil && it.right.Resolved()
}

// String implements the sql.Expression interface.
func (it *InTuple) String() string {
	if it.left == nil || it.right == nil {
		return "? IN ?"
	}
	return fmt.Sprintf("%s IN %s", it.left.String(), it.right.String())
}

// Type implements the sql.Expression interface.
func (it *InTuple) Type() sql.Type {
	return pgtypes.Bool
}

// WithChildren implements the sql.Expression interface.
func (it *InTuple) WithChildren(children ...sql.Expression) (sql.Expression, error) {
	if len(children) != 2 {
		return nil, sql.ErrInvalidChildrenNumber.New(it, len(children), 2)
	}
	rightTuple, ok := children[1].(expression.Tuple)
	if !ok {
		return nil, fmt.Errorf("%T: expected right child to be `%T` but has type `%T`", it, expression.Tuple{}, children[1])
	}
	return &InTuple{
		left:  children[0],
		right: rightTuple,
	}, nil
}

// WithResolvedChildren implements the vitess.InjectableExpression interface.
func (it *InTuple) WithResolvedChildren(children []any) (any, error) {
	if len(children) != 2 {
		return nil, fmt.Errorf("invalid vitess child count, expected `2` but got `%d`", len(children))
	}
	left, ok := children[0].(sql.Expression)
	if !ok {
		return nil, fmt.Errorf("expected vitess child to be an expression but has type `%T`", children[0])
	}
	right, ok := children[1].(expression.Tuple)
	if !ok {
		return nil, fmt.Errorf("expected vitess child to be an expression tuple but has type `%T`", children[1])
	}
	return it.WithChildren(left, right)
}

// Left implements the expression.BinaryExpression interface.
func (it *InTuple) Left() sql.Expression {
	return it.left
}

// Right implements the expression.BinaryExpression interface.
func (it *InTuple) Right() sql.Expression {
	return it.right
}
