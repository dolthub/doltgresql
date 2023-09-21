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
	initializeDefaultMessage(Parse{})
	addMessageHeader(Parse{})
}

// Parse represents a PostgreSQL message.
type Parse struct {
	PreparedStatement  string
	Query              string
	ParameterObjectIDs []int32
}

var parseDefault = Message{
	Name: "Parse",
	Fields: []*Field{
		{
			Name: "Header",
			Type: Byte1,
			Tags: Header,
			Data: int32('P'),
		},
		{
			Name: "MessageLength",
			Type: Int32,
			Tags: MessageLengthInclusive,
			Data: int32(0),
		},
		{
			Name: "PreparedStatement",
			Type: String,
			Data: "",
		},
		{
			Name: "Query",
			Type: String,
			Data: "",
		},
		{
			Name: "Parameters",
			Type: Int16,
			Data: int32(0),
			Children: [][]*Field{
				{
					{
						Name: "ObjectID",
						Type: Int32,
						Data: int32(0),
					},
				},
			},
		},
	},
}

var _ MessageType = Parse{}

// encode implements the interface MessageType.
func (m Parse) encode() (Message, error) {
	outputMessage := m.defaultMessage().Copy()
	outputMessage.Field("PreparedStatement").MustWrite(m.PreparedStatement)
	outputMessage.Field("Query").MustWrite(m.Query)
	for i, objectID := range m.ParameterObjectIDs {
		outputMessage.Field("Parameters").Child("ObjectID", i).MustWrite(objectID)
	}
	return outputMessage, nil
}

// decode implements the interface MessageType.
func (m Parse) decode(s Message) (MessageType, error) {
	if err := s.MatchesStructure(*m.defaultMessage()); err != nil {
		return nil, err
	}
	count := int(s.Field("Parameters").MustGet().(int32))
	objectIDs := make([]int32, count)
	for i := 0; i < count; i++ {
		objectIDs[i] = s.Field("Parameters").Child("ObjectID", i).MustGet().(int32)
	}
	return Parse{
		PreparedStatement:  s.Field("PreparedStatement").MustGet().(string),
		Query:              s.Field("Query").MustGet().(string),
		ParameterObjectIDs: objectIDs,
	}, nil
}

// defaultMessage implements the interface MessageType.
func (m Parse) defaultMessage() *Message {
	return &parseDefault
}
