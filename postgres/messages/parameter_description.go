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
	initializeDefaultMessage(ParameterDescription{})
}

// ParameterDescription represents a PostgreSQL message.
type ParameterDescription struct {
	ObjectIDs []int32
}

var parameterDescriptionDefault = MessageFormat{
	Name: "ParameterDescription",
	Fields: FieldGroup{
		{
			Name:  "Header",
			Type:  Byte1,
			Flags: Header,
			Data:  int32('t'),
		},
		{
			Name:  "MessageLength",
			Type:  Int32,
			Flags: MessageLengthInclusive,
			Data:  int32(0),
		},
		{
			Name: "Parameters",
			Type: Int16,
			Data: int32(0),
			Children: []FieldGroup{
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

var _ Message = ParameterDescription{}

// encode implements the interface Message.
func (m ParameterDescription) encode() (MessageFormat, error) {
	outputMessage := m.defaultMessage().Copy()
	for i, objectID := range m.ObjectIDs {
		outputMessage.Field("Parameters").Child("ObjectID", i).MustWrite(objectID)
	}
	return outputMessage, nil
}

// decode implements the interface Message.
func (m ParameterDescription) decode(s MessageFormat) (Message, error) {
	if err := s.MatchesStructure(*m.defaultMessage()); err != nil {
		return nil, err
	}
	count := int(s.Field("Parameters").MustGet().(int32))
	objectIDs := make([]int32, count)
	for i := 0; i < count; i++ {
		objectIDs[i] = s.Field("Parameters").Child("ObjectID", i).MustGet().(int32)
	}
	return ParameterDescription{
		ObjectIDs: objectIDs,
	}, nil
}

// defaultMessage implements the interface Message.
func (m ParameterDescription) defaultMessage() *MessageFormat {
	return &parameterDescriptionDefault
}
