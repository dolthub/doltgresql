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
	"math"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/shopspring/decimal"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// These functions can be gathered using the following query from a Postgres 15 instance:
// SELECT * FROM pg_operator o WHERE o.oprname = '*' ORDER BY o.oprcode::varchar;

// initBinaryMultiply registers the functions to the catalog.
func initBinaryMultiply() {
	framework.RegisterBinaryFunction(framework.Operator_BinaryMultiply, float4mul)
	framework.RegisterBinaryFunction(framework.Operator_BinaryMultiply, float48mul)
	framework.RegisterBinaryFunction(framework.Operator_BinaryMultiply, float8mul)
	framework.RegisterBinaryFunction(framework.Operator_BinaryMultiply, float84mul)
	framework.RegisterBinaryFunction(framework.Operator_BinaryMultiply, int2mul)
	framework.RegisterBinaryFunction(framework.Operator_BinaryMultiply, int24mul)
	framework.RegisterBinaryFunction(framework.Operator_BinaryMultiply, int28mul)
	framework.RegisterBinaryFunction(framework.Operator_BinaryMultiply, int4mul)
	framework.RegisterBinaryFunction(framework.Operator_BinaryMultiply, int42mul)
	framework.RegisterBinaryFunction(framework.Operator_BinaryMultiply, int48mul)
	framework.RegisterBinaryFunction(framework.Operator_BinaryMultiply, int8mul)
	framework.RegisterBinaryFunction(framework.Operator_BinaryMultiply, int82mul)
	framework.RegisterBinaryFunction(framework.Operator_BinaryMultiply, int84mul)
	framework.RegisterBinaryFunction(framework.Operator_BinaryMultiply, numeric_mul)
}

// float4mul represents the PostgreSQL function of the same name, taking the same parameters.
var float4mul = framework.Function2{
	Name:       "float4mul",
	Return:     pgtypes.Float32,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Float32, pgtypes.Float32},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		return val1.(float32) * val2.(float32), nil
	},
}

// float48mul represents the PostgreSQL function of the same name, taking the same parameters.
var float48mul = framework.Function2{
	Name:       "float48mul",
	Return:     pgtypes.Float64,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Float32, pgtypes.Float64},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		return float64(val1.(float32)) * val2.(float64), nil
	},
}

// float8mul represents the PostgreSQL function of the same name, taking the same parameters.
var float8mul = framework.Function2{
	Name:       "float8mul",
	Return:     pgtypes.Float64,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Float64, pgtypes.Float64},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		return val1.(float64) * val2.(float64), nil
	},
}

// float84mul represents the PostgreSQL function of the same name, taking the same parameters.
var float84mul = framework.Function2{
	Name:       "float84mul",
	Return:     pgtypes.Float64,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Float64, pgtypes.Float32},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		return val1.(float64) * float64(val2.(float32)), nil
	},
}

// int2mul represents the PostgreSQL function of the same name, taking the same parameters.
var int2mul = framework.Function2{
	Name:       "int2mul",
	Return:     pgtypes.Int16,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Int16, pgtypes.Int16},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		result := int64(val1.(int16)) * int64(val2.(int16))
		if result > math.MaxInt16 || result < math.MinInt16 {
			return nil, fmt.Errorf("smallint out of range")
		}
		return int16(result), nil
	},
}

// int24mul represents the PostgreSQL function of the same name, taking the same parameters.
var int24mul = framework.Function2{
	Name:       "int24mul",
	Return:     pgtypes.Int32,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Int16, pgtypes.Int32},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		result := int64(val1.(int16)) * int64(val2.(int32))
		if result > math.MaxInt16 || result < math.MinInt16 {
			return nil, fmt.Errorf("integer out of range")
		}
		return int32(result), nil
	},
}

// int28mul represents the PostgreSQL function of the same name, taking the same parameters.
var int28mul = framework.Function2{
	Name:       "int28mul",
	Return:     pgtypes.Int64,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Int16, pgtypes.Int64},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		return multiplyOverflow(int64(val1.(int16)), val2.(int64))
	},
}

// int4mul represents the PostgreSQL function of the same name, taking the same parameters.
var int4mul = framework.Function2{
	Name:       "int4mul",
	Return:     pgtypes.Int32,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Int32, pgtypes.Int32},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		result := int64(val1.(int32)) * int64(val2.(int32))
		if result > math.MaxInt32 || result < math.MinInt32 {
			return nil, fmt.Errorf("integer out of range")
		}
		return int32(result), nil
	},
}

// int42mul represents the PostgreSQL function of the same name, taking the same parameters.
var int42mul = framework.Function2{
	Name:       "int42mul",
	Return:     pgtypes.Int32,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Int32, pgtypes.Int16},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		result := int64(val1.(int32)) * int64(val2.(int16))
		if result > math.MaxInt32 || result < math.MinInt32 {
			return nil, fmt.Errorf("integer out of range")
		}
		return int32(result), nil
	},
}

// int48mul represents the PostgreSQL function of the same name, taking the same parameters.
var int48mul = framework.Function2{
	Name:       "int48mul",
	Return:     pgtypes.Int64,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Int32, pgtypes.Int64},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		return multiplyOverflow(int64(val1.(int32)), val2.(int64))
	},
}

// int8mul represents the PostgreSQL function of the same name, taking the same parameters.
var int8mul = framework.Function2{
	Name:       "int8mul",
	Return:     pgtypes.Int64,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Int64, pgtypes.Int64},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		return multiplyOverflow(val1.(int64), val2.(int64))
	},
}

// int82mul represents the PostgreSQL function of the same name, taking the same parameters.
var int82mul = framework.Function2{
	Name:       "int82mul",
	Return:     pgtypes.Int64,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Int64, pgtypes.Int16},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		return multiplyOverflow(val1.(int64), int64(val2.(int16)))
	},
}

// int84mul represents the PostgreSQL function of the same name, taking the same parameters.
var int84mul = framework.Function2{
	Name:       "int84mul",
	Return:     pgtypes.Int64,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Int64, pgtypes.Int32},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		return multiplyOverflow(val1.(int64), int64(val2.(int32)))
	},
}

// numeric_mul represents the PostgreSQL function of the same name, taking the same parameters.
var numeric_mul = framework.Function2{
	Name:       "numeric_mul",
	Return:     pgtypes.Numeric,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Numeric, pgtypes.Numeric},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		return val1.(decimal.Decimal).Mul(val2.(decimal.Decimal)), nil
	},
}

// multiplyOverflow is a convenience function that checks for overflow for int64 multiplication.
func multiplyOverflow(val1 int64, val2 int64) (any, error) {
	result := val1 * val2
	if val2 != 0 && result/val2 != val1 {
		return nil, fmt.Errorf("bigint out of range")
	}
	return result, nil
}
