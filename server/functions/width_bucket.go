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
	"math"

	"github.com/shopspring/decimal"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// init registers the functions to the catalog.
func init() {
	framework.RegisterFunction(width_bucket_float64_float64_float64_int64)
	framework.RegisterFunction(width_bucket_numeric_numeric_numeric_int64)
}

// width_bucket_float64_float64_float64_int64 represents the PostgreSQL function of the same name, taking the same parameters.
var width_bucket_float64_float64_float64_int64 = framework.Function4{
	Name:       "width_bucket",
	Return:     pgtypes.Int64,
	Parameters: []pgtypes.DoltgresType{pgtypes.Float64, pgtypes.Float64, pgtypes.Float64, pgtypes.Int64},
	Callable: func(ctx framework.Context, operandInterface any, lowInterface any, highInterface any, countInterface any) (any, error) {
		if operandInterface == nil || lowInterface == nil || highInterface == nil || countInterface == nil {
			return nil, nil
		}
		operand := operandInterface.(float64)
		low := lowInterface.(float64)
		high := highInterface.(float64)
		count := countInterface.(int64)
		if count <= 0 {
			return nil, fmt.Errorf("count must be greater than zero")
		}
		bucket := (high - low) / float64(count)
		result := int64(math.Ceil((operand - low) / bucket))
		if result < 0 {
			result = 0
		} else if result > count+1 {
			result = count + 1
		}
		return result, nil
	},
}

// width_bucket_numeric_numeric_numeric_int64 represents the PostgreSQL function of the same name, taking the same parameters.
var width_bucket_numeric_numeric_numeric_int64 = framework.Function4{
	Name:       "width_bucket",
	Return:     pgtypes.Int64,
	Parameters: []pgtypes.DoltgresType{pgtypes.Numeric, pgtypes.Numeric, pgtypes.Numeric, pgtypes.Int64},
	Callable: func(ctx framework.Context, operandInterface any, lowInterface any, highInterface any, countInterface any) (any, error) {
		if operandInterface == nil || lowInterface == nil || highInterface == nil || countInterface == nil {
			return nil, nil
		}
		operand := operandInterface.(decimal.Decimal)
		low := lowInterface.(decimal.Decimal)
		high := highInterface.(decimal.Decimal)
		count := countInterface.(int64)
		if count <= 0 {
			return nil, fmt.Errorf("count must be greater than zero")
		}
		bucket := high.Sub(low).Div(decimal.NewFromInt(count))
		result := operand.Sub(low).Div(bucket).Ceil()
		if result.LessThan(decimal.Zero) {
			result = decimal.Zero
		} else if result.GreaterThan(decimal.NewFromInt(count + 1)) {
			result = decimal.NewFromInt(count + 1)
		}
		return result, nil
	},
}
