// Copyright 2025 Dolthub, Inc.
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

import "strconv"

// InternalAccessMethod is an Internal wrapper for access methods. This wrapper must not be returned to the client.
type InternalAccessMethod Internal

// InternalCheck is an Internal wrapper for checks. This wrapper must not be returned to the client.
type InternalCheck Internal

// InternalCollation is an Internal wrapper for collations. This wrapper must not be returned to the client.
type InternalCollation Internal

// InternalColumnDefault is an Internal wrapper for column defaults. This wrapper must not be returned to the client.
type InternalColumnDefault Internal

// InternalDatabase is an Internal wrapper for databases. This wrapper must not be returned to the client.
type InternalDatabase Internal

// InternalEnumLabel is an Internal wrapper for enum labels. This wrapper must not be returned to the client.
type InternalEnumLabel Internal

// InternalForeignKey is an Internal wrapper for foreign keys. This wrapper must not be returned to the client.
type InternalForeignKey Internal

// InternalFunction is an Internal wrapper for functions. This wrapper must not be returned to the client.
type InternalFunction Internal

// InternalIndex is an Internal wrapper for indexes. This wrapper must not be returned to the client.
type InternalIndex Internal

// InternalNamespace is an Internal wrapper for schemas/namespaces. This wrapper must not be returned to the client.
type InternalNamespace Internal

// InternalOID is an Internal wrapper for OIDs. This wrapper must not be returned to the client.
type InternalOID Internal

// InternalSequence is an Internal wrapper for sequences. This wrapper must not be returned to the client.
type InternalSequence Internal

// InternalTable is an Internal wrapper for tables. This wrapper must not be returned to the client.
type InternalTable Internal

// InternalType is an Internal wrapper for types. This wrapper must not be returned to the client.
type InternalType Internal

// InternalView is an Internal wrapper for views. This wrapper must not be returned to the client.
type InternalView Internal

// NewInternalAccessMethod returns a new InternalAccessMethod. This wrapper must not be returned to the client.
func NewInternalAccessMethod(methodName string) InternalAccessMethod {
	if len(methodName) == 0 {
		return NullAccessMethod
	}
	return InternalAccessMethod(NewInternal(Section_AccessMethod, methodName))
}

// NewInternalCheck returns a new InternalCheck. This wrapper must not be returned to the client.
func NewInternalCheck(schemaName string, tableName string, checkName string) InternalCheck {
	if len(schemaName) == 0 && len(tableName) == 0 && len(checkName) == 0 {
		return NullCheck
	}
	return InternalCheck(NewInternal(Section_Check, schemaName, tableName, checkName))
}

// NewInternalCollation returns a new InternalCollation. This wrapper must not be returned to the client.
func NewInternalCollation(schemaName string, collationName string) InternalCollation {
	if len(schemaName) == 0 && len(collationName) == 0 {
		return NullCollation
	}
	return InternalCollation(NewInternal(Section_Collation, schemaName, collationName))
}

// NewInternalColumnDefault returns a new InternalColumnDefault. This wrapper must not be returned to the client.
func NewInternalColumnDefault(schemaName string, tableName string, columnName string) InternalColumnDefault {
	if len(schemaName) == 0 && len(tableName) == 0 && len(columnName) == 0 {
		return NullColumnDefault
	}
	return InternalColumnDefault(NewInternal(Section_ColumnDefault, schemaName, tableName, columnName))
}

// NewInternalDatabase returns a new InternalDatabase. This wrapper must not be returned to the client.
func NewInternalDatabase(dbName string) InternalDatabase {
	if len(dbName) == 0 {
		return NullDatabase
	}
	return InternalDatabase(NewInternal(Section_Database, dbName))
}

// NewInternalEnumLabel returns a new InternalEnumLabel. This wrapper must not be returned to the client.
func NewInternalEnumLabel(parent InternalType, label string) InternalEnumLabel {
	if len(parent) == 0 && len(label) == 0 {
		return NullEnumLabel
	}
	return InternalEnumLabel(NewInternal(Section_EnumLabel, string(parent), label))
}

// NewInternalForeignKey returns a new InternalForeignKey. This wrapper must not be returned to the client.
func NewInternalForeignKey(schemaName string, tableName string, fkName string) InternalForeignKey {
	if len(schemaName) == 0 && len(tableName) == 0 && len(fkName) == 0 {
		return NullForeignKey
	}
	return InternalForeignKey(NewInternal(Section_ForeignKey, schemaName, tableName, fkName))
}

// NewInternalFunction returns a new InternalFunction. This wrapper must not be returned to the client.
func NewInternalFunction(schemaName string, funcName string, params ...InternalType) InternalFunction {
	if len(schemaName) == 0 && len(funcName) == 0 && len(params) == 0 {
		return NullFunction
	}
	data := make([]string, len(params)+2)
	data[0] = schemaName
	data[1] = funcName
	for i := range params {
		data[2+i] = string(params[i])
	}
	return InternalFunction(NewInternal(Section_Function, data...))
}

