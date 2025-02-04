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

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/analyzer"

	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/core/id"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// Call runs the contained operations on the given runner.
func (iFunc InterpretedFunction) Call(ctx *sql.Context, runner analyzer.StatementRunner, paramsAndReturn []*pgtypes.DoltgresType, vals []any) (any, error) {
	// Set up the initial state of the function
	counter := -1 // We increment before accessing, so start at -1
	stack := NewInterpreterStack(runner)
	// Add the parameters
	if len(vals) != len(iFunc.ParameterTypes) {
		return nil, fmt.Errorf("parameter count mismatch: expected `%d` got %d`", len(iFunc.ParameterTypes), len(vals))
	}
	for i := range vals {
		stack.NewVariableWithValue(iFunc.ParameterNames[i], iFunc.ParameterTypes[i], vals[i])
	}
	// Run the statements
	for {
		counter++
		if counter >= len(iFunc.Statements) {
			break
		} else if counter < 0 {
			panic("negative function counter")
		}

		operation := iFunc.Statements[counter]
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
			retVal, err := iFunc.querySingleReturn(ctx, stack, operation.PrimaryData, iv.Type, operation.SecondaryData)
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
			typeCollection, err := core.GetTypesCollectionFromContext(ctx)
			if err != nil {
				return nil, err
			}
			resolvedType, exists := typeCollection.GetType(id.NewType("pg_catalog", operation.PrimaryData))
			if !exists {
				return nil, pgtypes.ErrTypeDoesNotExist.New(operation.PrimaryData)
			}
			stack.NewVariable(operation.Target, resolvedType)
		case OpCode_DeleteInto:
			// TODO: implement
		case OpCode_Exception:
			// TODO: implement
		case OpCode_Execute:
			rowIter, err := iFunc.queryMultiReturn(ctx, stack, operation.PrimaryData, operation.SecondaryData)
			if err != nil {
				return nil, err
			}
			if err = rowIter.Close(ctx); err != nil {
				return nil, err
			}
		case OpCode_For:
			// TODO: implement
		case OpCode_Foreach:
			// TODO: implement
		case OpCode_Get:
			// TODO: implement
		case OpCode_Goto:
			// We must compare to the index - 1, so that the increment hits our target
			if counter <= operation.Index {
				for ; counter < operation.Index-1; counter++ {
					switch iFunc.Statements[counter].OpCode {
					case OpCode_ScopeBegin:
						stack.PushScope()
					case OpCode_ScopeEnd:
						stack.PopScope()
					}
				}
			} else {
				for ; counter > operation.Index-1; counter-- {
					switch iFunc.Statements[counter].OpCode {
					case OpCode_ScopeBegin:
						stack.PopScope()
					case OpCode_ScopeEnd:
						stack.PushScope()
					}
				}
			}
		case OpCode_If:
			retVal, err := iFunc.querySingleReturn(ctx, stack, operation.PrimaryData, pgtypes.Bool, operation.SecondaryData)
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
		case OpCode_Loop:
			// TODO: implement
		case OpCode_Perform:
			rowIter, err := iFunc.queryMultiReturn(ctx, stack, operation.PrimaryData, operation.SecondaryData)
			if err != nil {
				return nil, err
			}
			if err = rowIter.Close(ctx); err != nil {
				return nil, err
			}
		case OpCode_Query:
			rowIter, err := iFunc.queryMultiReturn(ctx, stack, operation.PrimaryData, operation.SecondaryData)
			if err != nil {
				return nil, err
			}
			if err = rowIter.Close(ctx); err != nil {
				return nil, err
			}
		case OpCode_Return:
			if len(operation.PrimaryData) == 0 {
				return nil, nil
			}
			return iFunc.querySingleReturn(ctx, stack, operation.PrimaryData, iFunc.ReturnType, operation.SecondaryData)
		case OpCode_ScopeBegin:
			stack.PushScope()
		case OpCode_ScopeEnd:
			stack.PopScope()
		case OpCode_SelectInto:
			// TODO: implement
		case OpCode_When:
			// TODO: implement
		case OpCode_While:
			// TODO: implement
		case OpCode_UpdateInto:
			// TODO: implement
		default:
			panic("unimplemented opcode")
		}
	}
	return nil, nil
}
