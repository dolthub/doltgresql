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
	"fmt"

	"github.com/jackc/pgx/v5/pgproto3"
)

// MessageType represents the type of a message. New message types should be appended to the end, as we depend on this
// exact order for the serialized files.
type MessageType uint16

const (
	MessageType_AuthenticationCleartextPassword = iota
	MessageType_AuthenticationGSS
	MessageType_AuthenticationGSSContinue
	MessageType_AuthenticationMD5Password
	MessageType_AuthenticationOk
	MessageType_AuthenticationSASL
	MessageType_AuthenticationSASLContinue
	MessageType_AuthenticationSASLFinal
	MessageType_BackendKeyData
	MessageType_Bind
	MessageType_BindComplete
	MessageType_CancelRequest
	MessageType_Close
	MessageType_CloseComplete
	MessageType_CommandComplete
	MessageType_CopyBothResponse
	MessageType_CopyData
	MessageType_CopyDone
	MessageType_CopyFail
	MessageType_CopyInResponse
	MessageType_CopyOutResponse
	MessageType_DataRow
	MessageType_Describe
	MessageType_EmptyQueryResponse
	MessageType_ErrorResponse
	MessageType_Execute
	MessageType_Flush
	MessageType_FunctionCall
	MessageType_FunctionCallResponse
	MessageType_GSSEncRequest
	MessageType_GSSResponse
	MessageType_NoData
	MessageType_NoticeResponse
	MessageType_NotificationResponse
	MessageType_ParameterDescription
	MessageType_ParameterStatus
	MessageType_Parse
	MessageType_ParseComplete
	MessageType_PasswordMessage
	MessageType_PortalSuspended
	MessageType_Query
	MessageType_ReadyForQuery
	MessageType_RowDescription
	MessageType_SASLInitialResponse
	MessageType_SASLResponse
	MessageType_SSLRequest
	MessageType_StartupMessage
	MessageType_Sync
	MessageType_Terminate
)

// ToMessageType returns the MessageType of the given message.
func ToMessageType(message pgproto3.Message) (MessageType, error) {
	switch message.(type) {
	case *pgproto3.AuthenticationCleartextPassword:
		return MessageType_AuthenticationCleartextPassword, nil
	case *pgproto3.AuthenticationGSS:
		return MessageType_AuthenticationGSS, nil
	case *pgproto3.AuthenticationGSSContinue:
		return MessageType_AuthenticationGSSContinue, nil
	case *pgproto3.AuthenticationMD5Password:
		return MessageType_AuthenticationMD5Password, nil
	case *pgproto3.AuthenticationOk:
		return MessageType_AuthenticationOk, nil
	case *pgproto3.AuthenticationSASL:
		return MessageType_AuthenticationSASL, nil
	case *pgproto3.AuthenticationSASLContinue:
		return MessageType_AuthenticationSASLContinue, nil
	case *pgproto3.AuthenticationSASLFinal:
		return MessageType_AuthenticationSASLFinal, nil
	case *pgproto3.BackendKeyData:
		return MessageType_BackendKeyData, nil
	case *pgproto3.Bind:
		return MessageType_Bind, nil
	case *pgproto3.BindComplete:
		return MessageType_BindComplete, nil
	case *pgproto3.CancelRequest:
		return MessageType_CancelRequest, nil
	case *pgproto3.Close:
		return MessageType_Close, nil
	case *pgproto3.CloseComplete:
		return MessageType_CloseComplete, nil
	case *pgproto3.CommandComplete:
		return MessageType_CommandComplete, nil
	case *pgproto3.CopyBothResponse:
		return MessageType_CopyBothResponse, nil
	case *pgproto3.CopyData:
		return MessageType_CopyData, nil
	case *pgproto3.CopyDone:
		return MessageType_CopyDone, nil
	case *pgproto3.CopyFail:
		return MessageType_CopyFail, nil
	case *pgproto3.CopyInResponse:
		return MessageType_CopyInResponse, nil
	case *pgproto3.CopyOutResponse:
		return MessageType_CopyOutResponse, nil
	case *pgproto3.DataRow:
		return MessageType_DataRow, nil
	case *pgproto3.Describe:
		return MessageType_Describe, nil
	case *pgproto3.EmptyQueryResponse:
		return MessageType_EmptyQueryResponse, nil
	case *pgproto3.ErrorResponse:
		return MessageType_ErrorResponse, nil
	case *pgproto3.Execute:
		return MessageType_Execute, nil
	case *pgproto3.Flush:
		return MessageType_Flush, nil
	case *pgproto3.FunctionCall:
		return MessageType_FunctionCall, nil
	case *pgproto3.FunctionCallResponse:
		return MessageType_FunctionCallResponse, nil
	case *pgproto3.GSSEncRequest:
		return MessageType_GSSEncRequest, nil
	case *pgproto3.GSSResponse:
		return MessageType_GSSResponse, nil
	case *pgproto3.NoData:
		return MessageType_NoData, nil
	case *pgproto3.NoticeResponse:
		return MessageType_NoticeResponse, nil
	case *pgproto3.NotificationResponse:
		return MessageType_NotificationResponse, nil
	case *pgproto3.ParameterDescription:
		return MessageType_ParameterDescription, nil
	case *pgproto3.ParameterStatus:
		return MessageType_ParameterStatus, nil
	case *pgproto3.Parse:
		return MessageType_Parse, nil
	case *pgproto3.ParseComplete:
		return MessageType_ParseComplete, nil
	case *pgproto3.PasswordMessage:
		return MessageType_PasswordMessage, nil
	case *pgproto3.PortalSuspended:
		return MessageType_PortalSuspended, nil
	case *pgproto3.Query:
		return MessageType_Query, nil
	case *pgproto3.ReadyForQuery:
		return MessageType_ReadyForQuery, nil
	case *pgproto3.RowDescription:
		return MessageType_RowDescription, nil
	case *pgproto3.SASLInitialResponse:
		return MessageType_SASLInitialResponse, nil
	case *pgproto3.SASLResponse:
		return MessageType_SASLResponse, nil
	case *pgproto3.SSLRequest:
		return MessageType_SSLRequest, nil
	case *pgproto3.StartupMessage:
		return MessageType_StartupMessage, nil
	case *pgproto3.Sync:
		return MessageType_Sync, nil
	case *pgproto3.Terminate:
		return MessageType_Terminate, nil
	default:
		return 0, fmt.Errorf("unknown message type: %T", message)
	}
}

