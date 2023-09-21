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
	initializeDefaultMessage(AuthenticationSASLContinue{})
}

// AuthenticationSASLContinue represents a PostgreSQL message.
type AuthenticationSASLContinue struct {
	Data []byte
}

var authenticationSASLContinueDefault = Message{
	Name: "AuthenticationSASLContinue",
	Fields: []*Field{
		{
			Name: "Header",
			Type: Byte1,
			Tags: Header,
			Data: int32('R'),
		},
		{
			Name: "MessageLength",
			Type: Int32,
			Tags: MessageLengthInclusive,
			Data: int32(0),
		},
		{
			Name: "Status",
			Type: Int32,
			Data: int32(11),
		},
		{
			Name: "SASLData",
			Type: ByteN,
			Data: []byte{},
		},
	},
}

var _ MessageType = AuthenticationSASLContinue{}

// encode implements the interface MessageType.
func (m AuthenticationSASLContinue) encode() (Message, error) {
	outputMessage := m.defaultMessage().Copy()
	outputMessage.Field("SASLData").MustWrite(m.Data)
	return outputMessage, nil
}

// decode implements the interface MessageType.
func (m AuthenticationSASLContinue) decode(s Message) (MessageType, error) {
	if err := s.MatchesStructure(*m.defaultMessage()); err != nil {
		return nil, err
	}
	return AuthenticationSASLContinue{
		Data: s.Field("SASLData").MustGet().([]byte),
	}, nil
}

// defaultMessage implements the interface MessageType.
func (m AuthenticationSASLContinue) defaultMessage() *Message {
	return &authenticationSASLContinueDefault
}
