// Copyright 2024 Dolthub, Inc.
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

package types

import (
	"encoding/binary"
	"fmt"

	"github.com/dolthub/go-mysql-server/sql/types"
)

// SerializationID is an ID unique to Doltgres that can uniquely identify any type for the purposes of Serialization.
// These are different from OIDs, as they are unchanging and unique. If we need to add a new type that does not already
// have a pre-defined ID, then it must use a new number that has never been previously used.
type SerializationID uint16

// These are declared as constant numbers to signify their intent. Under no circumstances should we use iota, as that
// runs the risk of an accidental reordering potentially causing data loss. In addition, numbers for pre-existing IDs
// should never be changed.
const (
	SerializationID_Invalid               SerializationID = 0
	SerializationID_Bit                   SerializationID = 1
	SerializationID_BitArray              SerializationID = 2
	SerializationID_Bool                  SerializationID = 3
	SerializationID_BoolArray             SerializationID = 4
	SerializationID_Box                   SerializationID = 5
	SerializationID_BoxArray              SerializationID = 6
	SerializationID_Bytea                 SerializationID = 7
	SerializationID_ByteaArray            SerializationID = 8
	SerializationID_Char                  SerializationID = 9
	SerializationID_CharArray             SerializationID = 10
	SerializationID_Cidr                  SerializationID = 11
	SerializationID_CidrArray             SerializationID = 12
	SerializationID_Circle                SerializationID = 13
	SerializationID_CircleArray           SerializationID = 14
	SerializationID_Date                  SerializationID = 15
	SerializationID_DateArray             SerializationID = 16
	SerializationID_DateMultirange        SerializationID = 17
	SerializationID_DateRange             SerializationID = 18
	SerializationID_Enum                  SerializationID = 19
	SerializationID_EnumArray             SerializationID = 20
	SerializationID_Float32               SerializationID = 21
	SerializationID_Float32Array          SerializationID = 22
	SerializationID_Float64               SerializationID = 23
	SerializationID_Float64Array          SerializationID = 24
	SerializationID_Inet                  SerializationID = 25
	SerializationID_InetArray             SerializationID = 26
	SerializationID_Int16                 SerializationID = 27
	SerializationID_Int16Array            SerializationID = 28
	SerializationID_Int32                 SerializationID = 29
	SerializationID_Int32Array            SerializationID = 30
	SerializationID_Int32Multirange       SerializationID = 31
	SerializationID_Int32Range            SerializationID = 32
	SerializationID_Int64                 SerializationID = 33
	SerializationID_Int64Array            SerializationID = 34
	SerializationID_Int64Multirange       SerializationID = 35
	SerializationID_Int64Range            SerializationID = 36
	SerializationID_Interval              SerializationID = 37
	SerializationID_IntervalArray         SerializationID = 38
	SerializationID_Json                  SerializationID = 39
	SerializationID_JsonArray             SerializationID = 40
	SerializationID_JsonB                 SerializationID = 41
	SerializationID_JsonBArray            SerializationID = 42
	SerializationID_Line                  SerializationID = 43
	SerializationID_LineArray             SerializationID = 44
	SerializationID_LineSegment           SerializationID = 45
	SerializationID_LineSegmentArray      SerializationID = 46
	SerializationID_MacAddress            SerializationID = 47
	SerializationID_MacAddress8           SerializationID = 48
	SerializationID_MacAddress8Array      SerializationID = 49
	SerializationID_MacAddressArray       SerializationID = 50
	SerializationID_Money                 SerializationID = 51
	SerializationID_MoneyArray            SerializationID = 52
	SerializationID_Null                  SerializationID = 53
	SerializationID_Numeric               SerializationID = 54
	SerializationID_NumericArray          SerializationID = 55
	SerializationID_NumericMultirange     SerializationID = 56
	SerializationID_NumericRange          SerializationID = 57
	SerializationID_Path                  SerializationID = 58
	SerializationID_PathArray             SerializationID = 59
	SerializationID_Point                 SerializationID = 60
	SerializationID_PointArray            SerializationID = 61
	SerializationID_Polygon               SerializationID = 62
	SerializationID_PolygonArray          SerializationID = 63
	SerializationID_Text                  SerializationID = 64
	SerializationID_TextArray             SerializationID = 65
	SerializationID_Time                  SerializationID = 66
	SerializationID_TimeArray             SerializationID = 67
	SerializationID_TimeTZ                SerializationID = 68
	SerializationID_TimeTZArray           SerializationID = 69
	SerializationID_Timestamp             SerializationID = 70
	SerializationID_TimestampArray        SerializationID = 71
	SerializationID_TimestampMultirange   SerializationID = 72
	SerializationID_TimestampRange        SerializationID = 73
	SerializationID_TimestampTZ           SerializationID = 74
	SerializationID_TimestampTZArray      SerializationID = 75
	SerializationID_TimestampTZMultirange SerializationID = 76
	SerializationID_TimestampTZRange      SerializationID = 77
	SerializationID_TsQuery               SerializationID = 78
	SerializationID_TsQueryArray          SerializationID = 79
	SerializationID_TsVector              SerializationID = 80
	SerializationID_TsVectorArray         SerializationID = 81
	SerializationID_Uuid                  SerializationID = 82
	SerializationID_UuidArray             SerializationID = 83
	SerializationID_VarBit                SerializationID = 84
	SerializationID_VarBitArray           SerializationID = 85
	SerializationID_VarChar               SerializationID = 86
	SerializationID_VarCharArray          SerializationID = 87
	SerializationID_Xml                   SerializationID = 88
	SerializationID_XmlArray              SerializationID = 89
	SerializationID_Name                  SerializationID = 90
	SerializationID_NameArray             SerializationID = 91
)

