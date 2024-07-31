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

// initRtrim registers the functions to the catalog.
func initRtrim() {
	framework.RegisterFunction(rtrim_varchar)
	framework.RegisterFunction(rtrim_varchar_varchar)
}

// rtrim_varchar represents the PostgreSQL function of the same name, taking the same parameters.
var rtrim_varchar = framework.Function1{
	Name:       "rtrim",
	Return:     pgtypes.VarChar,
	Parameters: [1]pgtypes.DoltgresType{pgtypes.VarChar},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, val1 any) (any, error) {
		var unusedTypes [3]pgtypes.DoltgresType
		return rtrim_varchar_varchar.Callable(ctx, unusedTypes, val1, " ")
	},
}

// rtrim_varchar_varchar represents the PostgreSQL function of the same name, taking the same parameters.
var rtrim_varchar_varchar = framework.Function2{
	Name:       "rtrim",
	Return:     pgtypes.VarChar,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.VarChar, pgtypes.VarChar},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, str any, characters any) (any, error) {
		runes := []rune(str.(string))
		trimChars := make(map[rune]struct{})
		for _, c := range characters.(string) {
			trimChars[c] = struct{}{}
		}
		trimIdx := len(runes)
		for ; trimIdx > 0; trimIdx-- {
			if _, ok := trimChars[runes[trimIdx-1]]; !ok {
				break
			}
		}
		return string(runes[:trimIdx]), nil
	},
}
