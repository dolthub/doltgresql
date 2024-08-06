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

// Add adds the given function to the overload collection. Returns false if the new overload collides with an existing
// overload.
func (overloads *Overloads) Add(function FunctionInterface) bool {
	key := keyForParamTypes(function.GetParameters())
	if _, ok := overloads.ByParamType[key]; ok {
		return false
	}
	overloads.ByParamType[key] = function
	overloads.AllOverloads = append(overloads.AllOverloads, function)
	return true
}

// keyForParamTypes returns a string key that may be used in the `ExactMatches` field to find the desired overload.
func keyForParamTypes(types []pgtypes.DoltgresType) string {
	sb := strings.Builder{}
	for i, typ := range types {
		if i > 0 {
			sb.WriteByte(',')
		}
		sb.WriteString(typ.BaseID().String())
	}
	return sb.String()
}

// keyForParamTypes returns a string key that may be used in the `ExactMatches` field to find the desired overload.
func keyForBaseIds(types []pgtypes.DoltgresTypeBaseID) string {
	sb := strings.Builder{}
	for i, typ := range types {
		if i > 0 {
			sb.WriteByte(',')
		}
		sb.WriteString(typ.String())
	}
	return sb.String()
}

func baseIdsFortypes(types []pgtypes.DoltgresType) []pgtypes.DoltgresTypeBaseID {
	baseIds := make([]pgtypes.DoltgresTypeBaseID, len(types))
	for i, t := range types {
		baseIds[i] = t.BaseID()
	}
	return baseIds
}

// overloadsForParams returns all overloads matching the number of params given, without regard for types.
func (overloads *Overloads) overloadsForParams(numParams int) []functionOverload {
	extended := make([]functionOverload, len(overloads.AllOverloads))
	for permutationIdx, permutation := range overloads.AllOverloads {
		params := baseIdsFortypes(permutation.GetParameters())
		variadicIndex := permutation.VariadicIndex()
		if variadicIndex >= 0 && len(params) <= numParams {
			// Variadic functions may only match when the function is declared with parameters that are fewer or equal
			// to our target length. If our target length is less, then we cannot expand, so we do not treat it as
			// variadic.
			extendedParams := make([]pgtypes.DoltgresTypeBaseID, numParams)
			copy(extendedParams, params[:variadicIndex])
			// This is copying the parameters after the variadic index, so we need to add 1. We subtract the declared
			// parameter count from the target parameter count to obtain the additional parameter count.
			firstValueAfterVariadic := variadicIndex + 1 + (numParams - len(params))
			copy(extendedParams[firstValueAfterVariadic:], params[variadicIndex+1:])
			// ToArrayType immediately followed by BaseType is a way to get the base type without having to cast.
			// For array types, ToArrayType causes them to return themselves.
			variadicBaseType := permutation.GetParameters()[variadicIndex].ToArrayType().BaseType().BaseID()
			for variadicParamIdx := 0; variadicParamIdx < 1+(numParams-len(params)); variadicParamIdx++ {
				extendedParams[variadicParamIdx+variadicIndex] = variadicBaseType
			}
			extended[permutationIdx] = functionOverload{
				function:   permutation,
				paramTypes: params,
				argTypes:   extendedParams,
				variadic:   variadicIndex,
			}
		} else {
			extended[permutationIdx] = functionOverload{
				function:   permutation,
				paramTypes: params,
				argTypes:   params,
				variadic:   -1,
			}
		}
	}
	return extended
}

// ExactMatchForTypes returns the function that exactly matches the given parameter types, or nil if no overload with
// those types exists.
func (overloads *Overloads) ExactMatchForTypes(types []pgtypes.DoltgresType) (FunctionInterface, bool) {
	key := keyForParamTypes(types)
	fn, ok := overloads.ByParamType[key]
	return fn, ok
}

// ExactMatchForBaseIds returns the function that exactly matches the given parameter types, or nil if no overload with
// those types exists.
func (overloads *Overloads) ExactMatchForBaseIds(types ...pgtypes.DoltgresTypeBaseID) (FunctionInterface, bool) {
	key := keyForBaseIds(types)
	fn, ok := overloads.ByParamType[key]
	return fn, ok
}

// functionOverload is a single overload of a given function, used during evaluation to match the arguments provided
// to a particular overload.
type functionOverload struct {
	// function is the actual function to call to invoke this overload
	function FunctionInterface
	// paramTypes is the base IDs of the parameters that the function expects
	paramTypes []pgtypes.DoltgresTypeBaseID
	// argTypes is the base IDs of the parameters that the function expects, extended to match the number of args
	// provided in the case of a variadic function.
	argTypes []pgtypes.DoltgresTypeBaseID
	// variadic is the index of the variadic parameter, or -1 if the function is not variadic
	variadic int
}

// coalesceVariadicValues returns a new value set that coalesces all variadic parameters into an array parameter
func (p *functionOverload) coalesceVariadicValues(returnValues []any) []any {
	// If the overload is not variadic, then we don't need to do anything
	if p.variadic < 0 {
		return returnValues
	}
	coalescedValues := make([]any, len(p.paramTypes))
	copy(coalescedValues, returnValues[:p.variadic])
	// This is for the values after the variadic index, so we need to add 1. We subtract the extended parameter count
	// from the actual parameter count to obtain the additional parameter count.
	firstValueAfterVariadic := p.variadic + 1 + (len(p.argTypes) - len(p.paramTypes))
	copy(coalescedValues[p.variadic+1:], returnValues[firstValueAfterVariadic:])
	// We can just take the relevant slice out of the given values to represent our array, since all arrays use []any
	coalescedValues[p.variadic] = returnValues[p.variadic:firstValueAfterVariadic]
	return coalescedValues
}
