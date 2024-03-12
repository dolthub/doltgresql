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
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// init registers the functions to the catalog.
func init() {
	framework.RegisterFunction(rpad_varchar_int32)
	framework.RegisterFunction(rpad_varchar_int32_varchar)
}

// rpad_varchar_int32 represents the PostgreSQL function of the same name, taking the same parameters.
var rpad_varchar_int32 = framework.Function2{
	Name:       "rpad",
	Return:     pgtypes.VarCharMax,
	Parameters: []pgtypes.DoltgresType{pgtypes.VarCharMax, pgtypes.Int32},
	Callable: func(ctx framework.Context, val1 any, val2 any) (any, error) {
		return rpad_varchar_int32_varchar.Callable(framework.Context{
			Context:       ctx.Context,
			OriginalTypes: append(ctx.OriginalTypes, pgtypes.VarCharMax),
			Sources:       append(ctx.Sources, framework.Source_Constant),
		}, val1, val2, " ")
	},
}

// rpad_varchar_int32_varchar represents the PostgreSQL function of the same name, taking the same parameters.
var rpad_varchar_int32_varchar = framework.Function3{
	Name:       "rpad",
	Return:     pgtypes.VarCharMax,
	Parameters: []pgtypes.DoltgresType{pgtypes.VarCharMax, pgtypes.Int32, pgtypes.VarCharMax},
	Callable: func(ctx framework.Context, str any, length any, fill any) (any, error) {
		if str == nil || length == nil || fill == nil {
			return nil, nil
		}
		if length.(int32) <= 0 {
			return "", nil
		}
		runes := []rune(str.(string))
		fillRunes := []rune(fill.(string))
		for int32(len(runes)) < length.(int32) {
			runes = append(runes, fillRunes...)
		}
		return string(runes[:length.(int32)]), nil
	},
}
