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
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/server/extensions"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// compiledCatalog contains all of PostgreSQL functions in their compiled forms.
var compiledCatalog = map[string]sql.CreateFuncNArgs{}

// GetFunction returns the compiled function with the given name and parameters. Returns false if the function could not
// be found.
func GetFunction(ctx *sql.Context, functionName string, params ...sql.Expression) (*CompiledFunction, bool, error) {
	if createFunc, ok := compiledCatalog[functionName]; ok {
		expr, err := createFunc(ctx, params...)
		if err != nil {
			return nil, false, err
		}
		return expr.(*CompiledFunction), true, nil
	}
	return nil, false, nil
}

// dummyExpression is a simple expression that exists solely to capture type information for a parameter. This is used
// exclusively by the getQuickFunctionForTypes function.
type dummyExpression struct {
	t *pgtypes.DoltgresType
}

var _ sql.Expression = dummyExpression{}

func (d dummyExpression) Resolved() bool                   { return true }
func (d dummyExpression) String() string                   { return d.t.String() }
func (d dummyExpression) Type(ctx *sql.Context) sql.Type   { return d.t }
func (d dummyExpression) IsNullable(ctx *sql.Context) bool { return false }
func (d dummyExpression) Eval(ctx *sql.Context, row sql.Row) (interface{}, error) {
	panic("cannot Eval dummyExpression")
}
func (d dummyExpression) Children() []sql.Expression { return nil }
func (d dummyExpression) WithChildren(ctx *sql.Context, children ...sql.Expression) (sql.Expression, error) {
	return d, nil
}

// getQuickFunctionForTypes is used by the types package to load quick functions. This is declared here to work around
// import cycles. Returns nil if a QuickFunction could not be constructed.
func getQuickFunctionForTypes(ctx *sql.Context, schemaName string, functionName string, params []*pgtypes.DoltgresType) any {
	if schemaName != "pg_catalog" {
		return getQuickFunctionFromProvider(ctx, schemaName, functionName, params)
	}
	exprs := make([]sql.Expression, len(params))
	for i := range params {
		exprs[i] = dummyExpression{t: params[i]}
	}
	cf, ok, err := GetFunction(ctx, functionName, exprs...)
	if err != nil || !ok {
		return nil
	}
	return cf.GetQuickFunction(ctx)
}

// getQuickFunctionFromProvider resolves a user-defined function. Returns nil if it could not be resolved.
func getQuickFunctionFromProvider(ctx *sql.Context, schemaName string, functionName string, params []*pgtypes.DoltgresType) any {
	call := NewUserFunctionCall(ctx, schemaName, functionName, params)
	if call == nil {
		return nil
	}
	return &quickWrappedFunction{
		callable: func(ctx *sql.Context, resolvedTypes []*pgtypes.DoltgresType, args []any) (any, error) {
			compiled := *call.compiled
			compiled.callResolved = resolvedTypes
			return compiled.callFunction(ctx, args)
		},
		strict:        call.strict,
		resolvedTypes: call.compiled.callResolved,
	}
}

// getQuickExtensionFunction resolves an extension-provided function by name only. Returns nil if no registered
// extension declares a matching routine. Extension routines never read their resolved types, so the wrapper carries
// placeholders sized to the routine's signature.
func getQuickExtensionFunction(functionID id.Function) any {
	routine, ok := extensions.GetRoutine(functionID.FunctionName(), functionID.ParameterCount())
	if !ok {
		return nil
	}
	return &quickWrappedFunction{
		callable: func(ctx *sql.Context, _ []*pgtypes.DoltgresType, args []any) (any, error) {
			return routine.Impl(ctx, args...)
		},
		strict:        routine.Strict,
		resolvedTypes: make([]*pgtypes.DoltgresType, len(routine.Parameters)+1),
	}
}
