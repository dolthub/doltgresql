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
	"github.com/jackc/pgtype"
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
		num1 := val1.(pgtype.Numeric)
		num2 := val2.(pgtype.Numeric)
		if num1.NaN || num2.NaN {
			return pgtypes.NumericNaN, nil
		}
		var res pgtype.Numeric
		if num1.InfinityModifier == pgtype.Infinity {
			if num2.InfinityModifier == pgtype.Infinity || (num2.Int != nil && num2.Int.Sign() > 0) {
				return pgtypes.NumericInfinite, nil
			}
			if num2.InfinityModifier == pgtype.NegativeInfinity || (num2.Int != nil && num2.Int.Sign() < 0) {
				err := res.Set(0)
				return res, err
			}
			err := res.Set(1)
			return res, err
		} else if num1.InfinityModifier == pgtype.NegativeInfinity {
			even := false
			if num2.Int != nil {
				var i int64
				err := num2.AssignTo(&i)
				if err != nil {
					return nil, errors.Errorf(`a negative number raised to a non-integer power yields a complex result`)
				}
				even = i%2 == 0
			}
			if num2.InfinityModifier == pgtype.Infinity || (num2.Int != nil && num2.Int.Sign() > 0) {
				if even {
					return pgtypes.NumericInfinite, nil
				}
				return pgtypes.NumericNegativeInfinite, nil
			}
			if num2.InfinityModifier == pgtype.NegativeInfinity || (num2.Int != nil && num2.Int.Sign() < 0) {
				err := res.Set(0)
				return res, err
			}
			err := res.Set(1)
			return res, err
		} else if num1.Int != nil && num1.Int.Sign() == 0 {
			if num2.InfinityModifier == pgtype.Infinity {
				err := res.Set(0)
				return res, err
			}
			if num2.InfinityModifier == pgtype.NegativeInfinity || num2.Int != nil && num2.Int.Sign() < 0 {
				return nil, errPowerZeroToNegative
			}
			if num2.Int != nil && num2.Int.Sign() > 0 {
				err := res.Set("0.0000000000000000")
				return res, err
			}
		}
		// decimal.Pow() does not handle the zero exponent properly, so we special case it
		if num2.Int != nil && num2.Int.Sign() == 0 {
			err := res.Set("1.0000000000000000")
			return res, err
		}
		dec1 := pgtypes.NumericToDecimal(num1)
		if dec1.Equal(decimal.NewFromInt(1)) {
			err := res.Set("1.0000000000000000")
			return res, err
		}
		return pgtypes.AnyToNumeric(pgtypes.NumericToDecimal(num1).Pow(pgtypes.NumericToDecimal(num2)))
	},
}
