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
	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/core"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// UserFunctionCall is a resolved call to a user-defined function that takes argument values rather than argument
// expressions.
type UserFunctionCall struct {
	compiled *CompiledFunction
	strict   bool
}

// GetUserFunction returns the compiled call to the named function with the given argument expressions. Returns nil if
// the function does not exist.
func GetUserFunction(ctx *sql.Context, schemaName string, functionName string, args ...sql.Expression) (Function, error) {
	sqlFunc, ok := (&FunctionProvider{}).Function(ctx, schemaName, functionName)
	if !ok {
		return nil, nil
	}
	functionN, ok := sqlFunc.(sql.FunctionN)
	if !ok {
		return nil, nil
	}
	expr, err := functionN.Fn(ctx, args...)
	if err != nil {
		return nil, err
	}
	f, ok := expr.(Function)
	if !ok {
		return nil, errors.Errorf("function `%s` has an unexpected type: %T", functionName, expr)
	}
	return f, nil
}

// NewUserFunctionCall resolves the named function for the given parameter types. Returns nil if no overload matches.
func NewUserFunctionCall(ctx *sql.Context, schemaName string, functionName string, paramTypes []*pgtypes.DoltgresType) *UserFunctionCall {
	exprs := make([]sql.Expression, len(paramTypes))
	for i := range paramTypes {
		exprs[i] = dummyExpression{t: paramTypes[i]}
	}
	f, err := GetUserFunction(ctx, schemaName, functionName, exprs...)
	if err != nil {
		return nil
	}
	compiled, ok := f.(*CompiledFunction)
	if !ok || !compiled.Resolved() || !compiled.overload.Valid() {
		return nil
	}
	if runner, err := core.GetRunnerFromContext(ctx); err == nil && runner != nil {
		compiled = compiled.SetStatementRunner(ctx, runner).(*CompiledFunction)
	}
	return &UserFunctionCall{compiled: compiled, strict: compiled.IsStrict()}
}

// Call invokes the function with the given argument values.
func (u *UserFunctionCall) Call(ctx *sql.Context, args ...any) (any, error) {
	if u.strict {
		for _, arg := range args {
			if arg == nil {
				return nil, nil
			}
		}
	}
	return u.compiled.callFunction(ctx, args)
}
