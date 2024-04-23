// Copyright 2023 Dolthub, Inc.
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
	"github.com/dolthub/doltgresql/utils"
)

// initGcd registers the functions to the catalog.
func initGcd() {
	framework.RegisterFunction(gcd_int64_int64)
}

// gcd_int64_int64 represents the PostgreSQL function of the same name, taking the same parameters.
var gcd_int64_int64 = framework.Function2{
	Name:       "gcd",
	Return:     pgtypes.Int64,
	Parameters: []pgtypes.DoltgresType{pgtypes.Int64, pgtypes.Int64},
	Callable: func(ctx framework.Context, val1Interface any, val2Interface any) (any, error) {
		if val1Interface == nil || val2Interface == nil {
			return nil, nil
		}
		val1 := val1Interface.(int64)
		val2 := val2Interface.(int64)
		for val2 != 0 {
			temp := val2
			val2 = val1 % val2
			val1 = temp
		}
		return utils.Abs(val1), nil
	},
}
