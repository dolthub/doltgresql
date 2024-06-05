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

// DoltgresTypeBaseID is an ID that is common between all variations of a DoltgresType. For example, VARCHAR(3) and
// VARCHAR(6) are different types, however they will return the same DoltgresTypeBaseID. This ID is not suitable for
// serialization, as it may change over time. Many types use their SerializationID as their base ID, so for types that
// are not serializable (such as the "any" types), it is recommended that they start way after the largest
// SerializationID to prevent base ID conflicts.
type DoltgresTypeBaseID uint32

const (
	DoltgresTypeBaseID_Any DoltgresTypeBaseID = iota + 2147483648
	DoltgresTypeBaseID_AnyElement
	DoltgresTypeBaseID_AnyArray
	DoltgresTypeBaseID_AnyNonArray
	DoltgresTypeBaseID_AnyEnum
	DoltgresTypeBaseID_AnyRange
	DoltgresTypeBaseID_AnyMultirange
	DoltgresTypeBaseID_AnyCompatible
	DoltgresTypeBaseID_AnyCompatibleArray
	DoltgresTypeBaseID_AnyCompatibleNonArray
	DoltgresTypeBaseID_AnyCompatibleRange
	DoltgresTypeBaseID_AnyCompatibleMultirange
	DoltgresTypeBaseID_CString
	DoltgresTypeBaseID_Internal
	DoltgresTypeBaseID_Language_Handler
	DoltgresTypeBaseID_FDW_Handler
	DoltgresTypeBaseID_Table_AM_Handler
	DoltgresTypeBaseID_Index_AM_Handler
	DoltgresTypeBaseID_TSM_Handler
	DoltgresTypeBaseID_Record
	DoltgresTypeBaseID_Trigger
	DoltgresTypeBaseID_Event_Trigger
	DoltgresTypeBaseID_PG_DDL_Command
	DoltgresTypeBaseID_Void
	DoltgresTypeBaseID_Unknown
	DoltgresTypeBaseID_Int16Serial
	DoltgresTypeBaseID_Int32Serial
	DoltgresTypeBaseID_Int64Serial
)

const (
	DoltgresTypeBaseID_Bool        = DoltgresTypeBaseID(SerializationID_Bool)
	DoltgresTypeBaseID_Bytea       = DoltgresTypeBaseID(SerializationID_Bytea)
	DoltgresTypeBaseID_Char        = DoltgresTypeBaseID(SerializationID_Char)
	DoltgresTypeBaseID_Date        = DoltgresTypeBaseID(SerializationID_Date)
	DoltgresTypeBaseID_Float32     = DoltgresTypeBaseID(SerializationID_Float32)
	DoltgresTypeBaseID_Float64     = DoltgresTypeBaseID(SerializationID_Float64)
	DoltgresTypeBaseID_Int16       = DoltgresTypeBaseID(SerializationID_Int16)
	DoltgresTypeBaseID_Int32       = DoltgresTypeBaseID(SerializationID_Int32)
	DoltgresTypeBaseID_Int64       = DoltgresTypeBaseID(SerializationID_Int64)
	DoltgresTypeBaseID_Json        = DoltgresTypeBaseID(SerializationID_Json)
	DoltgresTypeBaseID_JsonB       = DoltgresTypeBaseID(SerializationID_JsonB)
	DoltgresTypeBaseID_Name        = DoltgresTypeBaseID(SerializationID_Name)
	DoltgresTypeBaseID_Null        = DoltgresTypeBaseID(SerializationID_Null)
	DoltgresTypeBaseID_Numeric     = DoltgresTypeBaseID(SerializationID_Numeric)
	DoltgresTypeBaseID_Oid         = DoltgresTypeBaseID(SerializationID_Oid)
	DoltgresTypeBaseID_Text        = DoltgresTypeBaseID(SerializationID_Text)
	DoltgresTypeBaseID_Time        = DoltgresTypeBaseID(SerializationID_Time)
	DoltgresTypeBaseID_Timestamp   = DoltgresTypeBaseID(SerializationID_Timestamp)
	DoltgresTypeBaseID_TimestampTZ = DoltgresTypeBaseID(SerializationID_TimestampTZ)
	DoltgresTypeBaseID_TimeTZ      = DoltgresTypeBaseID(SerializationID_TimeTZ)
	DoltgresTypeBaseID_Uuid        = DoltgresTypeBaseID(SerializationID_Uuid)
	DoltgresTypeBaseID_VarChar     = DoltgresTypeBaseID(SerializationID_VarChar)
	DoltgresTypeBaseID_Xid         = DoltgresTypeBaseID(SerializationID_Xid)
)

