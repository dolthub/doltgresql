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

// initPgIsOtherTempSchema registers the function to the catalog.
func initPgIsOtherTempSchema() {
	framework.RegisterFunction(pg_is_other_temp_schema)
}

// pg_is_other_temp_schema reports whether an OID identifies another session's temporary schema.
var pg_is_other_temp_schema = framework.Function1{
	Name:               "pg_is_other_temp_schema",
	Return:             pgtypes.Bool,
	Parameters:         [1]*pgtypes.DoltgresType{pgtypes.Oid},
	IsNonDeterministic: true,
	Strict:             true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		// Doltgres stores temporary tables in session-local state without creating
		// temporary schemas. No namespace can therefore belong to another session.
		// TODO: Check namespace ownership when session-specific temporary schemas are supported.
		return false, nil
	},
}
