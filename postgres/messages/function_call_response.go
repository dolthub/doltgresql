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
	connection.InitializeDefaultMessage(FunctionCallResponse{})
}

// FunctionCallResponse represents a PostgreSQL message.
type FunctionCallResponse struct {
	IsResultNull bool
	ResultValue  []byte
}

var functionCallResponseDefault = connection.MessageFormat{
	Name: "FunctionCallResponse",
	Fields: connection.FieldGroup{
		{
			Name:  "Header",
			Type:  connection.Byte1,
			Flags: connection.Header,
			Data:  int32('V'),
		},
		{
			Name:  "MessageLength",
			Type:  connection.Int32,
			Flags: connection.MessageLengthInclusive,
			Data:  int32(0),
		},
		{
			Name:  "ResultLength",
			Type:  connection.Int32,
			Flags: connection.ByteCount,
			Data:  int32(0),
		},
		{
			Name: "ResultValue",
			Type: connection.ByteN,
			Data: []byte{},
		},
	},
}

var _ connection.Message = FunctionCallResponse{}

// Encode implements the interface connection.Message.
func (m FunctionCallResponse) Encode() (connection.MessageFormat, error) {
	outputMessage := m.DefaultMessage().Copy()
	if m.IsResultNull {
		outputMessage.Field("ResultLength").MustWrite(-1)
	} else {
		if m.ResultValue == nil {
			m.ResultValue = []byte{}
		}
		outputMessage.Field("ResultLength").MustWrite(-1)
		outputMessage.Field("ResultValue").MustWrite(m.ResultValue)
	}
	return outputMessage, nil
}

// Decode implements the interface connection.Message.
func (m FunctionCallResponse) Decode(s connection.MessageFormat) (connection.Message, error) {
	if err := s.MatchesStructure(*m.DefaultMessage()); err != nil {
		return nil, err
	}
	isNull := s.Field("ResultLength").MustGet().(int32) == -1
	return FunctionCallResponse{
		IsResultNull: isNull,
		ResultValue:  s.Field("ResultValue").MustGet().([]byte),
	}, nil
}

// DefaultMessage implements the interface connection.Message.
func (m FunctionCallResponse) DefaultMessage() *connection.MessageFormat {
	return &functionCallResponseDefault
}
