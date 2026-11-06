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

// initRight registers the functions to the catalog.
func initRight() {
	framework.RegisterFunction(right_text_int32)
}

// right_text_int32 represents the PostgreSQL function of the same name, taking the same parameters.
var right_text_int32 = framework.Function2{
	Name:       "right",
	Return:     pgtypes.Text,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Text, pgtypes.Int32},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, strInt any, nInt any) (any, error) {
		str, err := framework.UnwrapString(ctx, strInt)
		if err != nil {
			return nil, err
		}
		runes := []rune(str)
		n := int(nInt.(int32))
		if n >= 0 {
			if len(runes)-n < 0 {
				return str, nil
			}
			return string(runes[len(runes)-n:]), nil
		} else {
			if -n > len(runes) {
				return "", nil
			}
			return string(runes[-n:]), nil
		}
	},
}
