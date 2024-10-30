// Copyright 2024 Dolthub, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
// http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package types

import (
	"testing"
)

//// TestSerialization operates as a line of defense to prevent accidental changes to pre-existing serialization IDs.
//// If this test fails, then a SerializationID was changed that should not have been changed.
//func TestSerialization(t *testing.T) {
//	ids := []struct {
//		SerializationID
//		ID   uint16
//		Name string
//	}{
//		{SerializationID_Invalid, 0, "Invalid"},
//		{SerializationID_Bit, 1, "Bit"},
//		{SerializationID_BitArray, 2, "BitArray"},
//		{SerializationID_Bool, 3, "Bool"},
//		{SerializationID_BoolArray, 4, "BoolArray"},
//		{SerializationID_Box, 5, "Box"},
//		{SerializationID_BoxArray, 6, "BoxArray"},
//		{SerializationID_Bytea, 7, "Bytea"},
//		{SerializationID_ByteaArray, 8, "ByteaArray"},
//		{SerializationID_Char, 9, "Char"},
//		{SerializationID_CharArray, 10, "CharArray"},
//		{SerializationID_Cidr, 11, "Cidr"},
//		{SerializationID_CidrArray, 12, "CidrArray"},
//		{SerializationID_Circle, 13, "Circle"},
//		{SerializationID_CircleArray, 14, "CircleArray"},
//		{SerializationID_Date, 15, "Date"},
//		{SerializationID_DateArray, 16, "DateArray"},
//		{SerializationID_DateMultirange, 17, "DateMultirange"},
//		{SerializationID_DateRange, 18, "DateRange"},
//		{SerializationID_Enum, 19, "Enum"},
//		{SerializationID_EnumArray, 20, "EnumArray"},
//		{SerializationID_Float32, 21, "Float32"},
//		{SerializationID_Float32Array, 22, "Float32Array"},
//		{SerializationID_Float64, 23, "Float64"},
//		{SerializationID_Float64Array, 24, "Float64Array"},
//		{SerializationID_Inet, 25, "Inet"},
//		{SerializationID_InetArray, 26, "InetArray"},
//		{SerializationID_Int16, 27, "Int16"},
//		{SerializationID_Int16Array, 28, "Int16Array"},
//		{SerializationID_Int32, 29, "Int32"},
//		{SerializationID_Int32Array, 30, "Int32Array"},
//		{SerializationID_Int32Multirange, 31, "Int32Multirange"},
//		{SerializationID_Int32Range, 32, "Int32Range"},
//		{SerializationID_Int64, 33, "Int64"},
//		{SerializationID_Int64Array, 34, "Int64Array"},
//		{SerializationID_Int64Multirange, 35, "Int64Multirange"},
//		{SerializationID_Int64Range, 36, "Int64Range"},
//		{SerializationID_Interval, 37, "Interval"},
//		{SerializationID_IntervalArray, 38, "IntervalArray"},
//		{SerializationID_Json, 39, "Json"},
//		{SerializationID_JsonArray, 40, "JsonArray"},
//		{SerializationID_JsonB, 41, "JsonB"},
//		{SerializationID_JsonBArray, 42, "JsonBArray"},
//		{SerializationID_Line, 43, "Line"},
//		{SerializationID_LineArray, 44, "LineArray"},
//		{SerializationID_LineSegment, 45, "LineSegment"},
//		{SerializationID_LineSegmentArray, 46, "LineSegmentArray"},
//		{SerializationID_MacAddress, 47, "MacAddress"},
//		{SerializationID_MacAddress8, 48, "MacAddress8"},
//		{SerializationID_MacAddress8Array, 49, "MacAddress8Array"},
//		{SerializationID_MacAddressArray, 50, "MacAddressArray"},
//		{SerializationID_Money, 51, "Money"},
//		{SerializationID_MoneyArray, 52, "MoneyArray"},
//		{SerializationID_Null, 53, "Null"},
//		{SerializationID_Numeric, 54, "Numeric"},
//		{SerializationID_NumericArray, 55, "NumericArray"},
//		{SerializationID_NumericMultirange, 56, "NumericMultirange"},
//		{SerializationID_NumericRange, 57, "NumericRange"},
//		{SerializationID_Path, 58, "Path"},
//		{SerializationID_PathArray, 59, "PathArray"},
//		{SerializationID_Point, 60, "Point"},
//		{SerializationID_PointArray, 61, "PointArray"},
//		{SerializationID_Polygon, 62, "Polygon"},
//		{SerializationID_PolygonArray, 63, "PolygonArray"},
//		{SerializationID_Text, 64, "Text"},
//		{SerializationID_TextArray, 65, "TextArray"},
//		{SerializationID_Time, 66, "Time"},
//		{SerializationID_TimeArray, 67, "TimeArray"},
//		{SerializationID_TimeTZ, 68, "TimeTZ"},
//		{SerializationID_TimeTZArray, 69, "TimeTZArray"},
//		{SerializationID_Timestamp, 70, "Timestamp"},
//		{SerializationID_TimestampArray, 71, "TimestampArray"},
//		{SerializationID_TimestampMultirange, 72, "TimestampMultirange"},
//		{SerializationID_TimestampRange, 73, "TimestampRange"},
//		{SerializationID_TimestampTZ, 74, "TimestampTZ"},
//		{SerializationID_TimestampTZArray, 75, "TimestampTZArray"},
//		{SerializationID_TimestampTZMultirange, 76, "TimestampTZMultirange"},
//		{SerializationID_TimestampTZRange, 77, "TimestampTZRange"},
//		{SerializationID_TsQuery, 78, "TsQuery"},
//		{SerializationID_TsQueryArray, 79, "TsQueryArray"},
//		{SerializationID_TsVector, 80, "TsVector"},
//		{SerializationID_TsVectorArray, 81, "TsVectorArray"},
//		{SerializationID_Uuid, 82, "Uuid"},
//		{SerializationID_UuidArray, 83, "UuidArray"},
//		{SerializationID_VarBit, 84, "VarBit"},
//		{SerializationID_VarBitArray, 85, "VarBitArray"},
//		{SerializationID_VarChar, 86, "VarChar"},
//		{SerializationID_VarCharArray, 87, "VarCharArray"},
//		{SerializationID_Xml, 88, "Xml"},
//		{SerializationID_XmlArray, 89, "XmlArray"},
//		{SerializationID_Name, 90, "Name"},
//		{SerializationID_NameArray, 91, "NameArray"},
//		{SerializationID_Oid, 92, "OID"},
//		{SerializationID_OidArray, 93, "OidArray"},
//		{SerializationID_Xid, 94, "Xid"},
//		{SerializationID_XidArray, 95, "XidArray"},
//		{SerializationID_InternalChar, 96, "InternalChar"},
//		{SerializationID_InternalCharArray, 97, "InternalCharArray"},
//		{SerializationId_Domain, 98, "Domain"},
//	}
//	allIds := make(map[uint16]string)
//	for _, id := range ids {
//		if uint16(id.SerializationID) != id.ID {
//			t.Logf("Serialization ID `%s` has been changed from its permanent value of `%d` to `%d`",
//				id.Name, id.ID, uint16(id.SerializationID))
//			t.Fail()
//		} else if existingName, ok := allIds[id.ID]; ok {
//			t.Logf("Serialization ID `%s` has the same value as `%s`: `%d`",
//				id.Name, existingName, id.ID)
//			t.Fail()
//		} else {
//			allIds[id.ID] = id.Name
//		}
//	}
//}
//
//// TestSerializationIDConsistency checks that all types use the same SerializationID that they report in
//// GetSerializationID and output in SerializeType.
//func TestSerializationIDConsistency(t *testing.T) {
//	for _, typ := range typesFromBaseID {
//		t.Run(typ.String(), func(t *testing.T) {
//			sID := typ.GetSerializationID()
//			if sID == SerializationID_Invalid {
//				_, err := typ.SerializeType()
//				require.Error(t, err)
//			} else {
//				serializedType, err := typ.SerializeType()
//				require.NoError(t, err)
//				require.True(t, len(serializedType) >= serializationIDHeaderSize)
//				idPrefix := sID.ToByteSlice(0)[:2]
//				require.Equal(t, idPrefix, serializedType[:2])
//			}
//		})
//	}
//}

// TestJsonValueType operates as a line of defense to prevent accidental changes to JSON type values. If this test
// fails, then a JsonValueType was changed that should not have been changed.
func TestJsonValueType(t *testing.T) {
	types := []struct {
		JsonValueType
		Value byte
		Name  string
	}{
		{JsonValueType_Object, 0, "Object"},
		{JsonValueType_Array, 1, "Array"},
		{JsonValueType_String, 2, "String"},
		{JsonValueType_Number, 3, "Number"},
		{JsonValueType_Boolean, 4, "Boolean"},
		{JsonValueType_Null, 5, "Null"},
	}
	allValues := make(map[byte]string)
	for _, typ := range types {
		if byte(typ.JsonValueType) != typ.Value {
			t.Logf("JSON value type `%s` has been changed from its permanent value of `%d` to `%d`",
				typ.Name, typ.Value, byte(typ.JsonValueType))
			t.Fail()
		} else if existingName, ok := allValues[typ.Value]; ok {
			t.Logf("JSON value type `%s` has the same value as `%s`: `%d`",
				typ.Name, existingName, typ.Value)
			t.Fail()
		} else {
			allValues[typ.Value] = typ.Name
		}
	}
}
