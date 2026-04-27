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
	"github.com/cockroachdb/apd/v3"
	errors "github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"

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
	Callable:   NumericDivCallable,
}

// NumericDivCallable is the callable logic for the numeric_div and div functions.
func NumericDivCallable(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
	num1 := val1.(apd.Decimal)
	num2 := val2.(apd.Decimal)
	if num1.Form == apd.NaN || num2.Form == apd.NaN ||
		(num1.Form == apd.Infinite && num2.Form == apd.Infinite) {
		return pgtypes.NumericNaN, nil
	}
	if num2.IsZero() {
		return nil, errors.Errorf("division by zero")
	}
	if num1.Form == apd.Infinite {
		return num1, nil
	}
	if num2.Form == apd.Infinite {
		return *apd.New(0, 0), nil
	}
	_, err := pgtypes.BaseContext.Quo(&num1, &num1, &num2)
	if err != nil {
		return nil, err
	}
	_, err = pgtypes.BaseContext.Quantize(&num1, &num1, -16)
	if err != nil {
		return nil, err
	}
	return num1, nil
}
