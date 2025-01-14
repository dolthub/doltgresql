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

package id

// Section represents a specific space that an Internal ID resides in. This makes it relatively simple to find the
// target of the ID, since each searchable space has its own section.
type Section uint8

// All new sections must be given an unused number, as these may be persisted in tables. Changing them would change
// pre-existing table data, potentially corrupting that table data. At most, there can be 127 sections, since the first
// bit is reserved to determine a format's encoding.
const (
	Section_Null                 Section = 0  // Represents a null ID
	Section_AccessMethod         Section = 1  // Refers to relation access methods
	Section_Cast                 Section = 2  // Refers to casts between types
	Section_Check                Section = 3  // Refers to checks on tables
	Section_Collation            Section = 4  // Refers to collations
	Section_ColumnDefault        Section = 5  // Refers to column defaults on tables
	Section_Database             Section = 6  // Refers to the database
	Section_EnumLabel            Section = 7  // Refers to a specific label in an ENUM type
	Section_EventTrigger         Section = 8  // Refers to event triggers
	Section_ExclusionConstraint  Section = 9  // Refers to exclusion constraints
	Section_Extension            Section = 10 // Refers to extensions
	Section_ForeignKey           Section = 11 // Refers to foreign keys on tables
	Section_ForeignDataWrapper   Section = 12 // Refers to foreign data wrappers
	Section_ForeignServer        Section = 13 // Refers to foreign servers
	Section_ForeignTable         Section = 14 // Refers to foreign tables
	Section_Function             Section = 15 // Refers to functions
	Section_FunctionLanguage     Section = 16 // Refers to the programming languages available for writing functions
	Section_Index                Section = 17 // Refers to indexes on tables
	Section_Namespace            Section = 18 // Namespaces are the underlying structure of a schema (basically the schema)
	Section_OID                  Section = 19 // Refers to a raw OID that is not actually attached to anything (ONLY used with reg types)
	Section_Operator             Section = 20 // Refers to operators (+, -, *, etc.)
	Section_OperatorClass        Section = 21 // Refers to operator classes
	Section_OperatorFamily       Section = 22 // Refers to operator families
	Section_PrimaryKey           Section = 23 // Refers to primary keys on tables
	Section_Procedure            Section = 24 // Refers to stored procedures
	Section_Publication          Section = 25 // Refers to publications
	Section_RowLevelSecurity     Section = 26 // Refers to row-level security polices on tables
	Section_Sequence             Section = 27 // Refers to sequences
	Section_Subscription         Section = 28 // Refers to logical replication subscriptions
	Section_Table                Section = 29 // Refers to tables
	Section_TextSearchConfig     Section = 30 // Refers to text search configuration
	Section_TextSearchDictionary Section = 31 // Refers to text search dictionaries
	Section_TextSearchParser     Section = 32 // Refers to text search parsers
	Section_TextSearchTemplate   Section = 33 // Refers to text search templates
	Section_Trigger              Section = 34 // Refers to triggers on tables and views
	Section_Type                 Section = 35 // Refers to types
	Section_UniqueKey            Section = 36 // Refers to unique keys on tables
	Section_User                 Section = 37 // Refers to users
	Section_View                 Section = 38 // Refers to views

	section_count uint8 = 39 // This is the number of sections, and should ALWAYS be kept up-to-date
)

// String returns the name of the Section.
func (section Section) String() string {
	switch section {
	case Section_Null:
		return "Null"
	case Section_AccessMethod:
		return "AccessMethod"
	case Section_Cast:
		return "Cast"
	case Section_Check:
		return "Check"
	case Section_Collation:
		return "Collation"
	case Section_ColumnDefault:
		return "ColumnDefault"
	case Section_Database:
		return "Database"
	case Section_EnumLabel:
		return "EnumLabel"
	case Section_EventTrigger:
		return "EventTrigger"
	case Section_ExclusionConstraint:
		return "ExclusionConstraint"
	case Section_Extension:
		return "Extension"
	case Section_ForeignKey:
		return "ForeignKey"
	case Section_ForeignDataWrapper:
		return "ForeignDataWrapper"
	case Section_ForeignServer:
		return "ForeignServer"
	case Section_ForeignTable:
		return "ForeignTable"
	case Section_Function:
		return "Function"
	case Section_FunctionLanguage:
		return "FunctionLanguage"
	case Section_Index:
		return "Index"
	case Section_Namespace:
		return "Namespace"
	case Section_OID:
		return "OID"
	case Section_Operator:
		return "Operator"
	case Section_OperatorClass:
		return "OperatorClass:"
	case Section_OperatorFamily:
		return "OperatorFamily"
	case Section_PrimaryKey:
		return "PrimaryKey"
	case Section_Procedure:
		return "Procedure"
	case Section_Publication:
		return "Publication"
	case Section_RowLevelSecurity:
		return "RowLevelSecurity"
	case Section_Sequence:
		return "Sequence"
	case Section_Subscription:
		return "Subscription"
	case Section_Table:
		return "Table"
	case Section_TextSearchConfig:
		return "TextSearchConfig"
	case Section_TextSearchDictionary:
		return "TextSearchDictionary"
	case Section_TextSearchParser:
		return "TextSearchParser"
	case Section_TextSearchTemplate:
		return "TextSearchTemplate"
	case Section_Trigger:
		return "Trigger"
	case Section_Type:
		return "Type"
	case Section_UniqueKey:
		return "UniqueKey"
	case Section_User:
		return "User"
	case Section_View:
		return "View"
	default:
		return "UNKNOWN_SECTION"
	}
}
