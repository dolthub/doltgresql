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
	initializeDefaultMessage(AuthenticationKerberosV5{})
}

// AuthenticationKerberosV5 represents a PostgreSQL message.
type AuthenticationKerberosV5 struct{}

var authenticationKerberosV5Default = MessageFormat{
	Name: "AuthenticationKerberosV5",
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
			Data:  int32(8),
		},
		{
			Name: "Status",
			Type: Int32,
			Data: int32(2),
		},
	},
}

var _ Message = AuthenticationKerberosV5{}

// encode implements the interface Message.
func (m AuthenticationKerberosV5) encode() (MessageFormat, error) {
	return m.defaultMessage().Copy(), nil
}

// decode implements the interface Message.
func (m AuthenticationKerberosV5) decode(s MessageFormat) (Message, error) {
	if err := s.MatchesStructure(*m.defaultMessage()); err != nil {
		return nil, err
	}
	return AuthenticationKerberosV5{}, nil
}

// defaultMessage implements the interface Message.
func (m AuthenticationKerberosV5) defaultMessage() *MessageFormat {
	return &authenticationKerberosV5Default
}
