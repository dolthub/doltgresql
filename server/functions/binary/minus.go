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

	"github.com/shopspring/decimal"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// These functions can be gathered using the following query from a Postgres 15 instance:
// SELECT * FROM pg_operator o WHERE o.oprname = '-' ORDER BY o.oprcode::varchar;

// initBinaryMinus registers the functions to the catalog.
func initBinaryMinus() {
	framework.RegisterBinaryFunction(framework.Operator_BinaryMinus, float4mi)
	framework.RegisterBinaryFunction(framework.Operator_BinaryMinus, float48mi)
	framework.RegisterBinaryFunction(framework.Operator_BinaryMinus, float8mi)
	framework.RegisterBinaryFunction(framework.Operator_BinaryMinus, float84mi)
	framework.RegisterBinaryFunction(framework.Operator_BinaryMinus, int2mi)
	framework.RegisterBinaryFunction(framework.Operator_BinaryMinus, int24mi)
	framework.RegisterBinaryFunction(framework.Operator_BinaryMinus, int28mi)
	framework.RegisterBinaryFunction(framework.Operator_BinaryMinus, int4mi)
	framework.RegisterBinaryFunction(framework.Operator_BinaryMinus, int42mi)
	framework.RegisterBinaryFunction(framework.Operator_BinaryMinus, int48mi)
	framework.RegisterBinaryFunction(framework.Operator_BinaryMinus, int8mi)
	framework.RegisterBinaryFunction(framework.Operator_BinaryMinus, int82mi)
	framework.RegisterBinaryFunction(framework.Operator_BinaryMinus, int84mi)
	framework.RegisterBinaryFunction(framework.Operator_BinaryMinus, numeric_sub)
}

// float4mi represents the PostgreSQL function of the same name, taking the same parameters.
var float4mi = framework.Function2{
	Name:       "float4mi",
	Return:     pgtypes.Float32,
	Parameters: []pgtypes.DoltgresType{pgtypes.Float32, pgtypes.Float32},
	Callable: func(ctx framework.Context, val1 any, val2 any) (any, error) {
		if val1 == nil || val2 == nil {
			return nil, nil
		}
		return val1.(float32) - val2.(float32), nil
	},
}

// float48mi represents the PostgreSQL function of the same name, taking the same parameters.
var float48mi = framework.Function2{
	Name:       "float48mi",
	Return:     pgtypes.Float64,
	Parameters: []pgtypes.DoltgresType{pgtypes.Float32, pgtypes.Float64},
	Callable: func(ctx framework.Context, val1 any, val2 any) (any, error) {
		if val1 == nil || val2 == nil {
			return nil, nil
		}
		return float64(val1.(float32)) - val2.(float64), nil
	},
}

// float8mi represents the PostgreSQL function of the same name, taking the same parameters.
var float8mi = framework.Function2{
	Name:       "float8mi",
	Return:     pgtypes.Float64,
	Parameters: []pgtypes.DoltgresType{pgtypes.Float64, pgtypes.Float64},
	Callable: func(ctx framework.Context, val1 any, val2 any) (any, error) {
		if val1 == nil || val2 == nil {
			return nil, nil
		}
		return val1.(float64) - val2.(float64), nil
	},
}

// float84mi represents the PostgreSQL function of the same name, taking the same parameters.
var float84mi = framework.Function2{
	Name:       "float84mi",
	Return:     pgtypes.Float64,
	Parameters: []pgtypes.DoltgresType{pgtypes.Float64, pgtypes.Float32},
	Callable: func(ctx framework.Context, val1 any, val2 any) (any, error) {
		if val1 == nil || val2 == nil {
			return nil, nil
		}
		return val1.(float64) - float64(val2.(float32)), nil
	},
}

// int2mi represents the PostgreSQL function of the same name, taking the same parameters.
var int2mi = framework.Function2{
	Name:       "int2mi",
	Return:     pgtypes.Int16,
	Parameters: []pgtypes.DoltgresType{pgtypes.Int16, pgtypes.Int16},
	Callable: func(ctx framework.Context, val1 any, val2 any) (any, error) {
		if val1 == nil || val2 == nil {
			return nil, nil
		}
		result := int64(val1.(int16)) - int64(val2.(int16))
		if result > math.MaxInt16 || result < math.MinInt16 {
			return nil, fmt.Errorf("smallint out of range")
		}
		return int16(result), nil
	},
}

// int24mi represents the PostgreSQL function of the same name, taking the same parameters.
var int24mi = framework.Function2{
	Name:       "int24mi",
	Return:     pgtypes.Int32,
	Parameters: []pgtypes.DoltgresType{pgtypes.Int16, pgtypes.Int32},
	Callable: func(ctx framework.Context, val1 any, val2 any) (any, error) {
		if val1 == nil || val2 == nil {
			return nil, nil
		}
		result := int64(val1.(int16)) - int64(val2.(int32))
		if result > math.MaxInt16 || result < math.MinInt16 {
			return nil, fmt.Errorf("integer out of range")
		}
		return int32(result), nil
	},
}