// FromMessageType returns a new message (ready to be decoded into) from the given MessageType.
func FromMessageType(messageType MessageType) (pgproto3.Message, error) {
	switch messageType {
	case MessageType_AuthenticationCleartextPassword:
		return &pgproto3.AuthenticationCleartextPassword{}, nil
	case MessageType_AuthenticationGSS:
		return &pgproto3.AuthenticationGSS{}, nil
	case MessageType_AuthenticationGSSContinue:
		return &pgproto3.AuthenticationGSSContinue{}, nil
	case MessageType_AuthenticationMD5Password:
		return &pgproto3.AuthenticationMD5Password{}, nil
	case MessageType_AuthenticationOk:
		return &pgproto3.AuthenticationOk{}, nil
	case MessageType_AuthenticationSASL:
		return &pgproto3.AuthenticationSASL{}, nil
	case MessageType_AuthenticationSASLContinue:
		return &pgproto3.AuthenticationSASLContinue{}, nil
	case MessageType_AuthenticationSASLFinal:
		return &pgproto3.AuthenticationSASLFinal{}, nil
	case MessageType_BackendKeyData:
		return &pgproto3.BackendKeyData{}, nil
	case MessageType_Bind:
		return &pgproto3.Bind{}, nil
	case MessageType_BindComplete:
		return &pgproto3.BindComplete{}, nil
	case MessageType_CancelRequest:
		return &pgproto3.CancelRequest{}, nil
	case MessageType_Close:
		return &pgproto3.Close{}, nil
	case MessageType_CloseComplete:
		return &pgproto3.CloseComplete{}, nil
	case MessageType_CommandComplete:
		return &pgproto3.CommandComplete{}, nil
	case MessageType_CopyBothResponse:
		return &pgproto3.CopyBothResponse{}, nil
	case MessageType_CopyData:
		return &pgproto3.CopyData{}, nil
	case MessageType_CopyDone:
		return &pgproto3.CopyDone{}, nil
	case MessageType_CopyFail:
		return &pgproto3.CopyFail{}, nil
	case MessageType_CopyInResponse:
		return &pgproto3.CopyInResponse{}, nil
	case MessageType_CopyOutResponse:
		return &pgproto3.CopyOutResponse{}, nil
	case MessageType_DataRow:
		return &pgproto3.DataRow{}, nil
	case MessageType_Describe:
		return &pgproto3.Describe{}, nil
	case MessageType_EmptyQueryResponse:
		return &pgproto3.EmptyQueryResponse{}, nil
	case MessageType_ErrorResponse:
		return &pgproto3.ErrorResponse{}, nil
	case MessageType_Execute:
		return &pgproto3.Execute{}, nil
	case MessageType_Flush:
		return &pgproto3.Flush{}, nil
	case MessageType_FunctionCall:
		return &pgproto3.FunctionCall{}, nil
	case MessageType_FunctionCallResponse:
		return &pgproto3.FunctionCallResponse{}, nil
	case MessageType_GSSEncRequest:
		return &pgproto3.GSSEncRequest{}, nil
	case MessageType_GSSResponse:
		return &pgproto3.GSSResponse{}, nil
	case MessageType_NoData:
		return &pgproto3.NoData{}, nil
	case MessageType_NoticeResponse:
		return &pgproto3.NoticeResponse{}, nil
	case MessageType_NotificationResponse:
		return &pgproto3.NotificationResponse{}, nil
	case MessageType_ParameterDescription:
		return &pgproto3.ParameterDescription{}, nil
	case MessageType_ParameterStatus:
		return &pgproto3.ParameterStatus{}, nil
	case MessageType_Parse:
		return &pgproto3.Parse{}, nil
	case MessageType_ParseComplete:
		return &pgproto3.ParseComplete{}, nil
	case MessageType_PasswordMessage:
		return &pgproto3.PasswordMessage{}, nil
	case MessageType_PortalSuspended:
		return &pgproto3.PortalSuspended{}, nil
	case MessageType_Query:
		return &pgproto3.Query{}, nil
	case MessageType_ReadyForQuery:
		return &pgproto3.ReadyForQuery{}, nil
	case MessageType_RowDescription:
		return &pgproto3.RowDescription{}, nil
	case MessageType_SASLInitialResponse:
		return &pgproto3.SASLInitialResponse{}, nil
	case MessageType_SASLResponse:
		return &pgproto3.SASLResponse{}, nil
	case MessageType_SSLRequest:
		return &pgproto3.SSLRequest{}, nil
	case MessageType_StartupMessage:
		return &pgproto3.StartupMessage{}, nil
	case MessageType_Sync:
		return &pgproto3.Sync{}, nil
	case MessageType_Terminate:
		return &pgproto3.Terminate{}, nil
	default:
		return nil, fmt.Errorf("unknown message type: %d", uint16(messageType))
	}
}
