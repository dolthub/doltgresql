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
	"sort"

	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// OverloadDeduction handles resolving which function to call by iterating over the parameter expressions. This also
// handles casting between types if an exact function match is not found.
type OverloadDeduction struct {
	Function  FunctionInterface
	Parameter map[pgtypes.DoltgresTypeBaseID]*OverloadDeduction
}

// Resolve returns an overload that either matches the given parameters exactly, or is a viable match after casting.
// Returns a nil OverloadDeduction if a viable match is not found.
func (overload *OverloadDeduction) Resolve(parameters []pgtypes.DoltgresType, sources []Source) (*OverloadDeduction, []TypeCastFunction, error) {
	// Call the recursive type resolver
	resultOverload := overload.resolveByType(parameters, sources)
	// If we receive a nil overload, then no valid overloads were found
	if resultOverload == nil {
		return nil, nil, nil
	}
	// If any of the result types are different from their originals, then we need to cast them to their resulting types
	// if it's possible.
	casts := make([]TypeCastFunction, len(parameters))
	for i, resultType := range resultOverload.Function.GetParameters() {
		casts[i] = GetExplicitCast(parameters[i].BaseID(), resultType.BaseID())
	}
	return resultOverload, casts, nil
}

// resolveByType returns the best matching overload for the given types. The result types represent the actual types
// used by the overload, which may differ from the calling types. It is up to the caller to cast the parameters to match
// the types expected by the returned overload. Returns a nil OverloadDeduction if a viable match is not found.
func (overload *OverloadDeduction) resolveByType(originalTypes []pgtypes.DoltgresType, sources []Source) *OverloadDeduction {
	if overload == nil {
		return nil
	}
	if len(originalTypes) == 0 {
		if overload.Function != nil {
			return overload
		}
		return nil
	}

	// Check if we're able to resolve the original type
	t := originalTypes[0]
	resultOverload := overload.Parameter[t.BaseID()].resolveByType(originalTypes[1:], sources[1:])
	if resultOverload != nil {
		return resultOverload
	}

	// We did not find a resolution for the original type, so we'll look through each type to find a possible cast.
	// Constants have a different set of considerations compared to other types of expressions, and string constants
	// are further specialized.
	var castFunc func(pgtypes.DoltgresTypeBaseID, pgtypes.DoltgresTypeBaseID) bool
	sourceStringLiteral := false
	if sources[0] == Source_Constant {
		castFunc = implicitOverloadCasts
		switch t.BaseID() {
		case pgtypes.DoltgresTypeBaseID_Char, pgtypes.DoltgresTypeBaseID_Text, pgtypes.DoltgresTypeBaseID_VarChar:
			sourceStringLiteral = true
		}
	} else {
		castFunc = numericUpcasts
	}
	for _, priority := range overload.castPriority(sourceStringLiteral) {
		if castFunc(priority, t.BaseID()) {
			resultOverload = overload.Parameter[priority].resolveByType(originalTypes[1:], sources[1:])
			if resultOverload != nil {
				return resultOverload
			}
		}
	}

	// We did not find any potential matches, so we'll return nil
	return nil
}

// castPriority returns the available types for the current overload position. These types are ordered by priority,
// which we try to match to the observed behavior of Postgres. The priorities are slightly different if we're casting
// from a string literal.
func (overload *OverloadDeduction) castPriority(sourceStringLiteral bool) []pgtypes.DoltgresTypeBaseID {
	// TODO: this should be precalculated during the overload construction
	types := make([]pgtypes.DoltgresTypeBaseID, len(overload.Parameter))
	idx := 0
	for k := range overload.Parameter {
		types[idx] = k
		idx++
	}
	sort.Slice(types, func(i, j int) bool {
		return castPriorityForType(types[i], sourceStringLiteral) < castPriorityForType(types[j], sourceStringLiteral)
	})
	return types
}
