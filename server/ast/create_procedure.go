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

package ast

import (
	"strings"

	"github.com/cockroachdb/errors"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/core/procedures"
	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
	"github.com/dolthub/doltgresql/postgres/parser/types"
	pgnodes "github.com/dolthub/doltgresql/server/node"
	"github.com/dolthub/doltgresql/server/plpgsql"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// nodeCreateProcedure handles *tree.CreateProcedure nodes.
func nodeCreateProcedure(ctx *Context, node *tree.CreateProcedure) (vitess.Statement, error) {
	options, err := validateRoutineOptions(ctx, node.Options)
	if err != nil {
		return nil, err
	}
	// Grab the general information that we'll need to create the procedure
	tableName := node.Name.ToTableName()
	paramNames := make([]string, len(node.Args))
	paramTypes := make([]*pgtypes.DoltgresType, len(node.Args))
	paramModes := make([]procedures.ParameterMode, len(node.Args))
	for i, arg := range node.Args {
		paramNames[i] = arg.Name.String()
		switch argType := arg.Type.(type) {
		case *types.T:
			paramTypes[i] = pgtypes.NewUnresolvedDoltgresType("", strings.ToLower(argType.Name()))
		default:
			paramTypes[i] = pgtypes.NewUnresolvedDoltgresType("", strings.ToLower(argType.SQLString()))
		}
		switch arg.Mode {
		case tree.RoutineArgModeIn:
			paramModes[i] = procedures.ParameterMode_IN
		case tree.RoutineArgModeVariadic:
			paramModes[i] = procedures.ParameterMode_VARIADIC
		case tree.RoutineArgModeOut:
			paramModes[i] = procedures.ParameterMode_OUT
		case tree.RoutineArgModeInout:
			paramModes[i] = procedures.ParameterMode_INOUT
		default:
			return nil, errors.Newf("unknown procedure argmode: `%v`", arg.Mode)
		}
	}
	// We only support PL/pgSQL, SQL and C for now, so we verify that here
	var parsedBody []plpgsql.InterpreterOperation
	var sqlDef string
	var sqlDefParsed vitess.Statement
	var extensionName, extensionSymbol string
	if languageOption, ok := options[tree.OptionLanguage]; ok {
		switch strings.ToLower(languageOption.Language) {
		case "plpgsql":
			// PL/pgSQL is different from standard Postgres SQL, so we have to use a special parser to handle it.
			// This parser also requires the full `CREATE PROCEDURE` string, so we'll pass that.
			parsedBody, err = plpgsql.Parse(ctx.originalQuery)
			if err != nil {
				return nil, err
			}
		case "sql":
			as, ok := options[tree.OptionAs1]
			if !ok {
				return nil, errors.Errorf("CREATE PROCEDURE definition needed for LANGUAGE SQL")
			}
			sqlDef, sqlDefParsed, err = handleLanguageSQL(as.Definition, paramNames, paramTypes)
			if err != nil {
				return nil, err
			}
		case "c":
			symbolOption, ok := options[tree.OptionAs2]
			if !ok {
				return nil, errors.Errorf("LANGUAGE C is only supported when providing both the module name and symbol")
			}
			extensionName = symbolOption.ObjFile
			extensionSymbol = symbolOption.LinkSymbol
		default:
			return nil, errors.Errorf("CREATE PROCEDURE only supports PL/pgSQL, C and SQL for now; others are not yet supported")
		}
	} else {
		return nil, errors.Errorf("CREATE PROCEDURE does not define an input language")
	}
	// Returns the stored procedure call with all options
	return vitess.InjectedStatement{
		Statement: pgnodes.NewCreateProcedure(
			tableName.Table(),
			tableName.Schema(),
			node.Replace,
			paramNames,
			paramTypes,
			paramModes,
			ctx.originalQuery,
			extensionName,
			extensionSymbol,
			parsedBody,
			sqlDef,
			sqlDefParsed,
		),
	}, nil
}
