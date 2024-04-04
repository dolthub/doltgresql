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
	framework.RegisterFunction(btrim_varchar_varchar)
}

// btrim_varchar_varchar represents the PostgreSQL function of the same name, taking the same parameters.
var btrim_varchar_varchar = framework.Function2{
	Name:       "btrim",
	Return:     pgtypes.VarChar,
	Parameters: []pgtypes.DoltgresType{pgtypes.VarChar, pgtypes.VarChar},
	Callable: func(ctx framework.Context, str any, characters any) (any, error) {
		if str == nil || characters == nil {
			return nil, nil
		}
		result, err := ltrim_varchar_varchar.Callable(ctx, str, characters)
		if err != nil {
			return nil, err
		}
		return rtrim_varchar_varchar.Callable(ctx, result, characters)
	},
}
