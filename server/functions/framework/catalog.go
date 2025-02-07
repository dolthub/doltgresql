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

	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/expression/function"

	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// Catalog contains all of the PostgreSQL functions.
var Catalog = map[string][]FunctionInterface{}

// initializedFunctions simply states whether Initialize has been called yet.
var initializedFunctions = false

// RegisterFunction registers the given function, so that it will be usable from a running server. This should be called
// from within an init().
func RegisterFunction(f FunctionInterface) {
	if initializedFunctions {
		// TODO: this should be able to handle overloads
		name := strings.ToLower(f.GetName())
		if err := validateFunction(name, []FunctionInterface{f}); err != nil {
			panic(err) // TODO: replace panics here with errors
		}
		compileNonOperatorFunction(name, []FunctionInterface{f})
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
	case InterpretedFunction:
		name := strings.ToLower(f.ID.FunctionName())
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

	pgtypes.LoadFunctionFromCatalog = getQuickFunctionForTypes
	replaceGmsBuiltIns()
	validateFunctions()
	compileFunctions()
}

// replaceGmsBuiltIns replaces all GMS built-ins that have conflicting names with PostgreSQL functions.
func replaceGmsBuiltIns() {
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
}

// validateFunctions panics if any functions are defined incorrectly or ambiguously.
func validateFunctions() {
	for funcName, overloads := range Catalog {
		if err := validateFunction(funcName, overloads); err != nil {
			panic(err)
		}
	}
}

// validateFunction validates whether functions are defined incorrectly or ambiguously.
func validateFunction(funcName string, overloads []FunctionInterface) error {
	// Verify that each function uses the correct Function overload
	for _, functionOverload := range overloads {
		if functionOverload.GetExpectedParameterCount() >= 0 &&
			len(functionOverload.GetParameters()) != functionOverload.GetExpectedParameterCount() {
			return errors.Errorf("function `%s` should have %d arguments but has %d arguments",
				funcName, functionOverload.GetExpectedParameterCount(), len(functionOverload.GetParameters()))
		}
	}
	// Verify that all overloads are unique
	for functionIndex, f1 := range overloads {
		for _, f2 := range overloads[functionIndex+1:] {
			sameCount := 0
			if f1.GetExpectedParameterCount() == f2.GetExpectedParameterCount() {
				f2Parameters := f2.GetParameters()
				for parameterIndex, f1Parameter := range f1.GetParameters() {
					if f1Parameter.Equals(f2Parameters[parameterIndex]) {
						sameCount++
					}
				}
			}
			if sameCount == f1.GetExpectedParameterCount() && f1.GetExpectedParameterCount() > 0 {
				return errors.Errorf("duplicate function overloads on `%s`", funcName)
			}
		}
	}
	return nil
}

// compileNonOperatorFunction creates a CompiledFunction for each overload of the given function.
func compileNonOperatorFunction(funcName string, overloads []FunctionInterface) {
	overloadTree := NewOverloads()
	for _, functionOverload := range overloads {
		if err := overloadTree.Add(functionOverload); err != nil {
			panic(err)
		}
	}

	// Store the compiled function into the engine's built-in functions
	// TODO: don't do this, use an actual contract for communicating these functions to the engine catalog
	createFunc := func(params ...sql.Expression) (sql.Expression, error) {
		return NewCompiledFunction(funcName, params, overloadTree, false), nil
	}
	function.BuiltIns = append(function.BuiltIns, sql.FunctionN{
		Name: funcName,
		Fn:   createFunc,
	})
	compiledCatalog[funcName] = createFunc
	namedCatalog[funcName] = overloads
}

// compileFunctions creates a CompiledFunction for each overload of each function in the catalog.
func compileFunctions() {
	for funcName, overloads := range Catalog {
		compileNonOperatorFunction(funcName, overloads)
	}

	// Build the overload for all unary and binary functions based on their operator. This will be used for fallback if
	// an exact match is not found. Compiled functions (which wrap the overload deducer) handle upcasting and other
	// special rules, so it's far more efficient to reuse it for operators. Operators are also a special case since they
	// all have different names, while standard overload deducers work on a function-name basis.
	for signature, functionOverload := range unaryFunctions {
		overloads, ok := unaryAggregateOverloads[signature.Operator]
		if !ok {
			overloads = NewOverloads()
			unaryAggregateOverloads[signature.Operator] = overloads
		}
		if err := overloads.Add(functionOverload); err != nil {
			panic(err)
		}
	}

	for signature, functionOverload := range binaryFunctions {
		overloads, ok := binaryAggregateOverloads[signature.Operator]
		if !ok {
			overloads = NewOverloads()
			binaryAggregateOverloads[signature.Operator] = overloads
		}
		if err := overloads.Add(functionOverload); err != nil {
			panic(err)
		}
	}

	// Add all permutations for the unary and binary operators
	for operator, overload := range unaryAggregateOverloads {
		unaryAggregatePermutations[operator] = overload.overloadsForParams(1)
	}
	for operator, overload := range binaryAggregateOverloads {
		binaryAggregatePermutations[operator] = overload.overloadsForParams(2)
	}
}
