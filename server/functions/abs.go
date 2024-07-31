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
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/shopspring/decimal"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
	"github.com/dolthub/doltgresql/utils"
)

// initAbs registers the functions to the catalog.
func initAbs() {
	framework.RegisterFunction(abs_int16)
	framework.RegisterFunction(abs_int32)
	framework.RegisterFunction(abs_int64)
	framework.RegisterFunction(abs_float64)
	framework.RegisterFunction(abs_numeric)
}

// abs_int16 represents the PostgreSQL function of the same name, taking the same parameters.
var abs_int16 = framework.Function1{
	Name:       "abs",
	Return:     pgtypes.Int16,
	Parameters: [1]pgtypes.DoltgresType{pgtypes.Int16},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, val1 any) (any, error) {
		return utils.Abs(val1.(int16)), nil
	},
}

// abs_int32 represents the PostgreSQL function of the same name, taking the same parameters.
var abs_int32 = framework.Function1{
	Name:       "abs",
	Return:     pgtypes.Int32,
	Parameters: [1]pgtypes.DoltgresType{pgtypes.Int32},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, val1 any) (any, error) {
		return utils.Abs(val1.(int32)), nil
	},
}

// abs_int64 represents the PostgreSQL function of the same name, taking the same parameters.
var abs_int64 = framework.Function1{
	Name:       "abs",
	Return:     pgtypes.Int64,
	Parameters: [1]pgtypes.DoltgresType{pgtypes.Int64},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, val1 any) (any, error) {
		return utils.Abs(val1.(int64)), nil
	},
}

// abs_float64 represents the PostgreSQL function of the same name, taking the same parameters.
var abs_float64 = framework.Function1{
	Name:       "abs",
	Return:     pgtypes.Float64,
	Parameters: [1]pgtypes.DoltgresType{pgtypes.Float64},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, val1 any) (any, error) {
		return utils.Abs(val1.(float64)), nil
	},
}

// abs_numeric represents the PostgreSQL function of the same name, taking the same parameters.
var abs_numeric = framework.Function1{
	Name:       "abs",
	Return:     pgtypes.Numeric,
	Parameters: [1]pgtypes.DoltgresType{pgtypes.Numeric},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, val1 any) (any, error) {
		return val1.(decimal.Decimal).Abs(), nil
	},
}
