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
	"strings"

	"github.com/cockroachdb/apd/v3"
	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initSqrt registers the functions to the catasqrt.
func initSqrt() {
	framework.RegisterFunction(sqrt_float64)
	framework.RegisterFunction(sqrt_numeric)
}

// sqrt_float64 represents the PostgreSQL function of the same name, taking the same parameters.
var sqrt_float64 = framework.Function1{
	Name:       "sqrt",
	Return:     pgtypes.Float64,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Float64},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		if val.(float64) < 0 {
			return nil, errors.Errorf("cannot take square root of a negative number")
		}
		return math.Sqrt(val.(float64)), nil
	},
}

// sqrt_numeric represents the PostgreSQL function of the same name, taking the same parameters.
var sqrt_numeric = framework.Function1{
	Name:       "sqrt",
	Return:     pgtypes.Numeric,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Numeric},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		dec := val.(apd.Decimal)
		if dec.Sign() < 0 {
			return nil, errors.Errorf("cannot take square root of a negative number")
		}

		// TODO: calculate precision and scale accurately
		s := dec.Text('f')
		parts := strings.Split(s, ".")

		exp := int32(-16)
		whole := int32(len(parts[0]) / 2)
		if dec.Exponent == 0 {
			exp = whole - 16
		} else if dec.Exponent < -16 {
			exp = dec.Exponent
		}
		p := uint32(whole) + 1
		if exp < 0 {
			p += uint32(-exp)
		}

		c := sql.DecimalCtx.WithPrecision(p)
		_, err := c.Sqrt(&dec, &dec)
		if err != nil {
			return nil, err
		}
		_, err = c.Quantize(&dec, &dec, exp)
		if err != nil {
			return nil, err
		}
		return dec, nil
	},
}
