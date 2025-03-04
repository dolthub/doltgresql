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

package plpgsql

import (
	"fmt"
	"strings"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/analyzer"

	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/core/typecollection"
	"github.com/dolthub/doltgresql/postgres/parser/types"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// InterpretedFunction is an interface that essentially mirrors the implementation of InterpretedFunction in the
// framework package.
type InterpretedFunction interface {
	GetParameters() []*pgtypes.DoltgresType
	GetParameterNames() []string
	GetReturn() *pgtypes.DoltgresType
	GetStatements() []InterpreterOperation
	QueryMultiReturn(ctx *sql.Context, stack InterpreterStack, stmt string, bindings []string) (rowIter sql.RowIter, err error)
	QuerySingleReturn(ctx *sql.Context, stack InterpreterStack, stmt string, targetType *pgtypes.DoltgresType, bindings []string) (val any, err error)
}

// GetTypesCollectionFromContext is declared within the core package, but is assigned to this variable to work around
// import cycles.
var GetTypesCollectionFromContext func(ctx *sql.Context) (*typecollection.TypeCollection, error)

// Call runs the contained operations on the given runner.
func Call(ctx *sql.Context, iFunc InterpretedFunction, runner analyzer.StatementRunner, paramsAndReturn []*pgtypes.DoltgresType, vals []any) (any, error) {
	// Set up the initial state of the function
	counter := -1 // We increment before accessing, so start at -1
	stack := NewInterpreterStack(runner)
	// Add the parameters
	parameterTypes := iFunc.GetParameters()
	parameterNames := iFunc.GetParameterNames()
	if len(vals) != len(parameterTypes) {
		return nil, fmt.Errorf("parameter count mismatch: expected %d got %d", len(parameterTypes), len(vals))
	}
	for i := range vals {
		stack.NewVariableWithValue(parameterNames[i], parameterTypes[i], vals[i])
	}
	// Run the statements
	statements := iFunc.GetStatements()
	for {
		counter++
		if counter >= len(statements) {
			break
		} else if counter < 0 {
			panic("negative function counter")
		}

		operation := statements[counter]
		switch operation.OpCode {
		case OpCode_Alias:
			iv := stack.GetVariable(operation.PrimaryData)
			if iv == nil {
				return nil, fmt.Errorf("variable `%s` could not be found", operation.PrimaryData)
			}
			stack.NewVariableAlias(operation.Target, iv)
		case OpCode_Assign:
			iv := stack.GetVariable(operation.Target)
			if iv == nil {
				return nil, fmt.Errorf("variable `%s` could not be found", operation.Target)
			}
			retVal, err := iFunc.QuerySingleReturn(ctx, stack, operation.PrimaryData, iv.Type, operation.SecondaryData)
			if err != nil {
				return nil, err
			}
			err = stack.SetVariable(ctx, operation.Target, retVal)
			if err != nil {
				return nil, err
			}
		case OpCode_Case:
			// TODO: implement
		case OpCode_Declare:
			typeCollection, err := GetTypesCollectionFromContext(ctx)
			if err != nil {
				return nil, err
			}

			// pg_query_go sets PrimaryData for implicit CASE statement variables to
			// `pg_catalog."integer"`, so we remove double-quotes and extract the schema name.
			typeName := operation.PrimaryData
			typeName = strings.ReplaceAll(typeName, `"`, "")
			schemaName := "pg_catalog"
			if strings.Contains(typeName, ".") {
				parts := strings.Split(typeName, ".")
				schemaName = parts[0]
				typeName = parts[1]
				// Check the NonKeyword type names to see if we're looking at
				// an alias of a type if we're in the pg_catalog schema.
				if schemaName == "pg_catalog" {
					typ, ok, _ := types.TypeForNonKeywordTypeName(typeName)
					if ok && typ != nil {
						typeName = typ.Name()
					}
				}
			}
			resolvedType, exists := typeCollection.GetType(id.NewType(schemaName, typeName))
			if !exists {
				return nil, pgtypes.ErrTypeDoesNotExist.New(operation.PrimaryData)
			}
			stack.NewVariable(operation.Target, resolvedType)
		case OpCode_DeleteInto:
			// TODO: implement
		case OpCode_Exception:
			// TODO: implement
		case OpCode_Execute:
			if len(operation.Target) > 0 {
				target := stack.GetVariable(operation.Target)
				if target == nil {
					return nil, fmt.Errorf("variable `%s` could not be found", operation.Target)
				}
				retVal, err := iFunc.QuerySingleReturn(ctx, stack, operation.PrimaryData, target.Type, operation.SecondaryData)
				if err != nil {
					return nil, err
				}
				err = stack.SetVariable(ctx, operation.Target, retVal)
				if err != nil {
					return nil, err
				}
			} else {
				rowIter, err := iFunc.QueryMultiReturn(ctx, stack, operation.PrimaryData, operation.SecondaryData)
				if err != nil {
					return nil, err
				}
				if _, err = sql.RowIterToRows(ctx, rowIter); err != nil {
					return nil, err
				}
			}
		case OpCode_Get:
			// TODO: implement
		case OpCode_Goto:
			// We must compare to the index - 1, so that the increment hits our target
			if counter <= operation.Index {
				for ; counter < operation.Index-1; counter++ {
					switch statements[counter].OpCode {
					case OpCode_ScopeBegin:
						stack.PushScope()
					case OpCode_ScopeEnd:
						stack.PopScope()
					}
				}
			} else {
				for ; counter > operation.Index-1; counter-- {
					switch statements[counter].OpCode {
					case OpCode_ScopeBegin:
						stack.PopScope()
					case OpCode_ScopeEnd:
						stack.PushScope()
					}
				}
			}
		case OpCode_If:
			retVal, err := iFunc.QuerySingleReturn(ctx, stack, operation.PrimaryData, pgtypes.Bool, operation.SecondaryData)
			if err != nil {
				return nil, err
			}
			if retVal.(bool) {
				// We're never changing the scope, so we can just assign it directly.
				// Also, we must assign to index-1, so that the increment hits our target.
				counter = operation.Index - 1
			}
		case OpCode_InsertInto:
			// TODO: implement
		case OpCode_Perform:
			rowIter, err := iFunc.QueryMultiReturn(ctx, stack, operation.PrimaryData, operation.SecondaryData)
			if err != nil {
				return nil, err
			}
			if _, err = sql.RowIterToRows(ctx, rowIter); err != nil {
				return nil, err
			}
		case OpCode_Return:
			if len(operation.PrimaryData) == 0 {
				return nil, nil
			}
			return iFunc.QuerySingleReturn(ctx, stack, operation.PrimaryData, iFunc.GetReturn(), operation.SecondaryData)
		case OpCode_ScopeBegin:
			stack.PushScope()
		case OpCode_ScopeEnd:
			stack.PopScope()
		case OpCode_SelectInto:
			// TODO: implement
		case OpCode_UpdateInto:
			// TODO: implement
		default:
			panic("unimplemented opcode")
		}
	}
	return nil, nil
}
