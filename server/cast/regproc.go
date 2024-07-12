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

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initRegproc handles all casts that are built-in. This comprises only the "From" types.
func initRegproc() {
	regprocAssignment()
	regprocImplicit()
}

// regprocAssignment registers all assignment casts. This comprises only the "From" types.
func regprocAssignment() {
	framework.MustAddAssignmentTypeCast(framework.TypeCast{
		FromType: pgtypes.Regproc,
		ToType:   pgtypes.Int32,
		Function: func(ctx *sql.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			return int32(val.(uint32)), nil
		},
	})
	framework.MustAddAssignmentTypeCast(framework.TypeCast{
		FromType: pgtypes.Regproc,
		ToType:   pgtypes.Int64,
		Function: func(ctx *sql.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			return int64(val.(uint32)), nil
		},
	})
}

// regprocImplicit registers all implicit casts. This comprises only the "From" types.
func regprocImplicit() {
	framework.MustAddImplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Regproc,
		ToType:   pgtypes.Oid,
		Function: func(ctx *sql.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			return val, nil
		},
	})
}
