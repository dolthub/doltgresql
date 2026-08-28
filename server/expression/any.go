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
	"context"
	"fmt"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/expression"
	"github.com/dolthub/go-mysql-server/sql/plan"

	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/core/casts"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// AnyExpr represents the ANY/SOME/ALL expression. ANY and SOME are synonyms (true if at least one comparison is
// true); ALL additionally requires every comparison to be true, which the eval methods handle via an isAll flag
// derived from name rather than as a separate type, since the two share all other resolution/traversal logic.
type AnyExpr struct {
	leftExpr    sql.Expression
	rightExpr   sql.Expression
	subOperator string
	name        string // ANY, SOME, or ALL

	subqueryAnyExpr   *subqueryAnyExpr
	expressionAnyExpr *expressionAnyExpr
}

// subqueryAnyExpr represents the resolved comparison functions for a plan.Subquery.
type subqueryAnyExpr struct {
	rightSub      *plan.Subquery
	staticLiteral *expression.Literal
	arrayLiterals []*expression.Literal
	compFuncs     []framework.Function
}

// expressionAnyExpr represents the resolved comparison function for a sql.Expression.
type expressionAnyExpr struct {
	rightExpr     sql.Expression
	arrCast       casts.Cast
	arrType       *pgtypes.DoltgresType
	staticLiteral *expression.Literal
	arrayLiteral  *expression.Literal
	compFunc      framework.Function
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
func (a *AnyExpr) IsNullable(ctx *sql.Context) bool {
	return a.leftExpr.IsNullable(ctx) || a.rightExpr.IsNullable(ctx)
}

// Type implements the Expression interface.
func (a *AnyExpr) Type(ctx *sql.Context) sql.Type {
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

// eval evaluates the comparison functions for subqueryAnyExpr. isAll distinguishes ALL's all-must-match semantics
// (vacuously true against zero rows; NULL if nothing failed but something was unknown) from ANY/SOME's
// any-may-match semantics (short-circuits true on the first match).
func (a *subqueryAnyExpr) eval(ctx *sql.Context, subOperator string, row sql.Row, left interface{}, isAll bool) (interface{}, error) {
	if len(a.compFuncs) == 0 {
		return nil, errors.Errorf("%T: cannot Eval as it has not been fully resolved", a)
	}

	// TODO: This sometimes panics in `evalMultiple` for subqueries that return
	// more than one row, when len(row) > len(iter.Next())
	rightValues, err := a.rightSub.EvalMultiple(ctx, row)
	if err != nil {
		return nil, err
	}

	if len(rightValues) == 0 {
		// ALL vacuously holds over an empty set; ANY/SOME cannot match anything.
		return isAll, nil
	}

	// TODO: This is a workaround some subqueries where the schema length does not
	// match the row length
	if len(a.arrayLiterals) == 1 && len(rightValues) != 1 {
		op, err := framework.GetOperatorFromString(subOperator)
		if err != nil {
			return nil, err
		}

		for i := len(a.arrayLiterals); i < len(rightValues); i++ {
			arrayLiteral := expression.NewLiteral(nil, a.arrayLiterals[0].Type(ctx))
			a.arrayLiterals = append(a.arrayLiterals, arrayLiteral)
			compFunc := framework.GetBinaryFunction(op).Compile(ctx, "internal_any_comparison", a.staticLiteral, a.arrayLiterals[i])
			a.compFuncs = append(a.compFuncs, compFunc)
		}
	}

	if len(a.arrayLiterals) != len(rightValues) {
		return nil, errors.Errorf("%T: expected right child to return `%d` values but returned `%d`", a, len(a.arrayLiterals), len(rightValues))
	}

	// Next we'll assign our evaluated values to the expressions that the comparison functions reference
	// Note that the compiled function has a reference to the staticLiteral and arrayLiterals, so we must alter them in place
	a.staticLiteral.Val = left
	for i, rightValue := range rightValues {
		a.arrayLiterals[i].Val = rightValue
	}

	if isAll {
		foundNull := false
		for _, compFunc := range a.compFuncs {
			result, err := compFunc.Eval(ctx, row)
			if err != nil {
				return nil, err
			}
			if result == nil {
				foundNull = true
				continue
			}
			if !result.(bool) {
				return false, nil
			}
		}
		if foundNull {
			return nil, nil
		}
		return true, nil
	}

	// Now we can loop over all comparison functions, as they'll reference their respective values
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

// eval evaluates the comparison function for expressionAnyExpr. See subqueryAnyExpr.eval for the meaning of isAll.
func (a *expressionAnyExpr) eval(ctx *sql.Context, row sql.Row, left interface{}, isAll bool) (interface{}, error) {
	if a.compFunc == nil {
		return nil, errors.Errorf("%T: cannot Eval as it has not been fully resolved", a)
	}

	rightInterface, err := a.rightExpr.Eval(ctx, row)
	if err != nil {
		return nil, err
	}

	if rightInterface == nil {
		return nil, nil
	}

	if rightType, ok := a.rightExpr.Type(ctx).(*pgtypes.DoltgresType); ok && rightType.ID == pgtypes.Unknown.ID && a.arrCast.ID.IsValid() {
		rightInterface, err = a.arrCast.Eval(ctx, rightInterface, rightType, a.arrType)
		if err != nil {
			return nil, err
		}
	}

	rightValues, ok := rightInterface.([]any)
	if !ok {
		return nil, errors.Errorf("%T: expected right child to return `%T` but returned `%T`", a, []any{}, rightInterface)
	}
	if len(rightValues) == 0 {
		// ALL vacuously holds over an empty array; ANY/SOME cannot match anything.
		return isAll, nil
	}

	// Next we'll assign our evaluated values to the expressions that the comparison function reference
	// Note that the compiled function has a reference to the staticLiteral and arrayLiteral, so we must alter them in place
	a.staticLiteral.Val = left

	if isAll {
		foundNull := false
		for _, rightValue := range rightValues {
			a.arrayLiteral.Val = rightValue
			result, err := a.compFunc.Eval(ctx, row)
			if err != nil {
				return nil, err
			}
			if result == nil {
				foundNull = true
				continue
			}
			if !result.(bool) {
				return false, nil
			}
		}
		if foundNull {
			return nil, nil
		}
		return true, nil
	}

	for _, rightValue := range rightValues {
		a.arrayLiteral.Val = rightValue
		result, err := a.compFunc.Eval(ctx, row)
		if err != nil {
			return nil, err
		}
		if result == nil {
			return nil, nil
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

	isAll := a.name == "ALL"
	if a.subqueryAnyExpr != nil {
		return a.subqueryAnyExpr.eval(ctx, a.subOperator, row, left, isAll)
	}

	if a.expressionAnyExpr != nil {
		return a.expressionAnyExpr.eval(ctx, row, left, isAll)
	}

	return nil, errors.Errorf("%T: cannot Eval as it has not been fully resolved", a)
}

// WithChildren implements the Expression interface.
func (a *AnyExpr) WithChildren(ctx *sql.Context, children ...sql.Expression) (sql.Expression, error) {
	if len(children) != 2 {
		return nil, sql.ErrInvalidChildrenNumber.New(a, len(children), 2)
	}

	leftExpr := children[0]
	rightExpr := children[1]
	// Unmodified BindVars use deferred type resolution, so we replace the deference with the left's type in array form
	if bv, ok := rightExpr.(*expression.BindVar); ok {
		if _, ok = bv.Typ.(*pgtypes.DoltgresType); !ok {
			if leftType, ok := leftExpr.Type(ctx).(*pgtypes.DoltgresType); ok {
				bv.Typ = leftType.ToArrayType()
			}
		}
	} else if bv, ok := leftExpr.(*expression.BindVar); ok {
		// Similarly, an untyped bind variable on the left (e.g. `$1 = ANY(arr_col)` or `$1 = ANY(SELECT ...)`) is
		// resolved from the right side's element type: the array's base type, or the subquery's first column type.
		if _, ok = bv.Typ.(*pgtypes.DoltgresType); !ok {
			if sub, ok := rightExpr.(*plan.Subquery); ok {
				if schema := sub.Query.Schema(ctx); len(schema) > 0 {
					if colType, ok := schema[0].Type.(*pgtypes.DoltgresType); ok {
						bv.Typ = colType
					}
				}
			} else if rightType, ok := rightExpr.Type(ctx).(*pgtypes.DoltgresType); ok && rightType.IsArrayType() {
				bv.Typ = rightType.BaseType()
			}
		}
	}
	anyExpr := &AnyExpr{
		leftExpr:    leftExpr,
		rightExpr:   rightExpr,
		subOperator: a.subOperator,
		name:        a.name,
	}

	if sub, ok := children[1].(*plan.Subquery); ok {
		return anySubqueryWithChildren(ctx, anyExpr, sub)
	}

	return anyExpressionWithChildren(ctx, anyExpr)
}

// WithResolvedChildren implements the Expression interface.
func (a *AnyExpr) WithResolvedChildren(ctx context.Context, children []any) (any, error) {
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
	return a.WithChildren(ctx.(*sql.Context), left, right)
}

// String implements the fmt.Stringer interface.
func (a *AnyExpr) String() string {
	if a.leftExpr == nil || a.rightExpr == nil {
		return fmt.Sprintf("? %s (?)", a.name)
	}
	return fmt.Sprintf("%s = %s (%s)", a.leftExpr, a.name, a.rightExpr)
}

// DebugString implements the Expression interface.
func (a *AnyExpr) DebugString(ctx *sql.Context) string {
	return fmt.Sprintf("%s %s (%s)", sql.DebugString(ctx, a.leftExpr), a.name, sql.DebugString(ctx, a.rightExpr))
}

// anySubqueryWithChildren resolves the comparison functions for a plan.Subquery.
func anySubqueryWithChildren(ctx *sql.Context, anyExpr *AnyExpr, sub *plan.Subquery) (sql.Expression, error) {
	schema := sub.Query.Schema(ctx)
	subTypes := make([]*pgtypes.DoltgresType, len(schema))
	for i, col := range schema {
		dgType, ok := col.Type.(*pgtypes.DoltgresType)
		if !ok {
			return nil, errors.Errorf("expected right child to be a DoltgresType but got `%T`", sub)
		}
		subTypes[i] = dgType
	}

	op, err := framework.GetOperatorFromString(anyExpr.subOperator)
	if err != nil {
		return nil, err
	}

	if leftType, ok := anyExpr.leftExpr.Type(ctx).(*pgtypes.DoltgresType); ok {
		// Resolve comparison functions once and reuse the functions in Eval.
		staticLiteral := expression.NewLiteral(nil, leftType)
		arrayLiterals := make([]*expression.Literal, len(subTypes))
		// Each expression may be a different type (which is valid), so we need a comparison function for each expression.
		compFuncs := make([]framework.Function, len(subTypes))
		for i, rightType := range subTypes {
			arrayLiterals[i] = expression.NewLiteral(nil, rightType)
			compFuncs[i] = framework.GetBinaryFunction(op).Compile(ctx, "internal_any_comparison", staticLiteral, arrayLiterals[i])
			if compFuncs[i] == nil {
				return nil, errors.Errorf("operator does not exist: %s = %s", leftType.String(), rightType.String())
			}
			if compFuncs[i].Type(ctx).(*pgtypes.DoltgresType).ID != pgtypes.Bool.ID {
				// This should never happen, but this is just to be safe
				return nil, errors.Errorf("%T: found equality comparison that does not return a bool", anyExpr)
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
func anyExpressionWithChildren(ctx *sql.Context, anyExpr *AnyExpr) (sql.Expression, error) {
	arrType, ok := anyExpr.rightExpr.Type(ctx).(*pgtypes.DoltgresType)
	if !ok {
		return nil, errors.Errorf("expected right child to be a DoltgresType but got `%T`", anyExpr.rightExpr)
	}
	leftType, ok := anyExpr.leftExpr.Type(ctx).(*pgtypes.DoltgresType)
	if !ok {
		return anyExpr, nil
	}

	var arrCast casts.Cast
	if arrType.ID == pgtypes.Unknown.ID {
		castsColl, err := core.GetCastsCollectionFromContext(ctx, "")
		if err != nil {
			return nil, err
		}
		// if array type is Unknown, use the left expr type for reference
		var arrCastToType *pgtypes.DoltgresType
		if leftType.ID == pgtypes.Unknown.ID {
			// if left expr type is also unknown, it's likely text array - TODO double check
			arrCastToType = pgtypes.TextArray
		} else {
			arrCastToType = leftType.ToArrayType()
		}
		arrCast, err = castsColl.GetImplicitCast(ctx, arrType, arrCastToType)
		if err != nil {
			return nil, err
		}
		arrType = arrCastToType
	}

	rightType := arrType.BaseType()
	op, err := framework.GetOperatorFromString(anyExpr.subOperator)
	if err != nil {
		return nil, err
	}

	// Resolve comparison function once and reuse the function in Eval.
	staticLiteral := expression.NewLiteral(nil, leftType)
	arrayLiteral := expression.NewLiteral(nil, rightType)
	compFunc := framework.GetBinaryFunction(op).Compile(ctx, "internal_any_comparison", staticLiteral, arrayLiteral)
	if compFunc == nil || compFunc.StashedError() != nil {
		return nil, errors.Errorf("operator does not exist: %s = %s", leftType.String(), rightType.String())
	}
	compFuncType := compFunc.Type(ctx)
	if compFuncType.(*pgtypes.DoltgresType).ID != pgtypes.Bool.ID {
		// This should never happen, but this is just to be safe
		return nil, errors.Errorf("%T: found equality comparison that does not return a bool", anyExpr)
	}
	anyExpr.expressionAnyExpr = &expressionAnyExpr{
		rightExpr:     anyExpr.rightExpr,
		arrCast:       arrCast,
		arrType:       arrType,
		staticLiteral: staticLiteral,
		arrayLiteral:  arrayLiteral,
		compFunc:      compFunc,
	}

	return anyExpr, nil
}
