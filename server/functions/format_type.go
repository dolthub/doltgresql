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

package functions

import (
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/lib/pq/oid"

	"github.com/dolthub/doltgresql/core/id"

	"github.com/dolthub/doltgresql/postgres/parser/types"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initFormatType registers the functions to the catalog.
func initFormatType() {
	framework.RegisterFunction(format_type)
}

// format_type represents the PostgreSQL system information function.
var format_type = framework.Function2{
	Name:       "format_type",
	Return:     pgtypes.Text,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Oid, pgtypes.Int32},
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1, val2 any) (any, error) {
		if val1 == nil {
			return nil, nil
		}
		toid := id.Cache().ToOID(val1.(id.Internal))
		if t, ok := types.OidToType[oid.Oid(toid)]; ok {
			if val2 == nil {
				return t.SQLStandardName(), nil
			} else {
				return t.SQLStandardNameWithTypmod(true, int(val2.(int32))), nil
			}
		}
		return "???", nil
	},
}
