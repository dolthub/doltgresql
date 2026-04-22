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

	"github.com/dolthub/doltgresql/core/casts"
	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/postgres/parser/duration"
	"github.com/dolthub/doltgresql/postgres/parser/timeofday"
	"github.com/dolthub/doltgresql/server/functions"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initInterval handles all casts that are built-in. This comprises only the source types.
func initInterval(builtInCasts map[id.Cast]casts.Cast) {
	intervalAssignment(builtInCasts)
	intervalImplicit(builtInCasts)
}

// intervalAssignment registers all assignment casts. This comprises only the source types.
func intervalAssignment(builtInCasts map[id.Cast]casts.Cast) {
	framework.MustAddAssignmentTypeCast(builtInCasts, framework.TypeCast{
		FromType: pgtypes.Interval,
		ToType:   pgtypes.Time,
		Function: func(ctx *sql.Context, val any, _, targetType *pgtypes.DoltgresType) (any, error) {
			dur := val.(duration.Duration)
			// the month and day of the duration are excluded
			return timeofday.FromInt(dur.Nanos() / functions.NanosPerMicro), nil
		},
	})
}

// intervalImplicit registers all implicit casts. This comprises only the source types.
func intervalImplicit(builtInCasts map[id.Cast]casts.Cast) {
	framework.MustAddImplicitTypeCast(builtInCasts, framework.TypeCast{
		FromType: pgtypes.Interval,
		ToType:   pgtypes.Interval,
		Function: func(ctx *sql.Context, val any, _, targetType *pgtypes.DoltgresType) (any, error) {
			return val.(duration.Duration), nil
		},
	})
}
