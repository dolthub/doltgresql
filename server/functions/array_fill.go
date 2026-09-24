// Copyright 2026 Dolthub, Inc.
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

	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

func initArrayFill() {
	framework.RegisterFunction(array_fill_two)
	framework.RegisterFunction(array_fill_three)
}

var array_fill_two = framework.Function2{
	Name:       "array_fill",
	Return:     pgtypes.AnyArray,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.AnyElement, pgtypes.Int32Array},
	Callable: func(ctx *sql.Context, t [3]*pgtypes.DoltgresType, val, dims any) (any, error) {
		return fillArray(val, dims, nil, false)
	},
}
var array_fill_three = framework.Function3{
	Name:       "array_fill",
	Return:     pgtypes.AnyArray,
	Parameters: [3]*pgtypes.DoltgresType{pgtypes.AnyElement, pgtypes.Int32Array, pgtypes.Int32Array},
	Callable: func(ctx *sql.Context, t [4]*pgtypes.DoltgresType, val, dims, bounds any) (any, error) {
		return fillArray(val, dims, bounds, true)
	},
}

// fillArray validates the dimensions before allocating the rectangular result. Non-default lower
// bounds cannot yet be persisted by the array representation.
func fillArray(val, dimensions, lowerBounds any, explicitBounds bool) (any, error) {
	if dimensions == nil || explicitBounds && lowerBounds == nil {
		return nil, pgerror.New(pgcode.NullValueNotAllowed, "dimension array or low bound array cannot be null")
	}
	dimsInput := dimensions.([]any)
	if len(pgtypes.ArrayDims(dimsInput, pgtypes.Int32)) > 1 {
		return nil, pgerror.New(pgcode.ArraySubscript, "wrong number of array subscripts")
	}
	if len(dimsInput) > 6 {
		return nil, pgerror.Newf(pgcode.ProgramLimitExceeded, "number of array dimensions (%d) exceeds the maximum allowed (6)", len(dimsInput))
	}
	dims := make([]int32, len(dimsInput))
	for i, d := range dimsInput {
		if d == nil {
			return nil, pgerror.New(pgcode.NullValueNotAllowed, "dimension values cannot be null")
		}
		dims[i] = d.(int32)
	}
	if explicitBounds {
		bounds := lowerBounds.([]any)
		if len(pgtypes.ArrayDims(bounds, pgtypes.Int32)) > 1 || len(bounds) != len(dims) {
			return nil, pgerror.New(pgcode.ArraySubscript, "wrong number of array subscripts")
		}
		for _, b := range bounds {
			if b == nil {
				return nil, pgerror.New(pgcode.NullValueNotAllowed, "lower bound values cannot be null")
			}
			if b.(int32) != 1 {
				return nil, pgerror.New(pgcode.FeatureNotSupported, "non-default array lower bounds are not yet supported")
			}
		}
	}
	count := int64(1)
	if len(dims) == 0 {
		count = 0
	}
	for _, d := range dims {
		if d < 0 || count*int64(d) > 134217727 {
			return nil, pgerror.New(pgcode.ProgramLimitExceeded, "array size exceeds the maximum allowed (134217727)")
		}
		count *= int64(d)
	}
	result := make([]any, int(count))
	for i := range result {
		result[i] = val
	}
	return pgtypes.InflateArray(result, dims), nil
}
