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

package binary

import (
	"fmt"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/shopspring/decimal"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// These functions can be gathered using the following query from a Postgres 15 instance:
// SELECT * FROM pg_operator o WHERE o.oprname = '%' ORDER BY o.oprcode::varchar;

// initBinaryMod registers the functions to the catalog.
func initBinaryMod() {
	framework.RegisterBinaryFunction(framework.Operator_BinaryMod, int2mod)
	framework.RegisterBinaryFunction(framework.Operator_BinaryMod, int4mod)
	framework.RegisterBinaryFunction(framework.Operator_BinaryMod, int8mod)
	framework.RegisterBinaryFunction(framework.Operator_BinaryMod, numeric_mod)
}

// int2mod represents the PostgreSQL function of the same name, taking the same parameters.
var int2mod = framework.Function2{
	Name:       "int2mod",
	Return:     pgtypes.Int16,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Int16, pgtypes.Int16},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any, varargs ...any) (any, error) {
		if val2.(int16) == 0 {
			return nil, fmt.Errorf("division by zero")
		}
		return val1.(int16) % val2.(int16), nil
	},
}

// int4mod represents the PostgreSQL function of the same name, taking the same parameters.
var int4mod = framework.Function2{
	Name:       "int4mod",
	Return:     pgtypes.Int32,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Int32, pgtypes.Int32},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any, varargs ...any) (any, error) {
		if val2.(int32) == 0 {
			return nil, fmt.Errorf("division by zero")
		}
		return val1.(int32) % val2.(int32), nil
	},
}

// int8mod represents the PostgreSQL function of the same name, taking the same parameters.
var int8mod = framework.Function2{
	Name:       "int8mod",
	Return:     pgtypes.Int64,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Int64, pgtypes.Int64},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any, varargs ...any) (any, error) {
		if val2.(int64) == 0 {
			return nil, fmt.Errorf("division by zero")
		}
		return val1.(int64) % val2.(int64), nil
	},
}

// numeric_mod represents the PostgreSQL function of the same name, taking the same parameters.
var numeric_mod = framework.Function2{
	Name:       "numeric_mod",
	Return:     pgtypes.Numeric,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Numeric, pgtypes.Numeric},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any, varargs ...any) (any, error) {
		if val2.(decimal.Decimal).Equal(decimal.Zero) {
			return nil, fmt.Errorf("division by zero")
		}
		return val1.(decimal.Decimal).Mod(val2.(decimal.Decimal)), nil
	},
}
