// Copyright 2025 Dolthub, Inc.
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

package merge

// ResolveMergeValues is a way to handle merging between "our" value and "their" value. This will always take the
// changed value if one side has changed from the "ancestor" while the other has not. If both have changed (or the
// ancestor does not exist), then this defers to a custom resolution function. This function is only called when both
// "our" and "their" values have changed from the ancestor.
func ResolveMergeValues[T comparable](ourVal, theirVal, ancVal T, hasAncestorValue bool, customResolve func(T, T) T) T {
	if hasAncestorValue {
		if ourVal == ancVal {
			return theirVal
		}
		if theirVal == ancVal {
			return ourVal
		}
	}
	if ourVal == theirVal {
		return ourVal
	}
	return customResolve(ourVal, theirVal)
}

// ResolveMergeValuesVariadic is the same as ResolveMergeValues, except that it will take a variadic custom resolution
// function. This is primarily for values that will use one of the variadic utility functions (Min, Max, etc.) as it
// will always receive two inputs. If Go expands how functions interact with generics, then this function can be removed.
func ResolveMergeValuesVariadic[T comparable](ourVal, theirVal, ancVal T, hasAncestorValue bool, customResolve func(...T) T) T {
	return ResolveMergeValues(ourVal, theirVal, ancVal, hasAncestorValue, func(t1, t2 T) T {
		return customResolve(t1, t2)
	})
}
