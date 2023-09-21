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
	initializeDefaultMessage(CopyFail{})
	addMessageHeader(CopyFail{})
}

// CopyFail represents a PostgreSQL message.
type CopyFail struct {
	ErrorMessage string
}

var copyFailDefault = Message{
	Name: "CopyFail",
	Fields: []*Field{
		{
			Name: "Header",
			Type: Byte1,
			Tags: Header,
			Data: int32('f'),
		},
		{
			Name: "MessageLength",
			Type: Int32,
			Tags: MessageLengthInclusive,
			Data: int32(0),
		},
		{
			Name: "ErrorMessage",
			Type: String,
			Data: "",
		},
	},
}

var _ MessageType = CopyFail{}

// encode implements the interface MessageType.
func (m CopyFail) encode() (Message, error) {
	outputMessage := m.defaultMessage().Copy()
	outputMessage.Field("ErrorMessage").MustWrite(m.ErrorMessage)
	return outputMessage, nil
}

// decode implements the interface MessageType.
func (m CopyFail) decode(s Message) (MessageType, error) {
	if err := s.MatchesStructure(*m.defaultMessage()); err != nil {
		return nil, err
	}
	return CopyFail{
		ErrorMessage: s.Field("ErrorMessage").MustGet().(string),
	}, nil
}

// defaultMessage implements the interface MessageType.
func (m CopyFail) defaultMessage() *Message {
	return &copyFailDefault
}
