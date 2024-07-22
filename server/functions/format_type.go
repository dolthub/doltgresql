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

	"github.com/dolthub/doltgresql/postgres/parser/types"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initFormatType registers the functions to the catalog.
func initFormatType() {
	framework.RegisterFunction(format_type)
	framework.RegisterFunction(format_type_null)
}

// format_type represents the PostgreSQL function of the same name, taking the same parameters.
var format_type = framework.Function2{
	Name:       "format_type",
	Return:     pgtypes.Text,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Oid, pgtypes.Int32},
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1, val2 any) (any, error) {
		toid := val1.(uint32)
		typemod := val2.(int32)
		if t, ok := types.OidToType[oid.Oid(toid)]; ok {
			return t.SQLStandardNameWithTypmod(true, int(typemod)), nil
		}
		return "???", nil
	},
}

// format_type_null represents the PostgreSQL function format_type, but with a null second parameter.
var format_type_null = framework.Function2{
	Name:       "format_type",
	Return:     pgtypes.Text,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Oid, pgtypes.Null},
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1, val2 any) (any, error) {
		toid := val1.(uint32)
		if t, ok := types.OidToType[oid.Oid(toid)]; ok {
			return t.SQLStandardName(), nil
		}
		return "???", nil
	},
}
