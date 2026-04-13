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
	"fmt"
	"strings"

	"github.com/cockroachdb/errors"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/core/procedures"
	"github.com/dolthub/doltgresql/postgres/parser/parser"
	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
	"github.com/dolthub/doltgresql/server/auth"
	pgnodes "github.com/dolthub/doltgresql/server/node"
	"github.com/dolthub/doltgresql/server/plpgsql"
)

// nodeCreateProcedure handles *tree.CreateProcedure nodes.
func nodeCreateProcedure(ctx *Context, node *tree.CreateProcedure) (vitess.Statement, error) {
	options, err := validateRoutineOptions(ctx, node.Options)
	if err != nil {
		return nil, err
	}
	// Grab the general information that we'll need to create the procedure
	tableName := node.Name.ToTableName()
	params := make([]pgnodes.RoutineParam, len(node.Args))
	var defaults []vitess.Expr
	for i, arg := range node.Args {
		// parameter name
		params[i].Name = arg.Name.String()
		// parameter type
		_, params[i].Type, err = nodeResolvableTypeReference(ctx, arg.Type, false)
		if err != nil {
			return nil, err
		}
		// parameter mode
		switch arg.Mode {
		case tree.RoutineArgModeIn:
			params[i].Mode = procedures.ParameterMode_IN
		case tree.RoutineArgModeVariadic:
			params[i].Mode = procedures.ParameterMode_VARIADIC
		case tree.RoutineArgModeOut:
			params[i].Mode = procedures.ParameterMode_OUT
		case tree.RoutineArgModeInout:
			params[i].Mode = procedures.ParameterMode_INOUT
		default:
			return nil, errors.Newf("unknown procedure argmode: `%v`", arg.Mode)
		}
		// parameter default
		if arg.Default != nil {
			params[i].HasDefault = true
			d, err := nodeExpr(ctx, arg.Default)
			if err != nil {
				return nil, err
			}
			defaults = append(defaults, d)
		}
	}
	// We only support PL/pgSQL, SQL and C for now, so we verify that here
	var parsedBody []plpgsql.InterpreterOperation
	var sqlDef string
	var sqlDefParsedStmts []vitess.Statement
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
			// parse types
			for i, op := range parsedBody {
				switch op.OpCode {
				case plpgsql.OpCode_Declare:
					// ParseType uses casting to parse the given type, but
					// some special types cannot be cast. Eg: `user_defined_table_type%ROWTYPE`
					if declareTyp, err := parser.ParseType(op.PrimaryData); err == nil {
						if _, dt, err := nodeResolvableTypeReference(ctx, declareTyp, false); err == nil && dt != nil {
							dtName := dt.Name()
							if dt.Schema() != "" {
								dtName = fmt.Sprintf("%s.%s", dt.Schema(), dtName)
							}
							parsedBody[i].PrimaryData = dtName
						}
					}
				}
			}
		case "sql":
			as, ok := options[tree.OptionAs1]
			if !ok {
				return nil, errors.Errorf("CREATE PROCEDURE definition needed for LANGUAGE SQL")
			}
			sqlDef, sqlDefParsedStmts, err = handleLanguageSQLAs(as.Definition, params)
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
			params,
			ctx.originalQuery,
			extensionName,
			extensionSymbol,
			parsedBody,
			sqlDef,
			sqlDefParsedStmts,
		),
		Auth: vitess.AuthInformation{
			AuthType:    auth.AuthType_CREATE,
			TargetType:  auth.AuthTargetType_SchemaIdentifiers,
			TargetNames: []string{tableName.Catalog(), tableName.Schema()},
		},
		Children: defaults,
	}, nil
}