// NewInternalIndex returns a new InternalIndex. This wrapper must not be returned to the client.
func NewInternalIndex(schemaName string, tableName string, indexName string) InternalIndex {
	if len(schemaName) == 0 && len(tableName) == 0 && len(indexName) == 0 {
		return NullIndex
	}
	return InternalIndex(NewInternal(Section_Index, schemaName, tableName, indexName))
}

// NewInternalNamespace returns a new InternalNamespace. This wrapper must not be returned to the client.
func NewInternalNamespace(schemaName string) InternalNamespace {
	if len(schemaName) == 0 {
		return NullNamespace
	}
	return InternalNamespace(NewInternal(Section_Namespace, schemaName))
}

// NewInternalOID returns a new InternalOID. This wrapper must not be returned to the client.
func NewInternalOID(val uint32) InternalOID {
	return InternalOID(NewInternal(Section_OID, strconv.FormatUint(uint64(val), 10)))
}

// NewInternalSequence returns a new InternalSequence. This wrapper must not be returned to the client.
func NewInternalSequence(schemaName string, sequenceName string) InternalSequence {
	if len(schemaName) == 0 && len(sequenceName) == 0 {
		return NullSequence
	}
	return InternalSequence(NewInternal(Section_Sequence, schemaName, sequenceName))
}

// NewInternalTable returns a new InternalTable. This wrapper must not be returned to the client.
func NewInternalTable(schemaName string, tableName string) InternalTable {
	if len(schemaName) == 0 && len(tableName) == 0 {
		return NullTable
	}
	return InternalTable(NewInternal(Section_Table, schemaName, tableName))
}

// NewInternalType returns a new InternalType. This wrapper must not be returned to the client.
func NewInternalType(schemaName string, typeName string) InternalType {
	if len(schemaName) == 0 && len(typeName) == 0 {
		return NullType
	}
	return InternalType(NewInternal(Section_Type, schemaName, typeName))
}

// NewInternalView returns a new InternalView. This wrapper must not be returned to the client.
func NewInternalView(schemaName string, viewName string) InternalView {
	if len(schemaName) == 0 && len(viewName) == 0 {
		return NullView
	}
	return InternalView(NewInternal(Section_View, schemaName, viewName))
}

// MethodName returns the method's name.
func (id InternalAccessMethod) MethodName() string {
	return Internal(id).Segment(0)
}

// CheckName returns the check's name.
func (id InternalCheck) CheckName() string {
	return Internal(id).Segment(2)
}

// SchemaName returns the schema name of the check.
func (id InternalCheck) SchemaName() string {
	return Internal(id).Segment(0)
}

// TableName returns the name of the table that the check belongs to.
func (id InternalCheck) TableName() string {
	return Internal(id).Segment(1)
}

// CollationName returns the collation's name.
func (id InternalCollation) CollationName() string {
	return Internal(id).Segment(1)
}

// SchemaName returns the schema name of the collation.
func (id InternalCollation) SchemaName() string {
	return Internal(id).Segment(0)
}

// ColumnName returns the column's name that the default belongs to.
func (id InternalColumnDefault) ColumnName() string {
	return Internal(id).Segment(2)
}

// SchemaName returns the schema name of the column default.
func (id InternalColumnDefault) SchemaName() string {
	return Internal(id).Segment(0)
}

// TableName returns the name of the table that the column belongs to.
func (id InternalColumnDefault) TableName() string {
	return Internal(id).Segment(1)
}

// DatabaseName returns the database's name.
func (id InternalDatabase) DatabaseName() string {
	return Internal(id).Segment(0)
}

// Parent returns the parent ENUM for the label.
func (id InternalEnumLabel) Parent() InternalType {
	return InternalType(Internal(id).Segment(0))
}

// Label returns the name of the label.
func (id InternalEnumLabel) Label() string {
	return Internal(id).Segment(1)
}

// ForeignKeyName returns the foreign key's name.
func (id InternalForeignKey) ForeignKeyName() string {
	return Internal(id).Segment(2)
}

// SchemaName returns the schema name of the foreign key.
func (id InternalForeignKey) SchemaName() string {
	return Internal(id).Segment(0)
}

// TableName returns the name of the table that the foreign key belongs to.
func (id InternalForeignKey) TableName() string {
	return Internal(id).Segment(1)
}

// FunctionName returns the function's name.
func (id InternalFunction) FunctionName() string {
	return Internal(id).Segment(1)
}

// Parameters returns the function's name.
func (id InternalFunction) Parameters() []InternalType {
	data := Internal(id).Data()[2:]
	params := make([]InternalType, len(data))
	for i := range data {
		params[i] = InternalType(data[i])
	}
	return params
}

// ParameterCount returns the function's name.
func (id InternalFunction) ParameterCount() int {
	return Internal(id).SegmentCount() - 2
}

// SchemaName returns the schema name of the function.
func (id InternalFunction) SchemaName() string {
	return Internal(id).Segment(0)
}

// IndexName returns the index's name.
func (id InternalIndex) IndexName() string {
	return Internal(id).Segment(2)
}

