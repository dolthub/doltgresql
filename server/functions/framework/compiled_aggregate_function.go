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
	"fmt"
	"strings"

	cerrors "github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/expression"

	"github.com/dolthub/doltgresql/server/plpgsql"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// AggregateFunction is an expression that represents CompiledAggregateFunction
type AggregateFunction interface {
	sql.FunctionExpression
	sql.Aggregation
	specificFuncImpl()
}

// CompiledFunction is an expression that represents a fully-analyzed PostgreSQL function.
type CompiledAggregateFunction struct {
	*CompiledFunction
	aggId sql.ColumnId
}

func (c *CompiledAggregateFunction) Id() sql.ColumnId {
	return c.aggId
}

func (c *CompiledAggregateFunction) WithId(id sql.ColumnId) sql.IdExpression {
	nc := *c
	nc.aggId = id
	return &nc
}

func (c *CompiledAggregateFunction) NewWindowFunction() (sql.WindowFunction, error) {
	panic("windows are not implemented yet")
}

func (c *CompiledAggregateFunction) WithWindow(window *sql.WindowDefinition) sql.WindowAdaptableExpression {
	panic("windows are not implemented yet")
}

func (c *CompiledAggregateFunction) Window() *sql.WindowDefinition {
	panic("windows are not implemented yet")
}

type arrayAggBuffer struct {
	elements []any	
}

func newArrayAggBuffer() *arrayAggBuffer {
	return &arrayAggBuffer{
		elements: make([]any, 0),
	}
}

func (a *arrayAggBuffer) Dispose() {}

func (a *arrayAggBuffer) Eval(context *sql.Context) (interface{}, error) {
	if len(a.elements) == 0 {
		return nil, nil
	}
	return a.elements, nil
}

func (a *arrayAggBuffer) Update(ctx *sql.Context, row sql.Row) error {
	a.elements = append(a.elements, row[0])
	return nil
}

func (c *CompiledAggregateFunction) NewBuffer() (sql.AggregationBuffer, error) {
	return newArrayAggBuffer(), nil
}

var _ AggregateFunction = (*CompiledAggregateFunction)(nil)

// NewCompiledAggregateFunction returns a newly compiled function.
func NewCompiledAggregateFunction(name string, args []sql.Expression, functions *Overloads) *CompiledAggregateFunction {
	return newCompiledAggregateFunctionInternal(name, args, functions, functions.overloadsForParams(len(args)))
}

// newCompiledFunctionInternal is called internally, which skips steps that may have already been processed.
func newCompiledAggregateFunctionInternal(
	name string,
	args []sql.Expression,
	overloads *Overloads,
	fnOverloads []Overload,
) *CompiledAggregateFunction {
	
	cf := newCompiledFunctionInternal(name, args, overloads, fnOverloads, false, nil)
	c := &CompiledAggregateFunction{
		CompiledFunction: cf,
	}
	
	return c
}

// FunctionName implements the interface sql.Expression.
func (c *CompiledAggregateFunction) FunctionName() string {
	return c.Name
}

// Description implements the interface sql.Expression.
func (c *CompiledAggregateFunction) Description() string {
	return fmt.Sprintf("The PostgreSQL function `%s`", c.Name)
}

// Resolved implements the interface sql.Expression.
func (c *CompiledAggregateFunction) Resolved() bool {
	for _, param := range c.Arguments {
		if !param.Resolved() {
			return false
		}
	}
	// We don't error until evaluation time, so we need to tell the engine we're resolved if there was a stashed error
	return c.stashedErr != nil || c.overload.Valid()
}

// StashedError returns the stashed error if one exists. Otherwise, returns nil.
func (c *CompiledAggregateFunction) StashedError() error {
	if c == nil {
		return nil
	}
	return c.stashedErr
}

// String implements the interface sql.Expression.
func (c *CompiledAggregateFunction) String() string {
	sb := strings.Builder{}
	sb.WriteString(c.Name + "(")
	for i, param := range c.Arguments {
		// Aliases will output the string "x as x", which is an artifact of how we build the AST, so we'll bypass it
		if alias, ok := param.(*expression.Alias); ok {
			param = alias.Child
		}
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(param.String())
	}
	sb.WriteString(")")
	return sb.String()
}

// OverloadString returns the name of the function represented by the given overload.
func (c *CompiledAggregateFunction) OverloadString(types []*pgtypes.DoltgresType) string {
	sb := strings.Builder{}
	sb.WriteString(c.Name + "(")
	for i, t := range types {
		if i > 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(t.String())
	}
	sb.WriteString(")")
	return sb.String()
}

// Type implements the interface sql.Expression.
func (c *CompiledAggregateFunction) Type() sql.Type {
	if len(c.callResolved) > 0 {
		return c.callResolved[len(c.callResolved)-1]
	}
	// Compilation must have errored, so we'll return the unknown type
	return pgtypes.Unknown
}

