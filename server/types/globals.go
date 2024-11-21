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

import "fmt"

// DoltgresTypeBaseID is an ID that is common between all variations of a DoltgresType. For example, VARCHAR(3) and
// VARCHAR(6) are different types, however they will return the same DoltgresTypeBaseID. This ID is not suitable for
// serialization, as it may change over time. Many types use their SerializationID as their base ID, so for types that
// are not serializable (such as the "any" types), it is recommended that they start way after the largest
// SerializationID to prevent base ID conflicts.
type DoltgresTypeBaseID uint32

//go:generate go run golang.org/x/tools/cmd/stringer -type=DoltgresTypeBaseID

const (
	DoltgresTypeBaseID_Any DoltgresTypeBaseID = iota + 8192
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
	DoltgresTypeBaseID_Regclass
	DoltgresTypeBaseID_Regcollation
	DoltgresTypeBaseID_Regconfig
	DoltgresTypeBaseID_Regdictionary
	DoltgresTypeBaseID_Regnamespace
	DoltgresTypeBaseID_Regoper
	DoltgresTypeBaseID_Regoperator
	DoltgresTypeBaseID_Regproc
	DoltgresTypeBaseID_Regprocedure
	DoltgresTypeBaseID_Regrole
	DoltgresTypeBaseID_Regtype
)

const (
	DoltgresTypeBaseID_Bool         = DoltgresTypeBaseID(SerializationID_Bool)
	DoltgresTypeBaseID_Bytea        = DoltgresTypeBaseID(SerializationID_Bytea)
	DoltgresTypeBaseID_Char         = DoltgresTypeBaseID(SerializationID_Char)
	DoltgresTypeBaseID_Date         = DoltgresTypeBaseID(SerializationID_Date)
	DoltgresTypeBaseID_Float32      = DoltgresTypeBaseID(SerializationID_Float32)
	DoltgresTypeBaseID_Float64      = DoltgresTypeBaseID(SerializationID_Float64)
	DoltgresTypeBaseID_Int16        = DoltgresTypeBaseID(SerializationID_Int16)
	DoltgresTypeBaseID_Int32        = DoltgresTypeBaseID(SerializationID_Int32)
	DoltgresTypeBaseID_Int64        = DoltgresTypeBaseID(SerializationID_Int64)
	DoltgresTypeBaseID_InternalChar = DoltgresTypeBaseID(SerializationID_InternalChar)
	DoltgresTypeBaseID_Interval     = DoltgresTypeBaseID(SerializationID_Interval)
	DoltgresTypeBaseID_Json         = DoltgresTypeBaseID(SerializationID_Json)
	DoltgresTypeBaseID_JsonB        = DoltgresTypeBaseID(SerializationID_JsonB)
	DoltgresTypeBaseID_Name         = DoltgresTypeBaseID(SerializationID_Name)
	DoltgresTypeBaseID_Null         = DoltgresTypeBaseID(SerializationID_Null)
	DoltgresTypeBaseID_Numeric      = DoltgresTypeBaseID(SerializationID_Numeric)
	DoltgresTypeBaseID_Oid          = DoltgresTypeBaseID(SerializationID_Oid)
	DoltgresTypeBaseID_Text         = DoltgresTypeBaseID(SerializationID_Text)
	DoltgresTypeBaseID_Time         = DoltgresTypeBaseID(SerializationID_Time)
	DoltgresTypeBaseID_Timestamp    = DoltgresTypeBaseID(SerializationID_Timestamp)
	DoltgresTypeBaseID_TimestampTZ  = DoltgresTypeBaseID(SerializationID_TimestampTZ)
	DoltgresTypeBaseID_TimeTZ       = DoltgresTypeBaseID(SerializationID_TimeTZ)
	DoltgresTypeBaseID_Uuid         = DoltgresTypeBaseID(SerializationID_Uuid)
	DoltgresTypeBaseID_VarChar      = DoltgresTypeBaseID(SerializationID_VarChar)
	DoltgresTypeBaseID_Xid          = DoltgresTypeBaseID(SerializationID_Xid)
	DoltgresTypeBaseId_Domain       = DoltgresTypeBaseID(SerializationId_Domain)
)

// TypeAlignment represents the alignment required when storing a value of this type.
type TypeAlignment string

const (
	TypeAlignment_Char   TypeAlignment = "c"
	TypeAlignment_Short  TypeAlignment = "s"
	TypeAlignment_Int    TypeAlignment = "i"
	TypeAlignment_Double TypeAlignment = "d"
)

// TypeCategory represents the type category that a type belongs to. These are used by Postgres to group similar types
// for parameter resolution, operator resolution, etc.
type TypeCategory string

