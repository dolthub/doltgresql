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

	"github.com/cockroachdb/apd/v3"
	"github.com/dolthub/go-mysql-server/sql"

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
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Float64},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		return math.Trunc(val.(float64)), nil
	},
}

// trunc_numeric represents the PostgreSQL function of the same name, taking the same parameters.
var trunc_numeric = framework.Function1{
	Name:       "trunc",
	Return:     pgtypes.Numeric,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Numeric},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		dec := val.(apd.Decimal)
		// TODO: calculate precision and scale accurately
		c := apd.BaseContext.WithPrecision(1000000)
		_, err := c.Quantize(&dec, &dec, 0)
		if err != nil {
			return nil, err
		}
		return dec, nil
	},
}

// trunc_numeric_int64 represents the PostgreSQL function of the same name, taking the same parameters.
var trunc_numeric_int64 = framework.Function2{
	Name:       "trunc",
	Return:     pgtypes.Numeric,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Numeric, pgtypes.Int32},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		//TODO: test for negative values in places
		dec := val1.(apd.Decimal)
		places := val2.(int32)
		// TODO: calculate precision and scale accurately
		c := apd.BaseContext.WithPrecision(1000000)
		_, err := c.Quantize(&dec, &dec, -places)
		if err != nil {
			return nil, err
		}
		return dec, nil
	},
}
