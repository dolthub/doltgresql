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
)

// Overloads is the collection of all overloads for a given function name.
type Overloads struct {
	// ByParamType contains all overloads for the function with this name, indexed by the key of the parameter types.
	ByParamType map[string]FunctionInterface
	// AllOverloads contains all overloads for the function with this name
	AllOverloads []FunctionInterface
}

// NewOverloads creates a new empty overload collection.
func NewOverloads() *Overloads {
	return &Overloads{
		ByParamType:  make(map[string]FunctionInterface),
		AllOverloads: make([]FunctionInterface, 0),
	}
}

// Add adds the given function to the overload collection. Returns an error if the there's a problem with the
// function's declaration.
func (o *Overloads) Add(function FunctionInterface) error {
	key := keyForParamTypes(function.GetParameters())
	if _, ok := o.ByParamType[key]; ok {
		return fmt.Errorf("duplicate function overload for `%s`", function.GetName())
	}

	if function.VariadicIndex() >= 0 {
		varArgsType := function.GetParameters()[function.VariadicIndex()]
		if !varArgsType.IsArrayType() {
			return fmt.Errorf("variadic parameter must be an array type for function `%s`", function.GetName())
		}
	}

	o.ByParamType[key] = function
	o.AllOverloads = append(o.AllOverloads, function)
	return nil
}

// keyForParamTypes returns a string key to match an overload with the given parameter types.
func keyForParamTypes(types []pgtypes.DoltgresType) string {
	sb := strings.Builder{}
	for i, typ := range types {
		if i > 0 {
			sb.WriteByte(',')
		}
		// TODO: check
		sb.WriteString(typ.String())
	}
	return sb.String()
}

// overloadsForParams returns all overloads matching the number of params given, without regard for types.
func (o *Overloads) overloadsForParams(numParams int) []Overload {
	results := make([]Overload, 0, len(o.AllOverloads))
	for _, overload := range o.AllOverloads {
		params := overload.GetParameters()
		variadicIndex := overload.VariadicIndex()
		if variadicIndex >= 0 && len(params) <= numParams {
			// Variadic functions may only match when the function is declared with parameters that are fewer or equal
			// to our target length. If our target length is less, then we cannot expand, so we do not treat it as
			// variadic.
			extendedParams := make([]pgtypes.DoltgresType, numParams)
			copy(extendedParams, params[:variadicIndex])
			// This is copying the parameters after the variadic index, so we need to add 1. We subtract the declared
			// parameter count from the target parameter count to obtain the additional parameter count.
			firstValueAfterVariadic := variadicIndex + 1 + (numParams - len(params))
			copy(extendedParams[firstValueAfterVariadic:], params[variadicIndex+1:])
			// ToArrayType immediately followed by BaseType is a way to get the base type without having to cast.
			// For array types, ToArrayType causes them to return themselves.
			arrType, _ := overload.GetParameters()[variadicIndex].ToArrayType()
			baseType, _ := arrType.ArrayBaseType()
			variadicBaseType := baseType
			for variadicParamIdx := 0; variadicParamIdx < 1+(numParams-len(params)); variadicParamIdx++ {
				extendedParams[variadicParamIdx+variadicIndex] = variadicBaseType
			}
			results = append(results, Overload{
				function:   overload,
				paramTypes: params,
				argTypes:   extendedParams,
				variadic:   variadicIndex,
			})
		} else if len(params) == numParams {
			results = append(results, Overload{
				function:   overload,
				paramTypes: params,
				argTypes:   params,
				variadic:   -1,
			})
		}
	}
	return results
}

// ExactMatchForTypes returns the function that exactly matches the given parameter types, or nil if no overload with
// those types exists.
func (o *Overloads) ExactMatchForTypes(types ...pgtypes.DoltgresType) (FunctionInterface, bool) {
	key := keyForParamTypes(types)
	fn, ok := o.ByParamType[key]
	return fn, ok
}

// Overload is a single overload of a given function, used during evaluation to match the arguments provided
// to a particular overload.
type Overload struct {
	// function is the actual function to call to invoke this overload
	function FunctionInterface
	// paramTypes is the base IDs of the parameters that the function expects
	paramTypes []pgtypes.DoltgresType
	// argTypes is the base IDs of the parameters that the function expects, extended to match the number of args
	// provided in the case of a variadic function.
	argTypes []pgtypes.DoltgresType
	// variadic is the index of the variadic parameter, or -1 if the function is not variadic
	variadic int
}

// coalesceVariadicValues returns a new value set that coalesces all variadic parameters into an array parameter
func (o *Overload) coalesceVariadicValues(returnValues []any) []any {
	// If the overload is not variadic, then we don't need to do anything
	if o.variadic < 0 {
		return returnValues
	}
	coalescedValues := make([]any, len(o.paramTypes))
	copy(coalescedValues, returnValues[:o.variadic])
	// This is for the values after the variadic index, so we need to add 1. We subtract the extended parameter count
	// from the actual parameter count to obtain the additional parameter count.
	firstValueAfterVariadic := o.variadic + 1 + (len(o.argTypes) - len(o.paramTypes))
	copy(coalescedValues[o.variadic+1:], returnValues[firstValueAfterVariadic:])
	// We can just take the relevant slice out of the given values to represent our array, since all arrays use []any
	coalescedValues[o.variadic] = returnValues[o.variadic:firstValueAfterVariadic]
	return coalescedValues
}

// overloadMatch is the result of a successful overload resolution, containing the types of the parameters as well
// as the type cast functions required to convert every argument to its appropriate parameter type
type overloadMatch struct {
	params Overload
	casts  []TypeCastFunction
}

// Valid returns whether this overload is valid (has a callable function)
func (o overloadMatch) Valid() bool {
	return o.params.function != nil
}

// Function returns the function for this overload
func (o overloadMatch) Function() FunctionInterface {
	return o.params.function
}
