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

// FunctionOverloadTree is a type tree used to resolve which overload of a given function to apply to a given
// parameter list. Each node in the tree represents a parameter in the function signature, and the leaves represent
// the function to call. Every node points to the set of possible next nodes via the type of the next
// expected parameter.
//
// Vararg functions are a special case: they are represented as a single node with the VarargType field set to the type
// of every argument.
type FunctionOverloadTree struct {
	// Function is the function to call for this overload (nil for non-leaf nodes)
	Function FunctionInterface
	// NextParam is the set of possible next nodes, keyed by the type of the next parameter.
	NextParam map[pgtypes.DoltgresTypeBaseID]*FunctionOverloadTree
	// Variadic is whether this node is variadic, which means that this node consumes all arguments with the last type.
	Variadic bool
}

// TODO: doc
type extendedParamPermutation struct {
	function   FunctionInterface
	paramTypes []pgtypes.DoltgresTypeBaseID
	argTypes   []pgtypes.DoltgresTypeBaseID
	variadic   int
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
func (overloads *Overloads) expandParameters(paramLength int) []extendedParamPermutation {
	extended := make([]extendedParamPermutation, len(overloads.Permutations))
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
			extended[permutationIdx] = extendedParamPermutation{
				function:   permutation,
				paramTypes: params,
				argTypes:   params,
				variadic:   variadicIndex,
			}
		} else {
			extended[permutationIdx] = extendedParamPermutation{
				function:   permutation,
				paramTypes: params,
				argTypes:   params,
				variadic:   -1,
			}
		}
	}
	return extended
}

type overloadParamPermutation struct {
	paramTypes []pgtypes.DoltgresTypeBaseID
	variadic   bool
}

func newOverloadParamPermutation(paramTypes []pgtypes.DoltgresTypeBaseID, variadic bool) overloadParamPermutation {
	return overloadParamPermutation{paramTypes, variadic}
}

func (opp overloadParamPermutation) copy() overloadParamPermutation {
	cpy := newOverloadParamPermutation(make([]pgtypes.DoltgresTypeBaseID, len(opp.paramTypes)), opp.variadic)
	copy(cpy.paramTypes, opp.paramTypes)
	return cpy
}

// collectOverloadPermutations collects all parameters, starting from the caller, such that we have a collection of
// slices containing all possible parameter combinations that lead to functions. For example, let's say we have the
// following function overloads:
//
// example(int4, int4)
//
// example(text, int8, int8)
//
// This would return two slices. The first would contain [int4, int4] while the second would contain [text, int8, int8].
func (overload *FunctionOverloadTree) collectOverloadPermutations() []overloadParamPermutation {
	var permutations []overloadParamPermutation
	overload.traverseOverloadTree(newOverloadParamPermutation(nil, false), &permutations)
	return permutations
}

// traverseOverloadTree walks the tree of overloads, persisting any paths that resolve to a function.
func (overload *FunctionOverloadTree) traverseOverloadTree(currentPermutation overloadParamPermutation, permutations *[]overloadParamPermutation) {
	// If we've hit a function, then we should persist the progress we've made so far
	if overload.Function != nil {
		perm := currentPermutation.copy()
		perm.variadic = overload.Variadic
		*permutations = append(*permutations, perm)
	}
	// Continue to walk the tree
	for baseID, child := range overload.NextParam {
		nextPermutation := newOverloadParamPermutation(append(currentPermutation.paramTypes, baseID), overload.Variadic)
		child.traverseOverloadTree(nextPermutation, permutations)
	}
}

// ExactMatch returns the function that exactly matches the given parameter types. If no exact match is found, then
// nil, false is returned.
func (overload *FunctionOverloadTree) ExactMatch(argTypes []pgtypes.DoltgresTypeBaseID) (FunctionInterface, bool) {
	// Base case: this is a leaf node and we're out of args
	if overload.Function != nil && len(argTypes) == 0 {
		if len(argTypes) == 0 {
			return overload.Function, true
		}
		return nil, false
	}

	for _, argType := range argTypes {
		nextNode, nextParamTypeMatches := overload.NextParam[argType]
		if !nextParamTypeMatches {
			continue
		}

		// If the next node is variadic, match the rest of the arguments with the current param type
		if nextNode.Variadic {
			// keep consuming the remainder of the varags
			for _, varargType := range argTypes[1:] {
				if _, ok := overload.NextParam[varargType]; !ok {
					return nil, false
				}
				return nextNode.Function, true
			}
		}

		// Otherwise, look for a match for the rest of the args
		matchingFunc, foundMatch := nextNode.ExactMatch(argTypes[1:])
		if foundMatch {
			return matchingFunc, true
		}
	}

	return nil, false
}

// ExactMatchForTypes returns the function that exactly matches the given parameter types. If no exact match is found, then
// nil, false is returned.
func (overload *FunctionOverloadTree) ExactMatchForTypes(paramTypes []pgtypes.DoltgresType) (FunctionInterface, bool) {
	baseTypeIds := make([]pgtypes.DoltgresTypeBaseID, len(paramTypes))
	for i, paramType := range paramTypes {
		baseTypeIds[i] = paramType.BaseID()
	}

	return overload.ExactMatch(baseTypeIds)
}
