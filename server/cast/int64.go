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
	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/shopspring/decimal"

	"github.com/dolthub/doltgresql/core/casts"
	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initInt64 handles all casts that are built-in. This comprises only the source types.
func initInt64(builtInCasts map[id.Cast]casts.Cast) {
	int64Assignment(builtInCasts)
	int64Implicit(builtInCasts)
}

// int64Assignment registers all assignment casts. This comprises only the source types.
func int64Assignment(builtInCasts map[id.Cast]casts.Cast) {
	framework.MustAddAssignmentTypeCast(builtInCasts, framework.TypeCast{
		FromType: pgtypes.Int64,
		ToType:   pgtypes.Int16,
		Function: func(ctx *sql.Context, val any, _, targetType *pgtypes.DoltgresType) (any, error) {
			if val.(int64) > 32767 || val.(int64) < -32768 {
				return nil, errors.Wrap(pgtypes.ErrCastOutOfRange, "smallint out of range")
			}
			return int16(val.(int64)), nil
		},
	})
	framework.MustAddAssignmentTypeCast(builtInCasts, framework.TypeCast{
		FromType: pgtypes.Int64,
		ToType:   pgtypes.Int32,
		Function: func(ctx *sql.Context, val any, _, targetType *pgtypes.DoltgresType) (any, error) {
			if val.(int64) > 2147483647 || val.(int64) < -2147483648 {
				return nil, errors.Wrap(pgtypes.ErrCastOutOfRange, "integer out of range")
			}
			return int32(val.(int64)), nil
		},
	})
}

// int64Implicit registers all implicit casts. This comprises only the source types.
func int64Implicit(builtInCasts map[id.Cast]casts.Cast) {
	framework.MustAddImplicitTypeCast(builtInCasts, framework.TypeCast{
		FromType: pgtypes.Int64,
		ToType:   pgtypes.Float32,
		Function: func(ctx *sql.Context, val any, _, targetType *pgtypes.DoltgresType) (any, error) {
			return float32(val.(int64)), nil
		},
	})
	framework.MustAddImplicitTypeCast(builtInCasts, framework.TypeCast{
		FromType: pgtypes.Int64,
		ToType:   pgtypes.Float64,
		Function: func(ctx *sql.Context, val any, _, targetType *pgtypes.DoltgresType) (any, error) {
			return float64(val.(int64)), nil
		},
	})
	framework.MustAddImplicitTypeCast(builtInCasts, framework.TypeCast{
		FromType: pgtypes.Int64,
		ToType:   pgtypes.Numeric,
		Function: func(ctx *sql.Context, val any, _, targetType *pgtypes.DoltgresType) (any, error) {
			return decimal.NewFromInt(val.(int64)), nil
		},
	})
	framework.MustAddImplicitTypeCast(builtInCasts, framework.TypeCast{
		FromType: pgtypes.Int64,
		ToType:   pgtypes.Oid,
		Function: func(ctx *sql.Context, val any, _, targetType *pgtypes.DoltgresType) (any, error) {
			if val.(int64) > pgtypes.MaxUint32 || val.(int64) < 0 {
				return nil, errOutOfRange.New(targetType.String())
			}
			if internalID := id.Cache().ToInternal(uint32(val.(int64))); internalID.IsValid() {
				return internalID, nil
			}
			return id.NewOID(uint32(val.(int64))).AsId(), nil
		},
	})
	framework.MustAddImplicitTypeCast(builtInCasts, framework.TypeCast{
		FromType: pgtypes.Int64,
		ToType:   pgtypes.Regclass,
		Function: func(ctx *sql.Context, val any, _, targetType *pgtypes.DoltgresType) (any, error) {
			if val.(int64) > pgtypes.MaxUint32 || val.(int64) < 0 {
				return nil, errOutOfRange.New(targetType.String())
			}
			if internalID := id.Cache().ToInternal(uint32(val.(int64))); internalID.IsValid() {
				return internalID, nil
			}
			return id.NewOID(uint32(val.(int64))).AsId(), nil
		},
	})
	framework.MustAddImplicitTypeCast(builtInCasts, framework.TypeCast{
		FromType: pgtypes.Int64,
		ToType:   pgtypes.Regproc,
		Function: func(ctx *sql.Context, val any, _, targetType *pgtypes.DoltgresType) (any, error) {
			if val.(int64) > pgtypes.MaxUint32 || val.(int64) < 0 {
				return nil, errOutOfRange.New(targetType.String())
			}
			if internalID := id.Cache().ToInternal(uint32(val.(int64))); internalID.IsValid() {
				return internalID, nil
			}
			return id.NewOID(uint32(val.(int64))).AsId(), nil
		},
	})
	framework.MustAddImplicitTypeCast(builtInCasts, framework.TypeCast{
		FromType: pgtypes.Int64,
		ToType:   pgtypes.Regtype,
		Function: func(ctx *sql.Context, val any, _, targetType *pgtypes.DoltgresType) (any, error) {
			if val.(int64) > pgtypes.MaxUint32 || val.(int64) < 0 {
				return nil, errOutOfRange.New(targetType.String())
			}
			if internalID := id.Cache().ToInternal(uint32(val.(int64))); internalID.IsValid() {
				return internalID, nil
			}
			return id.NewOID(uint32(val.(int64))).AsId(), nil
		},
	})
}
