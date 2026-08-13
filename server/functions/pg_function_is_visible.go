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

	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initPgFunctionIsVisible registers the functions to the catalog.
func initPgFunctionIsVisible() {
	framework.RegisterFunction(pg_function_is_visible_oid)
}

// pg_function_is_visible_oid represents the PostgreSQL system schema visibility inquiry function.
var pg_function_is_visible_oid = framework.Function1{
	Name:               "pg_function_is_visible",
	Return:             pgtypes.Bool,
	Parameters:         [1]*pgtypes.DoltgresType{pgtypes.Oid},
	IsNonDeterministic: true,
	Strict:             true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		oidVal := val.(id.Id)
		if oidVal.Section() != id.Section_Function && oidVal.Section() != id.Section_Procedure {
			return false, nil
		}
		// TODO: Postgres additionally checks that the function isn't shadowed by one with the same
		// name and argument types in an earlier schema of the search path.
		paths, err := core.SearchPath(ctx)
		if err != nil {
			return false, err
		}
		schemaName := oidVal.Segment(0)
		inPath := false
		for _, path := range paths {
			if path == schemaName {
				inPath = true
				break
			}
		}
		if !inPath {
			return false, nil
		}
		// Built-in functions are registered in the OID cache at startup, and don't appear in the
		// function and procedure collections that RunCallback iterates over.
		if id.Cache().Exists(oidVal) {
			return true, nil
		}
		isVisible := false
		err = RunCallback(ctx, oidVal, Callbacks{
			Function: func(ctx *sql.Context, schema ItemSchema, function ItemFunction) (cont bool, err error) {
				isVisible = true
				return false, nil
			},
			Procedure: func(ctx *sql.Context, schema ItemSchema, procedure ItemProcedure) (cont bool, err error) {
				isVisible = true
				return false, nil
			},
		})
		return isVisible, err
	},
}
