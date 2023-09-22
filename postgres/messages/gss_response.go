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
	connection.InitializeDefaultMessage(GSSResponse{})
}

// GSSResponse represents a PostgreSQL message.
type GSSResponse struct {
	Data []byte
}

var gSSResponseDefault = connection.MessageFormat{
	Name: "GSSResponse",
	Fields: connection.FieldGroup{
		{
			Name:  "Header",
			Type:  connection.Byte1,
			Flags: connection.Header,
			Data:  int32('p'),
		},
		{
			Name:  "MessageLength",
			Type:  connection.Int32,
			Flags: connection.MessageLengthInclusive,
			Data:  int32(0),
		},
		{
			Name: "Data",
			Type: connection.ByteN,
			Data: []byte{},
		},
	},
}

var _ connection.Message = GSSResponse{}

// Encode implements the interface connection.Message.
func (m GSSResponse) Encode() (connection.MessageFormat, error) {
	outputMessage := m.DefaultMessage().Copy()
	outputMessage.Field("Data").MustWrite(m.Data)
	return outputMessage, nil
}

// Decode implements the interface connection.Message.
func (m GSSResponse) Decode(s connection.MessageFormat) (connection.Message, error) {
	if err := s.MatchesStructure(*m.DefaultMessage()); err != nil {
		return nil, err
	}
	return GSSResponse{
		Data: s.Field("Data").MustGet().([]byte),
	}, nil
}

// DefaultMessage implements the interface connection.Message.
func (m GSSResponse) DefaultMessage() *connection.MessageFormat {
	return &gSSResponseDefault
}
