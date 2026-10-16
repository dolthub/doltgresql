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
	"strconv"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/core/casts"
	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/postgres/parser/ipaddr"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initInet handles all casts that are built-in. This comprises only the source types.
func initInet(builtInCasts map[id.Cast]casts.Cast) {
	inetAssignment(builtInCasts)
}

// inetAssignment registers all assignment casts. This comprises only the source types.
func inetAssignment(builtInCasts map[id.Cast]casts.Cast) {
	// Unlike the output function, casting to a string type always includes the mask length
	for _, targetType := range []*pgtypes.DoltgresType{pgtypes.BpChar, pgtypes.Text, pgtypes.VarChar} {
		framework.MustAddAssignmentTypeCast(builtInCasts, framework.TypeCast{
			FromType: pgtypes.Inet,
			ToType:   targetType,
			Function: func(ctx *sql.Context, val any, _, targetType *pgtypes.DoltgresType) (any, error) {
				ip := val.(ipaddr.IPAddr)
				mask := ip.Mask
				if ip.Family == ipaddr.IPv4family {
					ip.Mask = 32
				} else {
					ip.Mask = 128
				}
				str := ip.String() + "/" + strconv.Itoa(int(mask))
				if targetType.ID == pgtypes.Text.ID {
					return str, nil
				}
				return handleStringCast(str, targetType)
			},
		})
	}
}