// serializationIDToType is a map from each SerializationID to its matching DoltgresType.
var serializationIDToType = map[SerializationID]DoltgresType{}

// init sets the serialization and deserialization functions.
func init() {
	types.SetExtendedTypeSerializers(SerializeType, DeserializeType)
	for _, t := range typesFromBaseID {
		sID := t.GetSerializationID()
		if sID == SerializationID_Invalid {
			continue
		}
		if _, ok := serializationIDToType[sID]; ok {
			panic("duplicate serialization IDs in use")
		}
		serializationIDToType[sID] = t
	}
}

// SerializeType is able to serialize the given extended type into a byte slice. All extended types will be defined
// by DoltgreSQL.
func SerializeType(extendedType types.ExtendedType) ([]byte, error) {
	if doltgresType, ok := extendedType.(DoltgresType); ok {
		return doltgresType.SerializeType()
	}
	return nil, fmt.Errorf("unknown type to serialize")
}

// MustSerializeType internally calls SerializeType and panics on error. In general, panics should only occur when a
// type has not yet had its Serialization implemented yet.
func MustSerializeType(extendedType types.ExtendedType) []byte {
	// MustSerializeType is often used to efficiently compare any two types, so we'll make a special exception for types
	// that cannot be normally serialized. This is okay since these types cannot be deserialized, preventing them from
	// being used outside of comparisons.
	switch extendedType.(type) {
	case AnyArrayType:
		return []byte{0}
	case UnknownType:
		return []byte{1}
	}
	serializedType, err := SerializeType(extendedType)
	if err != nil {
		panic(err)
	}
	return serializedType
}

// DeserializeType is able to deserialize the given serialized type into an appropriate extended type. All extended
// types will be defined by DoltgreSQL.
func DeserializeType(serializedType []byte) (types.ExtendedType, error) {
	if len(serializedType) < serializationIDHeaderSize {
		return nil, fmt.Errorf("cannot deserialize an empty type")
	}
	serializationID, version := SerializationIDFromBytes(serializedType)
	targetType, ok := serializationIDToType[serializationID]
	if !ok {
		return nil, fmt.Errorf("serialization ID %d does not have a matching type for deserialization", serializationID)
	}
	return targetType.deserializeType(version, serializedType[serializationIDHeaderSize:])
}

// serializationIDHeaderSize is the size of the header that applies to all serialization IDs.
const serializationIDHeaderSize = 4

// ToByteSlice returns the ID as a byte slice.
func (id SerializationID) ToByteSlice(version uint16) []byte {
	b := make([]byte, serializationIDHeaderSize)
	binary.LittleEndian.PutUint16(b, uint16(id))
	binary.LittleEndian.PutUint16(b[2:], version)
	return b
}

// SerializationIDFromBytes reads a SerializationID and version from the given byte slice. The slice must have a length
// of at least 4 bytes. This function does not perform any validation, and is merely a convenience to ensure that the
// ID is read correctly.
func SerializationIDFromBytes(b []byte) (SerializationID, uint16) {
	return SerializationID(binary.LittleEndian.Uint16(b)), binary.LittleEndian.Uint16(b[2:])
}
