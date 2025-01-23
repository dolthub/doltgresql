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

	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"

	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// QuickFunction represents an optimized function expression that has specific criteria in order to streamline
// evaluation. This will only apply to very specific functions that are generally performance-critical.
type QuickFunction interface {
	Function
	// CallVariadic is the variadic form of the Call function that is specific to each implementation of QuickFunction.
	// The implementation will not verify that the correct number of arguments have been passed.
	CallVariadic(ctx *sql.Context, args ...any) (interface{}, error)
	// ResolvedTypes returns the types that were resolved with this function.
	ResolvedTypes() []*pgtypes.DoltgresType
	// WithResolvedTypes returns a new QuickFunction with the replaced resolved types. The implementation will not
	// verify that the new types are correct in any way. This returns a QuickFunction, however it's typed as "any" due
	// to potential import cycles.
	WithResolvedTypes(newTypes []*pgtypes.DoltgresType) any
}

// QuickFunction1 is an implementation of QuickFunction that handles a single parameter.
type QuickFunction1 struct {
	Name         string
	Argument     sql.Expression
	IsStrict     bool
	callResolved [2]*pgtypes.DoltgresType
	function     Function1
}

var _ QuickFunction = (*QuickFunction1)(nil)

// FunctionName implements the interface sql.Expression.
func (q *QuickFunction1) FunctionName() string {
	return q.Name
}

// Description implements the interface sql.Expression.
func (q *QuickFunction1) Description() string {
	return fmt.Sprintf("The PostgreSQL function `%s`", q.Name)
}

// Resolved implements the interface sql.Expression.
func (q *QuickFunction1) Resolved() bool {
	return true
}

// String implements the interface sql.Expression.
func (q *QuickFunction1) String() string {
	// We'll reuse the compiled function's output so that the logic is centralized
	c := CompiledFunction{
		Name:      q.Name,
		Arguments: []sql.Expression{q.Argument},
	}
	return c.String()
}

// Type implements the interface sql.Expression.
func (q *QuickFunction1) Type() sql.Type {
	return q.callResolved[1]
}

// IsNullable implements the interface sql.Expression.
func (q *QuickFunction1) IsNullable() bool {
	return true
}

// IsNonDeterministic implements the interface sql.NonDeterministicExpression.
func (q *QuickFunction1) IsNonDeterministic() bool {
	return q.function.IsNonDeterministic
}

// Eval implements the interface sql.Expression.
func (q *QuickFunction1) Eval(ctx *sql.Context, row sql.Row) (interface{}, error) {
	arg, err := q.Argument.Eval(ctx, row)
	if err != nil {
		return nil, err
	}
	if arg == nil && q.IsStrict {
		return nil, nil
	}
	return q.function.Callable(ctx, q.callResolved, arg)
}

// Call directly calls the underlying function with the given arguments. This does not perform any form of NULL checking
// as it is assumed that it was done prior to this call. It also does not validate any types. This exists purely for
// performance, when we can guarantee that the input is always valid and well-formed.
func (q *QuickFunction1) Call(ctx *sql.Context, arg0 any) (interface{}, error) {
	return q.function.Callable(ctx, q.callResolved, arg0)
}

// CallVariadic implements the interface QuickFunction.
func (q *QuickFunction1) CallVariadic(ctx *sql.Context, args ...any) (interface{}, error) {
	return q.function.Callable(ctx, q.callResolved, args[0])
}

// ResolvedTypes implements the interface QuickFunction.
func (q *QuickFunction1) ResolvedTypes() []*pgtypes.DoltgresType {
	return q.callResolved[:]
}

// WithResolvedTypes implements the interface QuickFunction.
func (q *QuickFunction1) WithResolvedTypes(newTypes []*pgtypes.DoltgresType) any {
	return &QuickFunction1{
		Name:         q.Name,
		Argument:     q.Argument,
		IsStrict:     q.IsStrict,
		callResolved: [2]*pgtypes.DoltgresType(newTypes),
		function:     q.function,
	}
}

