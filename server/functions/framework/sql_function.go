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
	"strconv"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/core/id"
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
	stmt := f.SqlStatement
	// TODO: safer to parse and replace expression instead of replacing string representation
	for i, name := range f.ParameterNames {
		formattedVar, err := f.ParameterTypes[i].FormatValue(args[i])
		if err != nil {
			return nil, err
		}
		if name == "" {
			// sanity check
			stmt = strings.Replace(stmt, "$"+strconv.Itoa(i+1), formattedVar, 1)
		} else {
			stmt = strings.Replace(stmt, name, formattedVar, -1)
		}
	}

	// TODO: handle single row or multiple row result
	targetType := f.ReturnType

	return sql.RunInterpreted(ctx, func(subCtx *sql.Context) (any, error) {
		sch, rowIter, _, err := runner.QueryWithBindings(subCtx, stmt, nil, nil, nil)
		if err != nil {
			return nil, err
		}
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
		if targetType == nil {
			return rows[0][0], nil
		}
		if rows[0][0] == nil {
			return nil, nil
		}
		fromType, ok := sch[0].Type.(*pgtypes.DoltgresType)
		if !ok {
			fromType, err = pgtypes.FromGmsTypeToDoltgresType(sch[0].Type)
			if err != nil {
				return nil, err
			}
		}
		castFunc := GetAssignmentCast(fromType, targetType)
		if castFunc == nil {
			// TODO: We're using assignment casting, but for some reason we have to use I/O casting here, which is incorrect?
			//  We need to dig into this and figure out exactly what's happening, as this is "wrong" according to what
			//  I understand. This lines up more with explicit casting, but it's supposed to be assignment.
			//  Maybe there are specific rules for pgsql?
			if fromType.TypCategory == pgtypes.TypeCategory_StringTypes {
				castFunc = func(ctx *sql.Context, val any, targetType *pgtypes.DoltgresType) (any, error) {
					if val == nil {
						return nil, nil
					}
					str, err := fromType.IoOutput(ctx, val)
					if err != nil {
						return nil, err
					}
					return targetType.IoInput(ctx, str)
				}
			} else {
				return nil, errors.New("no valid cast for return value")
			}
		}
		return castFunc(subCtx, rows[0][0], targetType)
	})

	//return sql.RunInterpreted(ctx, func(subCtx *sql.Context) ([]sql.Row, error) {
	//	_, rowIter, _, err := runner.QueryWithBindings(subCtx, stmt, nil, nil, nil)
	//	if err != nil {
	//		return nil, err
	//	}
	//	// TODO: we should come up with a good way of carrying the RowIter out of the function without needing to wrap
	//	//  each call to QueryMultiReturn with RunInterpreted. For now, we don't check the returned rows, so this is
	//	//  fine.
	//	return sql.RowIterToRows(subCtx, rowIter)
	//})
}
