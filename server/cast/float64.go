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

package cast

import (
	"fmt"
	"math"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/shopspring/decimal"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initFloat64 handles all casts that are built-in. This comprises only the "From" types.
func initFloat64() {
	float64Assignment()
}

// float64Assignment registers all assignment casts. This comprises only the "From" types.
func float64Assignment() {
	framework.MustAddAssignmentTypeCast(framework.TypeCast{
		FromType: pgtypes.Float64,
		ToType:   pgtypes.Float32,
		Function: func(ctx *sql.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			return float32(val.(float64)), nil
		},
	})
	framework.MustAddAssignmentTypeCast(framework.TypeCast{
		FromType: pgtypes.Float64,
		ToType:   pgtypes.Int16,
		Function: func(ctx *sql.Context, valInterface any, targetType pgtypes.DoltgresType) (any, error) {
			val := math.RoundToEven(valInterface.(float64))
			if val > 32767 || val < -32768 {
				return nil, fmt.Errorf("smallint out of range")
			}
			return int16(val), nil
		},
	})
	framework.MustAddAssignmentTypeCast(framework.TypeCast{
		FromType: pgtypes.Float64,
		ToType:   pgtypes.Int32,
		Function: func(ctx *sql.Context, valInterface any, targetType pgtypes.DoltgresType) (any, error) {
			val := math.RoundToEven(valInterface.(float64))
			if val > 2147483647 || val < -2147483648 {
				return nil, fmt.Errorf("integer out of range")
			}
			return int32(val), nil
		},
	})
	framework.MustAddAssignmentTypeCast(framework.TypeCast{
		FromType: pgtypes.Float64,
		ToType:   pgtypes.Int64,
		Function: func(ctx *sql.Context, valInterface any, targetType pgtypes.DoltgresType) (any, error) {
			val := math.RoundToEven(valInterface.(float64))
			if val > 9223372036854775807 || val < -9223372036854775808 {
				return nil, fmt.Errorf("bigint out of range")
			}
			return int64(val), nil
		},
	})
	framework.MustAddAssignmentTypeCast(framework.TypeCast{
		FromType: pgtypes.Float64,
		ToType:   pgtypes.Numeric,
		Function: func(ctx *sql.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			return pgtypes.GetNumericValueWithTypmod(decimal.NewFromFloat(val.(float64)), targetType.AttTypMod)
		},
	})
}
