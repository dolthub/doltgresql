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
	"strconv"

	"github.com/shopspring/decimal"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initInt64 handles all explicit and implicit casts that are built-in. This comprises only the "From" types.
func initInt64() {
	int64Explicit()
	int64Implicit()
}

// int64Explicit registers all explicit casts. This comprises only the "From" types.
func int64Explicit() {
	framework.MustAddExplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Int64,
		ToType:   pgtypes.BpChar,
		Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			str := strconv.FormatInt(val.(int64), 10)
			return handleCharExplicitCast(str, targetType)
		},
	})
	framework.MustAddExplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Int64,
		ToType:   pgtypes.Float32,
		Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			return float32(val.(int64)), nil
		},
	})
	framework.MustAddExplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Int64,
		ToType:   pgtypes.Float64,
		Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			return float64(val.(int64)), nil
		},
	})
	framework.MustAddExplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Int64,
		ToType:   pgtypes.Int16,
		Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			if val.(int64) > 32767 || val.(int64) < -32768 {
				return nil, fmt.Errorf("smallint out of range")
			}
			return int16(val.(int64)), nil
		},
	})
	framework.MustAddExplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Int64,
		ToType:   pgtypes.Int32,
		Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			if val.(int64) > 2147483647 || val.(int64) < -2147483648 {
				return nil, fmt.Errorf("integer out of range")
			}
			return int32(val.(int64)), nil
		},
	})
	framework.MustAddExplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Int64,
		ToType:   pgtypes.Int64,
		Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			return val, nil
		},
	})
	framework.MustAddExplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Int64,
		ToType:   pgtypes.Name,
		Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			str := strconv.FormatInt(val.(int64), 10)
			return handleCharExplicitCast(str, targetType)
		},
	})
	framework.MustAddExplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Int64,
		ToType:   pgtypes.Numeric,
		Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			return decimal.NewFromInt(val.(int64)), nil
		},
	})
	framework.MustAddExplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Int64,
		ToType:   pgtypes.Oid,
		Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			if val.(int64) > pgtypes.MaxUint32 || val.(int64) < 0 {
				return nil, errOutOfRange.New(targetType.String())
			}
			return uint32(val.(int64)), nil
		},
	})
	framework.MustAddExplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Int64,
		ToType:   pgtypes.Text,
		Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			return strconv.FormatInt(val.(int64), 10), nil
		},
	})
	framework.MustAddExplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Int64,
		ToType:   pgtypes.VarChar,
		Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			str := strconv.FormatInt(val.(int64), 10)
			return handleCharExplicitCast(str, targetType)
		},
	})
}

// int64Implicit registers all implicit casts. This comprises only the "From" types.
func int64Implicit() {
	framework.MustAddImplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Int64,
		ToType:   pgtypes.BpChar,
		Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			str := strconv.FormatInt(val.(int64), 10)
			return handleCharImplicitCast(str, targetType)
		},
	})
	framework.MustAddImplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Int64,
		ToType:   pgtypes.Float32,
		Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			return float32(val.(int64)), nil
		},
	})
	framework.MustAddImplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Int64,
		ToType:   pgtypes.Float64,
		Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			return float64(val.(int64)), nil
		},
	})
	framework.MustAddImplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Int64,
		ToType:   pgtypes.Int16,
		Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			if val.(int64) > 32767 || val.(int64) < -32768 {
				return nil, fmt.Errorf("smallint out of range")
			}
			return int16(val.(int64)), nil
		},
	})
	framework.MustAddImplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Int64,
		ToType:   pgtypes.Int32,
		Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			if val.(int64) > 2147483647 || val.(int64) < -2147483648 {
				return nil, fmt.Errorf("integer out of range")
			}
			return int32(val.(int64)), nil
		},
	})
	framework.MustAddImplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Int64,
		ToType:   pgtypes.Int64,
		Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			return val, nil
		},
	})
	framework.MustAddImplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Int64,
		ToType:   pgtypes.Name,
		Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			str := strconv.FormatInt(val.(int64), 10)
			return handleCharImplicitCast(str, targetType)
		},
	})
	framework.MustAddImplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Int64,
		ToType:   pgtypes.Numeric,
		Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			return decimal.NewFromInt(val.(int64)), nil
		},
	})
	framework.MustAddImplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Int64,
		ToType:   pgtypes.Oid,
		Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			if val.(int64) > pgtypes.MaxUint32 || val.(int64) < 0 {
				return nil, errOutOfRange.New(targetType.String())
			}
			return uint32(val.(int64)), nil
		},
	})
	framework.MustAddImplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Int64,
		ToType:   pgtypes.Text,
		Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			return strconv.FormatInt(val.(int64), 10), nil
		},
	})
	framework.MustAddImplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Int64,
		ToType:   pgtypes.VarChar,
		Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			str := strconv.FormatInt(val.(int64), 10)
			return handleCharImplicitCast(str, targetType)
		},
	})
}
