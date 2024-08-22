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

package functions

import (
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initArrayAppend registers the functions to the catalog.
func initArrayAppend() {
	framework.RegisterFunction(array_append_anyarray_anyelement)
}

// array_append_anyarray_anyelement represents the PostgreSQL function of the same name, taking the same parameters.
var array_append_anyarray_anyelement = framework.Function2{
	Name:       "array_append",
	Return:     pgtypes.AnyArray,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.AnyArray, pgtypes.AnyElement},
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		if val1 == nil {
			return []any{val2}, nil
		}
		array := val1.([]any)
		returnArray := make([]any, len(array)+1)
		copy(returnArray, array)
		returnArray[len(returnArray)-1] = val2
		return returnArray, nil
	},
}
