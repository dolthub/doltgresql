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
	initializeDefaultMessage(AuthenticationSASLFinal{})
}

// AuthenticationSASLFinal represents a PostgreSQL message.
type AuthenticationSASLFinal struct {
	AdditionalData []byte
}

var authenticationSASLFinalDefault = MessageFormat{
	Name: "AuthenticationSASLFinal",
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
			Data: int32(12),
		},
		{
			Name: "AdditionalData",
			Type: ByteN,
			Data: []byte{},
		},
	},
}

var _ Message = AuthenticationSASLFinal{}

// encode implements the interface Message.
func (m AuthenticationSASLFinal) encode() (MessageFormat, error) {
	outputMessage := m.defaultMessage().Copy()
	outputMessage.Field("AdditionalData").MustWrite(m.AdditionalData)
	return outputMessage, nil
}

// decode implements the interface Message.
func (m AuthenticationSASLFinal) decode(s MessageFormat) (Message, error) {
	if err := s.MatchesStructure(*m.defaultMessage()); err != nil {
		return nil, err
	}
	return AuthenticationSASLFinal{
		AdditionalData: s.Field("AdditionalData").MustGet().([]byte),
	}, nil
}

// defaultMessage implements the interface Message.
func (m AuthenticationSASLFinal) defaultMessage() *MessageFormat {
	return &authenticationSASLFinalDefault
}
