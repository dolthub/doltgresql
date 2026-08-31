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

	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/core/procedures"
	"github.com/dolthub/doltgresql/server/plpgsql"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// InterpretedFunction is the implementation of functions created using PL/pgSQL. The created functions are converted to
// a collection of operations, and an interpreter iterates over those operations to handle the logic.
type InterpretedFunction struct {
	ID                 id.Function
	ReturnType         *pgtypes.DoltgresType
	AllParams          []procedures.Parameter
	AllTypes           []*pgtypes.DoltgresType
	Variadic           bool
	IsNonDeterministic bool
	Strict             bool
	SRF                bool
	Statements         []plpgsql.InterpreterOperation
}

var _ FunctionInterface = InterpretedFunction{}
var _ plpgsql.InterpretedFunction = InterpretedFunction{}

// GetExpectedParameterCount implements the interface FunctionInterface.
func (iFunc InterpretedFunction) GetExpectedParameterCount() int {
	inputParamCount := 0
	for _, param := range iFunc.AllParams {
		switch param.Mode {
		case procedures.ParameterMode_IN, procedures.ParameterMode_INOUT, procedures.ParameterMode_VARIADIC:
			inputParamCount += 1
		}
	}
	return inputParamCount
}

// GetName implements the interface FunctionInterface.
func (iFunc InterpretedFunction) GetName() string {
	return iFunc.ID.FunctionName()
}

// GetAllNames implements the interface InterpretedFunction.
func (iFunc InterpretedFunction) GetAllNames() []string {
	var names []string
	for _, param := range iFunc.AllParams {
		names = append(names, param.Name)
	}
	return names
}

// GetInputParameterNamesAndTypes implements the interface InterpretedFunction.
func (iFunc InterpretedFunction) GetInputParameterNamesAndTypes() ([]string, []*pgtypes.DoltgresType) {
	var names []string
	var typs []*pgtypes.DoltgresType
	for i, param := range iFunc.AllParams {
		switch param.Mode {
		case procedures.ParameterMode_IN, procedures.ParameterMode_INOUT, procedures.ParameterMode_VARIADIC:
			names = append(names, param.Name)
			typs = append(typs, iFunc.AllTypes[i])
		}
	}
	return names, typs
}

// GetInputParameterTypes implements the interface FunctionInterface.
func (iFunc InterpretedFunction) GetInputParameterTypes() []*pgtypes.DoltgresType {
	var typs []*pgtypes.DoltgresType
	for i, param := range iFunc.AllParams {
		switch param.Mode {
		case procedures.ParameterMode_IN, procedures.ParameterMode_INOUT, procedures.ParameterMode_VARIADIC:
			typs = append(typs, iFunc.AllTypes[i])
		}
	}
	return typs
}

// GetOutputParameterNamesAndTypes implements the interface InterpretedFunction.
func (iFunc InterpretedFunction) GetOutputParameterNamesAndTypes() ([]string, []*pgtypes.DoltgresType) {
	var names []string
	var typs []*pgtypes.DoltgresType
	for i, param := range iFunc.AllParams {
		switch param.Mode {
		case procedures.ParameterMode_OUT, procedures.ParameterMode_INOUT:
			names = append(names, param.Name)
			typs = append(typs, iFunc.AllTypes[i])
		}
	}
	return names, typs
}

// GetOutParameters implements the interface FunctionInterface.
func (iFunc InterpretedFunction) GetOutParameters() sql.Schema {
	var outParams []*sql.Column
	for i, param := range iFunc.AllParams {
		switch param.Mode {
		case procedures.ParameterMode_OUT, procedures.ParameterMode_INOUT:
			outParams = append(outParams, &sql.Column{
				Name: param.Name,
				Type: iFunc.AllTypes[i],
				// TODO default val ?
			})
		}
	}
	return outParams
}

// GetReturn implements the interface FunctionInterface.
func (iFunc InterpretedFunction) GetReturn() *pgtypes.DoltgresType {
	return iFunc.ReturnType
}

