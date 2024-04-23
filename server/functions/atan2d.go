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
	"fmt"
	"math"

	"github.com/dolthub/doltgresql/server/functions/framework"

	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initAtan2d registers the functions to the catalog.
func initAtan2d() {
	framework.RegisterFunction(atan2d_float64)
}

// atan2_float64 represents the PostgreSQL function of the same name, taking the same parameters.
var atan2d_float64 = framework.Function2{
	Name:       "atan2d",
	Return:     pgtypes.Float64,
	Parameters: []pgtypes.DoltgresType{pgtypes.Float64, pgtypes.Float64},
	Callable: func(ctx framework.Context, y any, x any) (any, error) {
		if y == nil || x == nil {
			return nil, nil
		}
		r := math.Atan2(y.(float64), x.(float64))
		if math.IsNaN(r) {
			return nil, fmt.Errorf("input is out of range")
		}
		return toDegrees(r), nil
	},
}
