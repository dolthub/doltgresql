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
	initializeDefaultMessage(AuthenticationMD5Password{})
}

// AuthenticationMD5Password represents a PostgreSQL message.
type AuthenticationMD5Password struct {
	Salt int32
}

var authenticationMD5PasswordDefault = MessageFormat{
	Name: "AuthenticationMD5Password",
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
			Data:  int32(12),
		},
		{
			Name: "Status",
			Type: Int32,
			Data: int32(5),
		},
		{
			Name: "Salt",
			Type: Byte4,
			Data: int32(0),
		},
	},
}

var _ Message = AuthenticationMD5Password{}

// encode implements the interface Message.
func (m AuthenticationMD5Password) encode() (MessageFormat, error) {
	outputMessage := m.defaultMessage().Copy()
	outputMessage.Field("Salt").MustWrite(m.Salt)
	return outputMessage, nil
}

// decode implements the interface Message.
func (m AuthenticationMD5Password) decode(s MessageFormat) (Message, error) {
	if err := s.MatchesStructure(*m.defaultMessage()); err != nil {
		return nil, err
	}
	return AuthenticationMD5Password{
		Salt: s.Field("Salt").MustGet().(int32),
	}, nil
}

// defaultMessage implements the interface Message.
func (m AuthenticationMD5Password) defaultMessage() *MessageFormat {
	return &authenticationMD5PasswordDefault
}
