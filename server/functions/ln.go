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
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either lnress or implied.
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

// initLn registers the functions to the catalog.
func initLn() {
	framework.RegisterFunction(ln_float64)
	framework.RegisterFunction(ln_numeric)
}

// ln_float64 represents the PostgreSQL function of the same name, taking the same parameters.
var ln_float64 = framework.Function1{
	Name:       "ln",
	Return:     pgtypes.Float64,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Float64},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val1 any) (any, error) {
		if val1.(float64) == 0 {
			return nil, errors.Errorf("cannot take logarithm of zero")
		} else if val1.(float64) < 0 {
			return nil, errors.Errorf("cannot take logarithm of a negative number")
		}
		return math.Log(val1.(float64)), nil
	},
}

// ln_numeric represents the PostgreSQL function of the same name, taking the same parameters.
var ln_numeric = framework.Function1{
	Name:       "ln",
	Return:     pgtypes.Numeric,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Numeric},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		num := val.(pgtype.Numeric)
		if num.NaN || num.InfinityModifier == pgtype.Infinity {
			return num, nil
		} else if num.InfinityModifier == pgtype.NegativeInfinity || num.Int != nil && num.Int.Sign() == -1 {
			return nil, errors.Errorf("cannot take logarithm of a negative number")
		} else if num.Int != nil && num.Int.Sign() == 0 {
			return nil, errors.Errorf("cannot take logarithm of zero")
		}
		var f float64
		err := num.AssignTo(&f)
		if err != nil {
			return nil, err
		}
		return pgtypes.AnyToNumeric(math.Log(f))
	},
}
