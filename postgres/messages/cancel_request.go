// Copyright 2023 Dolthub, Inc.
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

package messages

import "github.com/dolthub/doltgresql/postgres/connection"

func init() {
	connection.InitializeDefaultMessage(CancelRequest{})
}

// CancelRequest represents a PostgreSQL message.
type CancelRequest struct {
	ProcessID int32
	SecretKey int32
}

var cancelRequestDefault = connection.MessageFormat{
	Name: "CancelRequest",
	Fields: connection.FieldGroup{
		{
			Name:  "MessageLength",
			Type:  connection.Int32,
			Flags: connection.MessageLengthInclusive,
			Data:  int32(0),
		},
		{
			Name: "RequestCode",
			Type: connection.Int32,
			Data: int32(80877102),
		},
		{
			Name: "ProcessID",
			Type: connection.Int32,
			Data: int32(0),
		},
		{
			Name: "SecretKey",
			Type: connection.Int32,
			Data: int32(0),
		},
	},
}

var _ connection.Message = CancelRequest{}

// Encode implements the interface connection.Message.
func (m CancelRequest) Encode() (connection.MessageFormat, error) {
	outputMessage := m.DefaultMessage().Copy()
	outputMessage.Field("ProcessID").MustWrite(m.ProcessID)
	outputMessage.Field("SecretKey").MustWrite(m.SecretKey)
	return outputMessage, nil
}

// Decode implements the interface connection.Message.
func (m CancelRequest) Decode(s connection.MessageFormat) (connection.Message, error) {
	if err := s.MatchesStructure(*m.DefaultMessage()); err != nil {
		return nil, err
	}
	return CancelRequest{
		ProcessID: s.Field("ProcessID").MustGet().(int32),
		SecretKey: s.Field("SecretKey").MustGet().(int32),
	}, nil
}

// DefaultMessage implements the interface connection.Message.
func (m CancelRequest) DefaultMessage() *connection.MessageFormat {
	return &cancelRequestDefault
}
