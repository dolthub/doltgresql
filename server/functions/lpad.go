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

// initLpad registers the functions to the catalog.
func initLpad() {
	framework.RegisterFunction(lpad_text_int32)
	framework.RegisterFunction(lpad_text_int32_text)
}

// lpad_text_int32 represents the PostgreSQL function of the same name, taking the same parameters.
var lpad_text_int32 = framework.Function2{
	Name:       "lpad",
	Return:     pgtypes.Text,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Text, pgtypes.Int32},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		var unusedTypes [4]pgtypes.DoltgresType
		return lpad_text_int32_text.Callable(ctx, unusedTypes, val1, val2, " ")
	},
}

// lpad_text_int32_text represents the PostgreSQL function of the same name, taking the same parameters.
var lpad_text_int32_text = framework.Function3{
	Name:       "lpad",
	Return:     pgtypes.Text,
	Parameters: [3]pgtypes.DoltgresType{pgtypes.Text, pgtypes.Int32, pgtypes.Text},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [4]pgtypes.DoltgresType, str any, length any, fill any) (any, error) {
		if length.(int32) <= 0 {
			return "", nil
		}
		runes := []rune(str.(string))
		fillTarget := length.(int32) - int32(len(runes))
		fillRunes := []rune(fill.(string))
		var result []rune
		if fillTarget > 0 {
			if len(fillRunes) > 0 {
				for int32(len(result)) < fillTarget {
					result = append(result, fillRunes...)
				}
			}
			result = result[:fillTarget]
		}
		result = append(result, runes...)
		return string(result[:length.(int32)]), nil
	},
}
