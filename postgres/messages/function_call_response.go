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

func init() {
	initializeDefaultMessage(FunctionCallResponse{})
}

// FunctionCallResponse represents a PostgreSQL message.
type FunctionCallResponse struct {
	IsResultNull bool
	ResultValue  []byte
}

var functionCallResponseDefault = MessageFormat{
	Name: "FunctionCallResponse",
	Fields: FieldGroup{
		{
			Name:  "Header",
			Type:  Byte1,
			Flags: Header,
			Data:  int32('V'),
		},
		{
			Name:  "MessageLength",
			Type:  Int32,
			Flags: MessageLengthInclusive,
			Data:  int32(0),
		},
		{
			Name:  "ResultLength",
			Type:  Int32,
			Flags: ByteCount,
			Data:  int32(0),
		},
		{
			Name: "ResultValue",
			Type: ByteN,
			Data: []byte{},
		},
	},
}

var _ Message = FunctionCallResponse{}

// encode implements the interface Message.
func (m FunctionCallResponse) encode() (MessageFormat, error) {
	outputMessage := m.defaultMessage().Copy()
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

// decode implements the interface Message.
func (m FunctionCallResponse) decode(s MessageFormat) (Message, error) {
	if err := s.MatchesStructure(*m.defaultMessage()); err != nil {
		return nil, err
	}
	isNull := s.Field("ResultLength").MustGet().(int32) == -1
	return FunctionCallResponse{
		IsResultNull: isNull,
		ResultValue:  s.Field("ResultValue").MustGet().([]byte),
	}, nil
}

// defaultMessage implements the interface Message.
func (m FunctionCallResponse) defaultMessage() *MessageFormat {
	return &functionCallResponseDefault
}
