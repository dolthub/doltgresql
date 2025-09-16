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
	"time"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/postgres/parser/pgdate"
	"github.com/dolthub/doltgresql/postgres/parser/timeofday"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initTimestampTZ handles all casts that are built-in. This comprises only the "From" types.
func initTimestampTZ() {
	timestampTZAssignment()
	timestampTZImplicit()
}

// timestampTZAssignment registers all assignment casts. This comprises only the "From" types.
func timestampTZAssignment() {
	framework.MustAddAssignmentTypeCast(framework.TypeCast{
		FromType: pgtypes.TimestampTZ,
		ToType:   pgtypes.Date,
		Function: func(ctx *sql.Context, val any, targetType *pgtypes.DoltgresType) (any, error) {
			return pgdate.MakeDateFromTime(val.(time.Time))
		},
	})
	framework.MustAddAssignmentTypeCast(framework.TypeCast{
		FromType: pgtypes.TimestampTZ,
		ToType:   pgtypes.Time,
		Function: func(ctx *sql.Context, val any, targetType *pgtypes.DoltgresType) (any, error) {
			return timeofday.FromTime(val.(time.Time)).ToTime(), nil
		},
	})
	framework.MustAddAssignmentTypeCast(framework.TypeCast{
		FromType: pgtypes.TimestampTZ,
		ToType:   pgtypes.Timestamp,
		Function: func(ctx *sql.Context, val any, targetType *pgtypes.DoltgresType) (any, error) {
			// TODO: check
			return val.(time.Time), nil
		},
	})
	framework.MustAddAssignmentTypeCast(framework.TypeCast{
		FromType: pgtypes.TimestampTZ,
		ToType:   pgtypes.TimeTZ,
		Function: func(ctx *sql.Context, val any, targetType *pgtypes.DoltgresType) (any, error) {
			return val.(time.Time), nil
		},
	})
}

// timestampTZImplicit registers all implicit casts. This comprises only the "From" types.
func timestampTZImplicit() {
	framework.MustAddImplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.TimestampTZ,
		ToType:   pgtypes.TimestampTZ,
		Function: func(ctx *sql.Context, val any, targetType *pgtypes.DoltgresType) (any, error) {
			return val.(time.Time), nil
		},
	})
}