// IsNullable implements the interface sql.Expression.
func (c *CompiledAggregateFunction) IsNullable() bool {
	// All functions seem to return NULL when given a NULL value
	return true
}

// IsNonDeterministic implements the interface sql.NonDeterministicExpression.
func (c *CompiledAggregateFunction) IsNonDeterministic() bool {
	if c.overload.Valid() {
		return c.overload.Function().NonDeterministic()
	}
	// Compilation must have errored, so we'll just return true
	return true
}

// IsVariadic returns whether this function has any variadic parameters.
func (c *CompiledAggregateFunction) IsVariadic() bool {
	if c.overload.Valid() {
		return c.overload.params.variadic != -1
	}
	// Compilation must have errored, so we'll just return true
	return true
}

// Eval implements the interface sql.Expression.
func (c *CompiledAggregateFunction) Eval(ctx *sql.Context, row sql.Row) (interface{}, error) {
	// TODO: probably should be an error?
	
	// If we have a stashed error, then we should return that now. Errors are stashed when they're supposed to be
	// returned during the call to Eval. This helps to ensure consistency with how errors are returned in Postgres.
	if c.stashedErr != nil {
		return nil, c.stashedErr
	}

	// Evaluate all arguments, returning immediately if we encounter a null argument and the function is marked STRICT
	var err error
	isStrict := c.overload.Function().IsStrict()
	args := make([]any, len(c.Arguments))
	for i, arg := range c.Arguments {
		args[i], err = arg.Eval(ctx, row)
		if err != nil {
			return nil, err
		}
		// TODO: once we remove GMS types from all of our expressions, we can remove this step which ensures the correct type
		if _, ok := arg.Type().(*pgtypes.DoltgresType); !ok {
			dt, err := pgtypes.FromGmsTypeToDoltgresType(arg.Type())
			if err != nil {
				return nil, err
			}
			args[i], _, _ = dt.Convert(ctx, args[i])
		}
		if args[i] == nil && isStrict {
			return nil, nil
		}
	}

	if len(c.overload.casts) > 0 {
		targetParamTypes := c.overload.Function().GetParameters()
		for i, arg := range args {
			// For variadic params, we need to identify the corresponding target type
			var targetType *pgtypes.DoltgresType
			isVariadicArg := c.overload.params.variadic >= 0 && i >= len(c.overload.params.paramTypes)-1
			if isVariadicArg {
				targetType = targetParamTypes[c.overload.params.variadic]
				if !targetType.IsArrayType() {
					// should be impossible, we check this at function compile time
					return nil, cerrors.Errorf("variadic arguments must be array types, was %T", targetType)
				}
				targetType = targetType.ArrayBaseType()
			} else {
				targetType = targetParamTypes[i]
			}

			if c.overload.casts[i] != nil {
				args[i], err = c.overload.casts[i](ctx, arg, targetType)
				if err != nil {
					return nil, err
				}
			} else {
				return nil, cerrors.Errorf("function %s is missing the appropriate implicit cast", c.OverloadString(c.originalTypes))
			}
		}
	}

	args = c.overload.params.coalesceVariadicValues(args)

	// Call the function
	switch f := c.overload.Function().(type) {
	case Function0:
		return f.Callable(ctx)
	case Function1:
		return f.Callable(ctx, ([2]*pgtypes.DoltgresType)(c.callResolved), args[0])
	case Function2:
		return f.Callable(ctx, ([3]*pgtypes.DoltgresType)(c.callResolved), args[0], args[1])
	case Function3:
		return f.Callable(ctx, ([4]*pgtypes.DoltgresType)(c.callResolved), args[0], args[1], args[2])
	case Function4:
		return f.Callable(ctx, ([5]*pgtypes.DoltgresType)(c.callResolved), args[0], args[1], args[2], args[3])
	case InterpretedFunction:
		return plpgsql.Call(ctx, f, c.runner, c.callResolved, args)
	default:
		return nil, cerrors.Errorf("unknown function type in CompiledFunction::Eval")
	}
}

// Children implements the interface sql.Expression.
func (c *CompiledAggregateFunction) Children() []sql.Expression {
	return c.Arguments
}

// WithChildren implements the interface sql.Expression.
func (c *CompiledAggregateFunction) WithChildren(children ...sql.Expression) (sql.Expression, error) {
	if len(children) != len(c.Arguments) {
		return nil, sql.ErrInvalidChildrenNumber.New(len(children), len(c.Arguments))
	}

	// We have to re-resolve here, since the change in children may require it (e.g. we have more type info than we did)
	return newCompiledAggregateFunctionInternal(c.Name, children, c.overloads, c.fnOverloads), nil
}

// specificFuncImpl implements the interface sql.Expression.
func (*CompiledAggregateFunction) specificFuncImpl() {}
