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

// initPgIndexesSize registers the functions to the catalog.
func initPgIndexesSize() {
	framework.RegisterFunction(pg_indexes_size_regclass)
}

// pg_indexes_size_regclass represents the PostgreSQL system administration functions.
var pg_indexes_size_regclass = framework.Function1{
	Name:               "pg_indexes_size",
	Return:             pgtypes.Int64,
	Parameters:         [1]pgtypes.DoltgresType{pgtypes.Regclass},
	IsNonDeterministic: true,
	Strict:             true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, val1 any) (any, error) {
		// TODO: Total disk space used by indexes attached to the specified table
		return 0, nil
	},
}
