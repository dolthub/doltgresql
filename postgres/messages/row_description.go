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

	"github.com/dolthub/doltgresql/postgres/connection"
)

func init() {
	connection.InitializeDefaultMessage(RowDescription{})
}

// RowDescription represents a RowDescription message intended for the client.
type RowDescription struct {
	Fields []*query.Field
}

var rowDescriptionDefault = connection.MessageFormat{
	Name: "RowDescription",
	Fields: connection.FieldGroup{
		{
			Name:  "Header",
			Type:  connection.Byte1,
			Flags: connection.Header,
			Data:  int32('T'),
		},
		{
			Name:  "MessageLength",
			Type:  connection.Int32,
			Flags: connection.MessageLengthInclusive,
			Data:  int32(0),
		},
		{
			Name: "Fields",
			Type: connection.Int16,
			Data: int32(0),
			Children: []connection.FieldGroup{
				{
					{
						Name: "ColumnName",
						Type: connection.String,
						Data: "",
					},
					{
						Name: "TableObjectID",
						Type: connection.Int32,
						Data: int32(0),
					},
					{
						Name: "ColumnAttributeNumber",
						Type: connection.Int16,
						Data: int32(0),
					},
					{
						Name: "DataTypeObjectID",
						Type: connection.Int32,
						Data: int32(0),
					},
					{
						Name: "DataTypeSize",
						Type: connection.Int16,
						Data: int32(0),
					},
					{
						Name: "DataTypeModifier",
						Type: connection.Int32,
						Data: int32(0),
					},
					{
						Name: "FormatCode",
						Type: connection.Int16,
						Data: int32(0),
					},
				},
			},
		},
	},
}

var _ connection.Message = RowDescription{}

// Encode implements the interface connection.Message.
func (m RowDescription) Encode() (connection.MessageFormat, error) {
	outputMessage := m.DefaultMessage().Copy()
	for i := 0; i < len(m.Fields); i++ {
		field := m.Fields[i]
		dataTypeObjectID, err := VitessFieldToDataTypeObjectID(field)
		if err != nil {
			return connection.MessageFormat{}, err
		}
		dataTypeSize, err := VitessFieldToDataTypeSize(field)
		if err != nil {
			return connection.MessageFormat{}, err
		}
		dataTypeModifier, err := VitessFieldToDataTypeModifier(field)
		if err != nil {
			return connection.MessageFormat{}, err
		}
		outputMessage.Field("Fields").Child("ColumnName", i).MustWrite(field.Name)
		outputMessage.Field("Fields").Child("DataTypeObjectID", i).MustWrite(dataTypeObjectID)
		outputMessage.Field("Fields").Child("DataTypeSize", i).MustWrite(dataTypeSize)
		outputMessage.Field("Fields").Child("DataTypeModifier", i).MustWrite(dataTypeModifier)
	}
	return outputMessage, nil
}

// Decode implements the interface connection.Message.
func (m RowDescription) Decode(s connection.MessageFormat) (connection.Message, error) {
	if err := s.MatchesStructure(*m.DefaultMessage()); err != nil {
		return nil, err
	}
	return nil, fmt.Errorf("RowDescription messages do not support decoding, as they're only sent from the server.")
}

// DefaultMessage implements the interface connection.Message.
func (m RowDescription) DefaultMessage() *connection.MessageFormat {
	return &rowDescriptionDefault
}

// VitessFieldToDataTypeObjectID returns a type, as defined by Vitess, into a type as defined by Postgres.
// OIDs can be obtained with the following query: `SELECT oid, typname FROM pg_type ORDER BY 1;`
func VitessFieldToDataTypeObjectID(field *query.Field) (int32, error) {
	switch field.Type {
	case query.Type_INT8:
		// Postgres doesn't make use of a small integer type for integer returns, which presents a bit of a conundrum.
		// GMS defines boolean operations as the smallest integer type, while Postgres has an explicit bool type.
		// We can't always assume that `INT8` means bool, since it could just be a small integer. As a result, we'll
		// always return this as though it's an `INT32`, which also means that we can't support bools right now.
		// OIDs 16 (bool) and 18 (char, ASCII only?) are the only single-byte types as far as I'm aware.
		return 23, nil
	case query.Type_INT16:
		// The technically correct OID is 21 (2-byte integer), however it seems like some clients don't actually expect
		// this, so I'm not sure when it's actually used by Postgres. Because of this, we'll just pretend it's an `INT32`.
		return 23, nil
	case query.Type_INT24:
		// Postgres doesn't have a 3-byte integer type, so just pretend it's `INT32`.
		return 23, nil
	case query.Type_INT32:
		return 23, nil
	case query.Type_INT64:
		return 20, nil
	case query.Type_FLOAT32:
		return 700, nil
	case query.Type_FLOAT64:
		return 701, nil
	case query.Type_DECIMAL:
		return 1700, nil
	case query.Type_CHAR:
		return 1042, nil
	case query.Type_VARCHAR:
		return 1043, nil
	case query.Type_TEXT:
		return 25, nil
	case query.Type_JSON:
		return 114, nil
	case query.Type_TIMESTAMP, query.Type_DATETIME:
		return 1114, nil
	case query.Type_DATE:
		return 1082, nil
	case query.Type_NULL_TYPE:
		return 25, nil // NULL is treated as TEXT on the wire
	default:
		return 0, fmt.Errorf("unsupported type returned from engine: %s", field.Type)
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
	case query.Type_FLOAT32:
		return 4, nil
	case query.Type_FLOAT64:
		return 8, nil
	case query.Type_DECIMAL:
		return -1, nil
	case query.Type_CHAR:
		return -1, nil
	case query.Type_VARCHAR:
		return -1, nil
	case query.Type_TEXT:
		return -1, nil
	case query.Type_JSON:
		return -1, nil
	case query.Type_TIMESTAMP, query.Type_DATETIME:
		return 8, nil
	case query.Type_DATE:
		return 4, nil
	case query.Type_NULL_TYPE:
		return -1, nil // NULL is treated as TEXT on the wire
	default:
		return 0, fmt.Errorf("unsupported type returned from engine: %s", field.Type)
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
	case query.Type_FLOAT32:
		return -1, nil
	case query.Type_FLOAT64:
		return -1, nil
	case query.Type_DECIMAL:
		// This is how we encode the precision and scale for some reason
		precision := int32(field.ColumnLength - 1)
		scale := int32(field.Decimals)
		if scale > 0 {
			precision--
		}
		// PostgreSQL adds 4 to the length for an unknown reason
		return (precision<<16 + scale) + 4, nil
	case query.Type_CHAR:
		// PostgreSQL adds 4 to the length for an unknown reason
		return int32(int64(field.ColumnLength)/sql.CharacterSetID(field.Charset).MaxLength()) + 4, nil
	case query.Type_VARCHAR:
		// PostgreSQL adds 4 to the length for an unknown reason
		return int32(int64(field.ColumnLength)/sql.CharacterSetID(field.Charset).MaxLength()) + 4, nil
	case query.Type_TEXT:
		return -1, nil
	case query.Type_JSON:
		return -1, nil
	case query.Type_TIMESTAMP, query.Type_DATETIME:
		return -1, nil
	case query.Type_DATE:
		return -1, nil
	case query.Type_NULL_TYPE:
		return -1, nil // NULL is treated as TEXT on the wire
	default:
		return 0, fmt.Errorf("unsupported type returned from engine: %s", field.Type)
	}
}
