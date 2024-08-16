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
	"github.com/dolthub/go-mysql-server/sql/types"
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

var nilKey, _ = sql.HashOf(sql.NewRow(nil))

// Children implements the sql.Expression interface.
func (in *InSubquery) Children() []sql.Expression {
	return []sql.Expression{in.leftExpr, in.rightExpr}
}

// Eval implements the sql.Expression interface.
func (in *InSubquery) Eval(ctx *sql.Context, row sql.Row) (any, error) {
	if len(in.compFuncs) == 0 {
		return nil, fmt.Errorf("%T: cannot Eval as it has not been fully resolved", in)
	}

	left, err := in.leftExpr.Eval(ctx, row)
	if err != nil {
		return nil, err
	}

	// The NULL handling for IN expressions is tricky. According to
	// https://dev.mysql.com/doc/refman/8.0/en/comparison-operators.html#operator_in:
	// To comply with the SQL standard, IN() returns NULL not only if the expression on the left hand side is NULL, but
	// also if no match is found in the list and one of the expressions in the list is NULL.
	// However, there's a strange edge case. NULL IN (empty list) return 0, not NULL.
	leftNull := left == nil

	if types.NumColumns(in.Left().Type()) != types.NumColumns(in.Right().Type()) {
		return nil, sql.ErrInvalidOperandColumns.New(types.NumColumns(in.Left().Type()), types.NumColumns(in.Right().Type()))
	}

	typ := in.rightExpr.Type()
	right := in.rightExpr

	// TODO: does this work for pg values?
	values, err := right.HashMultiple(ctx, row)
	if err != nil {
		return nil, err
	}

	// NULL IN (list) returns NULL. NULL IN (empty list) returns 0
	if leftNull {
		if values.Size() == 0 {
			return false, nil
		}
		return nil, nil
	}

	// convert left to right's type
	nLeft, _, err := typ.Convert(left)
	if err != nil {
		return false, nil
	}

	key, err := sql.HashOf(sql.NewRow(nLeft))
	if err != nil {
		return nil, err
	}

	val, notFoundErr := values.Get(key)
	if notFoundErr != nil {
		if _, nilValNotFoundErr := values.Get(nilKey); nilValNotFoundErr == nil {
			return nil, nil
		}
		return false, nil
	}

	// TODO: handle tuples
	return in.valuesEqual(ctx, sql.Row{val})
}

func (in *InSubquery) valuesEqual(ctx *sql.Context, row sql.Row) (interface{}, error) {
	for _, compFunc := range in.compFuncs {
		result, err := compFunc.Eval(ctx, row)
		if err != nil {
			return nil, err
		}
		if !result.(bool) {
			return false, nil
		}
	}
	return true, nil
}

// IsNullable implements the sql.Expression interface.
func (in *InSubquery) IsNullable() bool {
	return true
}

// Resolved implements the sql.Expression interface.
func (in *InSubquery) Resolved() bool {
	if in.leftExpr == nil || !in.leftExpr.Resolved() || in.rightExpr == nil || !in.rightExpr.Resolved() || len(in.compFuncs) == 0 {
		return false
	}
	for _, compFunc := range in.compFuncs {
		if !compFunc.Resolved() {
			return false
		}
	}
	return true
}

// String implements the sql.Expression interface.
func (in *InSubquery) String() string {
	if in.leftExpr == nil || in.rightExpr == nil {
		return "? IN ?"
	}
	return fmt.Sprintf("%s IN %s", in.leftExpr.String(), in.rightExpr.String())
}

// Type implements the sql.Expression interface.
func (in *InSubquery) Type() sql.Type {
	return pgtypes.Bool
}

// WithChildren implements the sql.Expression interface.
func (in *InSubquery) WithChildren(children ...sql.Expression) (sql.Expression, error) {
	if len(children) != 2 {
		return nil, sql.ErrInvalidChildrenNumber.New(in, len(children), 2)
	}
	sq, ok := children[1].(*plan.Subquery)
	if !ok {
		return nil, fmt.Errorf("%T: expected right child to be `%T` but has type `%T`", in, &plan.Subquery{}, children[1])
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

		// We need a comparison function for each type in the query result
		sch := sq.Query.Schema()
		staticLiteral := &Literal{typ: leftType}
		compFuncs := make([]*framework.CompiledFunction, len(sch))
		allValidChildren := true
		for i, rightCol := range sch {
			rightType, ok := rightCol.Type.(pgtypes.DoltgresType)
			if !ok {
				allValidChildren = false
				break
			}
			rightLit := &Literal{typ: rightType}
			compFuncs[i] = framework.GetBinaryFunction(framework.Operator_BinaryEqual).Compile("internal_in_comparison", staticLiteral, rightLit)
			if compFuncs[i] == nil {
				return nil, fmt.Errorf("operator does not exist: %s = %s", leftType.String(), rightType.String())
			}
			if compFuncs[i].Type().(pgtypes.DoltgresType).BaseID() != pgtypes.DoltgresTypeBaseID_Bool {
				// This should never happen, but this is just to be safe
				return nil, fmt.Errorf("%T: found equality comparison that does not return a bool", in)
			}
		}
		if allValidChildren {
			return &InSubquery{
				leftExpr:      children[0],
				rightExpr:     sq,
				staticLiteral: staticLiteral,
				compFuncs:     compFuncs,
			}, nil
		}
	}
	return &InSubquery{
		leftExpr:  children[0],
		rightExpr: sq,
	}, nil
}

// WithResolvedChildren implements the vitess.InjectableExpression interface.
func (in *InSubquery) WithResolvedChildren(children []any) (any, error) {
	if len(children) != 2 {
		return nil, fmt.Errorf("invalid vitess child count, expected `2` but got `%d`", len(children))
	}
	left, ok := children[0].(sql.Expression)
	if !ok {
		return nil, fmt.Errorf("expected vitess child to be an expression but has type `%T`", children[0])
	}
	right, ok := children[1].(*plan.Subquery)
	if !ok {
		return nil, fmt.Errorf("expected vitess child to be a *plan.Subquery but has type `%T`", children[1])
	}
	return in.WithChildren(left, right)
}

// Left implements the expression.BinaryExpression interface.
func (in *InSubquery) Left() sql.Expression {
	return in.leftExpr
}

// Right implements the expression.BinaryExpression interface.
func (in *InSubquery) Right() sql.Expression {
	return in.rightExpr
}
