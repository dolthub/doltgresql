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
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initVarBit handles all casts that are built-in. This comprises only the "From" types.
func initVarBit() {
	varBitImplicit()
}

// varBitImplicit registers all implicit casts. This comprises only the "From" types.
func varBitImplicit() {
	framework.MustAddImplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.VarBit,
		ToType:   pgtypes.Bit,
		Function: func(ctx *sql.Context, val any, targetType *pgtypes.DoltgresType) (any, error) {
			input := val.(string)
			array, err := tree.ParseDBitArray(input)
			if err != nil {
				return nil, err
			}
			expectedLength := pgtypes.GetCharLengthFromTypmod(targetType.GetAttTypMod())
			if array.BitLen() != uint(expectedLength) {
				return nil, pgtypes.ErrWrongLengthBit.New(len(input), expectedLength)
			}
			return tree.AsStringWithFlags(array, tree.FmtPgwireText), nil
		},
	})
	framework.MustAddImplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.VarBit,
		ToType:   pgtypes.VarBit,
		Function: func(ctx *sql.Context, val any, targetType *pgtypes.DoltgresType) (any, error) {
			input := val.(string)
			array, err := tree.ParseDBitArray(input)
			if err != nil {
				return nil, err
			}
			atttypmod := targetType.GetAttTypMod()
			if atttypmod != -1 {
				maxLength := pgtypes.GetCharLengthFromTypmod(atttypmod)
				if int32(array.BitLen()) > maxLength {
					return nil, pgtypes.ErrVarBitLengthExceeded.New(maxLength)
				}
			}
			return tree.AsStringWithFlags(array, tree.FmtPgwireText), nil
		},
	})
}
