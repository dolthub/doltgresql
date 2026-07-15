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
	"github.com/dolthub/go-mysql-server/sql/transform"
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
	aggId  sql.ColumnId
	window *sql.WindowDefinition
}

var _ AggregateFunction = (*CompiledAggregateFunction)(nil)

// NewCompiledAggregateFunction returns a newly compiled function.
func NewCompiledAggregateFunction(ctx *sql.Context, name string, args []sql.Expression, functions *Overloads) *CompiledAggregateFunction {
	return newCompiledAggregateFunctionInternal(ctx, name, args, functions, functions.overloadsForParams(len(args)))
}

// newCompiledAggregateFunctionInternal is called internally, which skips steps that may have already been processed.
func newCompiledAggregateFunctionInternal(ctx *sql.Context, name string, args []sql.Expression, overloads *Overloads, fnOverloads []Overload) *CompiledAggregateFunction {
	cf := newCompiledFunctionInternal(ctx, name, args, overloads, fnOverloads, false, nil)
	return &CompiledAggregateFunction{
		CompiledFunction: cf,
	}
}

// aggregateOverload returns the AggregateFunctionInterface that this function's overload resolution matched
// for its actual argument types. Each overload of a given function name carries its own NewBuffer/NewWindowFunc,
// so a name like "sum" that has separate int4/int8/numeric/etc. overloads gets the correct implementation for
// the arguments actually bound, rather than a single implementation shared across every overload of the name.
func (c *CompiledAggregateFunction) aggregateOverload() (AggregateFunctionInterface, error) {
	if !c.overload.Valid() {
		return nil, cerrors.Errorf("%s: no matching overload was resolved", c.Name)
	}
	agg, ok := c.overload.Function().(AggregateFunctionInterface)
	if !ok {
		return nil, cerrors.Errorf("%s: resolved overload is not an aggregate function", c.Name)
	}
	return agg, nil
}

// Eval implements the interface sql.Expression.
func (c *CompiledAggregateFunction) Eval(ctx *sql.Context, row sql.Row) (interface{}, error) {
	return nil, cerrors.New("Eval should not be called on CompiledAggregateFunction")
}

// Children implements the interface sql.Expression. When this aggregate is bound to a window (via WithWindow),
// the window's PartitionBy/OrderBy expressions are included after the aggregate's own arguments, mirroring
// GMS's own unaryAggBase.Children(): analyzer passes such as column pruning only discover a node's column
// dependencies by walking Children(), and window.PartitionBy/OrderBy are otherwise invisible to them since
// they aren't part of Arguments.
func (c *CompiledAggregateFunction) Children() []sql.Expression {
	children := append([]sql.Expression{}, c.Arguments...)
	if c.window != nil {
		children = append(children, c.window.ToExpressions()...)
	}
	return children
}

// WithChildren implements the interface sql.Expression.
func (c *CompiledAggregateFunction) WithChildren(ctx *sql.Context, children ...sql.Expression) (sql.Expression, error) {
	numArgs := len(c.Arguments)
	if len(children) < numArgs {
		return nil, sql.ErrInvalidChildrenNumber.New(len(children), numArgs)
	}

	// We have to re-resolve here, since the change in children may require it (e.g. we have more type info than we did)
	nc := newCompiledAggregateFunctionInternal(ctx, c.Name, children[:numArgs], c.overloads, c.fnOverloads)
	// Preserve the aggregate/window identity assigned by the planbuilder: a later analyzer pass that
	// rebuilds this expression via WithChildren (e.g. a generic bottom-up transform) must not silently
	// lose the WithId/WithWindow state that was already bound onto c.
	nc.aggId = c.aggId
	nc.window = c.window
	if len(children) > numArgs && c.window != nil {
		w, err := c.window.FromExpressions(ctx, children[numArgs:])
		if err != nil {
			return nil, err
		}
		nc.window = w
	}
	return nc, nil
}

// SetStatementRunner implements the interface analyzer.Interpreter.
func (c *CompiledAggregateFunction) SetStatementRunner(ctx *sql.Context, runner sql.StatementRunner) sql.Expression {
	nc := *c
	nc.runner = runner
	return &nc
}

// specificFuncImpl implements the interface sql.Expression.
func (*CompiledAggregateFunction) specificFuncImpl() {}

func (c *CompiledAggregateFunction) DebugString(ctx *sql.Context) string {
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
		sb.WriteString(sql.DebugString(ctx, param))
	}
	sb.WriteString(")")
	return sb.String()
}

// NewBuffer implements the interface sql.Aggregation.
func (c *CompiledAggregateFunction) NewBuffer(ctx *sql.Context) (sql.AggregationBuffer, error) {
	agg, err := c.aggregateOverload()
	if err != nil {
		return nil, err
	}
	args, err := cloneArguments(ctx, c.Arguments)
	if err != nil {
		return nil, err
	}
	return agg.NewBuffer(args)
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
func (c *CompiledAggregateFunction) NewWindowFunction(ctx *sql.Context) (sql.WindowFunction, error) {
	agg, err := c.aggregateOverload()
	if err != nil {
		return nil, err
	}
	newWindowFunc := agg.NewWindowFunc()
	if newWindowFunc == nil {
		return nil, cerrors.Errorf("aggregate function %s cannot be used as a window function", c.Name)
	}
	args, err := cloneArguments(ctx, c.Arguments)
	if err != nil {
		return nil, err
	}
	return newWindowFunc(args, c.window)
}

// cloneArguments returns a deep copy of args. Each partition/group gets its own AggregationBuffer or
// WindowFunction instance, but without cloning, they'd all share the same argument expressions - so a
// stateful expression like DISTINCT's dedup cache (which must reset per group/partition) would incorrectly
// carry state across them. This mirrors GMS's own aggregations (e.g. *aggregation.Sum.NewBuffer), which clone
// their child expression for the same reason.
func cloneArguments(ctx *sql.Context, args []sql.Expression) ([]sql.Expression, error) {
	cloned := make([]sql.Expression, len(args))
	for i, arg := range args {
		c, err := transform.Clone(ctx, arg)
		if err != nil {
			return nil, err
		}
		cloned[i] = c
	}
	return cloned, nil
}

// WithWindow implements the interface sql.WindowAdaptableExpression.
func (c *CompiledAggregateFunction) WithWindow(ctx *sql.Context, window *sql.WindowDefinition) sql.WindowAdaptableExpression {
	nc := *c
	nc.window = window
	return &nc
}

// Window implements the interface sql.WindowAdaptableExpression.
func (c *CompiledAggregateFunction) Window() *sql.WindowDefinition {
	return c.window
}
