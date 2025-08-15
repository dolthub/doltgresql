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

	"github.com/cockroachdb/errors"
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

var (
	// errPowerZeroToNegative is an error for raising zero to a negative power in the "power" functions.
	errPowerZeroToNegative = errors.New("zero raised to a negative power is undefined")
	// numericOne is equivalent to decimal.NewFromInt(1), but represented as a value for the sake of efficiency.
	numericOne = decimal.NewFromInt(1)
)

// power_float64_float64 represents the PostgreSQL function of the same name, taking the same parameters.
var power_float64_float64 = framework.Function2{
	Name:       "power",
	Return:     pgtypes.Float64,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Float64, pgtypes.Float64},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		f1 := val1.(float64)
		f2 := val2.(float64)
		if f1 == 0 && f2 < 0 {
			return nil, errPowerZeroToNegative
		}
		return math.Pow(f1, f2), nil
	},
}

// power_numeric_numeric represents the PostgreSQL function of the same name, taking the same parameters.
var power_numeric_numeric = framework.Function2{
	Name:       "power",
	Return:     pgtypes.Numeric,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Numeric, pgtypes.Numeric},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		if val1 == nil || val2 == nil {
			return nil, nil
		}
		d1 := val1.(decimal.Decimal)
		d2 := val2.(decimal.Decimal)
		if d1.Equal(numericOne) {
			return numericOne, nil
		}
		if d1.Equal(decimal.Zero) && d2.Cmp(decimal.Zero) == -1 {
			return nil, errPowerZeroToNegative
		}
		// TODO: this doesn't handle non-integer exponents
		return d1.Pow(d2), nil
	},
}
