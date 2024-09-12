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

// EncodeMessage encodes the given message to the byte representation needed by the Decode function. Each message's
// Encode function will often include additional information that the Decode function does not expect, hence it is
// removed in this function.
func EncodeMessage(message pgproto3.Message) ([]byte, error) {
	switch message := message.(type) {
	case *pgproto3.AuthenticationCleartextPassword:
		data, err := message.Encode(nil)
		return data[5:], err
	case *pgproto3.AuthenticationGSS:
		data, err := message.Encode(nil)
		return data[5:], err
	case *pgproto3.AuthenticationGSSContinue:
		data, err := message.Encode(nil)
		return data[5:], err
	case *pgproto3.AuthenticationMD5Password:
		data, err := message.Encode(nil)
		return data[5:], err
	case *pgproto3.AuthenticationOk:
		data, err := message.Encode(nil)
		return data[5:], err
	case *pgproto3.AuthenticationSASL:
		data, err := message.Encode(nil)
		return data[5:], err
	case *pgproto3.AuthenticationSASLContinue:
		data, err := message.Encode(nil)
		return data[5:], err
	case *pgproto3.AuthenticationSASLFinal:
		data, err := message.Encode(nil)
		return data[5:], err
	case *pgproto3.BackendKeyData:
		data, err := message.Encode(nil)
		return data[5:], err
	case *pgproto3.Bind:
		data, err := message.Encode(nil)
		return data[5:], err
	case *pgproto3.BindComplete:
		return nil, nil
	case *pgproto3.CancelRequest:
		data, err := message.Encode(nil)
		return data[4:], err
	case *pgproto3.Close:
		data, err := message.Encode(nil)
		return data[5:], err
	case *pgproto3.CloseComplete:
		return nil, nil
	case *pgproto3.CommandComplete:
		data, err := message.Encode(nil)
		return data[5:], err
	case *pgproto3.CopyBothResponse:
		data, err := message.Encode(nil)
		return data[5:], err
	case *pgproto3.CopyData:
		data, err := message.Encode(nil)
		return data[5:], err
	case *pgproto3.CopyDone:
		return nil, nil
	case *pgproto3.CopyFail:
		data, err := message.Encode(nil)
		return data[5:], err
	case *pgproto3.CopyInResponse:
		data, err := message.Encode(nil)
		return data[5:], err
	case *pgproto3.CopyOutResponse:
		data, err := message.Encode(nil)
		return data[5:], err
	case *pgproto3.DataRow:
		data, err := message.Encode(nil)
		return data[5:], err
	case *pgproto3.Describe:
		data, err := message.Encode(nil)
		return data[5:], err
	case *pgproto3.EmptyQueryResponse:
		return nil, nil
	case *pgproto3.ErrorResponse:
		data, err := message.Encode(nil)
		return data[5:], err
	case *pgproto3.Execute:
		data, err := message.Encode(nil)
		return data[5:], err
	case *pgproto3.Flush:
		return nil, nil
	case *pgproto3.FunctionCall:
		data, err := message.Encode(nil)
		return data[5:], err
	case *pgproto3.FunctionCallResponse:
		data, err := message.Encode(nil)
		return data[5:], err
	case *pgproto3.GSSEncRequest:
		data, err := message.Encode(nil)
		return data[4:], err
	case *pgproto3.GSSResponse:
		data, err := message.Encode(nil)
		return data[5:], err
	case *pgproto3.NoData:
		return nil, nil
	case *pgproto3.NoticeResponse:
		data, err := message.Encode(nil)
		return data[5:], err
	case *pgproto3.NotificationResponse:
		data, err := message.Encode(nil)
		return data[5:], err
	case *pgproto3.ParameterDescription:
		data, err := message.Encode(nil)
		return data[5:], err
	case *pgproto3.ParameterStatus:
		data, err := message.Encode(nil)
		return data[5:], err
	case *pgproto3.Parse:
		data, err := message.Encode(nil)
		return data[5:], err
	case *pgproto3.ParseComplete:
		return nil, nil
	case *pgproto3.PasswordMessage:
		data, err := message.Encode(nil)
		return data[5:], err
	case *pgproto3.PortalSuspended:
		return nil, nil
	case *pgproto3.Query:
		data, err := RewriteCopyToFileOnly(message).Encode(nil)
		return data[5:], err
	case *pgproto3.ReadyForQuery:
		data, err := message.Encode(nil)
		return data[5:], err
	case *pgproto3.RowDescription:
		data, err := message.Encode(nil)
		return data[5:], err
	case *pgproto3.SASLInitialResponse:
		data, err := message.Encode(nil)
		return data[5:], err
	case *pgproto3.SASLResponse:
		data, err := message.Encode(nil)
		return data[5:], err
	case *pgproto3.SSLRequest:
		data, err := message.Encode(nil)
		return data[4:], err
	case *pgproto3.StartupMessage:
		data, err := message.Encode(nil)
		return data[4:], err
	case *pgproto3.Sync:
		return nil, nil
	case *pgproto3.Terminate:
		return nil, nil
	default:
		return nil, fmt.Errorf("unknown message type: %T", message)
	}
}
