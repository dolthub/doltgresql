// Copyright 2025 Dolthub, Inc.
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

	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initPgGetFunctionResult registers the functions to the catalog.
func initPgGetFunctionResult() {
	framework.RegisterFunction(pg_get_function_result_oid)
}

// pg_get_function_result_oid represents the PostgreSQL system catalog information function.
var pg_get_function_result_oid = framework.Function1{
	Name:               "pg_get_function_result",
	Return:             pgtypes.Text,
	Parameters:         [1]*pgtypes.DoltgresType{pgtypes.Oid},
	IsNonDeterministic: true,
	Strict:             true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		var result any
		err := RunCallback(ctx, val.(id.Id), Callbacks{
			Function: func(ctx *sql.Context, schema ItemSchema, function ItemFunction) (cont bool, err error) {
				if function.Item.ReturnType.SchemaName() == "" {
					//TODO: show RETURNS TABLE functions as TABLE(...) once they can be called, as their columns are only stored in the type's name
					return false, nil
				}
				returnType, err := FormatType(ctx, function.Item.ReturnType.AsId(), nil)
				if err != nil {
					return false, err
				}
				if function.Item.SetOf {
					returnType = "SETOF " + returnType
				}
				result = returnType
				return false, nil
			},
		})
		return result, err
	},
}
