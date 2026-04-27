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
	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"

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
	numericOne = apd.New(1, 0)
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
		dec1 := val1.(apd.Decimal)
		dec2 := val2.(apd.Decimal)
		if dec1.Form == apd.NaN || dec2.Form == apd.NaN {
			return pgtypes.NumericNaN, nil
		}
		if dec1.Form == apd.Infinite && dec1.Negative {
			even := dec2.Form == apd.Infinite && !dec2.Negative
			if dec2.Form == apd.Finite {
				i, err := dec2.Int64()
				if err != nil {
					return nil, errors.Errorf(`a negative number raised to a non-integer power yields a complex result`)
				}
				even = i%2 == 0
			}

			if dec2.Sign() > 0 {
				// +inf will return neginf == fix!!
				if even {
					return pgtypes.NumericInf, nil
				}
				return pgtypes.NumericNegInf, nil
			}
			if (dec2.Form == apd.Infinite && dec2.Negative) || dec2.Sign() < 0 {
				return *apd.New(0, 0), nil
			}
			return *apd.New(1, 0), nil
		}
		if dec1.IsZero() {
			if dec2.Sign() < 0 {
				// includes neg inf
				return nil, errPowerZeroToNegative
			}
			if dec2.Form == apd.Infinite {
				return *apd.New(0, 0), nil
			}
			if dec2.Sign() > 0 {
				d := *apd.New(0, 0)
				_, _ = pgtypes.BaseContext.Quantize(&d, &d, -16)
				return d, nil
			}
		}
		// decimal.Pow() does not handle the zero exponent properly, so we special case it

		if dec2.IsZero() || dec1.Cmp(numericOne) == 0 {
			d := *apd.New(1, 0)
			_, _ = pgtypes.BaseContext.Quantize(&d, &d, -16)
			return d, nil
		}
		// give enough precision that we can round it to 16 exp
		_, err := pgtypes.BaseContext.WithPrecision(17).Pow(&dec1, &dec1, &dec2)
		if err != nil {
			return nil, err
		}
		_, err = pgtypes.BaseContext.Quantize(&dec1, &dec1, -16)
		if err != nil {
			return nil, err
		}
		return dec1, nil
	},
}
