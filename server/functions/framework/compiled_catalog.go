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

	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// compiledCatalog contains all of PostgreSQL functions in their compiled forms.
var compiledCatalog = map[string]sql.CreateFuncNArgs{}

// namedCatalog contains the definitions of every PostgreSQL function associated with the given name.
var namedCatalog = map[string][]FunctionInterface{}

// GetFunction returns the compiled function with the given name and parameters. Returns false if the function could not
// be found.
func GetFunction(functionName string, params ...sql.Expression) (*CompiledFunction, bool, error) {
	if createFunc, ok := compiledCatalog[functionName]; ok {
		expr, err := createFunc(params...)
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

func (d dummyExpression) Resolved() bool   { return true }
func (d dummyExpression) String() string   { return d.t.String() }
func (d dummyExpression) Type() sql.Type   { return d.t }
func (d dummyExpression) IsNullable() bool { return false }
func (d dummyExpression) Eval(ctx *sql.Context, row sql.Row) (interface{}, error) {
	panic("cannot Eval dummyExpression")
}
func (d dummyExpression) Children() []sql.Expression { return nil }
func (d dummyExpression) WithChildren(children ...sql.Expression) (sql.Expression, error) {
	return d, nil
}

// getQuickFunctionForTypes is used by the types package to load quick functions. This is declared here to work around
// import cycles. Returns nil if a QuickFunction could not be constructed.
func getQuickFunctionForTypes(functionName string, params []*pgtypes.DoltgresType) any {
	exprs := make([]sql.Expression, len(params))
	for i := range params {
		exprs[i] = dummyExpression{t: params[i]}
	}
	cf, ok, err := GetFunction(functionName, exprs...)
	if err != nil || !ok {
		return nil
	}
	return cf.GetQuickFunction()
}
