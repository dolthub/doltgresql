// Copyright 2026 Dolthub, Inc.
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

// WindowOnlyFunction is an expression that represents CompiledWindowFunction: a PostgreSQL function that may
// only be used as a window function (within an OVER(...) clause), such as row_number() or rank().
type WindowOnlyFunction interface {
	sql.FunctionExpression
	sql.WindowAggregation
	specificFuncImpl()
}

// CompiledWindowFunction is an expression that represents a fully-analyzed PostgreSQL window-only function.
type CompiledWindowFunction struct {
	*CompiledFunction
	windowId sql.ColumnId
	window   *sql.WindowDefinition
}

var _ WindowOnlyFunction = (*CompiledWindowFunction)(nil)

// NewCompiledWindowFunction returns a newly compiled function.
func NewCompiledWindowFunction(ctx *sql.Context, name string, args []sql.Expression, functions *Overloads) *CompiledWindowFunction {
	return newCompiledWindowFunctionInternal(ctx, name, args, functions, functions.overloadsForParams(len(args)))
}

// newCompiledWindowFunctionInternal is called internally, which skips steps that may have already been processed.
func newCompiledWindowFunctionInternal(ctx *sql.Context, name string, args []sql.Expression, overloads *Overloads, fnOverloads []Overload) *CompiledWindowFunction {
	cf := newCompiledFunctionInternal(ctx, name, args, overloads, fnOverloads, false, nil)
	return &CompiledWindowFunction{
		CompiledFunction: cf,
	}
}

// windowOverload returns the WindowFunctionInterface that this function's overload resolution matched for its
// actual argument types, mirroring CompiledAggregateFunction.aggregateOverload.
func (c *CompiledWindowFunction) windowOverload() (WindowFunctionInterface, error) {
	if !c.overload.Valid() {
		return nil, cerrors.Errorf("%s: no matching overload was resolved", c.Name)
	}
	fn, ok := c.overload.Function().(WindowFunctionInterface)
	if !ok {
		return nil, cerrors.Errorf("%s: resolved overload is not a window function", c.Name)
	}
	return fn, nil
}

// Eval implements the interface sql.Expression.
func (c *CompiledWindowFunction) Eval(ctx *sql.Context, row sql.Row) (interface{}, error) {
	return nil, cerrors.New("Eval should not be called on CompiledWindowFunction")
}

// Children implements the interface sql.Expression. When this function is bound to a window (via
// WithWindow), the window's PartitionBy/OrderBy expressions are included after its own arguments, mirroring
// GMS's own unaryAggBase.Children(): analyzer passes such as column pruning only discover a node's column
// dependencies by walking Children(), and window.PartitionBy/OrderBy are otherwise invisible to them since
// they aren't part of Arguments.
func (c *CompiledWindowFunction) Children() []sql.Expression {
	children := append([]sql.Expression{}, c.Arguments...)
	if c.window != nil {
		children = append(children, c.window.ToExpressions()...)
	}
	return children
}

// WithChildren implements the interface sql.Expression.
func (c *CompiledWindowFunction) WithChildren(ctx *sql.Context, children ...sql.Expression) (sql.Expression, error) {
	numArgs := len(c.Arguments)
	if len(children) < numArgs {
		return nil, sql.ErrInvalidChildrenNumber.New(len(children), numArgs)
	}

	// We have to re-resolve here, since the change in children may require it (e.g. we have more type info than we did)
	nc := newCompiledWindowFunctionInternal(ctx, c.Name, children[:numArgs], c.overloads, c.fnOverloads)
	// Preserve the identity assigned by the planbuilder: a later analyzer pass that rebuilds this
	// expression via WithChildren (e.g. a generic bottom-up transform) must not silently lose the
	// WithId/WithWindow state that was already bound onto c.
	nc.windowId = c.windowId
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
func (c *CompiledWindowFunction) SetStatementRunner(ctx *sql.Context, runner sql.StatementRunner) sql.Expression {
	nc := *c
	nc.runner = runner
	return &nc
}

// specificFuncImpl implements the interface WindowOnlyFunction.
func (*CompiledWindowFunction) specificFuncImpl() {}

func (c *CompiledWindowFunction) DebugString(ctx *sql.Context) string {
	sb := strings.Builder{}
	sb.WriteString("CompiledWindowFunction:")
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

// Id implements the interface sql.IdExpression.
func (c *CompiledWindowFunction) Id() sql.ColumnId {
	return c.windowId
}

// WithId implements the interface sql.IdExpression.
func (c *CompiledWindowFunction) WithId(id sql.ColumnId) sql.IdExpression {
	nc := *c
	nc.windowId = id
	return &nc
}

// NewWindowFunction implements the interface sql.WindowAdaptableExpression.
func (c *CompiledWindowFunction) NewWindowFunction(ctx *sql.Context) (sql.WindowFunction, error) {
	fn, err := c.windowOverload()
	if err != nil {
		return nil, err
	}
	newWindowFunc := fn.NewWindowFunc()
	if newWindowFunc == nil {
		return nil, cerrors.Errorf("function %s cannot be used as a window function", c.Name)
	}
	// See cloneArguments: each partition needs its own argument expression instances so stateful
	// expressions (e.g. DISTINCT's dedup cache) don't leak state across partitions.
	args, err := cloneArguments(ctx, c.Arguments)
	if err != nil {
		return nil, err
	}
	return newWindowFunc(args, c.window)
}

// WithWindow implements the interface sql.WindowAdaptableExpression.
func (c *CompiledWindowFunction) WithWindow(ctx *sql.Context, window *sql.WindowDefinition) sql.WindowAdaptableExpression {
	nc := *c
	nc.window = window
	return &nc
}

// Window implements the interface sql.WindowAdaptableExpression.
func (c *CompiledWindowFunction) Window() *sql.WindowDefinition {
	return c.window
}