// GetStatements returns the contained statements.
func (iFunc InterpretedFunction) GetStatements() []plpgsql.InterpreterOperation {
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

// IsSRF implements the interface FunctionInterface.
func (iFunc InterpretedFunction) IsSRF() bool {
	return iFunc.SRF
}

// NonDeterministic implements the interface FunctionInterface.
func (iFunc InterpretedFunction) NonDeterministic() bool {
	return iFunc.IsNonDeterministic
}

// IsCVariadic implements the FunctionInterface interface.
func (iFunc InterpretedFunction) IsCVariadic() bool {
	return false
}

// VariadicIndex implements the interface FunctionInterface.
func (iFunc InterpretedFunction) VariadicIndex() int {
	// TODO: implement variadic
	return -1
}

// QuerySingleReturn handles queries that are supposed to return a single value.
func (iFunc InterpretedFunction) QuerySingleReturn(ctx *sql.Context, stack plpgsql.InterpreterStack, stmt string, targetType *pgtypes.DoltgresType, bindings []string) (val any, err error) {
	stmt, _, err = iFunc.ApplyBindings(ctx, stack, stmt, bindings, true)
	if err != nil {
		return nil, err
	}

	return sql.RunInterpreted(ctx, func(subCtx *sql.Context) (any, error) {
		sch, rowIter, _, err := stack.Runner().QueryWithBindings(subCtx, stmt, nil, nil, nil)
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
		sourceType, ok := sch[0].Type.(*pgtypes.DoltgresType)
		if !ok {
			// TODO: We ensure we have a DoltgresType, but we should also convert the value to
			//       ensure it's in the correct form for the DoltgresType. This logic lives in
			//       pgexpressions.GMSCast, but need to be extracted to avoid a dependency cycle
			//       so it can be used here and from server.plpgsql.
			sourceType, err = pgtypes.FromGmsTypeToDoltgresType(sch[0].Type)
			if err != nil {
				return nil, err
			}
		}
		castsColl, err := core.GetCastsCollectionFromContext(ctx, "")
		if err != nil {
			return nil, err
		}
		cast, err := castsColl.GetAssignmentCast(ctx, sourceType, targetType)
		if err != nil {
			return nil, err
		}
		if !cast.ID.IsValid() {
			// TODO: We're using assignment casting, but for some reason we have to use I/O casting here, which is incorrect?
			//  We need to dig into this and figure out exactly what's happening, as this is "wrong" according to what
			//  I understand. This lines up more with explicit casting, but it's supposed to be assignment.
			//  Maybe there are specific rules for pgsql?
			if sourceType.TypCategory == pgtypes.TypeCategory_StringTypes {
				cast.ID = id.NewCast(sourceType.ID, targetType.ID)
				cast.UseInOut = true
			} else {
				return nil, errors.New("no valid cast for return value")
			}
		}
		return cast.Eval(subCtx, rows[0][0], sourceType, targetType)
	})
}

// QueryMultiReturn handles queries that may return multiple values over multiple rows.
func (iFunc InterpretedFunction) QueryMultiReturn(ctx *sql.Context, stack plpgsql.InterpreterStack, stmt string, bindings []string) (schema sql.Schema, rows []sql.Row, err error) {
	stmt, _, err = iFunc.ApplyBindings(ctx, stack, stmt, bindings, true)
	if err != nil {
		return nil, nil, err
	}

	rows, err = sql.RunInterpreted(ctx, func(subCtx *sql.Context) (rows []sql.Row, err error) {
		var rowIter sql.RowIter
		schema, rowIter, _, err = stack.Runner().QueryWithBindings(subCtx, stmt, nil, nil, nil)
		if err != nil {
			return nil, err
		}
		// TODO: we should come up with a good way of carrying the RowIter out of the function without needing to wrap
		//  each call to QueryMultiReturn with RunInterpreted. For now, we don't check the returned rows, so this is
		//  fine.
		return sql.RowIterToRows(subCtx, rowIter)
	})
	return schema, rows, err
}

// ApplyBindings applies the given bindings to the statement. If `varFound` is false, then the error will state that
// the variable was not found (which means the error may be ignored if you're only concerned with finding a variable).
// If `varFound` is true, then the error is related to formatting the variable. `enforceType` adds casting and quotes to
// ensure that the value is correctly represented in the string.
func (InterpretedFunction) ApplyBindings(ctx *sql.Context, stack plpgsql.InterpreterStack, stmt string, bindings []string, enforceType bool) (newStmt string, varFound bool, err error) {
	if len(bindings) == 0 {
		return stmt, false, nil
	}
	newStmt = stmt
	for i, bindingName := range bindings {
		variable, err := stack.GetVariableWithError(bindingName)
		if err != nil {
			// Only a name that matches no variable at all leaves the caller free to try something else; a
			// bad field on a variable that does exist is a real error that has to surface.
			return newStmt, !plpgsql.ErrVariableNotFound.Is(err), err
		}
		if variable.Type == nil {
			return newStmt, false, plpgsql.ErrVariableNotFound.New(bindingName)
		}
		var formattedVar string
		if *variable.Value != nil {
			formattedVar, err = variable.Type.FormatValueWithContext(ctx, *variable.Value)
			if err != nil {
				return newStmt, true, err
			}
			if enforceType {
				switch variable.Type.TypCategory {
				case pgtypes.TypeCategory_ArrayTypes, pgtypes.TypeCategory_CompositeTypes, pgtypes.TypeCategory_DateTimeTypes, pgtypes.TypeCategory_StringTypes, pgtypes.TypeCategory_UserDefinedTypes:
					formattedVar = pq.QuoteLiteral(formattedVar)
				}
			}
		} else {
			formattedVar = "NULL"
		}
		if enforceType {
			if variable.Type.TypCategory == pgtypes.TypeCategory_CompositeTypes {
				newStmt = strings.Replace(newStmt, "$"+strconv.Itoa(i+1), fmt.Sprintf(`(%s::%s)`, formattedVar, variable.Type.String()), 1)
			} else {
				newStmt = strings.Replace(newStmt, "$"+strconv.Itoa(i+1), fmt.Sprintf(`((%s)::%s)`, formattedVar, variable.Type.String()), 1)
			}
		} else {
			newStmt = strings.Replace(newStmt, "$"+strconv.Itoa(i+1), formattedVar, 1)
		}
	}
	return newStmt, true, nil
}

// enforceInterfaceInheritance implements the interface FunctionInterface.
func (iFunc InterpretedFunction) enforceInterfaceInheritance(error) {}
