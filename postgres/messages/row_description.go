package messages

import (
	"bytes"
	"fmt"

	"github.com/dolthub/vitess/go/vt/proto/query"
)

// RowDescription represents a RowDescription message intended for the client.
type RowDescription struct {
	Fields []RowDescriptionField
}

// NewRowDescription creates a new RowDescription from the given fields.
func NewRowDescription(fields []*query.Field) (RowDescription, error) {
	var err error
	rdFields := make([]RowDescriptionField, len(fields))
	for i, field := range fields {
		rdFields[i] = RowDescriptionField{
			TableObjectID:         0, // Unused for now
			ColumnAttributeNumber: 0, // Unused for now
			DataTypeModifier:      0, // Always -1 since we're supporting a narrow set of integers
			FormatCode:            0, // Always text for now
		}
		rdFields[i].Name = field.Name
		rdFields[i].DataTypeObjectID, err = VitessTypeToDataTypeObjectID(field.Type)
		if err != nil {
			return RowDescription{}, err
		}
		rdFields[i].DataTypeSize, err = VitessTypeToDataTypeSize(field.Type)
		if err != nil {
			return RowDescription{}, err
		}
	}
	return RowDescription{
		Fields: rdFields,
	}, nil
}

// RowDescriptionField represents a field in RowDescription.
type RowDescriptionField struct {
	Name                  string
	TableObjectID         int32
	ColumnAttributeNumber int16
	DataTypeObjectID      int32
	DataTypeSize          int16
	DataTypeModifier      int32
	FormatCode            int16
}

// Bytes returns RowDescription as a byte slice, ready to be returned to the client.
func (rd RowDescription) Bytes() []byte {
	buf := bytes.Buffer{}
	buf.WriteByte('T')          // Message Type
	WriteNumber(&buf, int32(0)) // Message length, will be corrected later
	WriteNumber(&buf, int16(len(rd.Fields)))
	for _, rdf := range rd.Fields {
		rdf.Bytes(&buf)
	}
	return WriteLength(buf.Bytes())
}

// Bytes writes the field into the given buffer.
func (rdf RowDescriptionField) Bytes(buf *bytes.Buffer) {
	buf.WriteString(rdf.Name)
	buf.WriteByte(0) // Trailing NULL character, denoting the end of the string
	WriteNumber(buf, rdf.TableObjectID)
	WriteNumber(buf, rdf.ColumnAttributeNumber)
	WriteNumber(buf, rdf.DataTypeObjectID)
	WriteNumber(buf, rdf.DataTypeSize)
	WriteNumber(buf, rdf.DataTypeModifier)
	WriteNumber(buf, rdf.FormatCode)
}

// VitessTypeToDataTypeObjectID returns a type, as defined by Vitess, into a type as defined by Postgres.
func VitessTypeToDataTypeObjectID(typ query.Type) (int32, error) {
	switch typ {
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
	default:
		return 0, fmt.Errorf("unsupported type returned from engine")
	}
}

// VitessTypeToDataTypeSize returns the type's size, as defined by Vitess, into the size as defined by Postgres.
func VitessTypeToDataTypeSize(typ query.Type) (int16, error) {
	switch typ {
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
	default:
		return 0, fmt.Errorf("unsupported type returned from engine")
	}
}
