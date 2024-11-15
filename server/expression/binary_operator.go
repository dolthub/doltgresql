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
)

// BinaryOperator represents a VALUE OPERATOR VALUE expression.
type BinaryOperator struct {
	operator     framework.Operator
	compiledFunc *framework.CompiledFunction
}

var _ vitess.Injectable = (*BinaryOperator)(nil)
var _ sql.Expression = (*BinaryOperator)(nil)
var _ expression.BinaryExpression = (*BinaryOperator)(nil)
var _ expression.Equality = (*BinaryOperator)(nil)

// NewBinaryOperator returns a new *BinaryOperator.
func NewBinaryOperator(operator framework.Operator) *BinaryOperator {
	return &BinaryOperator{operator: operator}
}

// Children implements the sql.Expression interface.
func (b *BinaryOperator) Children() []sql.Expression {
	return b.compiledFunc.Children()
}

// Eval implements the sql.Expression interface.
func (b *BinaryOperator) Eval(ctx *sql.Context, row sql.Row) (any, error) {
	return b.compiledFunc.Eval(ctx, row)
}

// IsNullable implements the sql.Expression interface.
func (b *BinaryOperator) IsNullable() bool {
	return b.compiledFunc.IsNullable()
}

// RepresentsEquality implements the expression.Equality interface.
func (b *BinaryOperator) RepresentsEquality() bool {
	return b.operator == framework.Operator_BinaryEqual
}

// Resolved implements the sql.Expression interface.
func (b *BinaryOperator) Resolved() bool {
	return b.compiledFunc.Resolved()
}

// String implements the sql.Expression interface.
func (b *BinaryOperator) String() string {
	if b.compiledFunc == nil {
		return fmt.Sprintf("? %s ?", b.operator.String())
	}
	// We know that we'll always have two parameters here
	return fmt.Sprintf("%s %s %s",
		b.compiledFunc.Arguments[0].String(), b.operator.String(), b.compiledFunc.Arguments[1].String())
}

// SwapParameters implements the expression.Equality interface.
func (b *BinaryOperator) SwapParameters(ctx *sql.Context) (expression.Equality, error) {
	// TODO: for now we'll assume this is valid, but we should check for the `COMMUTATOR` property on the operator
	f, err := b.WithResolvedChildren([]any{b.Right(), b.Left()})
	if err != nil {
		return nil, err
	}
	return f.(expression.Equality), nil
}

// ToComparer implements the expression.Equality interface.
func (b *BinaryOperator) ToComparer() (expression.Comparer, error) {
	return NewJoinComparator(b)
}

// Type implements the sql.Expression interface.
func (b *BinaryOperator) Type() sql.Type {
	return b.compiledFunc.Type()
}

// WithChildren implements the sql.Expression interface.
func (b *BinaryOperator) WithChildren(children ...sql.Expression) (sql.Expression, error) {
	if len(children) != 2 {
		return nil, sql.ErrInvalidChildrenNumber.New(b, len(children), 2)
	}
	compiledFunc, err := b.compiledFunc.WithChildren(children...)
	if err != nil {
		return nil, err
	}
	return &BinaryOperator{
		operator:     b.operator,
		compiledFunc: compiledFunc.(*framework.CompiledFunction),
	}, nil
}

// WithResolvedChildren implements the vitess.InjectableExpression interface.
func (b *BinaryOperator) WithResolvedChildren(children []any) (any, error) {
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
	funcName := "internal_binary_operator_func_" + b.operator.String()
	compiledFunc := framework.GetBinaryFunction(b.operator).Compile(funcName, left, right)
	if err := compiledFunc.StashedError(); err != nil {
		return nil, err
	}
	if compiledFunc == nil {
		return nil, fmt.Errorf("operator does not exist: %s %s %s",
			left.Type().String(), b.operator.String(), right.Type().String())
	}
	return &BinaryOperator{
		operator:     b.operator,
		compiledFunc: compiledFunc,
	}, nil
}

// Operator returns the operator that is used.
func (b *BinaryOperator) Operator() framework.Operator {
	return b.operator
}

// Left implements the expression.BinaryExpression interface.
func (b *BinaryOperator) Left() sql.Expression {
	// We know that we'll always have two parameters here
	return b.compiledFunc.Arguments[0]
}

// Right implements the expression.BinaryExpression interface.
func (b *BinaryOperator) Right() sql.Expression {
	// We know that we'll always have two parameters here
	return b.compiledFunc.Arguments[1]
}