// Children implements the interface sql.Expression.
func (q *QuickFunction1) Children() []sql.Expression {
	return []sql.Expression{q.Argument}
}

// WithChildren implements the interface sql.Expression.
func (q *QuickFunction1) WithChildren(children ...sql.Expression) (sql.Expression, error) {
	return nil, errors.Errorf("cannot change the children for `%T`", q)
}

// specificFuncImpl implements the interface sql.Expression.
func (*QuickFunction1) specificFuncImpl() {}

// QuickFunction2 is an implementation of QuickFunction that handles two parameters.
type QuickFunction2 struct {
	Name         string
	Arguments    [2]sql.Expression
	IsStrict     bool
	callResolved [3]*pgtypes.DoltgresType
	function     Function2
}

var _ QuickFunction = (*QuickFunction2)(nil)

// FunctionName implements the interface sql.Expression.
func (q *QuickFunction2) FunctionName() string {
	return q.Name
}

// Description implements the interface sql.Expression.
func (q *QuickFunction2) Description() string {
	return fmt.Sprintf("The PostgreSQL function `%s`", q.Name)
}

// Resolved implements the interface sql.Expression.
func (q *QuickFunction2) Resolved() bool {
	return true
}

// String implements the interface sql.Expression.
func (q *QuickFunction2) String() string {
	// We'll reuse the compiled function's output so that the logic is centralized
	c := CompiledFunction{
		Name:      q.Name,
		Arguments: q.Arguments[:],
	}
	return c.String()
}

// Type implements the interface sql.Expression.
func (q *QuickFunction2) Type() sql.Type {
	return q.callResolved[2]
}

// IsNullable implements the interface sql.Expression.
func (q *QuickFunction2) IsNullable() bool {
	return true
}

// IsNonDeterministic implements the interface sql.NonDeterministicExpression.
func (q *QuickFunction2) IsNonDeterministic() bool {
	return q.function.IsNonDeterministic
}

// Eval implements the interface sql.Expression.
func (q *QuickFunction2) Eval(ctx *sql.Context, row sql.Row) (interface{}, error) {
	var args [2]any
	for i := range q.Arguments {
		var err error
		args[i], err = q.Arguments[i].Eval(ctx, row)
		if err != nil {
			return nil, err
		}
		if args[i] == nil && q.IsStrict {
			return nil, nil
		}
	}
	return q.function.Callable(ctx, q.callResolved, args[0], args[1])
}

// Call directly calls the underlying function with the given arguments. This does not perform any form of NULL checking
// as it is assumed that it was done prior to this call. It also does not validate any types. This exists purely for
// performance, when we can guarantee that the input is always valid and well-formed.
func (q *QuickFunction2) Call(ctx *sql.Context, arg0 any, arg1 any) (interface{}, error) {
	return q.function.Callable(ctx, q.callResolved, arg0, arg1)
}

// CallVariadic implements the interface QuickFunction.
func (q *QuickFunction2) CallVariadic(ctx *sql.Context, args ...any) (interface{}, error) {
	return q.function.Callable(ctx, q.callResolved, args[0], args[1])
}

// ResolvedTypes implements the interface QuickFunction.
func (q *QuickFunction2) ResolvedTypes() []*pgtypes.DoltgresType {
	return q.callResolved[:]
}

// WithResolvedTypes implements the interface QuickFunction.
func (q *QuickFunction2) WithResolvedTypes(newTypes []*pgtypes.DoltgresType) any {
	return &QuickFunction2{
		Name:         q.Name,
		Arguments:    q.Arguments,
		IsStrict:     q.IsStrict,
		callResolved: [3]*pgtypes.DoltgresType(newTypes),
		function:     q.function,
	}
}

// Children implements the interface sql.Expression.
func (q *QuickFunction2) Children() []sql.Expression {
	return q.Arguments[:]
}

