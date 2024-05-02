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

// initXid handles all explicit and implicit casts that are built-in. This comprises only the "From" types.
func initXid() {
	xidExplicit()
	xidImplicit()
}

// xidExplicit registers all explicit casts. This comprises only the "From" types.
func xidExplicit() {
	framework.MustAddExplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Xid,
		ToType:   pgtypes.BpChar,
		Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			str := strconv.FormatInt(int64(val.(uint32)), 10)
			return handleCharExplicitCast(str, targetType)
		},
	})
	// framework.MustAddExplicitTypeCast(framework.TypeCast{
	// 	FromType: pgtypes.Xid,
	// 	ToType:   pgtypes.Float32,
	// 	Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
	// 		return float32(val.(uint32)), nil
	// 	},
	// })
	// framework.MustAddExplicitTypeCast(framework.TypeCast{
	// 	FromType: pgtypes.Xid,
	// 	ToType:   pgtypes.Float64,
	// 	Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
	// 		return float64(val.(uint32)), nil
	// 	},
	// })
	// framework.MustAddExplicitTypeCast(framework.TypeCast{
	// 	FromType: pgtypes.Xid,
	// 	ToType:   pgtypes.Int16,
	// 	Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
	// 		if val.(uint32) > 32767 {
	// 			return nil, fmt.Errorf("smallint out of range")
	// 		}
	// 		return int16(val.(uint32)), nil
	// 	},
	// })
	// framework.MustAddExplicitTypeCast(framework.TypeCast{
	// 	FromType: pgtypes.Xid,
	// 	ToType:   pgtypes.Int32,
	// 	Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
	// 		return int32(val.(uint32)), nil
	// 	},
	// })
	// framework.MustAddExplicitTypeCast(framework.TypeCast{
	// 	FromType: pgtypes.Xid,
	// 	ToType:   pgtypes.Int64,
	// 	Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
	// 		return int64(val.(uint32)), nil
	// 	},
	// })
	// framework.MustAddExplicitTypeCast(framework.TypeCast{
	// 	FromType: pgtypes.Xid,
	// 	ToType:   pgtypes.Numeric,
	// 	Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
	// 		return decimal.NewFromInt(int64(val.(uint32))), nil
	// 	},
	// })
	framework.MustAddExplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Xid,
		ToType:   pgtypes.Name,
		Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			return strconv.FormatInt(int64(val.(uint32)), 10), nil
		},
	})
	// framework.MustAddExplicitTypeCast(framework.TypeCast{
	// 	FromType: pgtypes.Xid,
	// 	ToType:   pgtypes.Oid,
	// 	Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
	// 		return val, nil
	// 	},
	// })
	framework.MustAddExplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Xid,
		ToType:   pgtypes.Text,
		Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			return strconv.FormatInt(int64(val.(uint32)), 10), nil
		},
	})
	framework.MustAddExplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Xid,
		ToType:   pgtypes.VarChar,
		Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			str := strconv.FormatInt(int64(val.(uint32)), 10)
			return handleCharExplicitCast(str, targetType)
		},
	})
	framework.MustAddExplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Xid,
		ToType:   pgtypes.Xid,
		Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			return val, nil
		},
	})
}

// xidImplicit registers all implicit casts. This comprises only the "From" types.
func xidImplicit() {
	framework.MustAddImplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Xid,
		ToType:   pgtypes.BpChar,
		Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			str := strconv.FormatInt(int64(val.(uint32)), 10)
			return handleCharImplicitCast(str, targetType)
		},
	})
	// framework.MustAddImplicitTypeCast(framework.TypeCast{
	// 	FromType: pgtypes.Xid,
	// 	ToType:   pgtypes.Float32,
	// 	Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
	// 		return float32(val.(uint32)), nil
	// 	},
	// })
	// framework.MustAddImplicitTypeCast(framework.TypeCast{
	// 	FromType: pgtypes.Xid,
	// 	ToType:   pgtypes.Float64,
	// 	Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
	// 		return float64(val.(uint32)), nil
	// 	},
	// })
	// framework.MustAddImplicitTypeCast(framework.TypeCast{
	// 	FromType: pgtypes.Xid,
	// 	ToType:   pgtypes.Int16,
	// 	Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
	// 		if val.(uint32) > 32767 {
	// 			return nil, fmt.Errorf("smallint out of range")
	// 		}
	// 		return int16(val.(uint32)), nil
	// 	},
	// })
	// framework.MustAddImplicitTypeCast(framework.TypeCast{
	// 	FromType: pgtypes.Xid,
	// 	ToType:   pgtypes.Int32,
	// 	Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
	// 		return int32(val.(uint32)), nil
	// 	},
	// })
	// framework.MustAddImplicitTypeCast(framework.TypeCast{
	// 	FromType: pgtypes.Xid,
	// 	ToType:   pgtypes.Int64,
	// 	Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
	// 		return int64(val.(uint32)), nil
	// 	},
	// })
	// framework.MustAddImplicitTypeCast(framework.TypeCast{
	// 	FromType: pgtypes.Xid,
	// 	ToType:   pgtypes.Numeric,
	// 	Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
	// 		return decimal.NewFromInt(int64(val.(uint32))), nil
	// 	},
	// })
	framework.MustAddImplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Xid,
		ToType:   pgtypes.Name,
		Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			return strconv.FormatInt(int64(val.(uint32)), 10), nil
		},
	})
	// framework.MustAddImplicitTypeCast(framework.TypeCast{
	// 	FromType: pgtypes.Xid,
	// 	ToType:   pgtypes.Oid,
	// 	Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
	// 		return val, nil
	// 	},
	// })
	framework.MustAddImplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Xid,
		ToType:   pgtypes.Text,
		Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			return strconv.FormatInt(int64(val.(uint32)), 10), nil
		},
	})
	framework.MustAddImplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Xid,
		ToType:   pgtypes.VarChar,
		Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			str := strconv.FormatInt(int64(val.(uint32)), 10)
			return handleCharImplicitCast(str, targetType)
		},
	})
	framework.MustAddImplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Xid,
		ToType:   pgtypes.Xid,
		Function: func(ctx framework.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			return val, nil
		},
	})
}
