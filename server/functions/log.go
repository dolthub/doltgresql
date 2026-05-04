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
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		f := val.(float64)
		if f == 0 {
			return nil, errors.Errorf("cannot take logarithm of zero")
		} else if f < 0 {
			return nil, errors.Errorf("cannot take logarithm of a negative number")
		}
		return math.Log10(f), nil
	},
}

// log_numeric represents the PostgreSQL function of the same name, taking the same parameters.
var log_numeric = framework.Function1{
	Name:       "log",
	Return:     pgtypes.Numeric,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Numeric},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		num := val.(pgtype.Numeric)
		if num.NaN || num.InfinityModifier == pgtype.Infinity {
			return num, nil
		} else if num.InfinityModifier == pgtype.NegativeInfinity || (num.Int != nil && num.Int.Sign() == -1) {
			return nil, errors.Errorf("cannot take logarithm of a negative number")
		} else if num.Int != nil && num.Int.Sign() == 0 {
			return nil, errors.Errorf("cannot take logarithm of zero")
		}

		// TODO: implement log for numeric instead of relying on float64
		var f float64
		err := num.AssignTo(&f)
		if err != nil {
			return nil, err
		}
		return pgtypes.AnyToNumeric(math.Log10(f))
	},
}

// log_numeric_numeric represents the PostgreSQL function of the same name, taking the same parameters.
var log_numeric_numeric = framework.Function2{
	Name:       "log",
	Return:     pgtypes.Numeric,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Numeric, pgtypes.Numeric},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		num1 := val1.(pgtype.Numeric)
		num2 := val2.(pgtype.Numeric)
		if num1.NaN || num2.NaN {
			// return NaN
			return num1, nil
		} else if num1.InfinityModifier == pgtype.NegativeInfinity || num2.InfinityModifier == pgtype.NegativeInfinity ||
			(num1.Int != nil && num1.Int.Sign() == -1) || (num2.Int != nil && num2.Int.Sign() == -1) {
			return nil, errors.Errorf("cannot take logarithm of a negative number")
		} else if (num1.Int != nil && num1.Int.Sign() == 0) || (num2.Int != nil && num2.Int.Sign() == 0) {
			return nil, errors.Errorf("cannot take logarithm of zero")
		}

		// TODO: implement log for numeric instead of relying on float64
		var base, num float64
		err := num1.AssignTo(&base)
		if err != nil {
			return nil, err
		}
		logBase := math.Log(base)
		if logBase == 0 {
			return nil, errors.Errorf("division by zero")
		}
		err = num2.AssignTo(&num)
		if err != nil {
			return nil, err
		}
		var l pgtype.Numeric
		err = l.Set(math.Log(num) / logBase)
		return l, err
	},
}
