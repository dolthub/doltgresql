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
	"github.com/jackc/pgx/v5/pgproto3"
)

// FilterMessages removes all "unnecessary" messages so that only the important messages remain.
func FilterMessages(messages []pgproto3.Message) []pgproto3.Message {
	filteredMessage := make([]pgproto3.Message, 0, len(messages))
	for _, message := range messages {
		switch message.(type) {
		case *pgproto3.AuthenticationCleartextPassword:
		case *pgproto3.AuthenticationGSS:
		case *pgproto3.AuthenticationGSSContinue:
		case *pgproto3.AuthenticationMD5Password:
		case *pgproto3.AuthenticationOk:
		case *pgproto3.AuthenticationSASL:
		case *pgproto3.AuthenticationSASLContinue:
		case *pgproto3.AuthenticationSASLFinal:
		case *pgproto3.BackendKeyData:
		case *pgproto3.Bind:
			filteredMessage = append(filteredMessage, message)
		case *pgproto3.BindComplete:
			filteredMessage = append(filteredMessage, message)
		case *pgproto3.CancelRequest:
			filteredMessage = append(filteredMessage, message)
		case *pgproto3.Close:
			filteredMessage = append(filteredMessage, message)
		case *pgproto3.CloseComplete:
			filteredMessage = append(filteredMessage, message)
		case *pgproto3.CommandComplete:
			filteredMessage = append(filteredMessage, message)
		case *pgproto3.CopyBothResponse:
			filteredMessage = append(filteredMessage, message)
		case *pgproto3.CopyData:
			filteredMessage = append(filteredMessage, message)
		case *pgproto3.CopyDone:
			filteredMessage = append(filteredMessage, message)
		case *pgproto3.CopyFail:
			filteredMessage = append(filteredMessage, message)
		case *pgproto3.CopyInResponse:
			filteredMessage = append(filteredMessage, message)
		case *pgproto3.CopyOutResponse:
			filteredMessage = append(filteredMessage, message)
		case *pgproto3.DataRow:
			filteredMessage = append(filteredMessage, message)
		case *pgproto3.Describe:
			filteredMessage = append(filteredMessage, message)
		case *pgproto3.EmptyQueryResponse:
			filteredMessage = append(filteredMessage, message)
		case *pgproto3.ErrorResponse:
			filteredMessage = append(filteredMessage, message)
		case *pgproto3.Execute:
			filteredMessage = append(filteredMessage, message)
		case *pgproto3.Flush:
		case *pgproto3.FunctionCall:
			filteredMessage = append(filteredMessage, message)
		case *pgproto3.FunctionCallResponse:
			filteredMessage = append(filteredMessage, message)
		case *pgproto3.GSSEncRequest:
		case *pgproto3.GSSResponse:
		case *pgproto3.NoData:
			filteredMessage = append(filteredMessage, message)
		case *pgproto3.NoticeResponse:
		case *pgproto3.NotificationResponse:
		case *pgproto3.ParameterDescription:
		case *pgproto3.ParameterStatus:
		case *pgproto3.Parse:
			filteredMessage = append(filteredMessage, message)
		case *pgproto3.ParseComplete:
			filteredMessage = append(filteredMessage, message)
		case *pgproto3.PasswordMessage:
		case *pgproto3.PortalSuspended:
		case *pgproto3.Query:
			filteredMessage = append(filteredMessage, message)
		case *pgproto3.ReadyForQuery:
			filteredMessage = append(filteredMessage, message)
		case *pgproto3.RowDescription:
			filteredMessage = append(filteredMessage, message)
		case *pgproto3.SASLInitialResponse:
		case *pgproto3.SASLResponse:
		case *pgproto3.SSLRequest:
		case *pgproto3.StartupMessage:
			filteredMessage = append(filteredMessage, message)
		case *pgproto3.Sync:
			filteredMessage = append(filteredMessage, message)
		case *pgproto3.Terminate:
			filteredMessage = append(filteredMessage, message)
		default:
			// We'll ignore any messages that we don't know about
		}
	}
	return filteredMessage
}

// SplitByTerminate splits the given set of messages by the Terminate message, which represents a test boundary.
func SplitByTerminate(messages []pgproto3.Message) [][]pgproto3.Message {
	var messageGroups [][]pgproto3.Message
	start := 0
	for i, message := range messages {
		if _, ok := message.(*pgproto3.Terminate); ok {
			messageGroups = append(messageGroups, messages[start:i+1])
			start = i + 1
		}
	}
	if start < len(messages)-1 {
		messageGroups = append(messageGroups, messages[start:])
	}
	return messageGroups
}

// CombineGroups combines message groups into a single group.
func CombineGroups(messageGroups ...[]pgproto3.Message) []pgproto3.Message {
	if len(messageGroups) == 0 {
		return nil
	}
	newSlice := append([]pgproto3.Message{}, messageGroups[0]...)
	for i := 1; i < len(messageGroups); i++ {
		newSlice = append(newSlice, messageGroups[i]...)
	}
	return newSlice
}
