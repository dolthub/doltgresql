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

// initLeft registers the functions to the catalog.
func initLeft() {
	framework.RegisterFunction(left_varchar_int32)
}

// left_varchar_int32 represents the PostgreSQL function of the same name, taking the same parameters.
var left_varchar_int32 = framework.Function2{
	Name:       "left",
	Return:     pgtypes.VarChar,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.VarChar, pgtypes.Int32},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, strInt any, nInt any) (any, error) {
		str := strInt.(string)
		n := nInt.(int32)
		if n >= 0 {
			if int(n) > len(str) {
				return str, nil
			}
			return str[:n], nil
		} else {
			if len(str)+int(n) <= 0 {
				return "", nil
			}
			return str[:len(str)+int(n)], nil
		}
	},
}
