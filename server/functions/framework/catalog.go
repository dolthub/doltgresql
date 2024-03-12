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

	pgtypes "github.com/dolthub/doltgresql/server/types"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/expression/function"
)

// Function is a name, along with a collection of functions, that represent a single PostgreSQL function with all of its
// overloads.
type Function struct {
	Name      string
	Overloads []any
}

// Catalog contains all of the PostgreSQL functions.
var Catalog = map[string][]FunctionInterface{}

// initializedFunctions simply states whether Initialize has been called yet.
var initializedFunctions = false

// RegisterFunction registers the given function, so that it will be usable from a running server. This should be called
// from within an init().
func RegisterFunction(f FunctionInterface) {
	if initializedFunctions {
		panic("attempted to register a function after the init() phase")
	}
	switch f := f.(type) {
	case Function0:
		name := strings.ToLower(f.Name)
		Catalog[name] = append(Catalog[name], f)
	case Function1:
		name := strings.ToLower(f.Name)
		Catalog[name] = append(Catalog[name], f)
	case Function2:
		name := strings.ToLower(f.Name)
		Catalog[name] = append(Catalog[name], f)
	case Function3:
		name := strings.ToLower(f.Name)
		Catalog[name] = append(Catalog[name], f)
	case Function4:
		name := strings.ToLower(f.Name)
		Catalog[name] = append(Catalog[name], f)
	default:
		panic("unhandled function type")
	}
}

// Initialize handles the initialization of the catalog by overwriting the built-in GMS functions, since they do not
// apply to PostgreSQL (and functions of the same name often have different behavior).
func Initialize() {
	// This should only be called once. We don't use sync.Once since we also want to panic if someone attempts to
	// register a function after initialization.
	if initializedFunctions {
		return
	}
	initializedFunctions = true

	// Flush all GMS built-ins that have conflicting names with PostgreSQL functions
	functionNames := make(map[string]struct{})
	for name := range Catalog {
		functionNames[strings.ToLower(name)] = struct{}{}
	}
	var newBuiltIns []sql.Function
	for _, f := range function.BuiltIns {
		if _, ok := functionNames[strings.ToLower(f.FunctionName())]; !ok {
			newBuiltIns = append(newBuiltIns, f)
		}
	}
	function.BuiltIns = newBuiltIns

	for funcName, catalogFunctions := range Catalog {
		funcName := funcName
		// Verify that each function uses the correct Function overload
		for _, functionOverload := range catalogFunctions {
			if len(functionOverload.GetParameters()) != functionOverload.GetExpectedParameterCount() {
				panic(fmt.Errorf("function `%s` should have %d arguments but has %d arguments",
					funcName, functionOverload.GetExpectedParameterCount(), len(functionOverload.GetParameters())))
			}
		}
		// Verify that all overloads are unique
		for functionIndex, f1 := range catalogFunctions {
			sameCount := 0
			for _, f2 := range catalogFunctions[functionIndex+1:] {
				if f1.GetExpectedParameterCount() == f2.GetExpectedParameterCount() {
					f2Parameters := f2.GetParameters()
					for parameterIndex, f1Parameter := range f1.GetParameters() {
						if f1Parameter.Equals(f2Parameters[parameterIndex]) {
							sameCount++
						}
					}
				}
			}
			if sameCount == f1.GetExpectedParameterCount() && f1.GetExpectedParameterCount() > 0 {
				panic(fmt.Errorf("duplicate function overloads on `%s`", funcName))
			}
		}
		// Build the overloads
		baseOverload := &OverloadDeduction{Parameter: make(map[pgtypes.DoltgresTypeBaseID]*OverloadDeduction)}
		for _, functionOverload := range catalogFunctions {
			// Loop through all of the parameters
			currentOverload := baseOverload
			for _, param := range functionOverload.GetParameters() {
				nextOverload := currentOverload.Parameter[param.BaseID()]
				if nextOverload == nil {
					nextOverload = &OverloadDeduction{Parameter: make(map[pgtypes.DoltgresTypeBaseID]*OverloadDeduction)}
					currentOverload.Parameter[param.BaseID()] = nextOverload
				}
				currentOverload = nextOverload
			}
			// This should never happen, but we'll check anyway to be safe
			if currentOverload.Function != nil {
				panic(fmt.Errorf("function `%s` somehow has duplicate overloads that weren't caught earlier", funcName))
			}
			currentOverload.Function = functionOverload
		}

		// Store the compiled function into the engine's built-in functions
		function.BuiltIns = append(function.BuiltIns, sql.FunctionN{
			Name: funcName,
			Fn: func(params ...sql.Expression) (sql.Expression, error) {
				return &CompiledFunction{
					Name:       funcName,
					Parameters: params,
					Functions:  baseOverload,
				}, nil
			},
		})
	}
}
