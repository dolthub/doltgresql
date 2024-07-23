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
	// VarargType is the type of the vararg parameter, if this overload is a vararg overload.
	VarargType pgtypes.DoltgresTypeBaseID
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
func (overload *FunctionOverloadTree) collectOverloadPermutations() [][]pgtypes.DoltgresTypeBaseID {
	var permutations [][]pgtypes.DoltgresTypeBaseID
	overload.traverseOverloadTree([]pgtypes.DoltgresTypeBaseID{}, &permutations)
	return permutations
}

// traverseOverloadTree walks the tree of overloads, persisting any paths that resolve to a function.
func (overload *FunctionOverloadTree) traverseOverloadTree(currentPermutation []pgtypes.DoltgresTypeBaseID, permutations *[][]pgtypes.DoltgresTypeBaseID) {
	// If we've hit a function, then we should persist the progress we've made so far
	if overload.Function != nil {
		pathCopy := make([]pgtypes.DoltgresTypeBaseID, len(currentPermutation))
		copy(pathCopy, currentPermutation)
		*permutations = append(*permutations, pathCopy)
	}
	// Continue to walk the tree
	for baseID, child := range overload.NextParam {
		child.traverseOverloadTree(append(currentPermutation, baseID), permutations)
	}
}

// ExactMatch returns the function that exactly matches the given parameter types. If no exact match is found, then
// nil, false is returned.
func (overload *FunctionOverloadTree) ExactMatch(paramTypes []pgtypes.DoltgresTypeBaseID) (FunctionInterface, bool) {
	if overload.Function != nil {
		if len(paramTypes) == 0 {
			return overload.Function, true
		}
		return nil, false
	}

	for _, paramType := range paramTypes {
		if nextNode, ok := overload.NextParam[paramType]; ok {
			return nextNode.ExactMatch(paramTypes[1:])
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
