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
		return truncNumeric(val.(*apd.Decimal), 0)
	},
}

// trunc_numeric_int64 represents the PostgreSQL function of the same name, taking the same parameters.
var trunc_numeric_int64 = framework.Function2{
	Name:       "trunc",
	Return:     pgtypes.Numeric,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Numeric, pgtypes.Int32},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, num any, places any) (any, error) {
		return truncNumeric(num.(*apd.Decimal), places.(int32))
	},
}

// truncNumeric truncates dec toward zero, keeping the given number of digits after the decimal point. A negative
// number of places truncates digits to the left of the decimal point. As in PostgreSQL, the result has
// max(places, 0) digits after the decimal point, so trunc(1.5, 3) returns 1.500 and trunc(123.456, -1) returns 120.
func truncNumeric(dec *apd.Decimal, places int32) (*apd.Decimal, error) {
	if dec.Form != apd.Finite {
		// NaN and infinities are returned unchanged
		return new(apd.Decimal).Set(dec), nil
	}
	// Quantize needs a precision large enough to hold every digit of the result
	p := dec.NumDigits()
	if dec.Exponent > 0 {
		p += int64(dec.Exponent)
	}
	if places > 0 {
		p += int64(places)
	}
	ctx := sql.DecimalCtx.WithPrecision(uint32(p))
	ctx.Rounding = apd.RoundDown
	res := new(apd.Decimal)
	if _, err := ctx.Quantize(res, dec, -places); err != nil {
		return nil, err
	}
	if places < 0 {
		// Quantize leaves the result with a positive exponent (123.456 becomes 12E+1), so we rescale it to have no
		// fractional digits, which is exact
		if _, err := ctx.Quantize(res, res, 0); err != nil {
			return nil, err
		}
	}
	if res.IsZero() {
		// PostgreSQL has no negative zero, so trunc(-0.5) returns 0
		res.Negative = false
	}
	return res, nil
}
