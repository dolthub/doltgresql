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

	subqueryAnyExpr   *subqueryAnyExpr
	expressionAnyExpr *expressionAnyExpr
}

// subqueryAnyExpr represents the resolved comparison functions for a plan.Subquery.
type subqueryAnyExpr struct {
	rightSub      *plan.Subquery
	staticLiteral *Literal
	arrayLiterals []*Literal
	compFuncs     []*framework.CompiledFunction
}

// expressionAnyExpr represents the resolved comparison function for a sql.Expression.
type expressionAnyExpr struct {
	rightExpr     sql.Expression
	staticLiteral *Literal
	arrayLiteral  *Literal
	compFunc      *framework.CompiledFunction
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
	if a.leftExpr == nil || !a.leftExpr.Resolved() || a.rightExpr == nil || !a.rightExpr.Resolved() {
		return false
	}
	if a.subqueryAnyExpr != nil {
		return a.subqueryAnyExpr.resolved()
	}
	if a.expressionAnyExpr != nil {
		return a.expressionAnyExpr.resolved()
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

// resolved checks if the comparison functions for subqueryAnyExpr is resolved.
func (a *subqueryAnyExpr) resolved() bool {
	if len(a.compFuncs) == 0 {
		return false
	}
	for _, compFunc := range a.compFuncs {
		if !compFunc.Resolved() {
			return false
		}
	}
	return true
}

// eval evaluates the comparison functions for subqueryAnyExpr.
func (a *subqueryAnyExpr) eval(ctx *sql.Context, subOperator string, row sql.Row, left interface{}) (interface{}, error) {
	if len(a.compFuncs) == 0 {
		return nil, fmt.Errorf("%T: cannot Eval as it has not been fully resolved", a)
	}

	// TODO: This sometimes panics in `evalMultiple` for subqueries that return
	// more than one row, when len(row) > len(iter.Next())
	rightValues, err := a.rightSub.EvalMultiple(ctx, row)
	if err != nil {
		return nil, err
	}

	if len(rightValues) == 0 {
		return nil, nil
	}

	// TODO: This is a workaround some subqueries where the schema length does not
	// match the row length
	if len(a.arrayLiterals) == 1 && len(rightValues) != 1 {
		op, err := framework.GetOperatorFromString(subOperator)
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

// resolved checks if the comparison function for expressionAnyExpr is resolved.
func (a *expressionAnyExpr) resolved() bool {
	if a.compFunc == nil || !a.compFunc.Resolved() {
		return false
	}
	return true
}

// eval evaluates the comparison function for expressionAnyExpr.
func (a *expressionAnyExpr) eval(ctx *sql.Context, row sql.Row, left interface{}) (interface{}, error) {
	if a.compFunc == nil {
		return nil, fmt.Errorf("%T: cannot Eval as it has not been fully resolved", a)
	}

	rightInterface, err := a.rightExpr.Eval(ctx, row)
	if err != nil {
		return nil, err
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

	// Next we'll assign our evaluated values to the expressions that the comparison function reference
	a.staticLiteral.value = left
	for _, rightValue := range rightValues {
		a.arrayLiteral.value = rightValue
		result, err := a.compFunc.Eval(ctx, row)
		if err != nil {
			return nil, err
		}
		if result.(bool) {
			return true, nil
		}
	}

	return false, nil
}

// Eval implements the Expression interface.
func (a *AnyExpr) Eval(ctx *sql.Context, row sql.Row) (interface{}, error) {
	left, err := a.leftExpr.Eval(ctx, row)
	if err != nil {
		return nil, err
	}

	if a.subqueryAnyExpr != nil {
		return a.subqueryAnyExpr.eval(ctx, a.subOperator, row, left)
	}

	if a.expressionAnyExpr != nil {
		return a.expressionAnyExpr.eval(ctx, row, left)
	}

	return nil, fmt.Errorf("%T: cannot Eval as it has not been fully resolved", a)
}

// WithChildren implements the Expression interface.
func (a *AnyExpr) WithChildren(children ...sql.Expression) (sql.Expression, error) {
	if len(children) != 2 {
		return nil, sql.ErrInvalidChildrenNumber.New(a, len(children), 2)
	}

	anyExpr := &AnyExpr{
		leftExpr:    children[0],
		rightExpr:   children[1],
		subOperator: a.subOperator,
		name:        a.name,
	}

	if _, ok := children[1].(*plan.Subquery); ok {
		// TODO: Fix subqueries and return anySubqueryWithChildren(anyExpr, sub)
		return nil, fmt.Errorf("%s does not support subqueries yet", a.name)
	}

	return anyExpressionWithChildren(anyExpr)
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

// AnySubqueryWithChildren resolves the comparison functions for a plan.Subquery.
func AnySubqueryWithChildren(anyExpr *AnyExpr, sub *plan.Subquery) (sql.Expression, error) {
	schema := sub.Query.Schema()
	subTypes := make([]pgtypes.DoltgresType, len(schema))
	for i, col := range schema {
		dgType, ok := col.Type.(pgtypes.DoltgresType)
		if !ok {
			return nil, fmt.Errorf("expected right child to be a DoltgresType but got `%T`", sub)
		}
		subTypes[i] = dgType
	}

	op, err := framework.GetOperatorFromString(anyExpr.subOperator)
	if err != nil {
		return nil, err
	}

	if leftType, ok := anyExpr.leftExpr.Type().(pgtypes.DoltgresType); ok {
		// Resolve comparison functions once and reuse the functions in Eval.
		staticLiteral := &Literal{typ: leftType}
		arrayLiterals := make([]*Literal, len(subTypes))
		// Each expression may be a different type (which is valid), so we need a comparison function for each expression.
		compFuncs := make([]*framework.CompiledFunction, len(subTypes))
		for i, rightType := range subTypes {
			arrayLiterals[i] = &Literal{typ: rightType}
			compFuncs[i] = framework.GetBinaryFunction(op).Compile("internal_any_comparison", staticLiteral, arrayLiterals[i])
			if compFuncs[i] == nil {
				return nil, fmt.Errorf("operator does not exist: %s = %s", leftType.String(), rightType.String())
			}
			if compFuncs[i].Type().(pgtypes.DoltgresType).BaseID() != pgtypes.DoltgresTypeBaseID_Bool {
				// This should never happen, but this is just to be safe
				return nil, fmt.Errorf("%T: found equality comparison that does not return a bool", anyExpr)
			}
		}

		anyExpr.subqueryAnyExpr = &subqueryAnyExpr{
			rightSub:      sub,
			staticLiteral: staticLiteral,
			arrayLiterals: arrayLiterals,
			compFuncs:     compFuncs,
		}
	}

	return anyExpr, nil
}

// anyExpressionWithChildren resolves the comparison functions for a sql.Expression.
func anyExpressionWithChildren(anyExpr *AnyExpr) (sql.Expression, error) {
	arrType, ok := anyExpr.rightExpr.Type().(pgtypes.DoltgresArrayType)
	if !ok {
		return nil, fmt.Errorf("expected right child to be a DoltgresType but got `%T`", anyExpr.rightExpr)
	}
	rightType := arrType.BaseType()

	op, err := framework.GetOperatorFromString(anyExpr.subOperator)
	if err != nil {
		return nil, err
	}

	if leftType, ok := anyExpr.leftExpr.Type().(pgtypes.DoltgresType); ok {
		// Resolve comparison function once and reuse the function in Eval.
		staticLiteral := &Literal{typ: leftType}
		arrayLiteral := &Literal{typ: rightType}
		compFunc := framework.GetBinaryFunction(op).Compile("internal_any_comparison", staticLiteral, arrayLiteral)
		if compFunc == nil {
			return nil, fmt.Errorf("operator does not exist: %s = %s", leftType.String(), rightType.String())
		}
		if compFunc.Type().(pgtypes.DoltgresType).BaseID() != pgtypes.DoltgresTypeBaseID_Bool {
			// This should never happen, but this is just to be safe
			return nil, fmt.Errorf("%T: found equality comparison that does not return a bool", anyExpr)
		}
		anyExpr.expressionAnyExpr = &expressionAnyExpr{
			rightExpr:     anyExpr.rightExpr,
			staticLiteral: staticLiteral,
			arrayLiteral:  arrayLiteral,
			compFunc:      compFunc,
		}
	}

	return anyExpr, nil
}
