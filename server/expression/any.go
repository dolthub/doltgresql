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
	"github.com/dolthub/go-mysql-server/sql/plan"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// AnyExpr represents the ANY/SOME expression.
type AnyExpr struct {
	leftExpr    sql.Expression
	rightExpr   sql.Expression
	subOperator string
	name        string // ANY or SOME
}

// NewAnyExpr creates a new AnyExpr expression.
func NewAnyExpr(subOperator string) *AnyExpr {
	return &AnyExpr{
		leftExpr:    nil,
		rightExpr:   nil,
		subOperator: subOperator,
		name:        "ANY",
	}
}

// NewSomeExpr creates a new AnyExpr expression for SOME.
func NewSomeExpr(subOperator string) *AnyExpr {
	return &AnyExpr{
		leftExpr:    nil,
		rightExpr:   nil,
		subOperator: subOperator,
		name:        "SOME",
	}
}

// Children implements the Expression interface.
func (a *AnyExpr) Children() []sql.Expression {
	return []sql.Expression{a.leftExpr, a.rightExpr}
}

// Resolved implements the Expression interface.
func (a *AnyExpr) Resolved() bool {
	if a.leftExpr == nil || !a.leftExpr.Resolved() || a.rightExpr == nil || !a.rightExpr.Resolved() {
		return false
	}
	return true
}

// IsNullable implements the Expression interface.
func (a *AnyExpr) IsNullable() bool {
	return a.leftExpr.IsNullable() || a.rightExpr.IsNullable()
}

// Type implements the Expression interface.
func (a *AnyExpr) Type() sql.Type {
	return pgtypes.Bool
}

// Eval implements the Expression interface.
func (a *AnyExpr) Eval(ctx *sql.Context, row sql.Row) (interface{}, error) {
	// First we'll evaluate everything before we do the comparisons
	left, err := a.leftExpr.Eval(ctx, row)
	if err != nil {
		return nil, err
	}

	var rightInterface interface{}
	if sub, ok := a.rightExpr.(*plan.Subquery); ok {
		rightInterface, err = sub.EvalMultiple(ctx, row)
		if err != nil {
			return nil, err
		}
	} else {
		rightInterface, err = a.rightExpr.Eval(ctx, row)
		if err != nil {
			return nil, err
		}
	}
	if rightInterface == nil {
		return false, nil
	}

	rightType, ok := a.rightExpr.Type().(pgtypes.DoltgresType)
	if !ok {
		return nil, fmt.Errorf("%T: cannot Eval as it has not been fully resolved", a)
	}
	if at, ok := rightType.(pgtypes.DoltgresArrayType); ok {
		rightType = at.BaseType()
	}

	rightValues, ok := rightInterface.([]any)
	if !ok {
		return nil, fmt.Errorf("%T: expected right child to return `%T` but returned `%T`", a, []any{}, rightInterface)
	}
	if len(rightValues) == 0 {
		return false, nil
	}

	if leftType, ok := a.leftExpr.Type().(pgtypes.DoltgresType); ok {
		for _, rightVal := range rightValues {
			op, err := a.getBinaryOperator()
			if err != nil {
				return nil, err
			}

			leftLiteral := &Literal{typ: leftType, value: left}
			rightLiteral := &Literal{typ: rightType, value: rightVal}
			compFunc := framework.GetBinaryFunction(op).Compile("internal_any_comparison", leftLiteral, rightLiteral)
			if compFunc == nil {
				return nil, fmt.Errorf("operator does not exist: %s = %s", leftType.String(), rightType.String())
			}
			if compFunc.Type().(pgtypes.DoltgresType).BaseID() != pgtypes.DoltgresTypeBaseID_Bool {
				// This should never happen, but this is just to be safe
				return nil, fmt.Errorf("%T: found equality comparison that does not return a bool", a)
			}

			result, err := compFunc.Eval(ctx, row)
			if err != nil {
				return nil, err
			}
			if result.(bool) {
				return true, nil
			}
		}
	}

	return false, nil
}

// WithChildren implements the Expression interface.
func (a *AnyExpr) WithChildren(children ...sql.Expression) (sql.Expression, error) {
	if len(children) != 2 {
		return nil, sql.ErrInvalidChildrenNumber.New(a, len(children), 2)
	}
	return &AnyExpr{
		leftExpr:    children[0],
		rightExpr:   children[1],
		subOperator: a.subOperator,
		name:        a.name,
	}, nil
}

// WithResolvedChildren implements the Expression interface.
func (a *AnyExpr) WithResolvedChildren(children []any) (any, error) {
	if len(children) != 2 {
		return nil, fmt.Errorf("invalid vitess child count, expected `2` but got `%d`", len(children))
	}
	left, ok := children[0].(sql.Expression)
	if !ok {
		return nil, fmt.Errorf("expected vitess child to be an expression but has type `%T`", children[0])
	}
	right, ok := children[1].(sql.Expression)
	if !ok {
		return nil, fmt.Errorf("expected vitess child to be an expression but has type `%T`", children[1])
	}
	return a.WithChildren(left, right)
}

// String implements the fmt.Stringer interface.
func (a *AnyExpr) String() string {
	if a.leftExpr == nil || a.rightExpr == nil {
		return fmt.Sprintf("? %s (?)", a.name)
	}
	return fmt.Sprintf("%s %s (%s)", a.leftExpr, a.name, a.rightExpr)
}

// DebugString implements the Expression interface.
func (a *AnyExpr) DebugString() string {
	return fmt.Sprintf("%s %s (%s)", sql.DebugString(a.leftExpr), a.name, sql.DebugString(a.rightExpr))
}

// getBinaryOperator returns the binary operator for the given subOperator.
func (a *AnyExpr) getBinaryOperator() (framework.Operator, error) {
	switch a.subOperator {
	case "=":
		return framework.Operator_BinaryEqual, nil
	case "<>", "!=":
		return framework.Operator_BinaryNotEqual, nil
	case "<":
		return framework.Operator_BinaryLessThan, nil
	case "<=":
		return framework.Operator_BinaryLessOrEqual, nil
	case ">":
		return framework.Operator_BinaryGreaterThan, nil
	case ">=":
		return framework.Operator_BinaryGreaterOrEqual, nil
	default:
		return 0, fmt.Errorf("unhandled SubOperator for %s `%s`", a.name, a.subOperator)
	}
}
