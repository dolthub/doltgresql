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

	// These variables are used so that we can resolve the comparison functions once and reuse them as we iterate over rows.
	// These are assigned in WithChildren, so refer there for more information.
	staticLiteral *Literal
	arrayLiterals []*Literal
	compFuncs     []*framework.CompiledFunction
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

// Children implements the Expression interface.
func (a *AnyExpr) Children() []sql.Expression {
	return []sql.Expression{a.leftExpr, a.rightExpr}
}

// Resolved implements the Expression interface.
func (a *AnyExpr) Resolved() bool {
	if a.leftExpr == nil || !a.leftExpr.Resolved() || a.rightExpr == nil || !a.rightExpr.Resolved() || len(a.compFuncs) == 0 {
		return false
	}
	for _, compFunc := range a.compFuncs {
		if !compFunc.Resolved() {
			return false
		}
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
	if len(a.compFuncs) == 0 {
		return nil, fmt.Errorf("%T: cannot Eval as it has not been fully resolved", a)
	}
	// First we'll evaluate everything before we do the comparisons
	left, err := a.leftExpr.Eval(ctx, row)
	if err != nil {
		return nil, err
	}

	var rightInterface interface{}
	if sub, ok := a.rightExpr.(*plan.Subquery); ok {
		// TODO: This sometimes panics in `evalMultiple` for subqueries that return
		// more than one row, when len(row) > len(iter.Next())
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
		return nil, nil
	}

	rightValues, ok := rightInterface.([]any)
	if !ok {
		return nil, fmt.Errorf("%T: expected right child to return `%T` but returned `%T`", a, []any{}, rightInterface)
	}
	if len(rightValues) == 0 {
		return nil, nil
	}

	// TODO: This is a workaround some expression types (Subquery, GetFields)
	// where we don't know the number of right values beforehand
	if len(a.arrayLiterals) == 1 && len(rightValues) != 1 {
		op, err := framework.GetOperatorFromString(a.subOperator)
		if err != nil {
			return nil, err
		}

		for i := len(a.arrayLiterals); i < len(rightValues); i++ {
			arrayLiteral := &Literal{typ: a.arrayLiterals[0].typ}
			a.arrayLiterals = append(a.arrayLiterals, arrayLiteral)
			compFunc := framework.GetBinaryFunction(op).Compile("internal_any_comparison", a.staticLiteral, a.arrayLiterals[i])
			a.compFuncs = append(a.compFuncs, compFunc)
		}
	}

	if len(a.arrayLiterals) != len(rightValues) {
		return nil, fmt.Errorf("%T: expected right child to return `%d` values but returned `%d`", a, len(a.arrayLiterals), len(rightValues))
	}

	// Next we'll assign our evaluated values to the expressions that the comparison functions reference
	a.staticLiteral.value = left
	for i, rightValue := range rightValues {
		a.arrayLiterals[i].value = rightValue
	}
	// Now we can loop over all of the comparison functions, as they'll reference their respective values
	for _, compFunc := range a.compFuncs {
		result, err := compFunc.Eval(ctx, row)
		if err != nil {
			return nil, err
		}
		if result.(bool) {
			return true, nil
		}
	}

	return false, nil
}

// WithChildren implements the Expression interface.
func (a *AnyExpr) WithChildren(children ...sql.Expression) (sql.Expression, error) {
	if len(children) != 2 {
		return nil, sql.ErrInvalidChildrenNumber.New(a, len(children), 2)
	}

	rightTypes, err := a.getRightTypes(children[1])
	if err != nil {
		return nil, err
	}

	op, err := framework.GetOperatorFromString(a.subOperator)
	if err != nil {
		return nil, err
	}

	if leftType, ok := children[0].Type().(pgtypes.DoltgresType); ok {
		// Resolve comparison functions once and reuse the functions in Eval.
		staticLiteral := &Literal{typ: leftType}
		arrayLiterals := make([]*Literal, len(rightTypes))
		// Each expression may be a different type (which is valid), so we need a comparison function for each expression.
		compFuncs := make([]*framework.CompiledFunction, len(rightTypes))
		for i, rightType := range rightTypes {
			arrayLiterals[i] = &Literal{typ: rightType}
			compFuncs[i] = framework.GetBinaryFunction(op).Compile("internal_any_comparison", staticLiteral, arrayLiterals[i])
			if compFuncs[i] == nil {
				return nil, fmt.Errorf("operator does not exist: %s = %s", leftType.String(), rightType.String())
			}
			if compFuncs[i].Type().(pgtypes.DoltgresType).BaseID() != pgtypes.DoltgresTypeBaseID_Bool {
				// This should never happen, but this is just to be safe
				return nil, fmt.Errorf("%T: found equality comparison that does not return a bool", a)
			}
		}
		return &AnyExpr{
			leftExpr:      children[0],
			rightExpr:     children[1],
			subOperator:   a.subOperator,
			name:          a.name,
			staticLiteral: staticLiteral,
			arrayLiterals: arrayLiterals,
			compFuncs:     compFuncs,
		}, nil
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

// getRightTypes returns the types of the right expression.
func (a *AnyExpr) getRightTypes(right sql.Expression) ([]pgtypes.DoltgresType, error) {
	if sub, ok := right.(*plan.Subquery); ok {
		return getSubQueryTypes(sub)
	}

	return getSqlExpressionTypes(right)
}

// getSubQueryTypes returns the types of the subquery schema.
func getSubQueryTypes(sub *plan.Subquery) ([]pgtypes.DoltgresType, error) {
	schema := sub.Query.Schema()
	subTypes := make([]pgtypes.DoltgresType, len(schema))
	for i, col := range schema {
		dgType, ok := col.Type.(pgtypes.DoltgresType)
		if !ok {
			return nil, fmt.Errorf("expected right child to be a DoltgresType but got `%T`", sub)
		}
		subTypes[i] = dgType
	}
	return subTypes, nil
}

// getSqlExpressionTypes returns the types of the right expression.
func getSqlExpressionTypes(expr sql.Expression) ([]pgtypes.DoltgresType, error) {
	var length int
	var dgType pgtypes.DoltgresType
	var ok bool
	switch r := expr.(type) {
	case *Literal:
		if val, valOk := r.Value().([]interface{}); valOk {
			length = len(val)
		}
		dgType, ok = r.Type().(pgtypes.DoltgresType)
	case *expression.GetField:
		length = 1
		dgType, ok = r.Type().(pgtypes.DoltgresType)
	default:
		dgType, ok = expr.Type().(pgtypes.DoltgresType)
		length = len(expr.Children())
	}
	if !ok {
		return nil, fmt.Errorf("expected right child to be a DoltgresType but got `%T`", expr)
	}
	if length == 0 {
		return nil, nil
	}

	if at, ok := dgType.(pgtypes.DoltgresArrayType); ok {
		dgType = at.BaseType()
	}

	rightTypes := make([]pgtypes.DoltgresType, length)
	for i := range length {
		rightTypes[i] = dgType
	}

	return rightTypes, nil
}
