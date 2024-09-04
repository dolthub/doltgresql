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

import (
	"fmt"
	"github.com/dolthub/doltgresql/postgres/connection"
)

func init() {
	connection.InitializeDefaultMessage(DataRow{})
}

// DataRow represents a row of data.
type DataRow struct {
	Values [][]byte
}

var dataRowDefault = connection.MessageFormat{
	Name: "DataRow",
	Fields: connection.FieldGroup{
		{
			Name:  "Header",
			Type:  connection.Byte1,
			Flags: connection.Header,
			Data:  int32('D'),
		},
		{
			Name:  "MessageLength",
			Type:  connection.Int32,
			Flags: connection.MessageLengthInclusive,
			Data:  int32(0),
		},
		{
			Name: "Columns",
			Type: connection.Int16,
			Data: int32(0),
			Children: []connection.FieldGroup{
				{
					{
						Name:  "ColumnLength",
						Type:  connection.Int32,
						Flags: connection.ByteCount,
						Data:  int32(0),
					},
					{
						Name: "ColumnData",
						Type: connection.ByteN,
						Data: []byte{},
					},
				},
			},
		},
	},
}

var _ connection.Message = DataRow{}

// Encode implements the interface connection.Message.
func (m DataRow) Encode() (connection.MessageFormat, error) {
	outputMessage := m.DefaultMessage().Copy()
	for i := 0; i < len(m.Values); i++ {
		if m.Values[i] == nil {
			outputMessage.Field("Columns").Child("ColumnLength", i).MustWrite(-1)
		} else {
			value := m.Values[i]
			valLen := len(value)
			outputMessage.Field("Columns").Child("ColumnLength", i).MustWrite(valLen)
			outputMessage.Field("Columns").Child("ColumnData", i).MustWrite(value[:valLen])
		}
	}
	return outputMessage, nil
}

// Decode implements the interface connection.Message.
func (m DataRow) Decode(s connection.MessageFormat) (connection.Message, error) {
	if err := s.MatchesStructure(*m.DefaultMessage()); err != nil {
		return nil, err
	}
	return nil, fmt.Errorf("DataRow messages do not support decoding, as they're only sent from the server.")
}

// DefaultMessage implements the interface connection.Message.
func (m DataRow) DefaultMessage() *connection.MessageFormat {
	return &dataRowDefault
}
