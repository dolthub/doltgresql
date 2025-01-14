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
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/core/id"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initRegtype handles all casts that are built-in. This comprises only the "From" types.
func initRegtype() {
	regtypeAssignment()
	regtypeImplicit()
}

// regtypeAssignment registers all assignment casts. This comprises only the "From" types.
func regtypeAssignment() {
	framework.MustAddAssignmentTypeCast(framework.TypeCast{
		FromType: pgtypes.Regtype,
		ToType:   pgtypes.Int32,
		Function: func(ctx *sql.Context, val any, targetType *pgtypes.DoltgresType) (any, error) {
			return int32(id.Cache().ToOID(val.(id.Id))), nil
		},
	})
	framework.MustAddAssignmentTypeCast(framework.TypeCast{
		FromType: pgtypes.Regtype,
		ToType:   pgtypes.Int64,
		Function: func(ctx *sql.Context, val any, targetType *pgtypes.DoltgresType) (any, error) {
			return int64(id.Cache().ToOID(val.(id.Id))), nil
		},
	})
}

// regtypeImplicit registers all implicit casts. This comprises only the "From" types.
func regtypeImplicit() {
	framework.MustAddImplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Regtype,
		ToType:   pgtypes.Oid,
		Function: func(ctx *sql.Context, val any, targetType *pgtypes.DoltgresType) (any, error) {
			return val, nil
		},
	})
}
