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

	"github.com/dolthub/doltgresql/core/casts"
	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/postgres/parser/pgdate"
	"github.com/dolthub/doltgresql/postgres/parser/timeofday"
	"github.com/dolthub/doltgresql/postgres/parser/timetz"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initTimestampTZ handles all casts that are built-in. This comprises only the source types.
func initTimestampTZ(builtInCasts map[id.Cast]casts.Cast) {
	timestampTZAssignment(builtInCasts)
	timestampTZImplicit(builtInCasts)
}

// timestampTZAssignment registers all assignment casts. This comprises only the source types.
func timestampTZAssignment(builtInCasts map[id.Cast]casts.Cast) {
	framework.MustAddAssignmentTypeCast(builtInCasts, framework.TypeCast{
		FromType: pgtypes.TimestampTZ,
		ToType:   pgtypes.Date,
		Function: func(ctx *sql.Context, val any, _, targetType *pgtypes.DoltgresType) (any, error) {
			d, err := pgdate.MakeDateFromTime(val.(time.Time))
			if err != nil {
				return nil, err
			}
			return d.ToTime()
		},
	})
	framework.MustAddAssignmentTypeCast(builtInCasts, framework.TypeCast{
		FromType: pgtypes.TimestampTZ,
		ToType:   pgtypes.Time,
		Function: func(ctx *sql.Context, val any, _, targetType *pgtypes.DoltgresType) (any, error) {
			return timeofday.FromTime(val.(time.Time)), nil
		},
	})
	framework.MustAddAssignmentTypeCast(builtInCasts, framework.TypeCast{
		FromType: pgtypes.TimestampTZ,
		ToType:   pgtypes.Timestamp,
		Function: func(ctx *sql.Context, val any, _, targetType *pgtypes.DoltgresType) (any, error) {
			// TODO: check
			return val.(time.Time), nil
		},
	})
	framework.MustAddAssignmentTypeCast(builtInCasts, framework.TypeCast{
		FromType: pgtypes.TimestampTZ,
		ToType:   pgtypes.TimeTZ,
		Function: func(ctx *sql.Context, val any, _, targetType *pgtypes.DoltgresType) (any, error) {
			return timetz.MakeTimeTZFromTime(val.(time.Time)), nil
		},
	})
}

// timestampTZImplicit registers all implicit casts. This comprises only the source types.
func timestampTZImplicit(builtInCasts map[id.Cast]casts.Cast) {
	framework.MustAddImplicitTypeCast(builtInCasts, framework.TypeCast{
		FromType: pgtypes.TimestampTZ,
		ToType:   pgtypes.TimestampTZ,
		Function: func(ctx *sql.Context, val any, _, targetType *pgtypes.DoltgresType) (any, error) {
			return val.(time.Time), nil
		},
	})
}
