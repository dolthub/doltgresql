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

// initLtrim registers the functions to the catalog.
func initLtrim() {
	framework.RegisterFunction(ltrim_varchar)
	framework.RegisterFunction(ltrim_varchar_varchar)
}

// ltrim_varchar represents the PostgreSQL function of the same name, taking the same parameters.
var ltrim_varchar = framework.Function1{
	Name:       "ltrim",
	Return:     pgtypes.VarChar,
	Parameters: []pgtypes.DoltgresType{pgtypes.VarChar},
	Callable: func(ctx *sql.Context, val1 any) (any, error) {
		return ltrim_varchar_varchar.Callable(ctx, val1, " ")
	},
	Strict: true,
}

// ltrim_varchar_varchar represents the PostgreSQL function of the same name, taking the same parameters.
var ltrim_varchar_varchar = framework.Function2{
	Name:       "ltrim",
	Return:     pgtypes.VarChar,
	Parameters: []pgtypes.DoltgresType{pgtypes.VarChar, pgtypes.VarChar},
	Callable: func(ctx *sql.Context, str any, characters any) (any, error) {
		runes := []rune(str.(string))
		trimChars := make(map[rune]struct{})
		for _, c := range characters.(string) {
			trimChars[c] = struct{}{}
		}
		trimIdx := 0
		for ; trimIdx < len(runes); trimIdx++ {
			if _, ok := trimChars[runes[trimIdx]]; !ok {
				break
			}
		}
		return string(runes[trimIdx:]), nil
	},
	Strict: true,
}
