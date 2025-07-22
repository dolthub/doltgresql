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

// initArrayCat registers the functions to the catalog.
func initArrayCat() {
	framework.RegisterFunction(array_cat_anyarray_anyarray)
}

// array_cat_anyarray_anyarray represents the PostgreSQL function of the same name, taking the same parameters.
var array_cat_anyarray_anyarray = framework.Function2{
	Name:       "array_cat",
	Return:     pgtypes.AnyArray,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.AnyArray, pgtypes.AnyArray},
	Strict:     false,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		if val1 == nil && val2 == nil {
			return nil, nil
		} else if val1 == nil {
			return val2, nil
		} else if val2 == nil {
			return val1, nil
		}

		array1 := val1.([]any)
		array2 := val2.([]any)

		// Concatenate the arrays
		result := make([]any, len(array1)+len(array2))
		copy(result, array1)
		copy(result[len(array1):], array2)

		return result, nil
	},
}
