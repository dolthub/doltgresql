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

package framework

import (
	"fmt"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/postgres/parser/parser"
	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// SQLFunction is the implementation of functions created using SQL.
type SQLFunction struct {
	ID                 id.Function
	ReturnType         *pgtypes.DoltgresType
	ParameterNames     []string
	ParameterTypes     []*pgtypes.DoltgresType
	Variadic           bool
	IsNonDeterministic bool
	Strict             bool
	SqlStatement       string
	SetOf              bool
	ReturnTableType    []*pgtypes.DoltgresType
}

var _ FunctionInterface = SQLFunction{}

// GetExpectedParameterCount implements the interface FunctionInterface.
func (sqlFunc SQLFunction) GetExpectedParameterCount() int {
	return len(sqlFunc.ParameterTypes)
}

// GetName implements the interface FunctionInterface.
func (sqlFunc SQLFunction) GetName() string {
	return sqlFunc.ID.FunctionName()
}

// GetParameters implements the interface FunctionInterface.
func (sqlFunc SQLFunction) GetParameters() []*pgtypes.DoltgresType {
	return sqlFunc.ParameterTypes
}

// GetReturn implements the interface FunctionInterface.
func (sqlFunc SQLFunction) GetReturn() *pgtypes.DoltgresType {
	return sqlFunc.ReturnType
}

// InternalID implements the interface FunctionInterface.
func (sqlFunc SQLFunction) InternalID() id.Id {
	return sqlFunc.ID.AsId()
}

// IsStrict implements the interface FunctionInterface.
func (sqlFunc SQLFunction) IsStrict() bool {
	return sqlFunc.Strict
}

// NonDeterministic implements the interface FunctionInterface.
func (sqlFunc SQLFunction) NonDeterministic() bool {
	return sqlFunc.IsNonDeterministic
}

// VariadicIndex implements the interface FunctionInterface.
func (sqlFunc SQLFunction) VariadicIndex() int {
	// TODO: implement variadic
	return -1
}

// IsSRF implements the interface FunctionInterface.
func (sqlFunc SQLFunction) IsSRF() bool {
	return false
}

// enforceInterfaceInheritance implements the interface FunctionInterface.
func (sqlFunc SQLFunction) enforceInterfaceInheritance(error) {}

// CallSqlFunction runs the given SQL definition inside the function on the given runner.
func CallSqlFunction(ctx *sql.Context, f SQLFunction, runner sql.StatementRunner, args []any) (any, error) {
	parsed, err := parser.ParseOne(f.SqlStatement)
	if err != nil {
		return "", err
	}

	paramMap := make(map[string]*ParamTypAndValue)
	for i, name := range f.ParameterNames {
		formattedVar, err := f.ParameterTypes[i].FormatValue(args[i])
		if err != nil {
			return nil, err
		}
		if name == "" {
			// sanity check
			name = fmt.Sprintf("$%d", i+1)
		}
		paramMap[name] = &ParamTypAndValue{
			Typ:    f.ParameterTypes[i],
			StrVal: formattedVar,
		}
	}

	err = ReplaceFunctionColumn(parsed.AST, paramMap)
	if err != nil {
		return nil, err
	}
	// stmt.AST is updated at this point with FunctionColumn
	return sql.RunInterpreted(ctx, func(subCtx *sql.Context) (any, error) {
		sch, rowIter, _, err := runner.QueryWithBindings(ctx, parsed.AST.String(), nil, nil, nil)
		if err != nil {
			return nil, err
		}

		if !f.SetOf {
			// single row result
			rows, err := sql.RowIterToRows(subCtx, rowIter)
			if err != nil {
				return nil, err
			}
			if len(sch) != 1 {
				return nil, errors.New("expression does not result in a single value")
			}
			if len(rows) != 1 {
				return nil, errors.New("expression returned multiple result sets")
			}
			if len(rows[0]) != 1 {
				return nil, errors.New("expression returned multiple results")
			}
			return rows[0][0], nil
		}
		// multiple row result
		return rowIter, nil
	})
}

// ParamTypAndValue contains the parameter type and
// string value of argument if applicable
type ParamTypAndValue struct {
	Typ    *pgtypes.DoltgresType
	StrVal string
}

// ReplaceFunctionColumn parses and replaces UnresolvedName and Placeholder expressions
// with FunctionColumn expression containing parameter type and arguments if applicable.
// It also replaces empty parameter name with binding variable name to match the name used in FunctionColumn.
// This function should be used for FUNCTION with SQL language statements only.
func ReplaceFunctionColumn(parsedAST tree.Statement, params map[string]*ParamTypAndValue) error {
	// Function's final statement must be SELECT or INSERT/UPDATE/DELETE RETURNING
	switch s := parsedAST.(type) {
	case *tree.Select:
		sc := s.Select.(*tree.SelectClause)
		for i, e := range sc.Exprs {
			sc.Exprs[i].Expr = ReplaceUnresolvedToFunctionColumn(params, e.Expr)
		}
		if sc.Where != nil {
			sc.Where.Expr = ReplaceUnresolvedToFunctionColumn(params, sc.Where.Expr)
		}
		return nil
	case *tree.Insert:
		if s.Returning != nil {
			return errors.Errorf("INSERT ... RETURNING statement in functions is not yet supported")
		}
	case *tree.Update:
		if s.Returning != nil {
			return errors.Errorf("UPDATE ... RETURNING statement in functions is not yet supported")
		}
	case *tree.Delete:
		if s.Returning != nil {
			return errors.Errorf("DELETE ... RETURNING statement in functions is not yet supported")
		}
	}
	return errors.Errorf("Function's final statement must be SELECT or INSERT/UPDATE/DELETE RETURNING")
}

// ReplaceUnresolvedToFunctionColumn replaces Placeholder and UnresolvedName expressions with FunctionColumn containing
// parameter type and argument value if applicable when the name of expression matches function parameter.
func ReplaceUnresolvedToFunctionColumn(paramMap map[string]*ParamTypAndValue, expr tree.Expr) tree.Expr {
	e, _ := tree.SimpleVisit(expr, func(visitingExpr tree.Expr) (recurse bool, newExpr tree.Expr, err error) {
		switch v := visitingExpr.(type) {
		case *tree.Placeholder:
			name := fmt.Sprintf("$%d", v.Idx+1)
			if tv, ok := paramMap[name]; ok {
				return false, tree.FunctionColumn{
					Name:   name,
					Typ:    tv.Typ,
					Idx:    uint16(v.Idx),
					StrVal: tv.StrVal,
				}, nil
			}
		case *tree.UnresolvedName:
			name := v.String()
			if tv, ok := paramMap[name]; ok {
				return false, tree.FunctionColumn{
					Name:   name,
					Typ:    tv.Typ,
					StrVal: tv.StrVal,
				}, nil
			}
		}
		return true, visitingExpr, nil
	})
	return e
}
