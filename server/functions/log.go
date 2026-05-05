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

// initLog registers the functions to the catalog.
func initLog() {
	framework.RegisterFunction(log_float64)
	framework.RegisterFunction(log_numeric)
	framework.RegisterFunction(log_numeric_numeric)
}

// log_float64 represents the PostgreSQL function of the same name, taking the same parameters.
var log_float64 = framework.Function1{
	Name:       "log",
	Return:     pgtypes.Float64,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Float64},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val1Interface any) (any, error) {
		val1 := val1Interface.(float64)
		if val1 == 0 {
			return nil, errors.Errorf("cannot take logarithm of zero")
		} else if val1 < 0 {
			return nil, errors.Errorf("cannot take logarithm of a negative number")
		}
		return math.Log10(val1), nil
	},
}

// log_numeric represents the PostgreSQL function of the same name, taking the same parameters.
var log_numeric = framework.Function1{
	Name:       "log",
	Return:     pgtypes.Numeric,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Numeric},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val1 any) (any, error) {
		dec := val1.(apd.Decimal)
		if dec.IsZero() {
			return nil, errors.Errorf("cannot take logarithm of zero")
		} else if dec.Sign() < 0 {
			return nil, errors.Errorf("cannot take logarithm of a negative number")
		}

		// TODO: calculate precision and scale accurately
		p := uint32(17)
		if dec.Exponent < 0 {
			p += uint32(-dec.Exponent)
		}
		c := sql.DecimalCtx.WithPrecision(p)
		_, err := c.Log10(&dec, &dec)
		if err != nil {
			return nil, err
		}
		return dec, nil
	},
}

// log_numeric_numeric represents the PostgreSQL function of the same name, taking the same parameters.
var log_numeric_numeric = framework.Function2{
	Name:       "log",
	Return:     pgtypes.Numeric,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Numeric, pgtypes.Numeric},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		base := val1.(apd.Decimal)
		num := val2.(apd.Decimal)
		if base.IsZero() || num.IsZero() {
			return nil, errors.Errorf("cannot take logarithm of zero")
		} else if base.Sign() < 0 || num.Sign() < 0 {
			return nil, errors.Errorf("cannot take logarithm of a negative number")
		}

		// TODO: calculate precision and scale accurately
		sNum := num.Text('f')
		sBase := base.Text('f')
		partsNum := strings.Split(sNum, ".")
		partsBase := strings.Split(sBase, ".")
		exp := int32(-16)
		if minExp := math.Min(float64(base.Exponent), float64(num.Exponent)); int32(minExp) < exp {
			exp = int32(minExp)
		}
		p := uint32(int32(math.Max(float64(len(partsNum[0])), float64(len(partsBase[0])))) + (-exp))
		c := sql.DecimalCtx.WithPrecision(p)

		lnBase := new(apd.Decimal)
		_, err := c.Ln(lnBase, &base)
		if err != nil {
			return nil, err
		}
		if lnBase.IsZero() {
			return nil, errors.Errorf("division by zero")
		}

		lnNum := new(apd.Decimal)
		_, err = c.Ln(lnNum, &num)
		if err != nil {
			return nil, err
		}
		if lnNum.IsZero() {
			return *apd.New(0, -16), nil
		}

		res := new(apd.Decimal)
		_, err = c.Quo(res, lnNum, lnBase)
		if err != nil {
			return nil, err
		}

		_, err = c.Quantize(res, res, exp)
		if err != nil {
			return nil, err
		}
		return *res, nil
	},
}
