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

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/vitess/go/vt/proto/query"
)

func init() {
	initializeDefaultMessage(RowDescription{})
}

// RowDescription represents a RowDescription message intended for the client.
type RowDescription struct {
	Fields []*query.Field
}

var rowDescriptionDefault = Message{
	Name: "RowDescription",
	Fields: []*Field{
		{
			Name: "Header",
			Type: Byte1,
			Tags: Header,
			Data: int32('T'),
		},
		{
			Name: "MessageLength",
			Type: Int32,
			Tags: MessageLengthInclusive,
			Data: int32(0),
		},
		{
			Name: "Fields",
			Type: Int16,
			Data: int32(0),
			Children: [][]*Field{
				{
					{
						Name: "ColumnName",
						Type: String,
						Data: "",
					},
					{
						Name: "TableObjectID",
						Type: Int32,
						Data: int32(0),
					},
					{
						Name: "ColumnAttributeNumber",
						Type: Int16,
						Data: int32(0),
					},
					{
						Name: "DataTypeObjectID",
						Type: Int32,
						Data: int32(0),
					},
					{
						Name: "DataTypeSize",
						Type: Int16,
						Data: int32(0),
					},
					{
						Name: "DataTypeModifier",
						Type: Int32,
						Data: int32(0),
					},
					{
						Name: "FormatCode",
						Type: Int16,
						Data: int32(0),
					},
				},
			},
		},
	},
}

var _ MessageType = RowDescription{}

// encode implements the interface MessageType.
func (m RowDescription) encode() (Message, error) {
	outputMessage := m.defaultMessage().Copy()
	for i := 0; i < len(m.Fields); i++ {
		field := m.Fields[i]
		dataTypeObjectID, err := VitessFieldToDataTypeObjectID(field)
		if err != nil {
			return Message{}, err
		}
		dataTypeSize, err := VitessFieldToDataTypeSize(field)
		if err != nil {
			return Message{}, err
		}
		dataTypeModifier, err := VitessFieldToDataTypeModifier(field)
		if err != nil {
			return Message{}, err
		}
		outputMessage.Field("Fields").Child("ColumnName", i).MustWrite(field.Name)
		outputMessage.Field("Fields").Child("DataTypeObjectID", i).MustWrite(dataTypeObjectID)
		outputMessage.Field("Fields").Child("DataTypeSize", i).MustWrite(dataTypeSize)
		outputMessage.Field("Fields").Child("DataTypeModifier", i).MustWrite(dataTypeModifier)
	}
	return outputMessage, nil
}

// decode implements the interface MessageType.
func (m RowDescription) decode(s Message) (MessageType, error) {
	if err := s.MatchesStructure(*m.defaultMessage()); err != nil {
		return nil, err
	}
	fieldCount := int(s.Field("Fields").MustGet().(int32))
	for i := 0; i < fieldCount; i++ {
		//TODO: decode the message in here
	}
	return RowDescription{
		Fields: nil,
	}, nil
}

// defaultMessage implements the interface MessageType.
func (m RowDescription) defaultMessage() *Message {
	return &rowDescriptionDefault
}

// VitessFieldToDataTypeObjectID returns a type, as defined by Vitess, into a type as defined by Postgres.
func VitessFieldToDataTypeObjectID(field *query.Field) (int32, error) {
	switch field.Type {
	case query.Type_INT8:
		return 17, nil
	case query.Type_INT16:
		return 21, nil
	case query.Type_INT24:
		return 23, nil
	case query.Type_INT32:
		return 23, nil
	case query.Type_INT64:
		return 20, nil
	case query.Type_CHAR:
		return 18, nil
	case query.Type_VARCHAR:
		return 1043, nil
	case query.Type_TEXT:
		return 25, nil
	default:
		return 0, fmt.Errorf("unsupported type returned from engine")
	}
}

// VitessFieldToDataTypeSize returns the type's size, as defined by Vitess, into the size as defined by Postgres.
func VitessFieldToDataTypeSize(field *query.Field) (int16, error) {
	switch field.Type {
	case query.Type_INT8:
		return 1, nil
	case query.Type_INT16:
		return 2, nil
	case query.Type_INT24:
		return 4, nil
	case query.Type_INT32:
		return 4, nil
	case query.Type_INT64:
		return 8, nil
	case query.Type_CHAR:
		return -1, nil
	case query.Type_VARCHAR:
		return -1, nil
	case query.Type_TEXT:
		return -1, nil
	default:
		return 0, fmt.Errorf("unsupported type returned from engine")
	}
}

// VitessFieldToDataTypeModifier returns the field's data type modifier as defined by Postgres.
func VitessFieldToDataTypeModifier(field *query.Field) (int32, error) {
	switch field.Type {
	case query.Type_INT8:
		return -1, nil
	case query.Type_INT16:
		return -1, nil
	case query.Type_INT24:
		return -1, nil
	case query.Type_INT32:
		return -1, nil
	case query.Type_INT64:
		return -1, nil
	case query.Type_CHAR:
		// PostgreSQL adds 4 to the length for an unknown reason
		return int32(int64(field.ColumnLength)/sql.CharacterSetID(field.Charset).MaxLength()) + 4, nil
	case query.Type_VARCHAR:
		// PostgreSQL adds 4 to the length for an unknown reason
		return int32(int64(field.ColumnLength)/sql.CharacterSetID(field.Charset).MaxLength()) + 4, nil
	case query.Type_TEXT:
		return -1, nil
	default:
		return 0, fmt.Errorf("unsupported type returned from engine")
	}
}
