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

	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/expression"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// InTuple represents a VALUE IN (<VALUES>) expression.
type InTuple struct {
	leftExpr  sql.Expression
	rightExpr expression.Tuple

	// These variables are used so that we can resolve the comparison functions once and reuse them as we iterate over rows.
	// These are assigned in WithChildren, so refer there for more information.
	staticLiteral *expression.Literal
	arrayLiterals []*expression.Literal
	compFuncs     []framework.Function
}

var _ vitess.Injectable = (*BinaryOperator)(nil)
var _ sql.Expression = (*BinaryOperator)(nil)
var _ expression.BinaryExpression = (*BinaryOperator)(nil)

// NewInTuple returns a new *InTuple.
func NewInTuple() *InTuple {
	return &InTuple{
		leftExpr:  nil,
		rightExpr: nil,
	}
}

// Children implements the sql.Expression interface.
func (it *InTuple) Children() []sql.Expression {
	return []sql.Expression{it.leftExpr, it.rightExpr}
}

// Decay returns the expression as a series of OR expressions. The behavior is not the same, however it allows some
// paths to simplify their expression handling (such as filters).
func (it *InTuple) Decay() sql.Expression {
	switch f := it.compFuncs[0].(type) {
	case *framework.CompiledFunction:
		f.Arguments = []sql.Expression{it.leftExpr, it.rightExpr[0]}
	case *framework.QuickFunction2:
		f.Arguments = [2]sql.Expression{it.leftExpr, it.rightExpr[0]}
	}
	var expr sql.Expression = &BinaryOperator{
		operator:     framework.Operator_BinaryEqual,
		compiledFunc: it.compFuncs[0],
	}
	for i := 1; i < len(it.rightExpr); i++ {
		switch f := it.compFuncs[i].(type) {
		case *framework.CompiledFunction:
			f.Arguments = []sql.Expression{it.leftExpr, it.rightExpr[i]}
		case *framework.QuickFunction2:
			f.Arguments = [2]sql.Expression{it.leftExpr, it.rightExpr[i]}
		}
		expr = expression.NewOr(expr, &BinaryOperator{
			operator:     framework.Operator_BinaryEqual,
			compiledFunc: it.compFuncs[i],
		})
	}
	return expr
}

// Eval implements the sql.Expression interface.
func (it *InTuple) Eval(ctx *sql.Context, row sql.Row) (any, error) {
	if len(it.compFuncs) == 0 {
		return nil, errors.Errorf("%T: cannot Eval as it has not been fully resolved", it)
	}
	// First we'll evaluate everything before we do the comparisons
	left, err := it.leftExpr.Eval(ctx, row)
	if err != nil {
		return nil, err
	}

	if left == nil {
		return nil, nil
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
			return nil, errors.Errorf("%T: expected right child to return `%T` but returned `%T`", it, []any{}, rightInterface)
		}
	}
	// Next we'll assign our evaluated values to the expressions that the comparison functions reference
	// Note that the compiled functions already have a reference to this literal, so we have to edit it in place
	it.staticLiteral.Val = left
	for i, rightValue := range rightValues {
		it.arrayLiterals[i].Val = rightValue
	}

	// Now we can loop over all of the comparison functions, as they'll reference their respective values
	// The rules for null comparisons are subtle: an IN expression that includes a NULL in the tuple will return null
	// instead of false if a match is not found, but true otherwise.
	sawNull := false
	for _, compFunc := range it.compFuncs {
		result, err := compFunc.Eval(ctx, row)
		if err != nil {
			return nil, err
		}

		if result == nil {
			sawNull = true
		} else if result.(bool) {
			return true, nil
		}
	}

	if sawNull {
		return nil, nil
	}

	return false, nil
}

// IsNullable implements the sql.Expression interface.
func (it *InTuple) IsNullable() bool {
	return true
}

// Resolved implements the sql.Expression interface.
func (it *InTuple) Resolved() bool {
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
func (it *InTuple) String() string {
	if it.leftExpr == nil || it.rightExpr == nil {
		return "? IN ?"
	}
	return fmt.Sprintf("%s IN %s", it.leftExpr.String(), it.rightExpr.String())
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
		return nil, errors.Errorf("%T: expected right child to be `%T` but has type `%T`", it, expression.Tuple{}, children[1])
	}
	if len(rightTuple) == 0 {
		return nil, errors.Errorf("IN must contain at least 1 expression")
	}
	// We'll only resolve the comparison functions once we have all Doltgres types.
	// We may see GMS types during some analyzer steps, so we should wait until those are done.
	if leftType, ok := children[0].Type().(*pgtypes.DoltgresType); ok {
		// Rather than finding and resolving a comparison function every time we call Eval, we resolve them once and
		// reuse the functions. We also want to avoid re-assigning the parameters of the comparison functions since that
		// will also cause the functions to resolve again. To do this, we store expressions within our struct that the
		// functions reference, so we can freely switch the values within the literals without changing anything
		// regarding the comparison functions. This is usually unsafe, but since we're verifying the types returned by
		// the parameters, and assigning the values to our own literals, we do not have to worry. This offers a
		// significant speedup as function resolution is very expensive, so we want to do it as few times as possible
		// (preferably once).
		staticLiteral := expression.NewLiteral(nil, leftType)
		arrayLiterals := make([]*expression.Literal, len(rightTuple))
		// Each expression may be a different type (which is valid), so we need a comparison function for each expression.
		compFuncs := make([]framework.Function, len(rightTuple))
		allValidChildren := true
		for i, rightExpr := range rightTuple {
			rightType, ok := rightExpr.Type().(*pgtypes.DoltgresType)
			if !ok {
				allValidChildren = false
				break
			}
			arrayLiterals[i] = expression.NewLiteral(nil, rightType)
			compFuncs[i] = framework.GetBinaryFunction(framework.Operator_BinaryEqual).Compile("internal_in_comparison", staticLiteral, arrayLiterals[i])
			if compFuncs[i] == nil {
				return nil, errors.Errorf("operator does not exist: %s = %s", leftType.String(), rightType.String())
			}
			if compFuncs[i].Type().(*pgtypes.DoltgresType).ID != pgtypes.Bool.ID {
				// This should never happen, but this is just to be safe
				return nil, errors.Errorf("%T: found equality comparison that does not return a bool", it)
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
func (it *InTuple) WithResolvedChildren(children []any) (any, error) {
	if len(children) != 2 {
		return nil, errors.Errorf("invalid vitess child count, expected `2` but got `%d`", len(children))
	}
	left, ok := children[0].(sql.Expression)
	if !ok {
		return nil, errors.Errorf("expected vitess child to be an expression but has type `%T`", children[0])
	}

	switch right := children[1].(type) {
	case expression.Tuple:
		return it.WithChildren(left, right)
	case *RecordExpr:
		// TODO: For now, if we see a RecordExpr come in, we convert it to a vitess Tuple representation, so that
		//       the existing in_tuple code can work with it. Alternatively, we could change in_tuple to always
		//       work directly with a Record expression.
		return it.WithChildren(left, expression.Tuple(right.exprs))
	default:
		return nil, errors.Errorf("expected child to be a RecordExpr or vitess Tuple but has type `%T`", children[1])
	}
}

// Left implements the expression.BinaryExpression interface.
func (it *InTuple) Left() sql.Expression {
	return it.leftExpr
}

// Right implements the expression.BinaryExpression interface.
func (it *InTuple) Right() sql.Expression {
	return it.rightExpr
}