// SchemaName returns the schema name of the index.
func (id InternalIndex) SchemaName() string {
	return Internal(id).Segment(0)
}

// TableName returns the name of the table that the index belongs to.
func (id InternalIndex) TableName() string {
	return Internal(id).Segment(1)
}

// SchemaName returns the schema name.
func (id InternalNamespace) SchemaName() string {
	return Internal(id).Segment(0)
}

// OID returns the contained uint32 value.
func (id InternalOID) OID() uint32 {
	val, _ := strconv.ParseUint(Internal(id).Segment(0), 10, 32)
	return uint32(val)
}

// SchemaName returns the schema name of the sequence.
func (id InternalSequence) SchemaName() string {
	return Internal(id).Segment(0)
}

// SequenceName returns the name of the sequence.
func (id InternalSequence) SequenceName() string {
	return Internal(id).Segment(1)
}

// SchemaName returns the schema name of the table.
func (id InternalTable) SchemaName() string {
	return Internal(id).Segment(0)
}

// TableName returns the table's name.
func (id InternalTable) TableName() string {
	return Internal(id).Segment(1)
}

// SchemaName returns the schema name of the type.
func (id InternalType) SchemaName() string {
	return Internal(id).Segment(0)
}

// TypeName returns the type's name.
func (id InternalType) TypeName() string {
	return Internal(id).Segment(1)
}

// SchemaName returns the schema name of the view.
func (id InternalView) SchemaName() string {
	return Internal(id).Segment(0)
}

// ViewName returns the view's name.
func (id InternalView) ViewName() string {
	return Internal(id).Segment(1)
}

// IsValid returns whether the ID is valid.
func (id InternalAccessMethod) IsValid() bool { return Internal(id).IsValid() }

// IsValid returns whether the ID is valid.
func (id InternalCheck) IsValid() bool { return Internal(id).IsValid() }

// IsValid returns whether the ID is valid.
func (id InternalCollation) IsValid() bool { return Internal(id).IsValid() }

// IsValid returns whether the ID is valid.
func (id InternalColumnDefault) IsValid() bool { return Internal(id).IsValid() }

// IsValid returns whether the ID is valid.
func (id InternalDatabase) IsValid() bool { return Internal(id).IsValid() }

// IsValid returns whether the ID is valid.
func (id InternalEnumLabel) IsValid() bool { return Internal(id).IsValid() }

// IsValid returns whether the ID is valid.
func (id InternalForeignKey) IsValid() bool { return Internal(id).IsValid() }

// IsValid returns whether the ID is valid.
func (id InternalFunction) IsValid() bool { return Internal(id).IsValid() }

// IsValid returns whether the ID is valid.
func (id InternalIndex) IsValid() bool { return Internal(id).IsValid() }

// IsValid returns whether the ID is valid.
func (id InternalNamespace) IsValid() bool { return Internal(id).IsValid() }

// IsValid returns whether the ID is valid.
func (id InternalOID) IsValid() bool { return Internal(id).IsValid() }

// IsValid returns whether the ID is valid.
func (id InternalSequence) IsValid() bool { return Internal(id).IsValid() }

// IsValid returns whether the ID is valid.
func (id InternalTable) IsValid() bool { return Internal(id).IsValid() }

// IsValid returns whether the ID is valid.
func (id InternalType) IsValid() bool { return Internal(id).IsValid() }

// IsValid returns whether the ID is valid.
func (id InternalView) IsValid() bool { return Internal(id).IsValid() }

// Internal returns the unwrapped ID.
func (id InternalAccessMethod) Internal() Internal { return Internal(id) }

// Internal returns the unwrapped ID.
func (id InternalCheck) Internal() Internal { return Internal(id) }

// Internal returns the unwrapped ID.
func (id InternalCollation) Internal() Internal { return Internal(id) }

// Internal returns the unwrapped ID.
func (id InternalColumnDefault) Internal() Internal { return Internal(id) }

// Internal returns the unwrapped ID.
func (id InternalDatabase) Internal() Internal { return Internal(id) }

// Internal returns the unwrapped ID.
func (id InternalEnumLabel) Internal() Internal { return Internal(id) }

// Internal returns the unwrapped ID.
func (id InternalForeignKey) Internal() Internal { return Internal(id) }

// Internal returns the unwrapped ID.
func (id InternalFunction) Internal() Internal { return Internal(id) }

// Internal returns the unwrapped ID.
func (id InternalIndex) Internal() Internal { return Internal(id) }

// Internal returns the unwrapped ID.
func (id InternalNamespace) Internal() Internal { return Internal(id) }

// Internal returns the unwrapped ID.
func (id InternalOID) Internal() Internal { return Internal(id) }

// Internal returns the unwrapped ID.
func (id InternalSequence) Internal() Internal { return Internal(id) }

// Internal returns the unwrapped ID.
func (id InternalTable) Internal() Internal { return Internal(id) }

// Internal returns the unwrapped ID.
func (id InternalType) Internal() Internal { return Internal(id) }

// Internal returns the unwrapped ID.
func (id InternalView) Internal() Internal { return Internal(id) }
