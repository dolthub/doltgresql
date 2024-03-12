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
	"strings"

	"github.com/dolthub/doltgresql/server/functions/framework"

	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// init registers the functions to the catalog.
func init() {
	framework.RegisterFunction(repeat_varchar_int32)
}

// repeat_varchar_int32 represents the PostgreSQL function of the same name, taking the same parameters.
var repeat_varchar_int32 = framework.Function2{
	Name:       "repeat",
	Return:     pgtypes.VarCharMax,
	Parameters: []pgtypes.DoltgresType{pgtypes.VarCharMax, pgtypes.Int32},
	Callable: func(ctx framework.Context, str any, num any) (any, error) {
		if str == nil || num == nil {
			return nil, nil
		}
		return strings.Repeat(str.(string), int(num.(int32))), nil
	},
}
