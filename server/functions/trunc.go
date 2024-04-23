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

	"github.com/shopspring/decimal"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initTrunc registers the functions to the catalog.
func initTrunc() {
	framework.RegisterFunction(trunc_float64)
	framework.RegisterFunction(trunc_numeric)
	framework.RegisterFunction(trunc_numeric_int64)
}

// trunc_float64 represents the PostgreSQL function of the same name, taking the same parameters.
var trunc_float64 = framework.Function1{
	Name:       "trunc",
	Return:     pgtypes.Float64,
	Parameters: []pgtypes.DoltgresType{pgtypes.Float64},
	Callable: func(ctx framework.Context, val1 any) (any, error) {
		if val1 == nil {
			return nil, nil
		}
		return math.Trunc(val1.(float64)), nil
	},
}

// trunc_numeric represents the PostgreSQL function of the same name, taking the same parameters.
var trunc_numeric = framework.Function1{
	Name:       "trunc",
	Return:     pgtypes.Numeric,
	Parameters: []pgtypes.DoltgresType{pgtypes.Numeric},
	Callable: func(ctx framework.Context, val1 any) (any, error) {
		if val1 == nil {
			return nil, nil
		}
		return decimal.NewFromInt(val1.(decimal.Decimal).IntPart()), nil
	},
}

// trunc_numeric_int64 represents the PostgreSQL function of the same name, taking the same parameters.
var trunc_numeric_int64 = framework.Function2{
	Name:       "trunc",
	Return:     pgtypes.Numeric,
	Parameters: []pgtypes.DoltgresType{pgtypes.Numeric, pgtypes.Int32},
	Callable: func(ctx framework.Context, num any, places any) (any, error) {
		if num == nil || places == nil {
			return nil, nil
		}
		//TODO: test for negative values in places
		return num.(decimal.Decimal).Truncate(places.(int32)), nil
	},
}
