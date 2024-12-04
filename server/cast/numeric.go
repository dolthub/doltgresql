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

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/shopspring/decimal"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initNumeric handles all casts that are built-in. This comprises only the "From" types.
func initNumeric() {
	numericAssignment()
	numericImplicit()
}

// numericAssignment registers all assignment casts. This comprises only the "From" types.
func numericAssignment() {
	framework.MustAddAssignmentTypeCast(framework.TypeCast{
		FromType: pgtypes.Numeric,
		ToType:   pgtypes.Int16,
		Function: func(ctx *sql.Context, val any, targetType *pgtypes.DoltgresType) (any, error) {
			d := val.(decimal.Decimal)
			if d.LessThan(pgtypes.NumericValueMinInt16) || d.GreaterThan(pgtypes.NumericValueMaxInt16) {
				return nil, fmt.Errorf("smallint out of range")
			}
			return int16(d.IntPart()), nil
		},
	})
	framework.MustAddAssignmentTypeCast(framework.TypeCast{
		FromType: pgtypes.Numeric,
		ToType:   pgtypes.Int32,
		Function: func(ctx *sql.Context, val any, targetType *pgtypes.DoltgresType) (any, error) {
			d := val.(decimal.Decimal)
			if d.LessThan(pgtypes.NumericValueMinInt32) || d.GreaterThan(pgtypes.NumericValueMaxInt32) {
				return nil, fmt.Errorf("integer out of range")
			}
			return int32(d.IntPart()), nil
		},
	})
	framework.MustAddAssignmentTypeCast(framework.TypeCast{
		FromType: pgtypes.Numeric,
		ToType:   pgtypes.Int64,
		Function: func(ctx *sql.Context, val any, targetType *pgtypes.DoltgresType) (any, error) {
			d := val.(decimal.Decimal)
			if d.LessThan(pgtypes.NumericValueMinInt64) || d.GreaterThan(pgtypes.NumericValueMaxInt64) {
				return nil, fmt.Errorf("bigint out of range")
			}
			return int64(d.IntPart()), nil
		},
	})
}

// numericImplicit registers all implicit casts. This comprises only the "From" types.
func numericImplicit() {
	framework.MustAddImplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Numeric,
		ToType:   pgtypes.Float32,
		Function: func(ctx *sql.Context, val any, targetType *pgtypes.DoltgresType) (any, error) {
			f, _ := val.(decimal.Decimal).Float64()
			return float32(f), nil
		},
	})
	framework.MustAddImplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Numeric,
		ToType:   pgtypes.Float64,
		Function: func(ctx *sql.Context, val any, targetType *pgtypes.DoltgresType) (any, error) {
			f, _ := val.(decimal.Decimal).Float64()
			return f, nil
		},
	})
	framework.MustAddImplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Numeric,
		ToType:   pgtypes.Numeric,
		Function: func(ctx *sql.Context, val any, targetType *pgtypes.DoltgresType) (any, error) {
			return pgtypes.GetNumericValueWithTypmod(val.(decimal.Decimal), targetType.AttTypMod)
		},
	})
}
