// Copyright 2026 Dolthub, Inc.
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

	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initBit handles all casts that are built-in. This comprises only the "From" types.
func initBit() {
	bitExplicit()
	bitImplicit()
}

// bitExplicit registers all explicit casts. This comprises only the "From" types.
func bitExplicit() {
	framework.MustAddExplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Bit,
		ToType:   pgtypes.Int32,
		Function: func(ctx *sql.Context, val any, targetType *pgtypes.DoltgresType) (any, error) {
			array, err := tree.ParseDBitArray(val.(string))
			if err != nil {
				return nil, err
			}
			if array.BitLen() > 32 {
				return nil, errors.Wrap(pgtypes.ErrCastOutOfRange, "integer out of range")
			}
			return int32(array.AsInt64(32)), nil
		},
	})
	framework.MustAddExplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Bit,
		ToType:   pgtypes.Int64,
		Function: func(ctx *sql.Context, val any, targetType *pgtypes.DoltgresType) (any, error) {
			array, err := tree.ParseDBitArray(val.(string))
			if err != nil {
				return nil, err
			}
			if array.BitLen() > 64 {
				return nil, errors.Wrap(pgtypes.ErrCastOutOfRange, "bigint out of range")
			}
			return array.AsInt64(64), nil
		},
	})
}

// bitImplicit registers all implicit casts. This comprises only the "From" types.
func bitImplicit() {
	framework.MustAddImplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Bit,
		ToType:   pgtypes.Bit,
		Function: func(ctx *sql.Context, val any, targetType *pgtypes.DoltgresType) (any, error) {
			input := val.(string)
			array, err := tree.ParseDBitArray(input)
			if err != nil {
				return nil, err
			}
			expectedLength := targetType.GetAttTypMod()
			if array.BitLen() != uint(expectedLength) {
				return nil, pgtypes.ErrWrongLengthBit.New(len(input), expectedLength)
			}
			return tree.AsStringWithFlags(array, tree.FmtPgwireText), nil
		},
	})
	framework.MustAddImplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Bit,
		ToType:   pgtypes.VarBit,
		Function: func(ctx *sql.Context, val any, targetType *pgtypes.DoltgresType) (any, error) {
			input := val.(string)
			array, err := tree.ParseDBitArray(input)
			if err != nil {
				return nil, err
			}
			atttypmod := targetType.GetAttTypMod()
			if atttypmod != -1 {
				if int32(array.BitLen()) > atttypmod {
					return nil, pgtypes.ErrVarBitLengthExceeded.New(atttypmod)
				}
			}
			return tree.AsStringWithFlags(array, tree.FmtPgwireText), nil
		},
	})
}
