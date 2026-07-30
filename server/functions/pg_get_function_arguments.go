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
	"strings"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/lib/pq/oid"

	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/core/procedures"
	"github.com/dolthub/doltgresql/postgres/parser/types"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initPgGetFunctionArguments registers the functions to the catalog.
func initPgGetFunctionArguments() {
	framework.RegisterFunction(pg_get_function_arguments_oid)
}

// pg_get_function_arguments_oid represents the PostgreSQL system catalog information function.
var pg_get_function_arguments_oid = framework.Function1{
	Name:               "pg_get_function_arguments",
	Return:             pgtypes.Text,
	Parameters:         [1]*pgtypes.DoltgresType{pgtypes.Oid},
	IsNonDeterministic: true,
	Strict:             true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		oidVal := val.(id.Id)
		result := ""
		err := RunCallback(ctx, oidVal, Callbacks{
			Function: func(ctx *sql.Context, schema ItemSchema, function ItemFunction) (cont bool, err error) {
				args := make([]string, len(function.Item.ParameterTypes))
				for i, paramType := range function.Item.ParameterTypes {
					arg := functionArgumentTypeName(paramType)
					if name := function.Item.ParameterNames[i]; name != "" {
						arg = name + " " + arg
					}
					// TODO: functions do not store per-parameter modes, so VARIADIC is not printed
					if def := function.Item.ParameterDefaults[i]; def != "" {
						arg += " DEFAULT " + def
					}
					args[i] = arg
				}
				result = strings.Join(args, ", ")
				return false, nil
			},
			Procedure: func(ctx *sql.Context, schema ItemSchema, procedure ItemProcedure) (cont bool, err error) {
				args := make([]string, len(procedure.Item.ParameterTypes))
				for i, paramType := range procedure.Item.ParameterTypes {
					arg := functionArgumentTypeName(paramType)
					if name := procedure.Item.ParameterNames[i]; name != "" {
						arg = name + " " + arg
					}
					switch procedure.Item.ParameterModes[i] {
					case procedures.ParameterMode_OUT:
						arg = "OUT " + arg
					case procedures.ParameterMode_INOUT:
						arg = "INOUT " + arg
					case procedures.ParameterMode_VARIADIC:
						arg = "VARIADIC " + arg
					}
					if def := procedure.Item.ParameterDefaults[i]; def != "" {
						arg += " DEFAULT " + def
					}
					args[i] = arg
				}
				result = strings.Join(args, ", ")
				return false, nil
			},
		})
		if err != nil {
			return "", err
		}
		return result, nil
	},
}

// functionArgumentTypeName returns the name of the given parameter type as it would appear in a function's argument
// list, preferring the SQL standard name (matching Postgres's format_type) when the type is a built-in.
func functionArgumentTypeName(paramType id.Type) string {
	if t, ok := types.OidToType[oid.Oid(id.Cache().ToOID(paramType.AsId()))]; ok {
		return t.SQLStandardName()
	}
	return paramType.TypeName()
}
