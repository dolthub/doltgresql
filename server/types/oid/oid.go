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

package oid

// Section represents a specific space that an OID subset resides in. This makes it relatively simple to find the target
// of an OID, since each searchable space has its own section.
type Section uint8

const (
	Section_BuiltIn       Section = iota // Contains all of the OIDs that are defined when creating a Postgres database
	Section_Check                        // Refers to checks on tables (the dataIndex is obtained by incrementing through all tables' checks)
	Section_Collation                    // Refers to collations
	Section_Database                     // Refers to the database (schemaIndex does not apply, so only dataIndex should be used)
	Section_ForeignKey                   // Refers to foreign keys on tables (the dataIndex is obtained by incrementing through all tables' foreign keys)
	Section_Function                     // Refers to functions only (no stored procedures, etc.)
	Section_Index                        // Refers to indexes on tables (the dataIndex is obtained by incrementing through all tables' indexes)
	Section_Namespace                    // Namespaces are the underlying structure of a schema, so these only use the dataIndex
	Section_Operator                     // Refers to operators (+, -, *, etc.)
	Section_Sequence                     // Refers to sequences
	Section_Table                        // Refers to tables
	Section_View                         // Refers to views
	Section_ColumnDefault                // Refers to column defaults on tables (the dataIndex is obtained by incrementing through all tables' column defaults)
	Section_Invalid                      // Represents an invalid OID
)

const (
	dataBitCount    = 20 // The size, in bits, of the dataIndex. 2^20 allows for 1048576 elements.
	schemaBitCount  = 8  // The size, in bits, of the schemaIndex. 2^8 allows for 256 schemas.
	sectionBitCount = 4  // The size, in bits, of the section. 2^4 allows for 16 sections.
)

const (
	dataMask   = uint32(1<<dataBitCount) - 1                       // Contains the mask for the dataIndex portion of an OID.
	schemaMask = (uint32(1<<(schemaBitCount)) - 1) << dataBitCount // Contains the mask for the schemaIndex portion of an OID.
)

// CreateOID returns an OID with the given contents. If any index is larger than the allotted bit size, then this
// returns an OID with the section set to Section_Invalid.
//
// The layout of an OID is as follows:
//
// [Section][SchemaIndex][DataIndex]
//
// The schemaIndex represents the schema position that the OID is referencing (for all sections except for
// Section_Namespace) when sorting all schemas in ascending order by name. The dataIndex represents the index of the
// data according to the Section given (which will use some deterministic algorithm for the index).
func CreateOID(section Section, schemaIndex int, dataIndex int) uint32 {
	if schemaIndex < 0 || dataIndex < 0 || schemaIndex >= (1<<schemaBitCount) || dataIndex >= (1<<dataBitCount) {
		return uint32(Section_Invalid) << (dataBitCount + schemaBitCount)
	}
	if (section == Section_Namespace || section == Section_Database) && schemaIndex != 0 {
		// Go doesn't allow for compile-time checks, so we'll panic here.
		// This shouldn't happen at run-time, as this should be caught by tests.
		panic("The Database and Namespace sections should only use the dataIndex, not the schemaIndex")
	}
	return (uint32(section) << (dataBitCount + schemaBitCount)) +
		(uint32(schemaIndex) << (dataBitCount)) +
		uint32(dataIndex)
}

// ParseOID deconstructs an OID that was previously created using CreateOID.
func ParseOID(oid uint32) (section Section, schemaIndex int, dataIndex int) {
	section = Section(oid >> (dataBitCount + schemaBitCount))
	if section >= Section_Invalid {
		return Section_Invalid, 0, 0
	}
	return section, int((oid & schemaMask) >> dataBitCount), int(oid & dataMask)
}

// init validates the bit field lengths for the OID
func init() {
	if dataBitCount+schemaBitCount+sectionBitCount != 32 {
		panic("OID bit counts do not cover the OID range")
	}
	if Section_Invalid >= (1 << sectionBitCount) {
		panic("There are more sections than the bit allocation allows")
	}
}
