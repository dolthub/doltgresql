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

import (
	"context"

	"github.com/dolthub/doltgresql/core/rootobject/objinterface"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/dolthub/dolt/go/libraries/doltcore/merge"
)

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

// DiffValues handles common comparisons for diffs. This makes an assumption that "our" and "their" values are always
// valid. Returns true when the diff represents a conflict. When false is returned, the diff's "our" value contains the
// merged value.
func DiffValues[T comparable](diff *objinterface.RootObjectDiff, ourVal, theirVal, ancVal T, hasAncestorValue bool) bool {
	return DiffValuesFunc(diff, ourVal, theirVal, ancVal, hasAncestorValue, func(v1 T, v2 T) bool {
		return v1 == v2
	})
}

// DiffValuesFunc is the same as DiffValues, except that this handles values that are not trivially comparable.
func DiffValuesFunc[T any](diff *objinterface.RootObjectDiff, ourVal, theirVal, ancVal T, hasAncestorValue bool, equals func(T, T) bool) bool {
	// Each check is ordered such that the successive checks rely on the failure of the previous checks
	if equals(ourVal, theirVal) {
		diff.OurValue = ourVal
		diff.TheirValue = theirVal
		diff.AncestorValue = nil
		diff.OurChange = objinterface.RootObjectDiffChange_NoChange
		diff.TheirChange = objinterface.RootObjectDiffChange_NoChange
		return false
	}
	if !hasAncestorValue {
		diff.OurValue = ourVal
		diff.TheirValue = theirVal
		diff.AncestorValue = nil
		diff.OurChange = objinterface.RootObjectDiffChange_Added
		diff.TheirChange = objinterface.RootObjectDiffChange_Added
		return true
	}
	if equals(ourVal, ancVal) {
		diff.OurValue = theirVal
		diff.TheirValue = theirVal
		diff.AncestorValue = ancVal
		diff.OurChange = objinterface.RootObjectDiffChange_NoChange
		diff.TheirChange = objinterface.RootObjectDiffChange_Modified
		return false
	}
	if equals(theirVal, ancVal) {
		diff.OurValue = ourVal
		diff.TheirValue = ourVal
		diff.AncestorValue = ancVal
		diff.OurChange = objinterface.RootObjectDiffChange_Modified
		diff.TheirChange = objinterface.RootObjectDiffChange_NoChange
		return false
	}
	diff.OurValue = ourVal
	diff.TheirValue = theirVal
	diff.AncestorValue = ancVal
	diff.OurChange = objinterface.RootObjectDiffChange_Modified
	diff.TheirChange = objinterface.RootObjectDiffChange_Modified
	return true
}

// CreateConflict handles conflict creation and is declared in a different package. It is assigned here by an Init
// function to get around import cycles.
var CreateConflict = func(ctx context.Context, rightSrc doltdb.Rootish, ours doltdb.RootObject, theirs doltdb.RootObject, ancestor doltdb.RootObject) (doltdb.RootObject, *merge.MergeStats, error) {
	return nil, nil, errors.New("CreateConflict was never initialized")
}
