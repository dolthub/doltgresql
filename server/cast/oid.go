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

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initOid handles all explicit and implicit casts that are built-in. This comprises only the "From" types.
func initOid() {
	oidExplicit()
	oidImplicit()
}

// oidExplicit registers all explicit casts. This comprises only the "From" types.
func oidExplicit() {
	framework.MustAddExplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Oid,
		ToType:   pgtypes.BpChar,
		Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			str := strconv.FormatInt(int64(val.(uint32)), 10)
			return handleCharExplicitCast(str, targetType)
		},
	})
	framework.MustAddExplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Oid,
		ToType:   pgtypes.Float32,
		Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			return nil, errCannotCast.New(pgtypes.Oid.String(), targetType.String())
		},
	})
	framework.MustAddExplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Oid,
		ToType:   pgtypes.Float64,
		Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			return nil, errCannotCast.New(pgtypes.Oid.String(), targetType.String())
		},
	})
	framework.MustAddExplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Oid,
		ToType:   pgtypes.Int16,
		Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			return nil, errCannotCast.New(pgtypes.Oid.String(), targetType.String())
		},
	})
	framework.MustAddExplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Oid,
		ToType:   pgtypes.Int32,
		Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			// Will return -1 for uint32 values greater than 2147483647
			return int32(val.(uint32)), nil
		},
	})
	framework.MustAddExplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Oid,
		ToType:   pgtypes.Int64,
		Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			return int64(val.(uint32)), nil
		},
	})
	framework.MustAddExplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Oid,
		ToType:   pgtypes.Name,
		Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			str := strconv.FormatInt(int64(val.(uint32)), 10)
			return handleCharExplicitCast(str, targetType)
		},
	})
	framework.MustAddExplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Oid,
		ToType:   pgtypes.Numeric,
		Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			return nil, errCannotCast.New(pgtypes.Oid.String(), targetType.String())
		},
	})
	framework.MustAddExplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Oid,
		ToType:   pgtypes.Oid,
		Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			return val, nil
		},
	})
	framework.MustAddExplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Oid,
		ToType:   pgtypes.Text,
		Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			return strconv.FormatInt(int64(val.(uint32)), 10), nil
		},
	})
	framework.MustAddExplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Oid,
		ToType:   pgtypes.VarChar,
		Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			str := strconv.FormatInt(int64(val.(uint32)), 10)
			return handleCharExplicitCast(str, targetType)
		},
	})
}

// oidImplicit registers all implicit casts. This comprises only the "From" types.
func oidImplicit() {
	framework.MustAddImplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Oid,
		ToType:   pgtypes.BpChar,
		Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			str := strconv.FormatInt(int64(val.(uint32)), 10)
			return handleCharImplicitCast(str, targetType)
		},
	})
	framework.MustAddImplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Oid,
		ToType:   pgtypes.Float32,
		Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			return nil, errCannotCast.New(pgtypes.Oid.String(), targetType.String())
		},
	})
	framework.MustAddImplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Oid,
		ToType:   pgtypes.Float64,
		Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			return nil, errCannotCast.New(pgtypes.Oid.String(), targetType.String())
		},
	})
	framework.MustAddImplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Oid,
		ToType:   pgtypes.Int16,
		Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			return nil, errCannotCast.New(pgtypes.Oid.String(), targetType.String())
		},
	})
	framework.MustAddImplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Oid,
		ToType:   pgtypes.Int32,
		Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			// Will return -1 for uint32 values greater than 2147483647
			return int32(val.(uint32)), nil
		},
	})
	framework.MustAddImplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Oid,
		ToType:   pgtypes.Int64,
		Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			return int64(val.(uint32)), nil
		},
	})
	framework.MustAddImplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Oid,
		ToType:   pgtypes.Name,
		Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			str := strconv.FormatInt(int64(val.(uint32)), 10)
			return handleCharExplicitCast(str, targetType)
		},
	})
	framework.MustAddImplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Oid,
		ToType:   pgtypes.Numeric,
		Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			return nil, errCannotCast.New(pgtypes.Oid.String(), targetType.String())
		},
	})
	framework.MustAddImplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Oid,
		ToType:   pgtypes.Oid,
		Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			return val, nil
		},
	})
	framework.MustAddImplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Oid,
		ToType:   pgtypes.Text,
		Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			return strconv.FormatInt(int64(val.(uint32)), 10), nil
		},
	})
	framework.MustAddImplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Oid,
		ToType:   pgtypes.VarChar,
		Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			str := strconv.FormatInt(int64(val.(uint32)), 10)
			return handleCharImplicitCast(str, targetType)
		},
	})
}
