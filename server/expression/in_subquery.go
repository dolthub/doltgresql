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
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// InSubquery represents a VALUE IN (SELECT ...) expression.
type InSubquery struct {
	leftExpr  sql.Expression
	rightExpr *plan.Subquery

	// These variables are used so that we can resolve the comparison functions once and reuse them as we iterate over rows.
	// These are assigned in WithChildren, so refer there for more information.
	staticLiteral *Literal
	arrayLiterals []*Literal
	compFuncs     []*framework.CompiledFunction
}

var _ vitess.Injectable = (*BinaryOperator)(nil)
var _ sql.Expression = (*BinaryOperator)(nil)
var _ expression.BinaryExpression = (*BinaryOperator)(nil)

// NewInSubquery returns a new *InSubquery.
func NewInSubquery() *InSubquery {
	return &InSubquery{}
}

// Children implements the sql.Expression interface.
func (it *InSubquery) Children() []sql.Expression {
	return []sql.Expression{it.leftExpr, it.rightExpr}
}

// Eval implements the sql.Expression interface.
func (it *InSubquery) Eval(ctx *sql.Context, row sql.Row) (any, error) {
	if len(it.compFuncs) == 0 {
		return nil, fmt.Errorf("%T: cannot Eval as it has not been fully resolved", it)
	}

	left, err := it.leftExpr.Eval(ctx, row)
	if err != nil {
		return nil, err
	}

	rightInterface, err := it.rightExpr.Eval(ctx, row)
	if err != nil {
		return nil, err
	}
	
	rightValues, ok := rightInterface.([]any)
	if !ok {
		// Tuples will return the value directly if it has a length of one, so we'll check for that first
		if len(it.rightExpr) == 1 {
			rightValues = []any{rightInterface}
		} else {
			return nil, fmt.Errorf("%T: expected right child to return `%T` but returned `%T`", it, []any{}, rightInterface)
		}
	}
	// Next we'll assign our evaluated values to the expressions that the comparison functions reference
	it.staticLiteral.value = left
	for i, rightValue := range rightValues {
		it.arrayLiterals[i].value = rightValue
	}
	// Now we can loop over all of the comparison functions, as they'll reference their respective values
	for _, compFunc := range it.compFuncs {
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

// IsNullable implements the sql.Expression interface.
func (it *InSubquery) IsNullable() bool {
	return true
}

// Resolved implements the sql.Expression interface.
func (it *InSubquery) Resolved() bool {
	if it.leftExpr == nil || !it.leftExpr.Resolved() || it.rightExpr == nil || !it.rightExpr.Resolved() || len(it.compFuncs) == 0 {
		return false
	}
	for _, compFunc := range it.compFuncs {
		if !compFunc.Resolved() {
			return false
		}
	}
	return true
}

// String implements the sql.Expression interface.
func (it *InSubquery) String() string {
	if it.leftExpr == nil || it.rightExpr == nil {
		return "? IN ?"
	}
	return fmt.Sprintf("%s IN %s", it.leftExpr.String(), it.rightExpr.String())
}

// Type implements the sql.Expression interface.
func (it *InSubquery) Type() sql.Type {
	return pgtypes.Bool
}

// WithChildren implements the sql.Expression interface.
func (it *InSubquery) WithChildren(children ...sql.Expression) (sql.Expression, error) {
	if len(children) != 2 {
		return nil, sql.ErrInvalidChildrenNumber.New(it, len(children), 2)
	}
	rightTuple, ok := children[1].(expression.Tuple)
	if !ok {
		return nil, fmt.Errorf("%T: expected right child to be `%T` but has type `%T`", it, expression.Tuple{}, children[1])
	}
	if len(rightTuple) == 0 {
		return nil, fmt.Errorf("IN must contain at least 1 expression")
	}
	// We'll only resolve the comparison functions once we have all Doltgres types.
	// We may see GMS types during some analyzer steps, so we should wait until those are done.
	if leftType, ok := children[0].Type().(pgtypes.DoltgresType); ok {
		// Rather than finding and resolving a comparison function every time we call Eval, we resolve them once and
		// reuse the functions. We also want to avoid re-assigning the parameters of the comparison functions since that
		// will also cause the functions to resolve again. To do this, we store expressions within our struct that the
		// functions reference, so we can freely switch the values within the literals without changing anything
		// regarding the comparison functions. This is usually unsafe, but since we're verifying the types returned by
		// the parameters, and assigning the values to our own literals, we do not have to worry. This offers a
		// significant speedup as function resolution is very expensive, so we want to do it as few times as possible
		// (preferably once).
		staticLiteral := &Literal{typ: leftType}
		arrayLiterals := make([]*Literal, len(rightTuple))
		// Each expression may be a different type (which is valid), so we need a comparison function for each expression.
		compFuncs := make([]*framework.CompiledFunction, len(rightTuple))
		allValidChildren := true
		for i, rightExpr := range rightTuple {
			rightType, ok := rightExpr.Type().(pgtypes.DoltgresType)
			if !ok {
				allValidChildren = false
				break
			}
			arrayLiterals[i] = &Literal{typ: rightType}
			compFuncs[i] = framework.GetBinaryFunction(framework.Operator_BinaryEqual).Compile("internal_in_comparison", staticLiteral, arrayLiterals[i])
			if compFuncs[i] == nil {
				return nil, fmt.Errorf("operator does not exist: %s = %s", leftType.String(), rightType.String())
			}
			if compFuncs[i].Type().(pgtypes.DoltgresType).BaseID() != pgtypes.DoltgresTypeBaseID_Bool {
				// This should never happen, but this is just to be safe
				return nil, fmt.Errorf("%T: found equality comparison that does not return a bool", it)
			}
		}
		if allValidChildren {
			return &InTuple{
				leftExpr:      children[0],
				rightExpr:     rightTuple,
				staticLiteral: staticLiteral,
				arrayLiterals: arrayLiterals,
				compFuncs:     compFuncs,
			}, nil
		}
	}
	return &InTuple{
		leftExpr:  children[0],
		rightExpr: rightTuple,
	}, nil
}

// WithResolvedChildren implements the vitess.InjectableExpression interface.
func (it *InSubquery) WithResolvedChildren(children []any) (any, error) {
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
func (it *InSubquery) Left() sql.Expression {
	return it.leftExpr
}

// Right implements the expression.BinaryExpression interface.
func (it *InSubquery) Right() sql.Expression {
	return it.rightExpr
}
