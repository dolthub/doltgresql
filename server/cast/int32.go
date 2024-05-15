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

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/shopspring/decimal"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initInt32 handles all explicit and implicit casts that are built-in. This comprises only the "From" types.
func initInt32() {
	int32Explicit()
	int32Implicit()
}

// int32Explicit registers all explicit casts. This comprises only the "From" types.
func int32Explicit() {
	framework.MustAddExplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Int32,
		ToType:   pgtypes.BpChar,
		Function: func(ctx *sql.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			str := strconv.FormatInt(int64(val.(int32)), 10)
			return handleCharExplicitCast(str, targetType)
		},
	})
	framework.MustAddExplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Int32,
		ToType:   pgtypes.Float32,
		Function: func(ctx *sql.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			return float32(val.(int32)), nil
		},
	})
	framework.MustAddExplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Int32,
		ToType:   pgtypes.Float64,
		Function: func(ctx *sql.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			return float64(val.(int32)), nil
		},
	})
	framework.MustAddExplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Int32,
		ToType:   pgtypes.Int16,
		Function: func(ctx *sql.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			if val.(int32) > 32767 || val.(int32) < -32768 {
				return nil, fmt.Errorf("smallint out of range")
			}
			return int16(val.(int32)), nil
		},
	})
	framework.MustAddExplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Int32,
		ToType:   pgtypes.Int32,
		Function: func(ctx *sql.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			return val, nil
		},
	})
	framework.MustAddExplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Int32,
		ToType:   pgtypes.Int64,
		Function: func(ctx *sql.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			return int64(val.(int32)), nil
		},
	})
	framework.MustAddExplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Int32,
		ToType:   pgtypes.Name,
		Function: func(ctx *sql.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			str := strconv.FormatInt(int64(val.(int32)), 10)
			return handleCharExplicitCast(str, targetType)
		},
	})
	framework.MustAddExplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Int32,
		ToType:   pgtypes.Numeric,
		Function: func(ctx *sql.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			return decimal.NewFromInt(int64(val.(int32))), nil
		},
	})
	framework.MustAddExplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Int32,
		ToType:   pgtypes.Oid,
		Function: func(ctx *sql.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			return uint32(val.(int32)), nil
		},
	})
	framework.MustAddExplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Int32,
		ToType:   pgtypes.Text,
		Function: func(ctx *sql.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			return strconv.FormatInt(int64(val.(int32)), 10), nil
		},
	})
	framework.MustAddExplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Int32,
		ToType:   pgtypes.VarChar,
		Function: func(ctx *sql.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			str := strconv.FormatInt(int64(val.(int32)), 10)
			return handleCharExplicitCast(str, targetType)
		},
	})
}

// int32Implicit registers all implicit casts. This comprises only the "From" types.
func int32Implicit() {
	framework.MustAddImplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Int32,
		ToType:   pgtypes.BpChar,
		Function: func(ctx *sql.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			str := strconv.FormatInt(int64(val.(int32)), 10)
			return handleCharImplicitCast(str, targetType)
		},
	})
	framework.MustAddImplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Int32,
		ToType:   pgtypes.Float32,
		Function: func(ctx *sql.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			return float32(val.(int32)), nil
		},
	})
	framework.MustAddImplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Int32,
		ToType:   pgtypes.Float64,
		Function: func(ctx *sql.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			return float64(val.(int32)), nil
		},
	})
	framework.MustAddImplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Int32,
		ToType:   pgtypes.Int16,
		Function: func(ctx *sql.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			if val.(int32) > 32767 || val.(int32) < -32768 {
				return nil, fmt.Errorf("smallint out of range")
			}
			return int16(val.(int32)), nil
		},
	})
	framework.MustAddImplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Int32,
		ToType:   pgtypes.Int32,
		Function: func(ctx *sql.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			return val, nil
		},
	})
	framework.MustAddImplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Int32,
		ToType:   pgtypes.Int64,
		Function: func(ctx *sql.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			return int64(val.(int32)), nil
		},
	})
	framework.MustAddImplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Int32,
		ToType:   pgtypes.Name,
		Function: func(ctx *sql.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			str := strconv.FormatInt(int64(val.(int32)), 10)
			return handleCharImplicitCast(str, targetType)
		},
	})
	framework.MustAddImplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Int32,
		ToType:   pgtypes.Numeric,
		Function: func(ctx *sql.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			return decimal.NewFromInt(int64(val.(int32))), nil
		},
	})
	framework.MustAddImplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Int32,
		ToType:   pgtypes.Oid,
		Function: func(ctx *sql.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			return uint32(val.(int32)), nil
		},
	})
	framework.MustAddImplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Int32,
		ToType:   pgtypes.Text,
		Function: func(ctx *sql.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			return strconv.FormatInt(int64(val.(int32)), 10), nil
		},
	})
	framework.MustAddImplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Int32,
		ToType:   pgtypes.VarChar,
		Function: func(ctx *sql.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			str := strconv.FormatInt(int64(val.(int32)), 10)
			return handleCharImplicitCast(str, targetType)
		},
	})
}
