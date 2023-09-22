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
	connection.InitializeDefaultMessage(Parse{})
	connection.AddMessageHeader(Parse{})
}

// Parse represents a PostgreSQL message.
type Parse struct {
	PreparedStatement  string
	Query              string
	ParameterObjectIDs []int32
}

var parseDefault = connection.MessageFormat{
	Name: "Parse",
	Fields: connection.FieldGroup{
		{
			Name:  "Header",
			Type:  connection.Byte1,
			Flags: connection.Header,
			Data:  int32('P'),
		},
		{
			Name:  "MessageLength",
			Type:  connection.Int32,
			Flags: connection.MessageLengthInclusive,
			Data:  int32(0),
		},
		{
			Name: "PreparedStatement",
			Type: connection.String,
			Data: "",
		},
		{
			Name: "Query",
			Type: connection.String,
			Data: "",
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

var _ connection.Message = Parse{}

// Encode implements the interface connection.Message.
func (m Parse) Encode() (connection.MessageFormat, error) {
	outputMessage := m.DefaultMessage().Copy()
	outputMessage.Field("PreparedStatement").MustWrite(m.PreparedStatement)
	outputMessage.Field("Query").MustWrite(m.Query)
	for i, objectID := range m.ParameterObjectIDs {
		outputMessage.Field("Parameters").Child("ObjectID", i).MustWrite(objectID)
	}
	return outputMessage, nil
}

// Decode implements the interface connection.Message.
func (m Parse) Decode(s connection.MessageFormat) (connection.Message, error) {
	if err := s.MatchesStructure(*m.DefaultMessage()); err != nil {
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

// DefaultMessage implements the interface connection.Message.
func (m Parse) DefaultMessage() *connection.MessageFormat {
	return &parseDefault
}
