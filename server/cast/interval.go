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
	"github.com/dolthub/doltgresql/postgres/parser/duration"
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initInterval handles all casts that are built-in. This comprises only the "From" types.
func initInterval() {
	intervalAssignment()
	intervalExplicit()
	intervalImplicit()
}

// intervalAssignment registers all assignment casts. This comprises only the "From" types.
func intervalAssignment() {
	framework.MustAddAssignmentTypeCast(framework.TypeCast{
		FromType: pgtypes.Interval,
		ToType:   pgtypes.BpChar,
		Function: func(ctx *sql.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			return handleStringCast(val.(duration.Duration).String(), targetType)
		},
	})
	framework.MustAddAssignmentTypeCast(framework.TypeCast{
		FromType: pgtypes.Interval,
		ToType:   pgtypes.VarChar,
		Function: func(ctx *sql.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			return handleStringCast(val.(duration.Duration).String(), targetType)
		},
	})
}

// intervalExplicit registers all implicit casts. This comprises only the "From" types.
func intervalExplicit() {
	framework.MustAddExplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Interval,
		ToType:   pgtypes.InternalChar,
		Function: func(ctx *sql.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			return handleStringCast(val.(duration.Duration).String(), targetType)
		},
	})
	framework.MustAddExplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Interval,
		ToType:   pgtypes.Name,
		Function: func(ctx *sql.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			return handleStringCast(val.(duration.Duration).String(), targetType)
		},
	})
}

// intervalImplicit registers all implicit casts. This comprises only the "From" types.
func intervalImplicit() {
	framework.MustAddImplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Interval,
		ToType:   pgtypes.Text,
		Function: func(ctx *sql.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			return val.(duration.Duration).String(), nil
		},
	})
}
