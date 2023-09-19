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
	initializeDefaultMessage(Query{})
	addMessageHeader(Query{})
}

// Query contains a query given by the client.
type Query struct {
	String string
}

var queryDefault = Message{
	Name: "Query",
	Fields: []*Field{
		{
			Name: "Header",
			Type: Byte1,
			Tags: Header,
			Data: int32('Q'),
		},
		{
			Name: "MessageLength",
			Type: Int32,
			Tags: MessageLengthInclusive,
			Data: int32(0),
		},
		{
			Name: "String",
			Type: String,
			Data: "",
		},
	},
}

var _ MessageType = Query{}

// encode implements the interface MessageType.
func (m Query) encode() (Message, error) {
	outputMessage := m.defaultMessage().Copy()
	outputMessage.Field("String").MustWrite(m.String)
	return outputMessage, nil
}

// decode implements the interface MessageType.
func (m Query) decode(s Message) (MessageType, error) {
	if err := s.MatchesStructure(*m.defaultMessage()); err != nil {
		return nil, err
	}
	return Query{
		String: s.Field("String").MustGet().(string),
	}, nil
}

// defaultMessage implements the interface MessageType.
func (m Query) defaultMessage() *Message {
	return &queryDefault
}
