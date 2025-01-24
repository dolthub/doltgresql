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
	"math"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/shopspring/decimal"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initWidthBucket registers the functions to the catalog.
func initWidthBucket() {
	framework.RegisterFunction(width_bucket_float64_float64_float64_int64)
	framework.RegisterFunction(width_bucket_numeric_numeric_numeric_int64)
}

// width_bucket_float64_float64_float64_int64 represents the PostgreSQL function of the same name, taking the same parameters.
var width_bucket_float64_float64_float64_int64 = framework.Function4{
	Name:       "width_bucket",
	Return:     pgtypes.Int32,
	Parameters: [4]*pgtypes.DoltgresType{pgtypes.Float64, pgtypes.Float64, pgtypes.Float64, pgtypes.Int32},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [5]*pgtypes.DoltgresType, operandInterface any, lowInterface any, highInterface any, countInterface any) (any, error) {
		operand := operandInterface.(float64)
		low := lowInterface.(float64)
		high := highInterface.(float64)
		if low == high {
			return nil, errors.Errorf("lower bound cannot equal upper bound")
		}
		count := countInterface.(int32)
		if count <= 0 {
			return nil, errors.Errorf("count must be greater than zero")
		}
		if operand == high {
			return count + 1, nil
		} else if operand == low {
			return int32(1), nil
		}
		bucket := (high - low) / float64(count)
		result := math.Ceil((operand - low) / bucket)
		if result < 0 {
			result = 0
		} else if result > float64(count+1) {
			result = float64(count + 1)
		}
		return int32(result), nil
	},
}

// width_bucket_numeric_numeric_numeric_int64 represents the PostgreSQL function of the same name, taking the same parameters.
var width_bucket_numeric_numeric_numeric_int64 = framework.Function4{
	Name:       "width_bucket",
	Return:     pgtypes.Int32,
	Parameters: [4]*pgtypes.DoltgresType{pgtypes.Numeric, pgtypes.Numeric, pgtypes.Numeric, pgtypes.Int32},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [5]*pgtypes.DoltgresType, operandInterface any, lowInterface any, highInterface any, countInterface any) (any, error) {
		operand := operandInterface.(decimal.Decimal)
		low := lowInterface.(decimal.Decimal)
		high := highInterface.(decimal.Decimal)
		if low.Cmp(high) == 0 {
			return nil, errors.Errorf("lower bound cannot equal upper bound")
		}
		count := countInterface.(int32)
		if count <= 0 {
			return nil, errors.Errorf("count must be greater than zero")
		}
		if operand.Equal(high) {
			return count + 1, nil
		} else if operand.Equal(low) {
			return int32(1), nil
		}
		bucket := high.Sub(low).Div(decimal.NewFromInt(int64(count)))
		result := operand.Sub(low).Div(bucket).Ceil()
		if result.LessThan(decimal.Zero) {
			result = decimal.Zero
		} else if result.GreaterThan(decimal.NewFromInt(int64(count + 1))) {
			result = decimal.NewFromInt(int64(count + 1))
		}
		i64 := result.IntPart()
		return int32(i64), nil
	},
}
