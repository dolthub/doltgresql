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
	initializeDefaultMessage(AuthenticationSCMCredential{})
}

// AuthenticationSCMCredential represents a PostgreSQL message.
type AuthenticationSCMCredential struct{}

var authenticationSCMCredentialDefault = Message{
	Name: "AuthenticationSCMCredential",
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
			Data: int32(8),
		},
		{
			Name: "Status",
			Type: Int32,
			Data: int32(6),
		},
	},
}

var _ MessageType = AuthenticationSCMCredential{}

// encode implements the interface MessageType.
func (m AuthenticationSCMCredential) encode() (Message, error) {
	return m.defaultMessage().Copy(), nil
}

// decode implements the interface MessageType.
func (m AuthenticationSCMCredential) decode(s Message) (MessageType, error) {
	if err := s.MatchesStructure(*m.defaultMessage()); err != nil {
		return nil, err
	}
	return AuthenticationSCMCredential{}, nil
}

// defaultMessage implements the interface MessageType.
func (m AuthenticationSCMCredential) defaultMessage() *Message {
	return &authenticationSCMCredentialDefault
}
