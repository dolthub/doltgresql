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

	"github.com/dolthub/doltgresql/postgres/parser/duration"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// These functions can be gathered using the following query from a Postgres 15 instance:
// SELECT * FROM pg_operator o WHERE o.oprname = '/' ORDER BY o.oprcode::varchar;

// initBinaryDivide registers the functions to the catalog.
func initBinaryDivide() {
	framework.RegisterBinaryFunction(framework.Operator_BinaryDivide, float4div)
	framework.RegisterBinaryFunction(framework.Operator_BinaryDivide, float48div)
	framework.RegisterBinaryFunction(framework.Operator_BinaryDivide, float8div)
	framework.RegisterBinaryFunction(framework.Operator_BinaryDivide, float84div)
	framework.RegisterBinaryFunction(framework.Operator_BinaryDivide, int2div)
	framework.RegisterBinaryFunction(framework.Operator_BinaryDivide, int24div)
	framework.RegisterBinaryFunction(framework.Operator_BinaryDivide, int28div)
	framework.RegisterBinaryFunction(framework.Operator_BinaryDivide, int4div)
	framework.RegisterBinaryFunction(framework.Operator_BinaryDivide, int42div)
	framework.RegisterBinaryFunction(framework.Operator_BinaryDivide, int48div)
	framework.RegisterBinaryFunction(framework.Operator_BinaryDivide, int8div)
	framework.RegisterBinaryFunction(framework.Operator_BinaryDivide, int82div)
	framework.RegisterBinaryFunction(framework.Operator_BinaryDivide, int84div)
	framework.RegisterBinaryFunction(framework.Operator_BinaryDivide, interval_div)
	framework.RegisterBinaryFunction(framework.Operator_BinaryDivide, numeric_div)
}

// float4div represents the PostgreSQL function of the same name, taking the same parameters.
var float4div = framework.Function2{
	Name:       "float4div",
	Return:     pgtypes.Float32,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Float32, pgtypes.Float32},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		if val2.(float32) == 0 {
			return nil, fmt.Errorf("division by zero")
		}
		return val1.(float32) / val2.(float32), nil
	},
}

// float48div represents the PostgreSQL function of the same name, taking the same parameters.
var float48div = framework.Function2{
	Name:       "float48div",
	Return:     pgtypes.Float64,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Float32, pgtypes.Float64},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		if val2.(float64) == 0 {
			return nil, fmt.Errorf("division by zero")
		}
		return float64(val1.(float32)) / val2.(float64), nil
	},
}

// float8div represents the PostgreSQL function of the same name, taking the same parameters.
var float8div = framework.Function2{
	Name:       "float8div",
	Return:     pgtypes.Float64,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Float64, pgtypes.Float64},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		if val2.(float64) == 0 {
			return nil, fmt.Errorf("division by zero")
		}
		return val1.(float64) / val2.(float64), nil
	},
}

// float84div represents the PostgreSQL function of the same name, taking the same parameters.
var float84div = framework.Function2{
	Name:       "float84div",
	Return:     pgtypes.Float64,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Float64, pgtypes.Float32},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		if val2.(float32) == 0 {
			return nil, fmt.Errorf("division by zero")
		}
		return val1.(float64) / float64(val2.(float32)), nil
	},
}

// int2div represents the PostgreSQL function of the same name, taking the same parameters.
var int2div = framework.Function2{
	Name:       "int2div",
	Return:     pgtypes.Int16,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Int16, pgtypes.Int16},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		if val2.(int16) == 0 {
			return nil, fmt.Errorf("division by zero")
		}
		return val1.(int16) / val2.(int16), nil
	},
}

// int24div represents the PostgreSQL function of the same name, taking the same parameters.
var int24div = framework.Function2{
	Name:       "int24div",
	Return:     pgtypes.Int32,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Int16, pgtypes.Int32},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		if val2.(int32) == 0 {
			return nil, fmt.Errorf("division by zero")
		}
		return int32(val1.(int16)) / val2.(int32), nil
	},
}

// int28div represents the PostgreSQL function of the same name, taking the same parameters.
var int28div = framework.Function2{
	Name:       "int28div",
	Return:     pgtypes.Int64,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Int16, pgtypes.Int64},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		if val2.(int64) == 0 {
			return nil, fmt.Errorf("division by zero")
		}
		return int64(val1.(int16)) / val2.(int64), nil
	},
}

// int4div represents the PostgreSQL function of the same name, taking the same parameters.
var int4div = framework.Function2{
	Name:       "int4div",
	Return:     pgtypes.Int32,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Int32, pgtypes.Int32},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		if val2.(int32) == 0 {
			return nil, fmt.Errorf("division by zero")
		}
		return val1.(int32) / val2.(int32), nil
	},
}

// int42div represents the PostgreSQL function of the same name, taking the same parameters.
var int42div = framework.Function2{
	Name:       "int42div",
	Return:     pgtypes.Int32,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Int32, pgtypes.Int16},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		if val2.(int16) == 0 {
			return nil, fmt.Errorf("division by zero")
		}
		return val1.(int32) / int32(val2.(int16)), nil
	},
}

// int48div represents the PostgreSQL function of the same name, taking the same parameters.
var int48div = framework.Function2{
	Name:       "int48div",
	Return:     pgtypes.Int64,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Int32, pgtypes.Int64},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		if val2.(int64) == 0 {
			return nil, fmt.Errorf("division by zero")
		}
		return int64(val1.(int32)) / val2.(int64), nil
	},
}

// int8div represents the PostgreSQL function of the same name, taking the same parameters.
var int8div = framework.Function2{
	Name:       "int8div",
	Return:     pgtypes.Int64,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Int64, pgtypes.Int64},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		if val2.(int64) == 0 {
			return nil, fmt.Errorf("division by zero")
		}
		return val1.(int64) / val2.(int64), nil
	},
}

// int82div represents the PostgreSQL function of the same name, taking the same parameters.
var int82div = framework.Function2{
	Name:       "int82div",
	Return:     pgtypes.Int64,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Int64, pgtypes.Int16},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		if val2.(int16) == 0 {
			return nil, fmt.Errorf("division by zero")
		}
		return val1.(int64) / int64(val2.(int16)), nil
	},
}

// int84div represents the PostgreSQL function of the same name, taking the same parameters.
var int84div = framework.Function2{
	Name:       "int84div",
	Return:     pgtypes.Int64,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Int64, pgtypes.Int32},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		if val2.(int32) == 0 {
			return nil, fmt.Errorf("division by zero")
		}
		return val1.(int64) / int64(val2.(int32)), nil
	},
}

// interval_div represents the PostgreSQL function of the same name, taking the same parameters.
var interval_div = framework.Function2{
	Name:       "interval_div",
	Return:     pgtypes.Interval,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Interval, pgtypes.Float64},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		if val2.(float64) == 0 {
			return nil, fmt.Errorf("division by zero")
		}
		return val1.(duration.Duration).DivFloat(val2.(float64)), nil
	},
}

// numeric_div represents the PostgreSQL function of the same name, taking the same parameters.
var numeric_div = framework.Function2{
	Name:       "numeric_div",
	Return:     pgtypes.Numeric,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Numeric, pgtypes.Numeric},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		if val2.(decimal.Decimal).Equal(decimal.Zero) {
			return nil, fmt.Errorf("division by zero")
		}
		return val1.(decimal.Decimal).Div(val2.(decimal.Decimal)), nil
	},
}
