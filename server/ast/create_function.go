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

	"github.com/dolthub/doltgresql/postgres/parser/parser"
	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
	"github.com/dolthub/doltgresql/postgres/parser/types"
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
	retType := pgtypes.Void
	if len(node.RetType) == 1 {
		switch typ := node.RetType[0].Type.(type) {
		case *types.T:
			retType = pgtypes.NewUnresolvedDoltgresType("", strings.ToLower(typ.Name()))
		default:
			sqlString := strings.ToLower(typ.SQLString())
			if sqlString == "trigger" {
				retType = pgtypes.Trigger
			} else {
				retType = pgtypes.NewUnresolvedDoltgresType("", sqlString)
			}
		}
	}
	paramNames := make([]string, len(node.Args))
	paramTypes := make([]*pgtypes.DoltgresType, len(node.Args))
	for i, arg := range node.Args {
		paramNames[i] = arg.Name.String()
		switch argType := arg.Type.(type) {
		case *types.T:
			paramTypes[i] = pgtypes.NewUnresolvedDoltgresType("", strings.ToLower(argType.Name()))
		default:
			paramTypes[i] = pgtypes.NewUnresolvedDoltgresType("", strings.ToLower(argType.SQLString()))
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
	var sqlDefParsed vitess.Statement
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
		case "sql":
			as, ok := options[tree.OptionAs1]
			if !ok {
				return nil, errors.Errorf("CREATE FUNCTION definition needed for LANGUAGE SQL")
			}
			stmts, err := parser.Parse(as.Definition)
			if err != nil {
				return nil, err
			}
			if len(stmts) > 1 {
				return nil, fmt.Errorf("only a single statement at a time is currently supported")
			}
			if len(stmts) == 0 {
				return nil, vitess.ErrEmpty
			}
			sqlDef = stmts[0].AST.String()
			paramNames, err = replaceFunctionColumnAndUpdateParamNames(paramNames, paramTypes, stmts[0].AST)
			if err != nil {
				return nil, err
			}
			// stmts[0].AST is updated at this point with FunctionColumn
			vitessAST, err := Convert(stmts[0])
			if err != nil {
				return nil, err
			}
			sqlDefParsed = vitessAST
		case "c":
			symbolOption, ok := options[tree.OptionAs2]
			if !ok {
				return nil, errors.Errorf("LANGUAGE C is only supported when providing both the module name and symbol")
			}
			extensionName = symbolOption.ObjFile
			extensionSymbol = symbolOption.LinkSymbol
		default:
			return nil, errors.Errorf("CREATE FUNCTION only supports PL/pgSQL for now")
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
			paramNames,
			paramTypes,
			strict,
			ctx.originalQuery,
			extensionName,
			extensionSymbol,
			parsedBody,
			sqlDef,
			sqlDefParsed,
			node.SetOf,
		),
	}, nil
}

// replaceFunctionColumnAndUpdateParamNames replaces UnresolvedName and Placeholder expressions with FunctionColumn expression.
// It also replaces empty parameter name with binding variable name to match the name used in FunctionColumn.
// This function should be used for FUNCTION with SQL language statements only.
func replaceFunctionColumnAndUpdateParamNames(paramNames []string, paramTypes []*pgtypes.DoltgresType, statement tree.Statement) ([]string, error) {
	paramMap := make(map[string]*pgtypes.DoltgresType, len(paramNames))
	if len(paramNames) != len(paramTypes) {
		return paramNames, errors.Errorf("expected %d parameters but got %d", len(paramNames), len(paramTypes))
	}
	for i, paramName := range paramNames {
		// placeholder name is empty
		if paramName == "\"\"" {
			n := fmt.Sprintf("$%v", i+1)
			paramMap[n] = paramTypes[i]
			paramNames[i] = n
		} else {
			paramMap[paramName] = paramTypes[i]
		}
	}

	// Function's final statement must be SELECT or INSERT/UPDATE/DELETE RETURNING
	switch s := statement.(type) {
	case *tree.Select:
		sc := s.Select.(*tree.SelectClause)
		for i, e := range sc.Exprs {
			sc.Exprs[i].Expr = replaceToFunctionColumn(paramMap, e.Expr)
		}
		if sc.Where != nil {
			sc.Where.Expr = replaceToFunctionColumn(paramMap, sc.Where.Expr)
		}
		return paramNames, nil
	case *tree.Insert:
		if s.Returning != nil {
			return paramNames, errors.Errorf("INSERT ... RETURNING statement is not supported in functions yet")
		}
	case *tree.Update:
		if s.Returning != nil {
			return paramNames, errors.Errorf("UPDATE ... RETURNING statement is not supported in functions yet")
		}
	case *tree.Delete:
		if s.Returning != nil {
			return paramNames, errors.Errorf("DELETE ... RETURNING statement is not supported in functions yet")
		}
	}
	return paramNames, errors.Errorf("Function's final statement must be SELECT or INSERT/UPDATE/DELETE RETURNING")
}

// replaceToFunctionColumn replaces Placeholder and UnresolvedName expressions with FunctionColumn if applicable
// when the name of expression matches parameter in paramMap.
func replaceToFunctionColumn(paramMap map[string]*pgtypes.DoltgresType, expr tree.Expr) tree.Expr {
	e, _ := tree.SimpleVisit(expr, func(visitingExpr tree.Expr) (recurse bool, newExpr tree.Expr, err error) {
		switch v := visitingExpr.(type) {
		case *tree.Placeholder:
			name := fmt.Sprintf("$%d", v.Idx+1)
			if typ, ok := paramMap[name]; ok {
				return false, tree.FunctionColumn{
					Name: name,
					Typ:  typ,
					Idx:  uint16(v.Idx),
				}, nil
			}
		case *tree.UnresolvedName:
			name := v.String()
			if typ, ok := paramMap[name]; ok {
				return false, tree.FunctionColumn{
					Name: name,
					Typ:  typ,
				}, nil
			}
		}
		return true, visitingExpr, nil
	})
	return e
}
