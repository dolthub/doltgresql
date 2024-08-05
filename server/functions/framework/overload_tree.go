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
	// Variadic is whether this node is variadic, which means that the
	Variadic bool
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

func (opp overloadParamPermutation) hasVariadic() bool {
	return opp.variadic
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
		*permutations = append(*permutations, currentPermutation.copy())
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
