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
	initializeDefaultMessage(AuthenticationSASL{})
}

// AuthenticationSASL represents a PostgreSQL message.
type AuthenticationSASL struct {
	Mechanisms []string
}

var authenticationSASLDefault = Message{
	Name: "AuthenticationSASL",
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
			Data: int32(10),
		},
		{
			Name: "Mechanisms",
			Type: Repeated,
			Tags: RepeatedTerminator,
			Data: int32(0),
			Children: [][]*Field{
				{
					{
						Name: "Mechanism",
						Type: String,
						Data: "",
					},
				},
			},
		},
	},
}

var _ MessageType = AuthenticationSASL{}

// encode implements the interface MessageType.
func (m AuthenticationSASL) encode() (Message, error) {
	outputMessage := m.defaultMessage().Copy()
	for i, mechanism := range m.Mechanisms {
		outputMessage.Field("Mechanisms").Child("Mechanism", i).MustWrite(mechanism)
	}
	return outputMessage, nil
}

// decode implements the interface MessageType.
func (m AuthenticationSASL) decode(s Message) (MessageType, error) {
	if err := s.MatchesStructure(*m.defaultMessage()); err != nil {
		return nil, err
	}
	count := int(s.Field("Mechanisms").MustGet().(int32))
	mechanisms := make([]string, count)
	for i := 0; i < count; i++ {
		mechanisms[i] = s.Field("Mechanisms").Child("Mechanism", i).MustGet().(string)
	}
	return AuthenticationSASL{
		Mechanisms: mechanisms,
	}, nil
}

// defaultMessage implements the interface MessageType.
func (m AuthenticationSASL) defaultMessage() *Message {
	return &authenticationSASLDefault
}
