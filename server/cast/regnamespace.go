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

// initRegnamespace handles all casts that are built-in. This comprises only the source types.
func initRegnamespace(builtInCasts map[id.Cast]casts.Cast) {
	regnamespaceAssignment(builtInCasts)
	regnamespaceImplicit(builtInCasts)
}

// regnamespaceAssignment registers all assignment casts. This comprises only the source types.
func regnamespaceAssignment(builtInCasts map[id.Cast]casts.Cast) {
	framework.MustAddAssignmentTypeCast(builtInCasts, framework.TypeCast{
		FromType: pgtypes.Regnamespace,
		ToType:   pgtypes.Int32,
		Function: func(ctx *sql.Context, val any, _, targetType *pgtypes.DoltgresType) (any, error) {
			return int32(id.Cache().ToOID(val.(id.Id))), nil
		},
	})
	framework.MustAddAssignmentTypeCast(builtInCasts, framework.TypeCast{
		FromType: pgtypes.Regnamespace,
		ToType:   pgtypes.Int64,
		Function: func(ctx *sql.Context, val any, _, targetType *pgtypes.DoltgresType) (any, error) {
			return int64(id.Cache().ToOID(val.(id.Id))), nil
		},
	})
}

// regnamespaceImplicit registers all implicit casts. This comprises only the source types.
func regnamespaceImplicit(builtInCasts map[id.Cast]casts.Cast) {
	framework.MustAddImplicitTypeCast(builtInCasts, framework.TypeCast{
		FromType: pgtypes.Regnamespace,
		ToType:   pgtypes.Oid,
		Function: func(ctx *sql.Context, val any, _, targetType *pgtypes.DoltgresType) (any, error) {
			return val, nil
		},
	})
}