// TypeCategory represents the type category that a type belongs to. These are used by Postgres to group similar types
// for parameter resolution, operator resolution, etc.
type TypeCategory uint8

const (
	TypeCategory_Unknown TypeCategory = iota
	TypeCategory_ArrayTypes
	TypeCategory_BooleanTypes
	TypeCategory_CompositeTypes
	TypeCategory_DateTimeTypes
	TypeCategory_EnumTypes
	TypeCategory_GeometricTypes
	TypeCategory_NetworkAddressTypes
	TypeCategory_NumericTypes
	TypeCategory_PseudoTypes
	TypeCategory_RangeTypes
	TypeCategory_StringTypes
	TypeCategory_TimespanTypes
	TypeCategory_UserDefinedTypes
	TypeCategory_BitStringTypes
	TypeCategory_XMLTypes
)

// baseIDArrayTypes contains a map of all base IDs that represent array variants.
var baseIDArrayTypes = map[DoltgresTypeBaseID]DoltgresArrayType{}

// baseIDCategories contains a map from all base IDs to their respective categories
// TODO: add all of the types to each category
var baseIDCategories = map[DoltgresTypeBaseID]TypeCategory{
	Bool.BaseID():    TypeCategory_BooleanTypes,
	BpChar.BaseID():  TypeCategory_StringTypes,
	Float32.BaseID(): TypeCategory_NumericTypes,
	Float64.BaseID(): TypeCategory_NumericTypes,
	Int16.BaseID():   TypeCategory_NumericTypes,
	Int32.BaseID():   TypeCategory_NumericTypes,
	Int64.BaseID():   TypeCategory_NumericTypes,
	Name.BaseID():    TypeCategory_StringTypes,
	Numeric.BaseID(): TypeCategory_NumericTypes,
	Oid.BaseID():     TypeCategory_NumericTypes,
	Text.BaseID():    TypeCategory_StringTypes,
	VarChar.BaseID(): TypeCategory_StringTypes,
}

// preferredTypeInCategory contains a map from each type category to that category's preferred type.
// TODO: add all of the preferred types
var preferredTypeInCategory = map[TypeCategory]DoltgresTypeBaseID{
	TypeCategory_BooleanTypes: Bool.BaseID(),
	TypeCategory_NumericTypes: Float64.BaseID(),
	TypeCategory_StringTypes:  Text.BaseID(),
}

// InitBaseIDs reads the list of all types and creates a mapping of the base ID for each array variant.
func InitBaseIDs() {
	for _, t := range typesFromBaseID {
		if dat, ok := t.(DoltgresArrayType); ok {
			baseIDArrayTypes[t.BaseID()] = dat
		}
	}
}

// IsBaseIDArrayType returns whether the base ID is an array type. If it is, it also returns the type.
func (id DoltgresTypeBaseID) IsBaseIDArrayType() (DoltgresArrayType, bool) {
	dat, ok := baseIDArrayTypes[id]
	return dat, ok
}

// GetTypeCategory returns the TypeCategory that this base ID belongs to. Returns Unknown if the ID does not belong to a
// category.
func (id DoltgresTypeBaseID) GetTypeCategory() TypeCategory {
	if tc, ok := baseIDCategories[id]; ok {
		return tc
	}
	return TypeCategory_Unknown
}

// GetRepresentativeType returns the representative type of the base ID. This is usually the unbounded version or
// equivalent.
func (id DoltgresTypeBaseID) GetRepresentativeType() DoltgresType {
	if t, ok := typesFromBaseID[id]; ok {
		return t
	}
	return Unknown
}

// GetPreferredType returns the preferred type for this TypeCategory. Returns Unknown if the category does not have a
// preferred type.
func (cat TypeCategory) GetPreferredType() DoltgresTypeBaseID {
	if id, ok := preferredTypeInCategory[cat]; ok {
		return id
	}
	return DoltgresTypeBaseID_Unknown
}
