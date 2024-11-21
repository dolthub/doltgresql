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

	"github.com/dolthub/go-mysql-server/sql"

	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// getFunctionAndEvaluateForTypes is a shortcut to getting CompiledFunction and evaluating the output result.
func getFunctionAndEvaluateForTypes(ctx *sql.Context, funcName string, paramTypes []pgtypes.DoltgresType, args []any) (any, error) {
	// get function and do Callable immediately
	overloads, ok := Catalog[funcName]
	if !ok {
		return nil, ErrFunctionDoesNotExist.New(funcName)
	}
	//There should only be one function
	if len(overloads) != 1 {
		return nil, fmt.Errorf("expected only one function named: %s", funcName)
	}
	function := overloads[0]

	if function.IsStrict() {
		for i := range args {
			if args[i] == nil {
				return nil, nil
			}
		}
	}

	funcTypes := append(paramTypes, function.GetReturn())
	// Call the function
	switch f := function.(type) {
	case Function0:
		return f.Callable(ctx)
	case Function1:
		return f.Callable(ctx, ([2]pgtypes.DoltgresType)(funcTypes), args[0])
	case Function2:
		return f.Callable(ctx, ([3]pgtypes.DoltgresType)(funcTypes), args[0], args[1])
	case Function3:
		return f.Callable(ctx, ([4]pgtypes.DoltgresType)(funcTypes), args[0], args[1], args[2])
	case Function4:
		return f.Callable(ctx, ([5]pgtypes.DoltgresType)(funcTypes), args[0], args[1], args[2], args[3])
	default:
		return nil, fmt.Errorf("unknown function type in type functions")
	}
}
