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
	"github.com/cockroachdb/errors"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initMod registers the functions to the catalog.
func initMod() {
	framework.RegisterFunction(mod_int16_int16)
	framework.RegisterFunction(mod_int32_int32)
	framework.RegisterFunction(mod_int64_int64)
	framework.RegisterFunction(mod_numeric_numeric)
}

// mod_int16_int16 represents the PostgreSQL function of the same name, taking the same parameters.
var mod_int16_int16 = framework.Function2{
	Name:       "mod",
	Return:     pgtypes.Int16,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Int16, pgtypes.Int16},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		if val2.(int16) == 0 {
			return nil, errors.Errorf("division by zero")
		}
		return val1.(int16) % val2.(int16), nil
	},
}

// mod_int32_int32 represents the PostgreSQL function of the same name, taking the same parameters.
var mod_int32_int32 = framework.Function2{
	Name:       "mod",
	Return:     pgtypes.Int32,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Int32, pgtypes.Int32},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		if val2.(int32) == 0 {
			return nil, errors.Errorf("division by zero")
		}
		return val1.(int32) % val2.(int32), nil
	},
}

// mod_int64_int64 represents the PostgreSQL function of the same name, taking the same parameters.
var mod_int64_int64 = framework.Function2{
	Name:       "mod",
	Return:     pgtypes.Int64,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Int64, pgtypes.Int64},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		if val2.(int64) == 0 {
			return nil, errors.Errorf("division by zero")
		}
		return val1.(int64) % val2.(int64), nil
	},
}

// mod_numeric_numeric represents the PostgreSQL function of the same name, taking the same parameters.
var mod_numeric_numeric = framework.Function2{
	Name:       "mod",
	Return:     pgtypes.Numeric,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Numeric, pgtypes.Numeric},
	Strict:     true,
	Callable:   NumericModCallable,
}

// NumericModCallable is the callable logic for the numeric_mod and mod functions.
func NumericModCallable(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
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
	_, err := pgtypes.BaseContext.Rem(&num1, &num1, &num2)
	if err != nil {
		return nil, err
	}
	return num1, nil
}
