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
	"fmt"

	"github.com/shopspring/decimal"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// init registers the functions to the catalog.
func init() {
	framework.RegisterFunction(mod_int16_int16)
	framework.RegisterFunction(mod_int32_int32)
	framework.RegisterFunction(mod_int64_int64)
	framework.RegisterFunction(mod_numeric_numeric)
}

// mod_int16_int16 represents the PostgreSQL function of the same name, taking the same parameters.
var mod_int16_int16 = framework.Function2{
	Name:       "mod",
	Return:     pgtypes.Int16,
	Parameters: []pgtypes.DoltgresType{pgtypes.Int16, pgtypes.Int16},
	Callable: func(ctx framework.Context, val1 any, val2 any) (any, error) {
		if val1 == nil || val2 == nil {
			return nil, nil
		}
		if val2.(int16) == 0 {
			return nil, fmt.Errorf("division by zero")
		}
		return val1.(int16) % val2.(int16), nil
	},
}

// mod_int32_int32 represents the PostgreSQL function of the same name, taking the same parameters.
var mod_int32_int32 = framework.Function2{
	Name:       "mod",
	Return:     pgtypes.Int32,
	Parameters: []pgtypes.DoltgresType{pgtypes.Int32, pgtypes.Int32},
	Callable: func(ctx framework.Context, val1 any, val2 any) (any, error) {
		if val1 == nil || val2 == nil {
			return nil, nil
		}
		if val2.(int32) == 0 {
			return nil, fmt.Errorf("division by zero")
		}
		return val1.(int32) % val2.(int32), nil
	},
}

// mod_int64_int64 represents the PostgreSQL function of the same name, taking the same parameters.
var mod_int64_int64 = framework.Function2{
	Name:       "mod",
	Return:     pgtypes.Int64,
	Parameters: []pgtypes.DoltgresType{pgtypes.Int64, pgtypes.Int64},
	Callable: func(ctx framework.Context, val1 any, val2 any) (any, error) {
		if val1 == nil || val2 == nil {
			return nil, nil
		}
		if val2.(int64) == 0 {
			return nil, fmt.Errorf("division by zero")
		}
		return val1.(int64) % val2.(int64), nil
	},
}

// mod_numeric_numeric represents the PostgreSQL function of the same name, taking the same parameters.
var mod_numeric_numeric = framework.Function2{
	Name:       "mod",
	Return:     pgtypes.Numeric,
	Parameters: []pgtypes.DoltgresType{pgtypes.Numeric, pgtypes.Numeric},
	Callable: func(ctx framework.Context, val1 any, val2 any) (any, error) {
		if val1 == nil || val2 == nil {
			return nil, nil
		}
		if val2.(decimal.Decimal).Cmp(decimal.Zero) == 0 {
			return nil, fmt.Errorf("division by zero")
		}
		return val1.(decimal.Decimal).Mod(val2.(decimal.Decimal)), nil
	},
}
