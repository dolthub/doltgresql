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

	"github.com/cockroachdb/apd/v3"
	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"

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
		operand := operandInterface.(apd.Decimal)
		low := lowInterface.(apd.Decimal)
		high := highInterface.(apd.Decimal)
		if low.Cmp(&high) == 0 {
			return nil, errors.Errorf("lower bound cannot equal upper bound")
		}
		count := countInterface.(int32)
		if count <= 0 {
			return nil, errors.Errorf("count must be greater than zero")
		}
		if operand.Cmp(&high) == 0 {
			return count + 1, nil
		} else if operand.Cmp(&low) == 0 {
			return int32(1), nil
		}
		bucket := new(apd.Decimal)
		_, err := sql.DecimalCtx.Sub(bucket, &high, &low)
		if err != nil {
			return nil, err
		}
		_, err = sql.DecimalCtx.Quo(bucket, bucket, apd.New(int64(count), 0))
		if err != nil {
			return nil, err
		}
		result := new(apd.Decimal)
		_, err = sql.DecimalCtx.Sub(result, &operand, &low)
		if err != nil {
			return nil, err
		}
		_, err = sql.DecimalCtx.Sub(result, result, bucket)
		if err != nil {
			return nil, err
		}
		_, err = sql.DecimalCtx.Ceil(result, result)
		if err != nil {
			return nil, err
		}
		if result.Sign() < 0 {
			result = apd.New(0, 0)
		} else if c1 := apd.New(int64(count+1), 0); result.Cmp(c1) > 0 {
			result = c1
		}
		i64, err := result.Int64()
		if err != nil {
			return nil, err
		}
		return int32(i64), nil
	},
}
