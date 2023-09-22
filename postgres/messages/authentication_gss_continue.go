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
	initializeDefaultMessage(AuthenticationGSSContinue{})
}

// AuthenticationGSSContinue represents a PostgreSQL message.
type AuthenticationGSSContinue struct {
	Data []byte
}

var authenticationGSSContinueDefault = MessageFormat{
	Name: "AuthenticationGSSContinue",
	Fields: FieldGroup{
		{
			Name:  "Header",
			Type:  Byte1,
			Flags: Header,
			Data:  int32('R'),
		},
		{
			Name:  "MessageLength",
			Type:  Int32,
			Flags: MessageLengthInclusive,
			Data:  int32(0),
		},
		{
			Name: "Status",
			Type: Int32,
			Data: int32(8),
		},
		{
			Name: "AuthenticationData",
			Type: ByteN,
			Data: []byte{},
		},
	},
}

var _ Message = AuthenticationGSSContinue{}

// encode implements the interface Message.
func (m AuthenticationGSSContinue) encode() (MessageFormat, error) {
	outputMessage := m.defaultMessage().Copy()
	outputMessage.Field("AuthenticationData").MustWrite(m.Data)
	return outputMessage, nil
}

// decode implements the interface Message.
func (m AuthenticationGSSContinue) decode(s MessageFormat) (Message, error) {
	if err := s.MatchesStructure(*m.defaultMessage()); err != nil {
		return nil, err
	}
	return AuthenticationGSSContinue{
		Data: s.Field("AuthenticationData").MustGet().([]byte),
	}, nil
}

// defaultMessage implements the interface Message.
func (m AuthenticationGSSContinue) defaultMessage() *MessageFormat {
	return &authenticationGSSContinueDefault
}
