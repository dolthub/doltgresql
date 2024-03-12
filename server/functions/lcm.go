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

// init registers the functions to the catalog.
func init() {
	framework.RegisterFunction(lcm_int64_int64)
}

// lcm_int64_int64 represents the PostgreSQL function of the same name, taking the same parameters.
var lcm_int64_int64 = framework.Function2{
	Name:       "lcm",
	Return:     pgtypes.Int64,
	Parameters: []pgtypes.DoltgresType{pgtypes.Int64, pgtypes.Int64},
	Callable: func(ctx framework.Context, val1 any, val2 any) (any, error) {
		if val1 == nil || val2 == nil {
			return nil, nil
		}
		gcdResultInterface, err := gcd_int64_int64.Callable(ctx, val1, val2)
		if err != nil {
			return nil, err
		}
		gcdResult := gcdResultInterface.(int64)
		if gcdResult == 0 {
			return int64(0), nil
		}
		return utils.Abs((val1.(int64) * val2.(int64)) / gcdResult), nil
	},
}
