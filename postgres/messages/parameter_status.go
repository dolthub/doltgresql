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
	initializeDefaultMessage(ParameterStatus{})
	addMessageHeader(ParameterStatus{})
}

// ParameterStatus reports various parameters to the client.
type ParameterStatus struct {
	Name  string
	Value string
}

var parameterStatusDefault = Message{
	Name: "ParameterStatus",
	Fields: []*Field{
		{
			Name: "Header",
			Type: Byte1,
			Tags: Header,
			Data: int32('S'),
		},
		{
			Name: "MessageLength",
			Type: Int32,
			Tags: MessageLengthInclusive,
			Data: int32(0),
		},
		{
			Name: "Name",
			Type: String,
			Data: "",
		},
		{
			Name: "Value",
			Type: String,
			Data: "",
		},
	},
}

var _ MessageType = ParameterStatus{}

// encode implements the interface MessageType.
func (m ParameterStatus) encode() (Message, error) {
	outputMessage := m.defaultMessage().Copy()
	outputMessage.Field("Name").MustWrite(m.Name)
	outputMessage.Field("Value").MustWrite(m.Value)
	return outputMessage, nil
}

// decode implements the interface MessageType.
func (m ParameterStatus) decode(s Message) (MessageType, error) {
	if err := s.MatchesStructure(*m.defaultMessage()); err != nil {
		return nil, err
	}
	return ParameterStatus{
		Name:  s.Field("Name").MustGet().(string),
		Value: s.Field("Value").MustGet().(string),
	}, nil
}

// defaultMessage implements the interface MessageType.
func (m ParameterStatus) defaultMessage() *Message {
	return &parameterStatusDefault
}
