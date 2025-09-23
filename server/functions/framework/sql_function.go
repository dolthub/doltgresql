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
	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/postgres/parser/parser"
	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
	pgtypes "github.com/dolthub/doltgresql/server/types"
	"github.com/dolthub/go-mysql-server/sql"
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
	paramMap := make(map[string]string)
	for i, name := range f.ParameterNames {
		formattedVar, err := f.ParameterTypes[i].FormatValue(args[i])
		if err != nil {
			return nil, err
		}
		if name == "" {
			// sanity check
			name = fmt.Sprintf("$%d", i+1)
		}
		paramMap[name] = formattedVar
	}

	query, err := parseAndReplaceFunctionColumn(ctx, f.SqlStatement, paramMap)
	if err != nil {
		return nil, err
	}

	return sql.RunInterpreted(ctx, func(subCtx *sql.Context) (any, error) {
		sch, rowIter, _, err := runner.QueryWithBindings(ctx, query, nil, nil, nil)
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

// parseAndReplaceFunctionColumn parses and replaces function parameter expressions with given arguments.
func parseAndReplaceFunctionColumn(ctx *sql.Context, q string, params map[string]string) (string, error) {
	parsed, err := parser.ParseOne(q)
	if err != nil {
		return "", err
	}

	// Function's final statement must be SELECT or INSERT/UPDATE/DELETE RETURNING
	switch s := parsed.AST.(type) {
	case *tree.Select:
		sc := s.Select.(*tree.SelectClause)
		for i, e := range sc.Exprs {
			sc.Exprs[i].Expr = replaceToFunctionColumn(params, e.Expr)
		}
		if sc.Where != nil {
			sc.Where.Expr = replaceToFunctionColumn(params, sc.Where.Expr)
		}
	}

	return parsed.AST.String(), nil
}

// replaceToFunctionColumn replaces Placeholder and UnresolvedName expressions with FunctionColumn containing
// argument value if applicable when the name of expression matches function parameter.
func replaceToFunctionColumn(paramMap map[string]string, expr tree.Expr) tree.Expr {
	e, _ := tree.SimpleVisit(expr, func(visitingExpr tree.Expr) (recurse bool, newExpr tree.Expr, err error) {
		switch v := visitingExpr.(type) {
		case *tree.Placeholder:
			name := fmt.Sprintf("$%d", v.Idx+1)
			if strval, ok := paramMap[name]; ok {
				return false, tree.FunctionColumn{
					Name:   name,
					Idx:    uint16(v.Idx),
					StrVal: strval,
				}, nil
			}
		case *tree.UnresolvedName:
			name := v.String()
			if strval, ok := paramMap[name]; ok {
				return false, tree.FunctionColumn{
					Name:   name,
					StrVal: strval,
				}, nil
			}
		}
		return true, visitingExpr, nil
	})
	return e
}
