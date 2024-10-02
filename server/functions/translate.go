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

// initTranslate registers the functions to the catalog.
func initTranslate() {
	framework.RegisterFunction(translate_text_text_text)
}

// translate_text_text_text represents the PostgreSQL function of the same name, taking the same parameters.
var translate_text_text_text = framework.Function3{
	Name:       "translate",
	Return:     pgtypes.Text,
	Parameters: [3]pgtypes.DoltgresType{pgtypes.Text, pgtypes.Text, pgtypes.Text},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [4]pgtypes.DoltgresType, val1, val2, val3 any) (any, error) {
		str := val1.(string)
		from := []rune(val2.(string))
		to := []rune(val3.(string))
		toLen := len(to)
		fromMap := make(map[string]int)
		for i, l := range from {
			fromMap[string(l)] = i
		}
		var newStr []rune
		for _, l := range str {
			if idx, exists := fromMap[string(l)]; exists {
				if idx < toLen {
					newStr = append(newStr, to[idx])
				}
			} else {
				newStr = append(newStr, l)
			}
		}
		return string(newStr), nil
	},
}
