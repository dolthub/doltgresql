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

package main

import (
	"bytes"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgproto3"
)

// ReplayOptions contain all of the options that may be given to Replay. This is a replacement for a long argument list.
type ReplayOptions struct {
	File         string
	Port         int
	Messages     []pgproto3.Message
	PrintQueries bool     // Prints both queries and file names to the CLI
	FailPSQL     bool     // Whether we should automatically fail PSQL commands, since they're slow and we fail them anyway
	FailQueries  []string // These are queries that cause catastrophic failures, like OOM errors, stack limits, etc.
}

// Replay will replay the given messages onto the Doltgres server running on the given port.
func Replay(options ReplayOptions) (*ReplayTracker, error) {
	tracker := NewReplayTracker(options.File)
	reader := NewMessageReader(FilterMessages(options.Messages))
	fmt.Println("-------------------- ", tracker.File, " --------------------")
ListenerLoop:
	for !reader.IsEmpty() {
		connection, err := NewConnection("127.0.0.1:"+strconv.Itoa(options.Port), reader, 15*time.Second)
		if err != nil {
			return nil, err
		}
		startupMessage, ok := reader.Next().(*pgproto3.StartupMessage)
		if !ok {
			return nil, fmt.Errorf("%s: first message is not StartupMessage (%T)", options.File, reader.Previous())
		}
		if _, ok = reader.Next().(*pgproto3.ReadyForQuery); !ok {
			return nil, fmt.Errorf("expected message after StartupMessage to be ReadyForQuery (%T)", reader.Previous())
		}
		if err = connection.SendNoSync(startupMessage); err != nil {
			return nil, err
		}
	StartupLoop:
		for {
			postgresMessage, err := connection.Receive()
			if err != nil {
				return nil, err
			}
			switch response := postgresMessage.(type) {
			case *pgproto3.AuthenticationOk:
			case *pgproto3.BackendKeyData:
			case *pgproto3.ErrorResponse:
				return nil, errors.New(response.Message)
			case *pgproto3.ParameterStatus:
			case *pgproto3.ReadyForQuery:
				break StartupLoop
			default:
				return nil, fmt.Errorf("unknown StartupMessage response type: %T", response)
			}
		}
	MessageLoop:
		for message := reader.Next(); message != nil; message = reader.Next() {
			switch message := message.(type) {
			case *pgproto3.CopyData:
				// TODO: messages have somehow gotten misordered in `copy2`, so need to fix that, then can remove this case
				reader.SyncToNextQuery()
			case *pgproto3.Describe:
				connection.Queue(message)
				if sync, ok := reader.Peek().(*pgproto3.Sync); ok {
					_ = reader.Next()
					connection.Queue(sync)
				}
				if err = connection.Send(); err != nil {
					tracker.Failed++
					tracker.AddFailure(ReplayTrackerItem{
						Query:           "DESCRIBE",
						UnexpectedError: err.Error(),
					})
					continue ListenerLoop
				}
				var expectedError *pgproto3.ErrorResponse
				var expectedRowDesc *pgproto3.RowDescription
			DescribeLoop:
				for {
					switch queryMessage := reader.Next().(type) {
					case *pgproto3.EmptyQueryResponse:
					case *pgproto3.ErrorResponse:
						expectedError = queryMessage
					case *pgproto3.NoData:
					case *pgproto3.ReadyForQuery:
						break DescribeLoop
					case *pgproto3.RowDescription:
						expectedRowDesc = queryMessage
					default:
						return nil, fmt.Errorf("unable to determine what to do with %T", queryMessage)
					}
				}
				var responseError *pgproto3.ErrorResponse
				var responseRowDesc *pgproto3.RowDescription
			DescribeResponseLoop:
				for {
					response, err := connection.Receive()
					if err != nil {
						tracker.Failed++
						tracker.AddFailure(ReplayTrackerItem{
							Query:           "DESCRIBE",
							UnexpectedError: err.Error(),
						})
						continue ListenerLoop
					}
					response = DuplicateMessage(response).(pgproto3.BackendMessage)
					switch response := response.(type) {
					case *pgproto3.EmptyQueryResponse:
					case *pgproto3.ErrorResponse:
						responseError = response
					case *pgproto3.NoData:
					case *pgproto3.NoticeResponse:
					case *pgproto3.ParameterDescription:
					case *pgproto3.ReadyForQuery:
						break DescribeResponseLoop
					case *pgproto3.RowDescription:
						responseRowDesc = response
					default:
						return nil, fmt.Errorf("unable to determine what to do with %T", message)
					}
				}
				if err = connection.EmptyReceiveBuffer(); err != nil {
					tracker.Failed++
					tracker.AddFailure(ReplayTrackerItem{
						Query:           "DESCRIBE",
						UnexpectedError: err.Error(),
					})
					continue MessageLoop
				}
				if expectedError == nil {
					if responseError != nil {
						tracker.Failed++
						tracker.AddFailure(ReplayTrackerItem{
							Query:           "DESCRIBE",
							UnexpectedError: responseError.Message,
						})
						continue MessageLoop
					}
					if expectedRowDesc == nil {
						if responseRowDesc == nil {
							tracker.Success++
						} else {
							tracker.Failed++
							tracker.AddFailure(ReplayTrackerItem{
								Query:           "DESCRIBE",
								UnexpectedError: "expected no row description but received a description",
							})
						}
						continue MessageLoop
					}
					if responseRowDesc == nil {
						tracker.Failed++
						tracker.AddFailure(ReplayTrackerItem{
							Query:           "DESCRIBE",
							UnexpectedError: "expected rows but received none",
						})
						continue MessageLoop
					}
					if len(expectedRowDesc.Fields) != len(responseRowDesc.Fields) {
						tracker.Failed++
						tracker.AddFailure(ReplayTrackerItem{
							Query: "DESCRIBE",
							UnexpectedError: fmt.Sprintf("expected column count %d but received %d",
								len(expectedRowDesc.Fields), len(responseRowDesc.Fields)),
						})
						continue MessageLoop
					}
					var partialSuccesses []string
					for i := range expectedRowDesc.Fields {
						expectedName := string(expectedRowDesc.Fields[i].Name)
						responseName := string(responseRowDesc.Fields[i].Name)
						if expectedName != responseName {
							partialSuccesses = append(partialSuccesses,
								fmt.Sprintf("expected column with name `%s` but received `%s`", expectedName, responseName))
						}
						// TODO: determine if we should also check column types
					}
					tracker.Success++
					if len(partialSuccesses) > 0 {
						tracker.PartialSuccess++
					}
				} else /* expectedError != nil */ {
					if responseError == nil {
						tracker.Failed++
						tracker.AddFailure(ReplayTrackerItem{
							Query:         "DESCRIBE",
							ExpectedError: expectedError.Message,
						})
						continue MessageLoop
					}
					tracker.Success++
					if expectedError.Message != responseError.Message {
						tracker.PartialSuccess++
						tracker.AddFailure(ReplayTrackerItem{
							Query:           "DESCRIBE",
							UnexpectedError: responseError.Message,
							ExpectedError:   expectedError.Message,
						})
					}
				}
			case *pgproto3.FunctionCall:
				if err = connection.Send(message); err != nil {
					tracker.Failed++
					tracker.AddFailure(ReplayTrackerItem{
						Query:           fmt.Sprintf("Function OID: %d", message.Function),
						UnexpectedError: err.Error(),
					})
					continue ListenerLoop
				}
				var expectedError *pgproto3.ErrorResponse
				var expectedData *pgproto3.FunctionCallResponse
			FunctionLoop:
				for {
					switch queryMessage := reader.Next().(type) {
					case *pgproto3.ErrorResponse:
						expectedError = queryMessage
					case *pgproto3.FunctionCallResponse:
						expectedData = queryMessage
					case *pgproto3.ReadyForQuery:
						break FunctionLoop
					default:
						return nil, fmt.Errorf("unable to determine what to do with %T", queryMessage)
					}
				}
				var responseError *pgproto3.ErrorResponse
				var responseData *pgproto3.FunctionCallResponse
			FunctionResponseLoop:
				for {
					response, err := connection.Receive()
					if err != nil {
						tracker.Failed++
						tracker.AddFailure(ReplayTrackerItem{
							Query:           fmt.Sprintf("Function OID: %d", message.Function),
							UnexpectedError: err.Error(),
						})
						continue ListenerLoop
					}
					response = DuplicateMessage(response).(pgproto3.BackendMessage)
					switch response := response.(type) {
					case *pgproto3.EmptyQueryResponse:
					case *pgproto3.ErrorResponse:
						responseError = response
					case *pgproto3.FunctionCallResponse:
						responseData = response
					case *pgproto3.NoticeResponse:
					case *pgproto3.ReadyForQuery:
						break FunctionResponseLoop
					default:
						return nil, fmt.Errorf("unable to determine what to do with %T", message)
					}
				}
				if err = connection.EmptyReceiveBuffer(); err != nil {
					tracker.Failed++
					tracker.AddFailure(ReplayTrackerItem{
						Query:           fmt.Sprintf("Function OID: %d", message.Function),
						UnexpectedError: err.Error(),
					})
					continue MessageLoop
				}
				if expectedError == nil {
					if responseError != nil {
						tracker.Failed++
						tracker.AddFailure(ReplayTrackerItem{
							Query:           fmt.Sprintf("Function OID: %d", message.Function),
							UnexpectedError: responseError.Message,
						})
						continue MessageLoop
					}
					if expectedData != nil {
						if responseData == nil {
							tracker.Failed++
							tracker.AddFailure(ReplayTrackerItem{
								Query:           fmt.Sprintf("Function OID: %d", message.Function),
								UnexpectedError: "expected a result but received no result",
							})
							continue MessageLoop
						}
						if !bytes.Equal(expectedData.Result, responseData.Result) {
							tracker.Failed++
							tracker.AddFailure(ReplayTrackerItem{
								Query:           fmt.Sprintf("Function OID: %d", message.Function),
								UnexpectedError: "result is incorrect",
							})
							continue MessageLoop
						}
					} else /* expectedData == nil */ {
						if responseData != nil {
							tracker.Failed++
							tracker.AddFailure(ReplayTrackerItem{
								Query:           fmt.Sprintf("Function OID: %d", message.Function),
								UnexpectedError: "expected no result but received a result",
							})
							continue MessageLoop
						}
					}
					tracker.Success++
					tracker.AddSuccess(ReplayTrackerItem{
						Query: fmt.Sprintf("Function OID: %d", message.Function),
					})
				} else /* expectedError != nil */ {
					if responseError == nil {
						tracker.Failed++
						tracker.AddFailure(ReplayTrackerItem{
							Query:         fmt.Sprintf("Function OID: %d", message.Function),
							ExpectedError: expectedError.Message,
						})
						continue MessageLoop
					}
					tracker.Success++
					tracker.AddSuccess(ReplayTrackerItem{
						Query: fmt.Sprintf("Function OID: %d", message.Function),
					})
					if expectedError.Message != responseError.Message {
						tracker.PartialSuccess++
						tracker.AddFailure(ReplayTrackerItem{
							Query:           fmt.Sprintf("Function OID: %d", message.Function),
							UnexpectedError: responseError.Message,
							ExpectedError:   expectedError.Message,
						})
					}
				}
			case *pgproto3.Parse:
				connection.Queue(message)
				if sync, ok := reader.Peek().(*pgproto3.Sync); ok {
					_ = reader.Next()
					connection.Queue(sync)
				}
				if err = connection.Send(); err != nil {
					tracker.Failed++
					tracker.AddFailure(ReplayTrackerItem{
						Query:           message.Query,
						UnexpectedError: err.Error(),
					})
					continue ListenerLoop
				}
				var expectedError *pgproto3.ErrorResponse
			ParseLoop:
				for {
					switch queryMessage := reader.Next().(type) {
					case *pgproto3.EmptyQueryResponse:
					case *pgproto3.ErrorResponse:
						expectedError = queryMessage
					case *pgproto3.NoData:
					case *pgproto3.ParseComplete:
					case *pgproto3.ReadyForQuery:
						break ParseLoop
					default:
						return nil, fmt.Errorf("unable to determine what to do with %T", queryMessage)
					}
				}
				var responseError *pgproto3.ErrorResponse
			ParseResponseLoop:
				for {
					response, err := connection.Receive()
					if err != nil {
						tracker.Failed++
						tracker.AddFailure(ReplayTrackerItem{
							Query:           message.Query,
							UnexpectedError: err.Error(),
						})
						continue ListenerLoop
					}
					response = DuplicateMessage(response).(pgproto3.BackendMessage)
					switch response := response.(type) {
					case *pgproto3.EmptyQueryResponse:
					case *pgproto3.ErrorResponse:
						responseError = response
					case *pgproto3.NoData:
					case *pgproto3.NoticeResponse:
					case *pgproto3.ParseComplete:
					case *pgproto3.ReadyForQuery:
						break ParseResponseLoop
					default:
						return nil, fmt.Errorf("unable to determine what to do with %T", message)
					}
				}
				if err = connection.EmptyReceiveBuffer(); err != nil {
					tracker.Failed++
					tracker.AddFailure(ReplayTrackerItem{
						Query:           message.Query,
						UnexpectedError: err.Error(),
					})
					continue MessageLoop
				}
				if expectedError == nil {
					if responseError != nil {
						tracker.Failed++
						tracker.AddFailure(ReplayTrackerItem{
							Query:           message.Query,
							UnexpectedError: responseError.Message,
						})
						continue MessageLoop
					}
					tracker.Success++
					tracker.AddSuccess(ReplayTrackerItem{
						Query: message.Query,
					})
				} else /* expectedError != nil */ {
					if responseError == nil {
						tracker.Failed++
						tracker.AddFailure(ReplayTrackerItem{
							Query:         message.Query,
							ExpectedError: expectedError.Message,
						})
						continue MessageLoop
					}
					tracker.Success++
					tracker.AddSuccess(ReplayTrackerItem{
						Query: message.Query,
					})
					if expectedError.Message != responseError.Message {
						tracker.PartialSuccess++
						tracker.AddFailure(ReplayTrackerItem{
							Query:           message.Query,
							UnexpectedError: responseError.Message,
							ExpectedError:   expectedError.Message,
						})
					}
				}
			case *pgproto3.Query:
				if options.PrintQueries {
					fmt.Println("QUERY: " + message.String)
				}
				if options.FailPSQL {
					if strings.HasPrefix(message.String, "SELECT c2.relname, i.indisprimary, i.indisunique, i.indisclustered, i.indisvalid, pg_catalog.pg_get_indexdef(i.indexrelid, 0, true),") {
						tracker.Failed++
						tracker.AddFailure(ReplayTrackerItem{
							Query:           message.String,
							UnexpectedError: "set to automatically fail PSQL commands",
						})
						reader.SyncToNextQuery()
						continue MessageLoop
					}
				}
				for _, failQuery := range options.FailQueries {
					if message.String == failQuery {
						tracker.Failed++
						tracker.AddFailure(ReplayTrackerItem{
							Query:           message.String,
							UnexpectedError: "set to automatically fail due to catastrophic error (OOM, stack limit, etc.)",
						})
						reader.SyncToNextQuery()
						continue MessageLoop
					}
				}
				if err = connection.Send(message); err != nil {
					tracker.Failed++
					tracker.AddFailure(ReplayTrackerItem{
						Query:           message.String,
						UnexpectedError: err.Error(),
					})
					continue ListenerLoop
				}
				var expectedError *pgproto3.ErrorResponse
				var expectedRowDesc *pgproto3.RowDescription
				var expectedDataRows []*pgproto3.DataRow
				var expectedCopyData []*pgproto3.CopyData
			QueryLoop:
				for {
					switch queryMessage := reader.Next().(type) {
					case *pgproto3.CommandComplete:
					case *pgproto3.CopyData:
						expectedCopyData = append(expectedCopyData, queryMessage)
					case *pgproto3.CopyDone:
					case *pgproto3.CopyInResponse:
					case *pgproto3.CopyOutResponse:
					case *pgproto3.DataRow:
						expectedDataRows = append(expectedDataRows, queryMessage)
					case *pgproto3.EmptyQueryResponse:
					case *pgproto3.ErrorResponse:
						expectedError = queryMessage
					case *pgproto3.ReadyForQuery:
						break QueryLoop
					case *pgproto3.RowDescription:
						expectedRowDesc = queryMessage
					default:
						return nil, fmt.Errorf("unable to determine what to do with %T", queryMessage)
					}
				}
				var responseError *pgproto3.ErrorResponse
				var responseRowDesc *pgproto3.RowDescription
				var responseDataRows []*pgproto3.DataRow
			ResponseLoop:
				for {
					response, err := connection.Receive()
					if err != nil {
						tracker.Failed++
						tracker.AddFailure(ReplayTrackerItem{
							Query:           message.String,
							UnexpectedError: err.Error(),
						})
						continue ListenerLoop
					}
					response = DuplicateMessage(response).(pgproto3.BackendMessage)
					switch response := response.(type) {
					case *pgproto3.CommandComplete:
					case *pgproto3.CopyInResponse:
						for _, copyData := range expectedCopyData {
							connection.Queue(copyData)
						}
						if err = connection.SendNoSync(&pgproto3.CopyDone{}); err != nil {
							tracker.Failed++
							tracker.AddFailure(ReplayTrackerItem{
								Query:           message.String,
								UnexpectedError: err.Error(),
							})
							continue ListenerLoop
						}
					case *pgproto3.DataRow:
						responseDataRows = append(responseDataRows, response)
					case *pgproto3.EmptyQueryResponse:
					case *pgproto3.ErrorResponse:
						responseError = response
					case *pgproto3.NoticeResponse:
					case *pgproto3.ReadyForQuery:
						break ResponseLoop
					case *pgproto3.RowDescription:
						responseRowDesc = response
					default:
						return nil, fmt.Errorf("unable to determine what to do with %T", message)
					}
				}
				if err = connection.EmptyReceiveBuffer(); err != nil {
					tracker.Failed++
					tracker.AddFailure(ReplayTrackerItem{
						Query:           message.String,
						UnexpectedError: err.Error(),
					})
					continue MessageLoop
				}
				if expectedError == nil {
					if responseError != nil {
						tracker.Failed++
						tracker.AddFailure(ReplayTrackerItem{
							Query:           message.String,
							UnexpectedError: responseError.Message,
						})
						continue MessageLoop
					}
					if expectedRowDesc == nil {
						if responseRowDesc == nil {
							tracker.Success++
							tracker.AddSuccess(ReplayTrackerItem{
								Query: message.String,
							})
						} else {
							tracker.Failed++
							tracker.AddFailure(ReplayTrackerItem{
								Query:           message.String,
								UnexpectedError: "expected no rows but received rows",
							})
						}
						continue MessageLoop
					}
					if responseRowDesc == nil {
						tracker.Failed++
						tracker.AddFailure(ReplayTrackerItem{
							Query:           message.String,
							UnexpectedError: "expected rows but received none",
						})
						continue MessageLoop
					}
					if len(expectedRowDesc.Fields) != len(responseRowDesc.Fields) {
						tracker.Failed++
						tracker.AddFailure(ReplayTrackerItem{
							Query: message.String,
							UnexpectedError: fmt.Sprintf("expected column count %d but received %d",
								len(expectedRowDesc.Fields), len(responseRowDesc.Fields)),
						})
						continue MessageLoop
					}
					var partialSuccesses []string
					for i := range expectedRowDesc.Fields {
						expectedName := string(expectedRowDesc.Fields[i].Name)
						responseName := string(responseRowDesc.Fields[i].Name)
						if expectedName != responseName {
							partialSuccesses = append(partialSuccesses,
								fmt.Sprintf("expected column with name `%s` but received `%s`", expectedName, responseName))
						}
						// TODO: determine if we should also check column types
					}
					if len(expectedDataRows) != len(responseDataRows) {
						tracker.Failed++
						tracker.AddFailure(ReplayTrackerItem{
							Query: message.String,
							UnexpectedError: fmt.Sprintf("expected row count %d but received %d",
								len(expectedDataRows), len(responseDataRows)),
						})
						continue MessageLoop
					}
					if strings.Contains(strings.ToLower(message.String), "order by") {
						// There's an ORDER BY, so we need to check based on the order
						if err = CompareRowsOrdered(expectedRowDesc, responseRowDesc, expectedDataRows, responseDataRows); err != nil {
							tracker.Failed++
							tracker.AddFailure(ReplayTrackerItem{
								Query:           message.String,
								UnexpectedError: err.Error(),
							})
							continue MessageLoop
						}
					} else {
						// There's no ORDER BY, so our native row order may differ from Postgres.
						if err = CompareRowsUnordered(expectedRowDesc, responseRowDesc, expectedDataRows, responseDataRows); err != nil {
							tracker.Failed++
							tracker.AddFailure(ReplayTrackerItem{
								Query:           message.String,
								UnexpectedError: err.Error(),
							})
							continue MessageLoop
						}
					}
					tracker.Success++
					tracker.AddSuccess(ReplayTrackerItem{
						Query: message.String,
					})
					if len(partialSuccesses) > 0 {
						tracker.PartialSuccess++
					}
				} else /* expectedError != nil */ {
					if responseError == nil {
						tracker.Failed++
						tracker.AddFailure(ReplayTrackerItem{
							Query:         message.String,
							ExpectedError: expectedError.Message,
						})
						continue MessageLoop
					}
					tracker.Success++
					tracker.AddSuccess(ReplayTrackerItem{
						Query: message.String,
					})
					if expectedError.Message != responseError.Message {
						tracker.PartialSuccess++
						tracker.AddFailure(ReplayTrackerItem{
							Query:           message.String,
							UnexpectedError: responseError.Message,
							ExpectedError:   expectedError.Message,
						})
					}
				}
			case *pgproto3.Terminate:
				if err = connection.SendNoSync(message); err != nil {
					return nil, err
				}
				break MessageLoop
			default:
				return nil, fmt.Errorf("unable to determine what to do with %T", message)
			}
		}
		connection.Close()
	}
	return tracker, nil
}
