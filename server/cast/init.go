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
	"github.com/dolthub/doltgresql/core/casts"
	"github.com/dolthub/doltgresql/core/id"
)

// Init initializes all casts in this package.
func Init(builtInCasts map[id.Cast]casts.Cast) {
	initBit(builtInCasts)
	initBool(builtInCasts)
	initChar(builtInCasts)
	initDate(builtInCasts)
	initFloat32(builtInCasts)
	initFloat64(builtInCasts)
	initInt16(builtInCasts)
	initInt32(builtInCasts)
	initInt64(builtInCasts)
	initInternalChar(builtInCasts)
	initInterval(builtInCasts)
	initJson(builtInCasts)
	initJsonB(builtInCasts)
	initName(builtInCasts)
	initNumeric(builtInCasts)
	initOid(builtInCasts)
	initRegclass(builtInCasts)
	initRegproc(builtInCasts)
	initRegtype(builtInCasts)
	initText(builtInCasts)
	initTime(builtInCasts)
	initTimestamp(builtInCasts)
	initTimestampTZ(builtInCasts)
	initTimeTZ(builtInCasts)
	initVarBit(builtInCasts)
	initVarChar(builtInCasts)
}
