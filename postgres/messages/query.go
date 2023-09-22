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
	connection.InitializeDefaultMessage(Query{})
	connection.AddMessageHeader(Query{})
}

// Query contains a query given by the client.
type Query struct {
	String string
}

var queryDefault = connection.MessageFormat{
	Name: "Query",
	Fields: connection.FieldGroup{
		{
			Name:  "Header",
			Type:  connection.Byte1,
			Flags: connection.Header,
			Data:  int32('Q'),
		},
		{
			Name:  "MessageLength",
			Type:  connection.Int32,
			Flags: connection.MessageLengthInclusive,
			Data:  int32(0),
		},
		{
			Name: "String",
			Type: connection.String,
			Data: "",
		},
	},
}

var _ connection.Message = Query{}

// Encode implements the interface connection.Message.
func (m Query) Encode() (connection.MessageFormat, error) {
	outputMessage := m.DefaultMessage().Copy()
	outputMessage.Field("String").MustWrite(m.String)
	return outputMessage, nil
}

// Decode implements the interface connection.Message.
func (m Query) Decode(s connection.MessageFormat) (connection.Message, error) {
	if err := s.MatchesStructure(*m.DefaultMessage()); err != nil {
		return nil, err
	}
	return Query{
		String: s.Field("String").MustGet().(string),
	}, nil
}

// DefaultMessage implements the interface connection.Message.
func (m Query) DefaultMessage() *connection.MessageFormat {
	return &queryDefault
}