const (
	TypeCategory_ArrayTypes          TypeCategory = "A"
	TypeCategory_BooleanTypes        TypeCategory = "B"
	TypeCategory_CompositeTypes      TypeCategory = "C"
	TypeCategory_DateTimeTypes       TypeCategory = "D"
	TypeCategory_EnumTypes           TypeCategory = "E"
	TypeCategory_GeometricTypes      TypeCategory = "G"
	TypeCategory_NetworkAddressTypes TypeCategory = "I"
	TypeCategory_NumericTypes        TypeCategory = "N"
	TypeCategory_PseudoTypes         TypeCategory = "P"
	TypeCategory_RangeTypes          TypeCategory = "R"
	TypeCategory_StringTypes         TypeCategory = "S"
	TypeCategory_TimespanTypes       TypeCategory = "T"
	TypeCategory_UserDefinedTypes    TypeCategory = "U"
	TypeCategory_BitStringTypes      TypeCategory = "V"
	TypeCategory_UnknownTypes        TypeCategory = "X"
	TypeCategory_InternalUseTypes    TypeCategory = "Z"
)

// TypeStorage represents the storage strategy for storing `varlena` (typlen = -1) types.
type TypeStorage string

const (
	TypeStorage_Plain    TypeStorage = "p"
	TypeStorage_External TypeStorage = "e"
	TypeStorage_Main     TypeStorage = "m"
	TypeStorage_Extended TypeStorage = "x"
)

// TypeType represents the type of types that can be created/used.
// This includes 'base', 'composite', 'domain', 'enum', 'shell', 'range' and  'multirange'.
type TypeType string

const (
	TypeType_Base       TypeType = "b"
	TypeType_Composite  TypeType = "c"
	TypeType_Domain     TypeType = "d"
	TypeType_Enum       TypeType = "e"
	TypeType_Pseudo     TypeType = "p"
	TypeType_Range      TypeType = "r"
	TypeType_MultiRange TypeType = "m"
)

// baseIDArrayTypes contains a map of all base IDs that represent array variants.
var baseIDArrayTypes = map[DoltgresTypeBaseID]DoltgresArrayType{}

// baseIDCategories contains a map from all base IDs to their respective categories
// TODO: add all of the types to each category
var baseIDCategories = map[DoltgresTypeBaseID]TypeCategory{}

// preferredTypeInCategory contains a map from each type category to that category's preferred type.
// TODO: add all of the preferred types
var preferredTypeInCategory = map[TypeCategory][]DoltgresTypeBaseID{}

// oidToType holds a reference from a given OID to its type.
var oidToType = map[uint32]DoltgresType{}

// Init reads the list of all types and creates mappings that will be used by various functions.
func Init() {
	for baseID, t := range typesFromBaseID {
		if dat, ok := t.(DoltgresArrayType); ok {
			baseIDArrayTypes[t.BaseID()] = dat
		}
		if t.IsPreferredType() {
			preferredTypeInCategory[t.Category()] = append(preferredTypeInCategory[t.Category()], t.BaseID())
		}
		// Add the types to the OID map
		if baseID.HasUniqueOID() {
			if existingType, ok := oidToType[t.OID()]; ok {
				panic(fmt.Errorf("OID (%d) type conflict: `%s` and `%s`", t.OID(), existingType.String(), t.String()))
			}
			oidToType[t.OID()] = t
			baseIDCategories[t.BaseID()] = t.Category()
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
	return TypeCategory_UnknownTypes
}

// GetRepresentativeType returns the representative type of the base ID. This is usually the unbounded version or
// equivalent.
func (id DoltgresTypeBaseID) GetRepresentativeType() DoltgresType {
	if t, ok := typesFromBaseID[id]; ok {
		return t
	}
	return Unknown
}

// HasUniqueOID returns whether the type belonging to the base ID has a unique OID. This will be true for most types.
// Examples of types that do not have unique OIDs are the serial types, since they're not actual types.
func (id DoltgresTypeBaseID) HasUniqueOID() bool {
	switch id {
	case DoltgresTypeBaseID_Null,
		DoltgresTypeBaseID_Int16Serial,
		DoltgresTypeBaseID_Int32Serial,
		DoltgresTypeBaseID_Int64Serial:
		return false
	default:
		return true
	}
}

// IsPreferredType returns whether the type passed is a preferred type for this TypeCategory.
func (cat TypeCategory) IsPreferredType(p DoltgresTypeBaseID) bool {
	if pts, ok := preferredTypeInCategory[cat]; ok {
		for _, pt := range pts {
			if pt == p {
				return true
			}
		}
	}
	return false
}

// GetTypeByOID returns the DoltgresType matching the given OID. If the OID does not match a type, then nil is returned.
func GetTypeByOID(oid uint32) DoltgresType {
	t, ok := oidToType[oid]
	if !ok {
		return nil
	}
	return t
}