// WithChildren implements the interface sql.Expression.
func (q *QuickFunction2) WithChildren(children ...sql.Expression) (sql.Expression, error) {
	return nil, errors.Errorf("cannot change the children for `%T`", q)
}

// specificFuncImpl implements the interface sql.Expression.
func (*QuickFunction2) specificFuncImpl() {}

// QuickFunction3 is an implementation of QuickFunction that handles three parameters.
type QuickFunction3 struct {
	Name         string
	Arguments    [3]sql.Expression
	IsStrict     bool
	callResolved [4]*pgtypes.DoltgresType
	function     Function3
}

var _ QuickFunction = (*QuickFunction3)(nil)

// FunctionName implements the interface sql.Expression.
func (q *QuickFunction3) FunctionName() string {
	return q.Name
}

// Description implements the interface sql.Expression.
func (q *QuickFunction3) Description() string {
	return fmt.Sprintf("The PostgreSQL function `%s`", q.Name)
}

// Resolved implements the interface sql.Expression.
func (q *QuickFunction3) Resolved() bool {
	return true
}

// String implements the interface sql.Expression.
func (q *QuickFunction3) String() string {
	// We'll reuse the compiled function's output so that the logic is centralized
	c := CompiledFunction{
		Name:      q.Name,
		Arguments: q.Arguments[:],
	}
	return c.String()
}

// Type implements the interface sql.Expression.
func (q *QuickFunction3) Type() sql.Type {
	return q.callResolved[3]
}

// IsNullable implements the interface sql.Expression.
func (q *QuickFunction3) IsNullable() bool {
	return true
}

// IsNonDeterministic implements the interface sql.NonDeterministicExpression.
func (q *QuickFunction3) IsNonDeterministic() bool {
	return q.function.IsNonDeterministic
}

// Eval implements the interface sql.Expression.
func (q *QuickFunction3) Eval(ctx *sql.Context, row sql.Row) (interface{}, error) {
	var args [3]any
	for i := range q.Arguments {
		var err error
		args[i], err = q.Arguments[i].Eval(ctx, row)
		if err != nil {
			return nil, err
		}
		if args[i] == nil && q.IsStrict {
			return nil, nil
		}
	}
	return q.function.Callable(ctx, q.callResolved, args[0], args[1], args[2])
}

// Call directly calls the underlying function with the given arguments. This does not perform any form of NULL checking
// as it is assumed that it was done prior to this call. It also does not validate any types. This exists purely for
// performance, when we can guarantee that the input is always valid and well-formed.
func (q *QuickFunction3) Call(ctx *sql.Context, arg0 any, arg1 any, arg2 any) (interface{}, error) {
	return q.function.Callable(ctx, q.callResolved, arg0, arg1, arg2)
}

// CallVariadic implements the interface QuickFunction.
func (q *QuickFunction3) CallVariadic(ctx *sql.Context, args ...any) (interface{}, error) {
	return q.function.Callable(ctx, q.callResolved, args[0], args[1], args[2])
}

// ResolvedTypes implements the interface QuickFunction.
func (q *QuickFunction3) ResolvedTypes() []*pgtypes.DoltgresType {
	return q.callResolved[:]
}

// WithResolvedTypes implements the interface QuickFunction.
func (q *QuickFunction3) WithResolvedTypes(newTypes []*pgtypes.DoltgresType) any {
	return &QuickFunction3{
		Name:         q.Name,
		Arguments:    q.Arguments,
		IsStrict:     q.IsStrict,
		callResolved: [4]*pgtypes.DoltgresType(newTypes),
		function:     q.function,
	}
}

// Children implements the interface sql.Expression.
func (q *QuickFunction3) Children() []sql.Expression {
	return q.Arguments[:]
}

// WithChildren implements the interface sql.Expression.
func (q *QuickFunction3) WithChildren(children ...sql.Expression) (sql.Expression, error) {
	return nil, errors.Errorf("cannot change the children for `%T`", q)
}

// specificFuncImpl implements the interface sql.Expression.
func (*QuickFunction3) specificFuncImpl() {}
