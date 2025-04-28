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
	"strconv"
	"strings"

	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/dsess"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/analyzer"
	"github.com/jackc/pgx/v5/pgproto3"

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
	QueryMultiReturn(ctx *sql.Context, stack InterpreterStack, stmt string, bindings []string) (rows []sql.Row, err error)
	QuerySingleReturn(ctx *sql.Context, stack InterpreterStack, stmt string, targetType *pgtypes.DoltgresType, bindings []string) (val any, err error)
}

// GetTypesCollectionFromContext is declared within the core package, but is assigned to this variable to work around
// import cycles.
var GetTypesCollectionFromContext func(ctx *sql.Context) (*typecollection.TypeCollection, error)

// Call runs the contained operations on the given runner.
func Call(ctx *sql.Context, iFunc InterpretedFunction, runner analyzer.StatementRunner, paramsAndReturn []*pgtypes.DoltgresType, vals []any) (any, error) {
	// Set up the initial state of the function
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
	return call(ctx, iFunc, stack)
}

// TriggerCall runs the contained trigger operations on the given runner.
func TriggerCall(ctx *sql.Context, iFunc InterpretedFunction, runner analyzer.StatementRunner, sch sql.Schema, oldRow sql.Row, newRow sql.Row) (any, error) {
	// Set up the initial state of the function
	stack := NewInterpreterStack(runner)
	// Add the special variables
	// TODO: there are way more than just NEW and OLD -> https://www.postgresql.org/docs/15/plpgsql-trigger.html
	stack.NewRecord("OLD", sch, oldRow)
	stack.NewRecord("NEW", sch, newRow)
	return call(ctx, iFunc, stack)
}

// call runs the contained operations on the given runner.
func call(ctx *sql.Context, iFunc InterpretedFunction, stack InterpreterStack) (any, error) {
	// We increment before accessing, so start at -1
	counter := -1
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
			if iv.Type == nil {
				return nil, fmt.Errorf("variable `%s` could not be found", operation.PrimaryData)
			}
			stack.NewVariableAlias(operation.Target, operation.PrimaryData)
		case OpCode_Assign:
			iv := stack.GetVariable(operation.Target)
			if iv.Type == nil {
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
			resolvedType, err := typeCollection.GetType(ctx, id.NewType(schemaName, typeName))
			if err != nil {
				return nil, err
			}
			if resolvedType == nil {
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
				if target.Type == nil {
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
				_, err := iFunc.QueryMultiReturn(ctx, stack, operation.PrimaryData, operation.SecondaryData)
				if err != nil {
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
			_, err := iFunc.QueryMultiReturn(ctx, stack, operation.PrimaryData, operation.SecondaryData)
			if err != nil {
				return nil, err
			}
		case OpCode_Raise:
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

			if err = applyNoticeOptions(ctx, noticeResponse, operation.Options); err != nil {
				return nil, err
			}
			sess := dsess.DSessFromSess(ctx.Session)
			sess.Notice(noticeResponse)
		case OpCode_Return:
			if len(operation.PrimaryData) == 0 {
				return nil, nil
			}
			// TODO: handle record types properly, we'll special case triggers for now
			if iFunc.GetReturn().ID == pgtypes.Trigger.ID && len(operation.SecondaryData) == 1 {
				normalized := strings.ReplaceAll(strings.ToLower(operation.PrimaryData), " ", "")
				if normalized == "select$1;" {
					if strings.EqualFold(operation.SecondaryData[0], "new") {
						return *stack.GetVariable("NEW").Value, nil
					} else if strings.EqualFold(operation.SecondaryData[0], "old") {
						return *stack.GetVariable("OLD").Value, nil
					}
				}
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

// applyNoticeOptions adds the specified |options| to the |noticeResponse|.
func applyNoticeOptions(ctx *sql.Context, noticeResponse *pgproto3.NoticeResponse, options map[string]string) error {
	for key, value := range options {
		i, err := strconv.Atoi(key)
		if err != nil {
			return err
		}

		switch NoticeOptionType(i) {
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
			ctx.GetLogger().Warnf("unhandled notice option type: %s", key)
		}
	}
	return nil
}

// evaluteNoticeMessage evaluates the message for a RAISE NOTICE statement, including
// evaluating any specified parameters and plugging them into the message in place of
// the % placeholders.
func evaluteNoticeMessage(ctx *sql.Context, iFunc InterpretedFunction,
	operation InterpreterOperation, stack InterpreterStack) (string, error) {
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
