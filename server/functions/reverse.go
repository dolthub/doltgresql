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
	framework.RegisterFunction(reverse_varchar)
}

// reverse_varchar represents the PostgreSQL function of the same name, taking the same parameters.
var reverse_varchar = framework.Function1{
	Name:       "reverse",
	Return:     pgtypes.VarCharMax,
	Parameters: []pgtypes.DoltgresType{pgtypes.VarCharMax},
	Callable: func(ctx framework.Context, val1 any) (any, error) {
		if val1 == nil {
			return nil, nil
		}
		runes := []rune(val1.(string))
		for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
			runes[i], runes[j] = runes[j], runes[i]
		}
		return string(runes), nil
	},
}