// int28mi represents the PostgreSQL function of the same name, taking the same parameters.
var int28mi = framework.Function2{
	Name:       "int28mi",
	Return:     pgtypes.Int64,
	Parameters: []pgtypes.DoltgresType{pgtypes.Int16, pgtypes.Int64},
	Callable: func(ctx framework.Context, val1 any, val2 any) (any, error) {
		if val1 == nil || val2 == nil {
			return nil, nil
		}
		return minusOverflow(int64(val1.(int16)), val2.(int64))
	},
}

// int4mi represents the PostgreSQL function of the same name, taking the same parameters.
var int4mi = framework.Function2{
	Name:       "int4mi",
	Return:     pgtypes.Int32,
	Parameters: []pgtypes.DoltgresType{pgtypes.Int32, pgtypes.Int32},
	Callable: func(ctx framework.Context, val1 any, val2 any) (any, error) {
		if val1 == nil || val2 == nil {
			return nil, nil
		}
		result := int64(val1.(int32)) - int64(val2.(int32))
		if result > math.MaxInt32 || result < math.MinInt32 {
			return nil, fmt.Errorf("integer out of range")
		}
		return int32(result), nil
	},
}

// int42mi represents the PostgreSQL function of the same name, taking the same parameters.
var int42mi = framework.Function2{
	Name:       "int42mi",
	Return:     pgtypes.Int32,
	Parameters: []pgtypes.DoltgresType{pgtypes.Int32, pgtypes.Int16},
	Callable: func(ctx framework.Context, val1 any, val2 any) (any, error) {
		if val1 == nil || val2 == nil {
			return nil, nil
		}
		result := int64(val1.(int32)) - int64(val2.(int16))
		if result > math.MaxInt32 || result < math.MinInt32 {
			return nil, fmt.Errorf("integer out of range")
		}
		return int32(result), nil
	},
}

// int48mi represents the PostgreSQL function of the same name, taking the same parameters.
var int48mi = framework.Function2{
	Name:       "int48mi",
	Return:     pgtypes.Int64,
	Parameters: []pgtypes.DoltgresType{pgtypes.Int32, pgtypes.Int64},
	Callable: func(ctx framework.Context, val1 any, val2 any) (any, error) {
		if val1 == nil || val2 == nil {
			return nil, nil
		}
		return minusOverflow(int64(val1.(int32)), val2.(int64))
	},
}

// int8mi represents the PostgreSQL function of the same name, taking the same parameters.
var int8mi = framework.Function2{
	Name:       "int8mi",
	Return:     pgtypes.Int64,
	Parameters: []pgtypes.DoltgresType{pgtypes.Int64, pgtypes.Int64},
	Callable: func(ctx framework.Context, val1 any, val2 any) (any, error) {
		if val1 == nil || val2 == nil {
			return nil, nil
		}
		return minusOverflow(val1.(int64), val2.(int64))
	},
}

// int82mi represents the PostgreSQL function of the same name, taking the same parameters.
var int82mi = framework.Function2{
	Name:       "int82mi",
	Return:     pgtypes.Int64,
	Parameters: []pgtypes.DoltgresType{pgtypes.Int64, pgtypes.Int16},
	Callable: func(ctx framework.Context, val1 any, val2 any) (any, error) {
		if val1 == nil || val2 == nil {
			return nil, nil
		}
		return minusOverflow(val1.(int64), int64(val2.(int16)))
	},
}

// int84mi represents the PostgreSQL function of the same name, taking the same parameters.
var int84mi = framework.Function2{
	Name:       "int84mi",
	Return:     pgtypes.Int64,
	Parameters: []pgtypes.DoltgresType{pgtypes.Int64, pgtypes.Int32},
	Callable: func(ctx framework.Context, val1 any, val2 any) (any, error) {
		if val1 == nil || val2 == nil {
			return nil, nil
		}
		return minusOverflow(val1.(int64), int64(val2.(int32)))
	},
}

// numeric_sub represents the PostgreSQL function of the same name, taking the same parameters.
var numeric_sub = framework.Function2{
	Name:       "numeric_sub",
	Return:     pgtypes.Numeric,
	Parameters: []pgtypes.DoltgresType{pgtypes.Numeric, pgtypes.Numeric},
	Callable: func(ctx framework.Context, val1 any, val2 any) (any, error) {
		if val1 == nil || val2 == nil {
			return nil, nil
		}
		return val1.(decimal.Decimal).Sub(val2.(decimal.Decimal)), nil
	},
}

// minusOverflow is a convenience function that checks for overflow for int64 subtraction.
func minusOverflow(val1 int64, val2 int64) (any, error) {
	if val2 > 0 {
		if val1 < math.MinInt64+val2 {
			return nil, fmt.Errorf("bigint out of range")
		}
	} else {
		if val1 > math.MaxInt64+val2 {
			return nil, fmt.Errorf("bigint out of range")
		}
	}
	return val1 - val2, nil
}
