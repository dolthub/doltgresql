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
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initBool handles all explicit and implicit casts that are built-in. This comprises only the "From" types.
func initBool() {
	boolExplicit()
	boolImplicit()
}

// boolExplicit registers all explicit casts. This comprises only the "From" types.
func boolExplicit() {
	framework.MustAddExplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Bool,
		ToType:   pgtypes.Bool,
		Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			return val, nil
		},
	})
	framework.MustAddExplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Bool,
		ToType:   pgtypes.BpChar,
		Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			str := "false"
			if val.(bool) {
				str = "true"
			}
			return handleCharExplicitCast(str, targetType)
		},
	})
	framework.MustAddExplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Bool,
		ToType:   pgtypes.Int32,
		Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			if val.(bool) {
				return int32(1), nil
			} else {
				return int32(0), nil
			}
		},
	})
	framework.MustAddExplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Bool,
		ToType:   pgtypes.Name,
		Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			str := "false"
			if val.(bool) {
				str = "true"
			}
			return handleCharExplicitCast(str, targetType)
		},
	})
	framework.MustAddExplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Bool,
		ToType:   pgtypes.Oid,
		Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			if val.(bool) {
				return uint32(1), nil
			} else {
				return uint32(0), nil
			}
		},
	})
	framework.MustAddExplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Bool,
		ToType:   pgtypes.Text,
		Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			if val.(bool) {
				return "true", nil
			} else {
				return "false", nil
			}
		},
	})
	framework.MustAddExplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Bool,
		ToType:   pgtypes.VarChar,
		Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			str := "false"
			if val.(bool) {
				str = "true"
			}
			return handleCharExplicitCast(str, targetType)
		},
	})
	framework.MustAddExplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Bool,
		ToType:   pgtypes.Xid,
		Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			if val.(bool) {
				return int32(1), nil
			} else {
				return int32(0), nil
			}
		},
	})
}

// boolImplicit registers all implicit casts. This comprises only the "From" types.
func boolImplicit() {
	framework.MustAddImplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Bool,
		ToType:   pgtypes.Bool,
		Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			return val, nil
		},
	})
	framework.MustAddImplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Bool,
		ToType:   pgtypes.BpChar,
		Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			str := "false"
			if val.(bool) {
				str = "true"
			}
			return handleCharImplicitCast(str, targetType)
		},
	})
	framework.MustAddImplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Bool,
		ToType:   pgtypes.Name,
		Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			str := "false"
			if val.(bool) {
				str = "true"
			}
			return handleCharImplicitCast(str, targetType)
		},
	})
	framework.MustAddImplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Bool,
		ToType:   pgtypes.Text,
		Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			if val.(bool) {
				return "true", nil
			} else {
				return "false", nil
			}
		},
	})
	framework.MustAddImplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Bool,
		ToType:   pgtypes.VarChar,
		Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			str := "false"
			if val.(bool) {
				str = "true"
			}
			return handleCharImplicitCast(str, targetType)
		},
	})
}
