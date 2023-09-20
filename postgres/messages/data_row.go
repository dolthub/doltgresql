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
	"github.com/dolthub/vitess/go/sqltypes"
)

func init() {
	initializeDefaultMessage(DataRow{})
	addMessageHeader(DataRow{})
}

// DataRow represents a row of data.
type DataRow struct {
	Values []sqltypes.Value
}

var dataRowDefault = Message{
	Name: "DataRow",
	Fields: []*Field{
		{
			Name: "Header",
			Type: Byte1,
			Tags: Header,
			Data: int32('D'),
		},
		{
			Name: "MessageLength",
			Type: Int32,
			Tags: MessageLengthInclusive,
			Data: int32(0),
		},
		{
			Name: "Columns",
			Type: Int16,
			Data: int32(0),
			Children: [][]*Field{
				{
					{
						Name: "ColumnLength",
						Type: Int32,
						Tags: ByteCount,
						Data: int32(0),
					},
					{
						Name: "ColumnData",
						Type: ByteN,
						Data: []byte{},
					},
				},
			},
		},
	},
}

var _ MessageType = DataRow{}

// encode implements the interface MessageType.
func (m DataRow) encode() (Message, error) {
	outputMessage := m.defaultMessage().Copy()
	for i := 0; i < len(m.Values); i++ {
		if m.Values[i].IsNull() {
			outputMessage.Field("Columns").Child("ColumnLength", i).MustWrite(-1)
		} else {
			value := []byte(m.Values[i].ToString())
			outputMessage.Field("Columns").Child("ColumnLength", i).MustWrite(len(value))
			outputMessage.Field("Columns").Child("ColumnData", i).MustWrite(value)
		}
	}
	return outputMessage, nil
}

// decode implements the interface MessageType.
func (m DataRow) decode(s Message) (MessageType, error) {
	if err := s.MatchesStructure(*m.defaultMessage()); err != nil {
		return nil, err
	}
	columnCount := int(s.Field("Columns").MustGet().(int32))
	for i := 0; i < columnCount; i++ {
		//TODO: decode the message in here
	}
	return DataRow{
		Values: nil,
	}, nil
}

// defaultMessage implements the interface MessageType.
func (m DataRow) defaultMessage() *Message {
	return &dataRowDefault
}
