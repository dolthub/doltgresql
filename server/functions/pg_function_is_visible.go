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
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
	"github.com/dolthub/go-mysql-server/sql"
)

// initPgFunctionIsVisible registers the functions to the catalog.
func initPgFunctionIsVisible() {
	framework.RegisterFunction(pg_function_is_visible)
}

// pg_function_is_visible represents the PostgreSQL system schema visibility inquiry function.
var pg_function_is_visible = framework.Function1{
	Name:               "pg_function_is_visible",
	Return:             pgtypes.Bool,
	Parameters:         [1]pgtypes.DoltgresType{pgtypes.Oid},
	IsNonDeterministic: true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, val any) (any, error) {
		// TODO: Functions are not contained within a schema for now
		return true, nil
	},
	Strict: true,
}
