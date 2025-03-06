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
	"strconv"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/lib/pq"

	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/core/interpreter"
	"github.com/dolthub/doltgresql/server/plpgsql"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// InterpretedFunction is the implementation of functions created using PL/pgSQL. The created functions are converted to
// a collection of operations, and an interpreter iterates over those operations to handle the logic.
type InterpretedFunction struct {
	ID                 id.Function
	ReturnType         *pgtypes.DoltgresType
	ParameterNames     []string
	ParameterTypes     []*pgtypes.DoltgresType
	Variadic           bool
	IsNonDeterministic bool
	Strict             bool
	Statements         []interpreter.InterpreterOperation
}

var _ FunctionInterface = InterpretedFunction{}
var _ plpgsql.InterpretedFunction = InterpretedFunction{}

// GetExpectedParameterCount implements the interface FunctionInterface.
func (iFunc InterpretedFunction) GetExpectedParameterCount() int {
	return len(iFunc.ParameterTypes)
}

// GetName implements the interface FunctionInterface.
func (iFunc InterpretedFunction) GetName() string {
	return iFunc.ID.FunctionName()
}

// GetParameters implements the interface FunctionInterface.
func (iFunc InterpretedFunction) GetParameters() []*pgtypes.DoltgresType {
	return iFunc.ParameterTypes
}

// GetParameterNames returns the names of all parameters.
func (iFunc InterpretedFunction) GetParameterNames() []string {
	return iFunc.ParameterNames
}

// GetReturn implements the interface FunctionInterface.
func (iFunc InterpretedFunction) GetReturn() *pgtypes.DoltgresType {
	return iFunc.ReturnType
}

// GetStatements returns the contained statements.
func (iFunc InterpretedFunction) GetStatements() []interpreter.InterpreterOperation {
	return iFunc.Statements
}

// InternalID implements the interface FunctionInterface.
func (iFunc InterpretedFunction) InternalID() id.Id {
	return iFunc.ID.AsId()
}

// IsStrict implements the interface FunctionInterface.
func (iFunc InterpretedFunction) IsStrict() bool {
	return iFunc.Strict
}

// Return implements the interface plan.Interpreter.
func (iFunc InterpretedFunction) Return(ctx *sql.Context) sql.Type {
	return iFunc.ReturnType
}

// NonDeterministic implements the interface FunctionInterface.
func (iFunc InterpretedFunction) NonDeterministic() bool {
	return iFunc.IsNonDeterministic
}

// VariadicIndex implements the interface FunctionInterface.
func (iFunc InterpretedFunction) VariadicIndex() int {
	// TODO: implement variadic
	return -1
}

// QuerySingleReturn handles queries that are supposed to return a single value.
func (InterpretedFunction) QuerySingleReturn(ctx *sql.Context, stack plpgsql.InterpreterStack, stmt string, targetType *pgtypes.DoltgresType, bindings []string) (val any, err error) {
	if len(bindings) > 0 {
		for i, bindingName := range bindings {
			variable := stack.GetVariable(bindingName)
			if variable == nil {
				return nil, fmt.Errorf("variable `%s` could not be found", bindingName)
			}
			formattedVar, err := variable.Type.FormatValue(variable.Value)
			if err != nil {
				return nil, err
			}
			switch variable.Type.TypCategory {
			case pgtypes.TypeCategory_ArrayTypes, pgtypes.TypeCategory_StringTypes:
				formattedVar = pq.QuoteLiteral(formattedVar)
			}
			stmt = strings.Replace(stmt, "$"+strconv.Itoa(i+1), formattedVar, 1)
		}
	}
	sch, rowIter, _, err := stack.Runner().QueryWithBindings(ctx, stmt, nil, nil, nil)
	if err != nil {
		return nil, err
	}
	rows, err := sql.RowIterToRows(ctx, rowIter)
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
	fromType, ok := sch[0].Type.(*pgtypes.DoltgresType)
	if !ok {
		fromType, err = pgtypes.FromGmsTypeToDoltgresType(sch[0].Type)
		if err != nil {
			return nil, err
		}
	}
	castFunc := GetAssignmentCast(fromType, targetType)
	if castFunc == nil {
		// TODO: try I/O casting
		return nil, errors.New("no valid cast for return value")
	}
	return castFunc(ctx, rows[0][0], targetType)
}

// QueryMultiReturn handles queries that may return multiple values over multiple rows.
func (InterpretedFunction) QueryMultiReturn(ctx *sql.Context, stack plpgsql.InterpreterStack, stmt string, bindings []string) (rowIter sql.RowIter, err error) {
	if len(bindings) > 0 {
		for i, bindingName := range bindings {
			variable := stack.GetVariable(bindingName)
			if variable == nil {
				return nil, fmt.Errorf("variable `%s` could not be found", bindingName)
			}
			formattedVar, err := variable.Type.FormatValue(variable.Value)
			if err != nil {
				return nil, err
			}
			switch variable.Type.TypCategory {
			case pgtypes.TypeCategory_ArrayTypes, pgtypes.TypeCategory_StringTypes:
				formattedVar = pq.QuoteLiteral(formattedVar)
			}
			stmt = strings.Replace(stmt, "$"+strconv.Itoa(i+1), formattedVar, 1)
		}
	}
	_, rowIter, _, err = stack.Runner().QueryWithBindings(ctx, stmt, nil, nil, nil)
	return rowIter, err
}

// enforceInterfaceInheritance implements the interface FunctionInterface.
func (iFunc InterpretedFunction) enforceInterfaceInheritance(error) {}
