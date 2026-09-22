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

	"github.com/dolthub/doltgresql/core/casts"
	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initXml handles all casts that are built-in. This comprises only the source types.
func initXml(builtInCasts map[id.Cast]casts.Cast) {
	xmlAssignment(builtInCasts)
}

// xmlAssignment registers all assignment casts. This comprises only the source types.
func xmlAssignment(builtInCasts map[id.Cast]casts.Cast) {
	framework.MustAddAssignmentTypeCast(builtInCasts, framework.TypeCast{
		FromType: pgtypes.Xml,
		ToType:   pgtypes.BpChar,
		Function: func(ctx *sql.Context, val any, _, targetType *pgtypes.DoltgresType) (any, error) {
			str, err := framework.UnwrapString(ctx, val)
			if err != nil {
				return nil, err
			}
			return handleStringCast(str, targetType)
		},
	})
	framework.MustAddAssignmentTypeCast(builtInCasts, framework.TypeCast{
		FromType: pgtypes.Xml,
		ToType:   pgtypes.Text,
		Function: func(ctx *sql.Context, val any, _, targetType *pgtypes.DoltgresType) (any, error) {
			return framework.UnwrapString(ctx, val)
		},
	})
	framework.MustAddAssignmentTypeCast(builtInCasts, framework.TypeCast{
		FromType: pgtypes.Xml,
		ToType:   pgtypes.VarChar,
		Function: func(ctx *sql.Context, val any, _, targetType *pgtypes.DoltgresType) (any, error) {
			str, err := framework.UnwrapString(ctx, val)
			if err != nil {
				return nil, err
			}
			return handleStringCast(str, targetType)
		},
	})
}
