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

// OverloadDeduction handles resolving which function to call by iterating over the parameter expressions. This also
// handles casting between types if an exact function match is not found.
type OverloadDeduction struct {
	Function  FunctionInterface
	Parameter map[pgtypes.DoltgresTypeBaseID]*OverloadDeduction
}

// Resolve returns an overload that either matches the given parameters exactly, or is a viable match after casting.
// This will modify the parameter slice in-place. Returns a nil OverloadDeduction if a viable match is not found.
func (overload *OverloadDeduction) Resolve(parameters []pgtypes.DoltgresType) (*OverloadDeduction, []TypeCastFunction, error) {
	// Create a slice of types that will be modified in-place to contain the resulting types that the function requires.
	resultTypes := make([]pgtypes.DoltgresType, len(parameters))
	copy(resultTypes, parameters)
	// Call the recursive type resolver
	resultOverload := overload.resolveByType(parameters, resultTypes)
	// If we receive a nil overload, then no valid overloads were found
	if resultOverload == nil {
		return nil, nil, nil
	}
	// If any of the result types are different from their originals, then we need to cast them to their resulting types
	// if it's possible.
	casts := make([]TypeCastFunction, len(parameters))
	for i, t := range resultTypes {
		if parameters[i].Equals(t) {
			continue
		}
		casts[i] = GetCast(parameters[i].BaseID(), resultTypes[i].BaseID())
	}
	return resultOverload, casts, nil
}

// resolveByType returns the best matching overload for the given types. The result types represent the actual types
// used by the overload, which may differ from the calling types. It is up to the caller to cast the parameters to match
// the types expected by the returned overload. Returns a nil OverloadDeduction if a viable match is not found.
func (overload *OverloadDeduction) resolveByType(originalTypes []pgtypes.DoltgresType, resultTypes []pgtypes.DoltgresType) *OverloadDeduction {
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
	resultOverload := overload.Parameter[t.BaseID()].resolveByType(originalTypes[1:], resultTypes[1:])
	if resultOverload != nil {
		resultTypes[0] = t
		return resultOverload
	}

	// We did not find a resolution for the original type, so we'll look through each cast
	for _, cast := range GetPotentialCasts(t.BaseID()) {
		resultOverload = overload.Parameter[cast.BaseID()].resolveByType(originalTypes[1:], resultTypes[1:])
		if resultOverload != nil {
			resultTypes[0] = cast
			return resultOverload
		}
	}
	// We did not find any potential matches, so we'll return nil
	return nil
}
