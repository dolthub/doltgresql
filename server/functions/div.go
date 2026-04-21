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
	"github.com/cockroachdb/errors"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/jackc/pgtype"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initDiv registers the functions to the catalog.
func initDiv() {
	framework.RegisterFunction(div_numeric)
}

// div_numeric represents the PostgreSQL function of the same name, taking the same parameters.
var div_numeric = framework.Function2{
	Name:       "div",
	Return:     pgtypes.Numeric,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Numeric, pgtypes.Numeric},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		num1, num2 := val1.(pgtype.Numeric), val2.(pgtype.Numeric)
		return NumericDiv(num1, num2)
	},
}

// NumericDiv takes two pgtype.Numeric arguments and returns division result of them.
func NumericDiv(n1, n2 pgtype.Numeric) (pgtype.Numeric, error) {
	if n1.NaN || n2.NaN {
		return pgtypes.NumericNaN, nil
	}
	if n2.Int != nil && n2.Int.Sign() == 0 {
		return pgtype.Numeric{}, errors.Errorf("division by zero")
	}
	if (n1.InfinityModifier == pgtype.Infinity || n1.InfinityModifier == pgtype.NegativeInfinity) &&
		(n2.InfinityModifier == pgtype.Infinity || n2.InfinityModifier == pgtype.NegativeInfinity) {
		return pgtypes.NumericNaN, nil
	}
	if n1.InfinityModifier == pgtype.Infinity || n1.InfinityModifier == pgtype.NegativeInfinity {
		return n1, nil
	}
	if n2.InfinityModifier == pgtype.Infinity || n2.InfinityModifier == pgtype.NegativeInfinity {
		return pgtypes.NumericZeroo(), nil
	}
	return pgtypes.AnyToNumeric(pgtypes.NumericToDecimal(n1).Div(pgtypes.NumericToDecimal(n2)))
}
