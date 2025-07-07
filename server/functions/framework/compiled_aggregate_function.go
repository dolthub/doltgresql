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

package framework

import (
	"strings"

	cerrors "github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/expression"
)

// AggregateFunction is an expression that represents CompiledAggregateFunction
type AggregateFunction interface {
	sql.FunctionExpression
	sql.Aggregation
	specificFuncImpl()
}

type NewBufferFn func([]sql.Expression) (sql.AggregationBuffer, error)

// CompiledAggregateFunction is an expression that represents a fully-analyzed PostgreSQL aggregate function.
type CompiledAggregateFunction struct {
	*CompiledFunction
	aggId     sql.ColumnId
	newBuffer NewBufferFn
}

var _ AggregateFunction = (*CompiledAggregateFunction)(nil)

// NewCompiledAggregateFunction returns a newly compiled function.
// TODO: newBuffer probably needs to be parameterized in the overloads
func NewCompiledAggregateFunction(name string, args []sql.Expression, functions *Overloads, newBuffer NewBufferFn) *CompiledAggregateFunction {
	return newCompiledAggregateFunctionInternal(name, args, functions, functions.overloadsForParams(len(args)), newBuffer)
}

// newCompiledAggregateFunctionInternal is called internally, which skips steps that may have already been processed.
func newCompiledAggregateFunctionInternal(name string, args []sql.Expression, overloads *Overloads, fnOverloads []Overload, newBuffer NewBufferFn) *CompiledAggregateFunction {
	cf := newCompiledFunctionInternal(name, args, overloads, fnOverloads, false, nil)
	c := &CompiledAggregateFunction{
		CompiledFunction: cf,
		newBuffer:        newBuffer,
	}

	return c
}

// Eval implements the interface sql.Expression.
func (c *CompiledAggregateFunction) Eval(ctx *sql.Context, row sql.Row) (interface{}, error) {
	return nil, cerrors.New("Eval should not be called on CompiledAggregateFunction")
}

// WithChildren implements the interface sql.Expression.
func (c *CompiledAggregateFunction) WithChildren(children ...sql.Expression) (sql.Expression, error) {
	if len(children) != len(c.Arguments) {
		return nil, sql.ErrInvalidChildrenNumber.New(len(children), len(c.Arguments))
	}

	// We have to re-resolve here, since the change in children may require it (e.g. we have more type info than we did)
	return newCompiledAggregateFunctionInternal(c.Name, children, c.overloads, c.fnOverloads, c.newBuffer), nil
}

// SetStatementRunner implements the interface analyzer.Interpreter.
func (c *CompiledAggregateFunction) SetStatementRunner(ctx *sql.Context, runner sql.StatementRunner) sql.Expression {
	nc := *c
	nc.runner = runner
	return &nc
}

// specificFuncImpl implements the interface sql.Expression.
func (*CompiledAggregateFunction) specificFuncImpl() {}

func (c *CompiledAggregateFunction) DebugString() string {
	sb := strings.Builder{}
	sb.WriteString("CompiledAggregateFunction:")
	sb.WriteString(c.Name + "(")
	for i, param := range c.Arguments {
		// Aliases will output the string "x as x", which is an artifact of how we build the AST, so we'll bypass it
		if alias, ok := param.(*expression.Alias); ok {
			param = alias.Child
		}
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(sql.DebugString(param))
	}
	sb.WriteString(")")
	return sb.String()
}

// NewBuffer implements the interface sql.Aggregation.
func (c *CompiledAggregateFunction) NewBuffer() (sql.AggregationBuffer, error) {
	return c.newBuffer(c.Arguments)
}

// Id implements the interface sql.Aggregation.
func (c *CompiledAggregateFunction) Id() sql.ColumnId {
	return c.aggId
}

// WithId implements the interface sql.Aggregation.
func (c *CompiledAggregateFunction) WithId(id sql.ColumnId) sql.IdExpression {
	nc := *c
	nc.aggId = id
	return &nc
}

// NewWindowFunction implements the interface sql.WindowAdaptableExpression.
func (c *CompiledAggregateFunction) NewWindowFunction() (sql.WindowFunction, error) {
	panic("windows are not implemented yet")
}

// WithWindow implements the interface sql.WindowAdaptableExpression.
func (c *CompiledAggregateFunction) WithWindow(window *sql.WindowDefinition) sql.WindowAdaptableExpression {
	panic("windows are not implemented yet")
}

// Window implements the interface sql.WindowAdaptableExpression.
func (c *CompiledAggregateFunction) Window() *sql.WindowDefinition {
	panic("windows are not implemented yet")
}
