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

	"github.com/dolthub/doltgresql/server/functions"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initTime handles all casts that are built-in. This comprises only the "From" types.
func initTime() {
	timeExplicit()
}

// timeExplicit registers all implicit casts. This comprises only the "From" types.
func timeExplicit() {
	framework.MustAddExplicitTypeCast(framework.TypeCast{
		FromType: pgtypes.Time,
		ToType:   pgtypes.Interval,
		Function: func(ctx *sql.Context, val any, targetType pgtypes.DoltgresType) (any, error) {
			t := val.(time.Time)
			dur := functions.GetIntervalDurationFromTimeComponents(0, 0, 0, int64(t.Hour()), int64(t.Minute()), int64(t.Second()), int64(t.Nanosecond()))
			return dur, nil
		},
	})
}
