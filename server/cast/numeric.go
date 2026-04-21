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
	"github.com/jackc/pgtype"

	"github.com/dolthub/go-mysql-server/sql"

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
			var i int16
			num := val.(pgtype.Numeric)
			err := num.AssignTo(&i)
			if err != nil {
				return nil, errors.Errorf("smallint out of range")
			}
			return i, nil
		},
	})
	framework.MustAddAssignmentTypeCast(framework.TypeCast{
		FromType: pgtypes.Numeric,
		ToType:   pgtypes.Int32,
		Function: func(ctx *sql.Context, val any, targetType *pgtypes.DoltgresType) (any, error) {
			var i int32
			num := val.(pgtype.Numeric)
			err := num.AssignTo(&i)
			if err != nil {
				return nil, errors.Wrap(pgtypes.ErrCastOutOfRange, "integer out of range")
			}
			return i, nil
		},
	})
	framework.MustAddAssignmentTypeCast(framework.TypeCast{
		FromType: pgtypes.Numeric,
		ToType:   pgtypes.Int64,
		Function: func(ctx *sql.Context, val any, targetType *pgtypes.DoltgresType) (any, error) {
			var i int64
			num := val.(pgtype.Numeric)
			err := num.AssignTo(&i)
			if err != nil {
				return nil, errors.Wrap(pgtypes.ErrCastOutOfRange, "bigint out of range")
			}
			return i, nil
		},
	})
}

// numericImplicit registers all implicit casts. This comprises only the "From" types.
func numericImplicit() {
	framework.MustAddImplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Numeric,
		ToType:   pgtypes.Float32,
		Function: func(ctx *sql.Context, val any, targetType *pgtypes.DoltgresType) (any, error) {
			var f float32
			num := val.(pgtype.Numeric)
			err := num.AssignTo(&f)
			if err != nil {
				return pgtype.Numeric{}, err
			}
			return f, nil
		},
	})
	framework.MustAddImplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Numeric,
		ToType:   pgtypes.Float64,
		Function: func(ctx *sql.Context, val any, targetType *pgtypes.DoltgresType) (any, error) {
			var f float64
			num := val.(pgtype.Numeric)
			err := num.AssignTo(&f)
			if err != nil {
				return pgtype.Numeric{}, err
			}
			return f, nil
		},
	})
	framework.MustAddImplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Numeric,
		ToType:   pgtypes.Numeric,
		Function: func(ctx *sql.Context, val any, targetType *pgtypes.DoltgresType) (any, error) {
			return pgtypes.GetNumericValueWithTypmod(val, targetType.GetAttTypMod())
		},
	})
}
