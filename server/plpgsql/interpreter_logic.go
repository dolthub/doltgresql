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
	"github.com/jackc/pgx/v5/pgproto3"

	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/core/interpreter"
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
	GetStatements() []interpreter.InterpreterOperation
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
		case interpreter.OpCode_Alias:
			iv := stack.GetVariable(operation.PrimaryData)
			if iv == nil {
				return nil, fmt.Errorf("variable `%s` could not be found", operation.PrimaryData)
			}
			stack.NewVariableAlias(operation.Target, iv)
		case interpreter.OpCode_Assign:
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
		case interpreter.OpCode_Declare:
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
		case interpreter.OpCode_DeleteInto:
			// TODO: implement
		case interpreter.OpCode_Exception:
			// TODO: implement
		case interpreter.OpCode_Execute:
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
		case interpreter.OpCode_Get:
			// TODO: implement
		case interpreter.OpCode_Goto:
			// We must compare to the index - 1, so that the increment hits our target
			if counter <= operation.Index {
				for ; counter < operation.Index-1; counter++ {
					switch statements[counter].OpCode {
					case interpreter.OpCode_ScopeBegin:
						stack.PushScope()
					case interpreter.OpCode_ScopeEnd:
						stack.PopScope()
					}
				}
			} else {
				for ; counter > operation.Index-1; counter-- {
					switch statements[counter].OpCode {
					case interpreter.OpCode_ScopeBegin:
						stack.PopScope()
					case interpreter.OpCode_ScopeEnd:
						stack.PushScope()
					}
				}
			}
		case interpreter.OpCode_If:
			retVal, err := iFunc.QuerySingleReturn(ctx, stack, operation.PrimaryData, pgtypes.Bool, operation.SecondaryData)
			if err != nil {
				return nil, err
			}
			if retVal.(bool) {
				// We're never changing the scope, so we can just assign it directly.
				// Also, we must assign to index-1, so that the increment hits our target.
				counter = operation.Index - 1
			}
		case interpreter.OpCode_InsertInto:
			// TODO: implement
		case interpreter.OpCode_Perform:
			rowIter, err := iFunc.QueryMultiReturn(ctx, stack, operation.PrimaryData, operation.SecondaryData)
			if err != nil {
				return nil, err
			}
			if _, err = sql.RowIterToRows(ctx, rowIter); err != nil {
				return nil, err
			}
		case interpreter.OpCode_Raise:
			backend, err := core.GetBackend(ctx)
			if err != nil {
				return nil, err
			}

			// TODO: Use the client_min_messages config param to determine which
			//       notice levels to send to the client.
			// https://www.postgresql.org/docs/current/runtime-config-client.html#GUC-CLIENT-MIN-MESSAGES

			// TODO: Notices at the EXCEPTION level should also abort the current tx.

			message, err := evaluteNoticeMessage(ctx, iFunc, operation, stack)
			if err != nil {
				return nil, err
			}

			noticeResponse := &pgproto3.NoticeResponse{
				Severity: operation.PrimaryData,
				Message:  message,
			}

			applyNoticeOptions(ctx, noticeResponse, operation.Options.(map[uint8]string))
			backend.Send(noticeResponse)
			if err = backend.Flush(); err != nil {
				return nil, err
			}
		case interpreter.OpCode_Return:
			if len(operation.PrimaryData) == 0 {
				return nil, nil
			}
			return iFunc.QuerySingleReturn(ctx, stack, operation.PrimaryData, iFunc.GetReturn(), operation.SecondaryData)
		case interpreter.OpCode_ScopeBegin:
			stack.PushScope()
		case interpreter.OpCode_ScopeEnd:
			stack.PopScope()
		case interpreter.OpCode_SelectInto:
			// TODO: implement
		case interpreter.OpCode_UpdateInto:
			// TODO: implement
		default:
			panic("unimplemented opcode")
		}
	}
	return nil, nil
}

// applyNoticeOptions adds the specified |options| to the |noticeResponse|.
func applyNoticeOptions(ctx *sql.Context, noticeResponse *pgproto3.NoticeResponse, options map[uint8]string) {
	for key, value := range options {
		switch NoticeOptionType(key) {
		case NoticeOptionTypeErrCode:
			noticeResponse.Code = value
		case NoticeOptionTypeMessage:
			noticeResponse.Message = value
		case NoticeOptionTypeDetail:
			noticeResponse.Detail = value
		case NoticeOptionTypeHint:
			noticeResponse.Hint = value
		case NoticeOptionTypeConstraint:
			noticeResponse.ConstraintName = value
		case NoticeOptionTypeDataType:
			noticeResponse.DataTypeName = value
		case NoticeOptionTypeTable:
			noticeResponse.TableName = value
		case NoticeOptionTypeSchema:
			noticeResponse.SchemaName = value
		default:
			ctx.GetLogger().Warnf("unhandled notice option type: %d", key)
		}
	}
}

// evaluteNoticeMessage evaluates the message for a RAISE NOTICE statement, including
// evaluating any specified parameters and plugging them into the message in place of
// the % placeholders.
func evaluteNoticeMessage(ctx *sql.Context, iFunc InterpretedFunction,
	operation interpreter.InterpreterOperation, stack InterpreterStack) (string, error) {
	message := operation.SecondaryData[0]
	if len(operation.SecondaryData) > 1 {
		params := operation.SecondaryData[1:]
		currentParam := 0

		parts := strings.Split(message, "%%")
		for i, part := range parts {
			for strings.Contains(part, "%") {
				retVal, err := iFunc.QuerySingleReturn(ctx, stack, "SELECT "+params[currentParam], nil, nil)
				if err != nil {
					return "", err
				}
				currentParam += 1

				s := fmt.Sprintf("%v", retVal)
				part = strings.Replace(part, "%", s, 1)
			}
			parts[i] = part
		}
		message = strings.Join(parts, "%")
	}
	return message, nil
}
