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
	"math"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/shopspring/decimal"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initPower registers the functions to the catalog.
func initPower() {
	framework.RegisterFunction(power_float64_float64)
	framework.RegisterFunction(power_numeric_numeric)
}

// power_float64_float64 represents the PostgreSQL function of the same name, taking the same parameters.
var power_float64_float64 = framework.Function2{
	Name:       "power",
	Return:     pgtypes.Float64,
	Parameters: []pgtypes.DoltgresType{pgtypes.Float64, pgtypes.Float64},
	Callable: func(ctx *sql.Context, val1 any, val2 any) (any, error) {
		if val1 == nil || val2 == nil {
			return nil, nil
		}
		return math.Pow(val1.(float64), val2.(float64)), nil
	},
}

// power_numeric_numeric represents the PostgreSQL function of the same name, taking the same parameters.
var power_numeric_numeric = framework.Function2{
	Name:       "power",
	Return:     pgtypes.Numeric,
	Parameters: []pgtypes.DoltgresType{pgtypes.Numeric, pgtypes.Numeric},
	Callable: func(ctx *sql.Context, val1 any, val2 any) (any, error) {
		if val1 == nil || val2 == nil {
			return nil, nil
		}
		one := decimal.NewFromInt(1)
		if val1.(decimal.Decimal).Cmp(one) == 0 {
			return one, nil
		}
		return val1.(decimal.Decimal).Pow(val2.(decimal.Decimal)), nil
	},
}
