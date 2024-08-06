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

// TODO: doc
type overloadParamPermutation struct {
	function   FunctionInterface
	paramTypes []pgtypes.DoltgresTypeBaseID
	argTypes   []pgtypes.DoltgresTypeBaseID
	variadic   int
}

// coalesceVariadicValues returns a new value set that coalesces all variadic parameters into their actual array parameter.
func (p *overloadParamPermutation) coalesceVariadicValues(values []any) []any {
	// If the overload is not variadic, then we don't need to do anything
	if p.variadic < 0 {
		return values
	}
	coalescedValues := make([]any, len(p.paramTypes))
	copy(coalescedValues, values[:p.variadic])
	// This is for the values after the variadic index, so we need to add 1. We subtract the extended parameter count
	// from the actual parameter count to obtain the additional parameter count.
	firstValueAfterVariadic := p.variadic + 1 + (len(p.argTypes) - len(p.paramTypes))
	copy(coalescedValues[p.variadic+1:], values[firstValueAfterVariadic:])
	// We can just take the relevant slice out of the given values to represent our array, since all arrays use []any
	coalescedValues[p.variadic] = values[p.variadic:firstValueAfterVariadic]
	return coalescedValues
}

// Overloads represents the collection of all valid overloads for a function.
type Overloads struct {
	// ExactMatches uses the concatenation of the parameters' base IDs to construct an exact lookup key for each
	// overload. This is only used by a specific step during function resolution, and therefore does not apply to other
	// steps (such as resolving the `unknown` type, etc.).
	ExactMatches map[string]FunctionInterface
	// Permutations contains all of the valid permutations for the function.
	Permutations []FunctionInterface
}

// Add adds the given function to the overload collection. Returns false if the new overload collides with an existing
// overload.
func (overloads *Overloads) Add(function FunctionInterface) bool {
	key := keyForParamTypes(function.GetParameters())
	if _, ok := overloads.ExactMatches[key]; ok {
		return false
	}
	overloads.ExactMatches[key] = function
	overloads.Permutations = append(overloads.Permutations, function)
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

func baseIdsFortypes(types []pgtypes.DoltgresType) []pgtypes.DoltgresTypeBaseID {
	baseIds := make([]pgtypes.DoltgresTypeBaseID, len(types))
	for i, t := range types {
		baseIds[i] = t.BaseID()
	}
	return baseIds
}

// expandParameters returns all of the permutations, while using the given parameter length to determine the length
// target for variadic functions.
func (overloads *Overloads) expandParameters(paramLength int) []overloadParamPermutation {
	extended := make([]overloadParamPermutation, len(overloads.Permutations))
	for permutationIdx, permutation := range overloads.Permutations {
		params := baseIdsFortypes(permutation.GetParameters())
		variadicIndex := -1 // permutation.VariadicIndex()
		if variadicIndex >= 0 && len(params) <= paramLength {
			// Variadic functions may only match when the function is declared with parameters that are fewer or equal
			// to our target length. If our target length is less, then we cannot expand, so we do not treat it as
			// variadic.
			extendedParams := make([]pgtypes.DoltgresTypeBaseID, paramLength)
			copy(extendedParams, params[:variadicIndex])
			// This is copying the parameters after the variadic index, so we need to add 1. We subtract the declared
			// parameter count from the target parameter count to obtain the additional parameter count.
			firstValueAfterVariadic := variadicIndex + 1 + (paramLength - len(params))
			copy(extendedParams[firstValueAfterVariadic:], params[variadicIndex+1:])
			// ToArrayType immediately followed by BaseType is a way to get the base type without having to cast.
			// For array types, ToArrayType causes them to return themselves.
			variadicBaseType := permutation.GetParameters()[variadicIndex].ToArrayType().BaseType().BaseID()
			for variadicParamIdx := 0; variadicParamIdx < 1+(paramLength-len(params)); variadicParamIdx++ {
				extendedParams[variadicParamIdx+variadicIndex] = variadicBaseType
			}
			extended[permutationIdx] = overloadParamPermutation{
				function:   permutation,
				paramTypes: params,
				argTypes:   params,
				variadic:   variadicIndex,
			}
		} else {
			extended[permutationIdx] = overloadParamPermutation{
				function:   permutation,
				paramTypes: params,
				argTypes:   params,
				variadic:   -1,
			}
		}
	}
	return extended
}

func (overloads *Overloads) ExactMatchForTypes(types []pgtypes.DoltgresType) (FunctionInterface, bool) {
	key := keyForParamTypes(types)
	fn, ok := overloads.ExactMatches[key]
	return fn, ok
}
