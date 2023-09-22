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

import "github.com/dolthub/doltgresql/postgres/connection"

func init() {
	connection.InitializeDefaultMessage(ParameterDescription{})
}

// ParameterDescription represents a PostgreSQL message.
type ParameterDescription struct {
	ObjectIDs []int32
}

var parameterDescriptionDefault = connection.MessageFormat{
	Name: "ParameterDescription",
	Fields: connection.FieldGroup{
		{
			Name:  "Header",
			Type:  connection.Byte1,
			Flags: connection.Header,
			Data:  int32('t'),
		},
		{
			Name:  "MessageLength",
			Type:  connection.Int32,
			Flags: connection.MessageLengthInclusive,
			Data:  int32(0),
		},
		{
			Name: "Parameters",
			Type: connection.Int16,
			Data: int32(0),
			Children: []connection.FieldGroup{
				{
					{
						Name: "ObjectID",
						Type: connection.Int32,
						Data: int32(0),
					},
				},
			},
		},
	},
}

var _ connection.Message = ParameterDescription{}

// Encode implements the interface connection.Message.
func (m ParameterDescription) Encode() (connection.MessageFormat, error) {
	outputMessage := m.DefaultMessage().Copy()
	for i, objectID := range m.ObjectIDs {
		outputMessage.Field("Parameters").Child("ObjectID", i).MustWrite(objectID)
	}
	return outputMessage, nil
}

// Decode implements the interface connection.Message.
func (m ParameterDescription) Decode(s connection.MessageFormat) (connection.Message, error) {
	if err := s.MatchesStructure(*m.DefaultMessage()); err != nil {
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

// DefaultMessage implements the interface connection.Message.
func (m ParameterDescription) DefaultMessage() *connection.MessageFormat {
	return &parameterDescriptionDefault
}
