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
	"strconv"

	"github.com/shopspring/decimal"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initInt16 handles all explicit and implicit casts that are built-in. This comprises only the "From" types.
func initInt16() {
	int16Explicit()
	int16Implicit()
}

// int16Explicit registers all explicit casts. This comprises only the "From" types.
func int16Explicit() {
	framework.MustAddExplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Int16,
		ToType:   pgtypes.BpChar,
		Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			str := strconv.FormatInt(int64(val.(int16)), 10)
			return handleCharExplicitCast(str, targetType)
		},
	})
	framework.MustAddExplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Int16,
		ToType:   pgtypes.Float32,
		Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			return float32(val.(int16)), nil
		},
	})
	framework.MustAddExplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Int16,
		ToType:   pgtypes.Float64,
		Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			return float64(val.(int16)), nil
		},
	})
	framework.MustAddExplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Int16,
		ToType:   pgtypes.Int16,
		Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			return val, nil
		},
	})
	framework.MustAddExplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Int16,
		ToType:   pgtypes.Int32,
		Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			return int32(val.(int16)), nil
		},
	})
	framework.MustAddExplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Int16,
		ToType:   pgtypes.Int64,
		Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			return int64(val.(int16)), nil
		},
	})
	framework.MustAddExplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Int16,
		ToType:   pgtypes.Numeric,
		Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			return decimal.NewFromInt(int64(val.(int16))), nil
		},
	})
	framework.MustAddExplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Int16,
		ToType:   pgtypes.Text,
		Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			return strconv.FormatInt(int64(val.(int16)), 10), nil
		},
	})
	framework.MustAddExplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Int16,
		ToType:   pgtypes.VarChar,
		Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			str := strconv.FormatInt(int64(val.(int16)), 10)
			return handleCharExplicitCast(str, targetType)
		},
	})
}

// int16Implicit registers all implicit casts. This comprises only the "From" types.
func int16Implicit() {
	framework.MustAddImplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Int16,
		ToType:   pgtypes.BpChar,
		Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			str := strconv.FormatInt(int64(val.(int16)), 10)
			return handleCharImplicitCast(str, targetType)
		},
	})
	framework.MustAddImplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Int16,
		ToType:   pgtypes.Float32,
		Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			return float32(val.(int16)), nil
		},
	})
	framework.MustAddImplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Int16,
		ToType:   pgtypes.Float64,
		Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			return float64(val.(int16)), nil
		},
	})
	framework.MustAddImplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Int16,
		ToType:   pgtypes.Int16,
		Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			return val, nil
		},
	})
	framework.MustAddImplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Int16,
		ToType:   pgtypes.Int32,
		Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			return int32(val.(int16)), nil
		},
	})
	framework.MustAddImplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Int16,
		ToType:   pgtypes.Int64,
		Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			return int64(val.(int16)), nil
		},
	})
	framework.MustAddImplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Int16,
		ToType:   pgtypes.Numeric,
		Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			return decimal.NewFromInt(int64(val.(int16))), nil
		},
	})
	framework.MustAddImplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Int16,
		ToType:   pgtypes.Text,
		Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			return strconv.FormatInt(int64(val.(int16)), 10), nil
		},
	})
	framework.MustAddImplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Int16,
		ToType:   pgtypes.VarChar,
		Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			str := strconv.FormatInt(int64(val.(int16)), 10)
			return handleCharImplicitCast(str, targetType)
		},
	})
}
