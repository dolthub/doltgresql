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

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initPgTableSize registers the functions to the catalog.
func initPgTableSize() {
	framework.RegisterFunction(pg_table_size_regclass)
}

// pg_table_size_regclass represents the PostgreSQL function of the same name, taking the same parameters.
var pg_table_size_regclass = framework.Function1{
	Name:               "pg_table_size",
	Return:             pgtypes.Int64,
	Parameters:         [1]*pgtypes.DoltgresType{pgtypes.Regclass},
	IsNonDeterministic: true,
	Strict:             true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		// TODO: Disk space used by the specified table, excluding indexes (but including TOAST, free space map, and visibility map)
		return int64(0), nil
	},
}
