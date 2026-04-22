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
	"unicode"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/core/casts"
	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initInternalChar handles all casts that are built-in. This comprises only the source types.
func initInternalChar(builtInCasts map[id.Cast]casts.Cast) {
	internalCharAssignment(builtInCasts)
	internalCharExplicit(builtInCasts)
	internalCharImplicit(builtInCasts)
}

// internalCharAssignment registers all assignment casts. This comprises only the source types.
func internalCharAssignment(builtInCasts map[id.Cast]casts.Cast) {
	framework.MustAddAssignmentTypeCast(builtInCasts, framework.TypeCast{
		FromType: pgtypes.InternalChar,
		ToType:   pgtypes.BpChar,
		Function: func(ctx *sql.Context, val any, _, targetType *pgtypes.DoltgresType) (any, error) {
			return targetType.IoInput(ctx, val.(string))
		},
	})
	framework.MustAddAssignmentTypeCast(builtInCasts, framework.TypeCast{
		FromType: pgtypes.InternalChar,
		ToType:   pgtypes.VarChar,
		Function: func(ctx *sql.Context, val any, _, targetType *pgtypes.DoltgresType) (any, error) {
			return handleStringCast(val.(string), targetType)
		},
	})
}

// internalCharExplicit registers all explicit casts. This comprises only the source types.
func internalCharExplicit(builtInCasts map[id.Cast]casts.Cast) {
	framework.MustAddExplicitTypeCast(builtInCasts, framework.TypeCast{
		FromType: pgtypes.InternalChar,
		ToType:   pgtypes.Int32,
		Function: func(ctx *sql.Context, val any, _, targetType *pgtypes.DoltgresType) (any, error) {
			s := val.(string)
			if len(s) == 0 {
				return int32(0), nil
			}
			if unicode.IsLetter(rune(s[0])) {
				return int32(s[0]), nil
			}
			i, err := strconv.ParseInt(s, 10, 32)
			if err != nil {
				return 0, err
			}
			return int32(i), nil
		},
	})
}

// internalCharImplicit registers all implicit casts. This comprises only the source types.
func internalCharImplicit(builtInCasts map[id.Cast]casts.Cast) {
	framework.MustAddImplicitTypeCast(builtInCasts, framework.TypeCast{
		FromType: pgtypes.InternalChar,
		ToType:   pgtypes.Text,
		Function: func(ctx *sql.Context, val any, _, targetType *pgtypes.DoltgresType) (any, error) {
			return val, nil
		},
	})
}
