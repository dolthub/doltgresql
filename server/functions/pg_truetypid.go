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

package functions

import (
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initPgTrueTypID registers the functions to the catalog.
func initPgTrueTypID() {
	framework.RegisterFunction(_pg_truetypid)
}

// _pg_truetypid returns a domain's immediate base type or the attribute's type.
// https://github.com/postgres/postgres/blob/REL_15_STABLE/src/backend/catalog/information_schema.sql
var _pg_truetypid = framework.Function2{
	// TODO: Support schema-aware built-in registration and lookup for information_schema functions.
	// Schema: "information_schema",
	Name:       "_pg_truetypid",
	Return:     pgtypes.Oid,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.PgAttribute, pgtypes.PgType},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1, val2 any) (any, error) {
		attribute := val1.([]pgtypes.RecordValue)
		typ := val2.([]pgtypes.RecordValue)
		if typ[6].Value == string(pgtypes.TypeType_Domain) { // typtype
			return typ[25].Value, nil // typbasetype
		}
		return attribute[2].Value, nil // atttypid
	},
}
