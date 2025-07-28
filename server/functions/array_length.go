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

// initArrayLength registers the functions to the catalog.
func initArrayLength() {
	framework.RegisterFunction(array_length_anyarray_int32)
}

// array_length_anyarray_int32 represents the PostgreSQL function of the same name, taking the same parameters.
var array_length_anyarray_int32 = framework.Function2{
	Name:       "array_length",
	Return:     pgtypes.Int32,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.AnyArray, pgtypes.Int32},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		array := val1.([]any)
		dimension := val2.(int32)

		// PostgreSQL arrays are 1-dimensional in this implementation
		// Dimension 0 is invalid, dimensions > 1 return null for 1D arrays
		if dimension <= 0 {
			return nil, nil
		}

		if dimension == 1 {
			if len(array) == 0 {
				return nil, nil
			}
			return int32(len(array)), nil
		}

		// For dimensions other than 1, return null (multi-dimensional arrays not fully supported)
		return nil, nil
	},
}
