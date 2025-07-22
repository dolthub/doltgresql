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

package functions

import (
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initArrayPosition registers the functions to the catalog.
func initArrayPosition() {
	framework.RegisterFunction(array_position_anyarray_anyelement)
	framework.RegisterFunction(array_position_anyarray_anyelement_int32)
	framework.RegisterFunction(array_positions_anyarray_anyelement)
}

// array_position_anyarray_anyelement represents the PostgreSQL function of the same name, taking the same parameters.
var array_position_anyarray_anyelement = framework.Function2{
	Name:       "array_position",
	Return:     pgtypes.Int32,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.AnyArray, pgtypes.AnyElement},
	Strict:     false,
	Callable: func(ctx *sql.Context, t [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		if val1 == nil {
			return nil, nil
		}

		array := val1.([]any)
		searchElement := val2
		arrayType := t[0]
		baseType := arrayType.ArrayBaseType()

		// Search for the element starting from position 1 (1-indexed)
		for i, element := range array {
			cmp, err := baseType.Compare(ctx, element, searchElement)
			if err != nil {
				return nil, err
			}
			if cmp == 0 {
				return int32(i + 1), nil // PostgreSQL uses 1-indexed arrays
			}
		}

		// Element not found
		return nil, nil
	},
}

// array_position_anyarray_anyelement_int32 represents the PostgreSQL function of the same name, taking the same parameters.
var array_position_anyarray_anyelement_int32 = framework.Function3{
	Name:       "array_position",
	Return:     pgtypes.Int32,
	Parameters: [3]*pgtypes.DoltgresType{pgtypes.AnyArray, pgtypes.AnyElement, pgtypes.Int32},
	Strict:     false,
	Callable: func(ctx *sql.Context, t [4]*pgtypes.DoltgresType, val1 any, val2 any, val3 any) (any, error) {
		if val1 == nil {
			return nil, nil
		}

		array := val1.([]any)
		searchElement := val2
		start := val3.(int32)
		arrayType := t[0]
		baseType := arrayType.ArrayBaseType()

		// Convert 1-indexed start position to 0-indexed
		startIdx := int(start - 1)
		if startIdx < 0 {
			startIdx = 0
		}
		if startIdx >= len(array) {
			return nil, nil
		}

		// Search for the element starting from the specified position
		for i := startIdx; i < len(array); i++ {
			cmp, err := baseType.Compare(ctx, array[i], searchElement)
			if err != nil {
				return nil, err
			}
			if cmp == 0 {
				return int32(i + 1), nil // PostgreSQL uses 1-indexed arrays
			}
		}

		// Element not found
		return nil, nil
	},
}

// array_positions_anyarray_anyelement represents the PostgreSQL function of the same name, taking the same parameters.
var array_positions_anyarray_anyelement = framework.Function2{
	Name:       "array_positions",
	Return:     pgtypes.AnyArray,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.AnyArray, pgtypes.AnyElement},
	Strict:     false,
	Callable: func(ctx *sql.Context, t [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		if val1 == nil {
			return nil, nil
		}

		array := val1.([]any)
		searchElement := val2
		arrayType := t[0]
		baseType := arrayType.ArrayBaseType()
		var positions []any

		// Search for all occurrences of the element
		for i, element := range array {
			cmp, err := baseType.Compare(ctx, element, searchElement)
			if err != nil {
				return nil, err
			}
			if cmp == 0 {
				positions = append(positions, int32(i+1)) // PostgreSQL uses 1-indexed arrays
			}
		}

		// Return array of positions, or empty array if no matches
		return positions, nil
	},
}
