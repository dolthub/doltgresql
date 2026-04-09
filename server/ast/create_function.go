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
	"context"
	"fmt"
	"strings"

	"github.com/cockroachdb/errors"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/postgres/parser/parser"
	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
	"github.com/dolthub/doltgresql/server/auth"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgnodes "github.com/dolthub/doltgresql/server/node"
	"github.com/dolthub/doltgresql/server/plpgsql"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// nodeCreateFunction handles *tree.CreateFunction nodes.
func nodeCreateFunction(ctx *Context, node *tree.CreateFunction) (vitess.Statement, error) {
	options, err := validateRoutineOptions(ctx, node.Options)
	if err != nil {
		return nil, err
	}
	// Grab the general information that we'll need to create the function
	tableName := node.Name.ToTableName()
	var retType *pgtypes.DoltgresType
	if len(node.RetType) == 0 {
		retType = pgtypes.Void
	} else if !node.ReturnsTable {
		// Return types may specify "trigger", but this doesn't apply elsewhere
		_, retType, err = nodeResolvableTypeReference(ctx, node.RetType[0].Type, true)
		if err != nil {
			return nil, err
		}
	} else {
		retType = createAnonymousCompositeType(node.RetType)
	}

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
	var strict bool
	if nullInputOption, ok := options[tree.OptionNullInput]; ok {
		if nullInputOption.NullInput == tree.ReturnsNullOnNullInput || nullInputOption.NullInput == tree.StrictNullInput {
			strict = true
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
			// This parser also requires the full `CREATE FUNCTION` string, so we'll pass that.
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
			if ok {
				sqlDef, sqlDefParsedStmts, err = handleLanguageSQLAs(as.Definition, params)
				if err != nil {
					return nil, err
				}
				break
			}
			sqlBody, ok := options[tree.OptionSqlBody]
			if ok {
				beginAtomic, ok := sqlBody.SqlBody.(*tree.BeginEndBlock)
				if !ok {
					return nil, errors.Errorf("Expected BEGIN in CREATE FUNCTION definition, got %T", sqlBody.SqlBody)
				}
				stmts := make([]parser.Statement, len(beginAtomic.Statements))
				for i, s := range beginAtomic.Statements {
					stmts[i] = parser.Statement{
						AST: s,
						SQL: s.String(),
					}
				}
				sqlDef, sqlDefParsedStmts, err = convertSQLStmts(stmts, params)
				if err != nil {
					return nil, err
				}
				break
			}
			return nil, errors.Errorf("CREATE FUNCTION definition needed for LANGUAGE SQL")
		case "c":
			symbolOption, ok := options[tree.OptionAs2]
			if !ok {
				return nil, errors.Errorf("LANGUAGE C is only supported when providing both the module name and symbol")
			}
			extensionName = symbolOption.ObjFile
			extensionSymbol = symbolOption.LinkSymbol
		default:
			return nil, errors.Errorf("CREATE FUNCTION only supports PL/pgSQL, C and SQL for now; others are not yet supported")
		}
	} else {
		return nil, errors.Errorf("CREATE FUNCTION does not define an input language")
	}
	// Returns the stored procedure call with all options
	return vitess.InjectedStatement{
		Statement: pgnodes.NewCreateFunction(
			tableName.Table(),
			tableName.Schema(),
			node.Replace,
			retType,
			params,
			strict,
			ctx.originalQuery,
			extensionName,
			extensionSymbol,
			parsedBody,
			sqlDef,
			sqlDefParsedStmts,
			node.ReturnsSetOf,
		),
		Auth: vitess.AuthInformation{
			AuthType:    auth.AuthType_CREATE,
			TargetType:  auth.AuthTargetType_SchemaIdentifiers,
			TargetNames: []string{tableName.Catalog(), tableName.Schema()},
		},
		Children: defaults,
	}, nil
}

// createAnonymousCompositeType creates a new DoltgresType for the anonymous composite return
// type for a function, as represented by the |fieldTypes| that were specified in the function
// definition.
func createAnonymousCompositeType(fieldTypes []tree.SimpleColumnDef) *pgtypes.DoltgresType {
	attrs := make([]pgtypes.CompositeAttribute, len(fieldTypes))
	for i, fieldType := range fieldTypes {
		attrs[i] = pgtypes.NewCompositeAttribute(nil, id.Null, fieldType.Name.String(),
			id.NewType("", fieldType.Type.SQLString()), int16(i), "")
	}

	typeIdString := "table("
	for i, attr := range attrs {
		if i > 0 {
			typeIdString += ","
		}
		typeIdString += attr.Name
		typeIdString += ":"
		typeIdString += attr.TypeID.TypeName()
	}
	typeIdString += ")"

	// NOTE: there is no schema needed, since these types are anonymous and can't be directly referenced
	typeId := id.NewType("", typeIdString)

	return pgtypes.NewCompositeType(context.Background(), id.Null, id.NullType, typeId, attrs)
}

// handleLanguageSQLAs handles parsing SQL definition strings in both CREATE FUNCTION and CREATE PROCEDURE
// and returns converted the sql statements into vitess statements.
func handleLanguageSQLAs(definition string, params []pgnodes.RoutineParam) (string, []vitess.Statement, error) {
	stmts, err := parser.Parse(definition)
	if err != nil {
		return "", nil, err
	}

	return convertSQLStmts(stmts, params)
}

// convertSQLStmts takes parser.Statements and routine parameters and
// returns converted to string representation and vitess statements.
func convertSQLStmts(stmts parser.Statements, params []pgnodes.RoutineParam) (string, []vitess.Statement, error) {
	paramMap := make(map[string]*framework.ParamTypAndValue, len(params))
	for i, param := range params {
		tv := &framework.ParamTypAndValue{
			Typ:    param.Type,
			StrVal: "", // must be empty string
		}
		// placeholder name is empty
		if param.Name == "\"\"" {
			n := fmt.Sprintf("$%d", i+1)
			paramMap[n] = tv
			params[i].Name = n
		} else {
			paramMap[param.Name] = tv
		}
	}

	var sqlDefs = make([]string, len(stmts))
	var vitessASTs = make([]vitess.Statement, len(stmts))
	for i, stmt := range stmts {
		sqlDefs[i] = stmt.AST.String()
		err := framework.ReplaceFunctionColumn(stmt.AST, paramMap)
		if err != nil {
			return "", nil, err
		}
		// stmt.AST is updated at this point with FunctionColumn
		vitessASTs[i], err = Convert(stmt)
		if err != nil {
			return "", nil, err
		}
	}
	return strings.Join(sqlDefs, ";"), vitessASTs, nil
}

// validateRoutineOptions ensures that each option is defined only once. Returns a map containing all options, or an
// error if an option is invalid or is defined multiple times.
func validateRoutineOptions(ctx *Context, options []tree.RoutineOption) (map[tree.FunctionOption]tree.RoutineOption, error) {
	var optDefined = make(map[tree.FunctionOption]tree.RoutineOption)
	for _, opt := range options {
		if _, ok := optDefined[opt.OptionType]; ok {
			return nil, errors.Errorf("ERROR:  conflicting or redundant options")
		} else {
			optDefined[opt.OptionType] = opt
		}
	}
	return optDefined, nil
}
