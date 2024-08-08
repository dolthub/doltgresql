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
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/server/functions/framework"
)

// UnaryOperator represents a VALUE OPERATOR VALUE expression.
type UnaryOperator struct {
	operator     framework.Operator
	compiledFunc *framework.CompiledFunction
}

var _ vitess.Injectable = (*UnaryOperator)(nil)
var _ sql.Expression = (*UnaryOperator)(nil)

// NewUnaryOperator returns a new *UnaryOperator.
func NewUnaryOperator(operator framework.Operator) *UnaryOperator {
	return &UnaryOperator{operator: operator}
}

// Children implements the sql.Expression interface.
func (b *UnaryOperator) Children() []sql.Expression {
	return b.compiledFunc.Children()
}

// Eval implements the sql.Expression interface.
func (b *UnaryOperator) Eval(ctx *sql.Context, row sql.Row) (any, error) {
	return b.compiledFunc.Eval(ctx, row)
}

// IsNullable implements the sql.Expression interface.
func (b *UnaryOperator) IsNullable() bool {
	return b.compiledFunc.IsNullable()
}

// Resolved implements the sql.Expression interface.
func (b *UnaryOperator) Resolved() bool {
	return b.compiledFunc.Resolved()
}

// String implements the sql.Expression interface.
func (b *UnaryOperator) String() string {
	if b.compiledFunc == nil {
		return fmt.Sprintf("%s?", b.operator.String())
	}
	// We know that we'll always have one parameter here
	return fmt.Sprintf("%s%s", b.operator.String(), b.compiledFunc.Arguments[0].String())
}

// Type implements the sql.Expression interface.
func (b *UnaryOperator) Type() sql.Type {
	return b.compiledFunc.Type()
}

// WithChildren implements the sql.Expression interface.
func (b *UnaryOperator) WithChildren(children ...sql.Expression) (sql.Expression, error) {
	if len(children) != 1 {
		return nil, sql.ErrInvalidChildrenNumber.New(b, len(children), 1)
	}
	compiledFunc, err := b.compiledFunc.WithChildren(children...)
	if err != nil {
		return nil, err
	}
	return &UnaryOperator{
		operator:     b.operator,
		compiledFunc: compiledFunc.(*framework.CompiledFunction),
	}, nil
}

// WithResolvedChildren implements the vitess.InjectableExpression interface.
func (b *UnaryOperator) WithResolvedChildren(children []any) (any, error) {
	if len(children) != 1 {
		return nil, fmt.Errorf("invalid vitess child count, expected `1` but got `%d`", len(children))
	}
	child, ok := children[0].(sql.Expression)
	if !ok {
		return nil, fmt.Errorf("expected vitess child to be an expression but has type `%T`", children[0])
	}
	funcName := "internal_unary_operator_func_" + b.operator.String()
	compiledFunc := framework.GetUnaryFunction(b.operator).Compile(funcName, child)
	if compiledFunc == nil {
		return nil, fmt.Errorf("operator does not exist: %s%s", b.operator.String(), child.Type().String())
	}
	return &UnaryOperator{
		operator:     b.operator,
		compiledFunc: compiledFunc,
	}, nil
}
