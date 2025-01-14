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

// AccessMethod is an Id wrapper for access methods. This wrapper must not be returned to the client.
type AccessMethod Id

// Check is an Id wrapper for checks. This wrapper must not be returned to the client.
type Check Id

// Collation is an Id wrapper for collations. This wrapper must not be returned to the client.
type Collation Id

// ColumnDefault is an Id wrapper for column defaults. This wrapper must not be returned to the client.
type ColumnDefault Id

// Database is an Id wrapper for databases. This wrapper must not be returned to the client.
type Database Id

// EnumLabel is an Id wrapper for enum labels. This wrapper must not be returned to the client.
type EnumLabel Id

// ForeignKey is an Id wrapper for foreign keys. This wrapper must not be returned to the client.
type ForeignKey Id

// Function is an Id wrapper for functions. This wrapper must not be returned to the client.
type Function Id

// Index is an Id wrapper for indexes. This wrapper must not be returned to the client.
type Index Id

// Namespace is an Id wrapper for schemas/namespaces. This wrapper must not be returned to the client.
type Namespace Id

// Oid is an Id wrapper for OIDs. This wrapper must not be returned to the client.
type Oid Id

// Sequence is an Id wrapper for sequences. This wrapper must not be returned to the client.
type Sequence Id

// Table is an Id wrapper for tables. This wrapper must not be returned to the client.
type Table Id

// Type is an Id wrapper for types. This wrapper must not be returned to the client.
type Type Id

// View is an Id wrapper for views. This wrapper must not be returned to the client.
type View Id

// NewAccessMethod returns a new AccessMethod. This wrapper must not be returned to the client.
func NewAccessMethod(methodName string) AccessMethod {
	if len(methodName) == 0 {
		return NullAccessMethod
	}
	return AccessMethod(NewId(Section_AccessMethod, methodName))
}

// NewCheck returns a new Check. This wrapper must not be returned to the client.
func NewCheck(schemaName string, tableName string, checkName string) Check {
	if len(schemaName) == 0 && len(tableName) == 0 && len(checkName) == 0 {
		return NullCheck
	}
	return Check(NewId(Section_Check, schemaName, tableName, checkName))
}

// NewCollation returns a new Collation. This wrapper must not be returned to the client.
func NewCollation(schemaName string, collationName string) Collation {
	if len(schemaName) == 0 && len(collationName) == 0 {
		return NullCollation
	}
	return Collation(NewId(Section_Collation, schemaName, collationName))
}

// NewColumnDefault returns a new ColumnDefault. This wrapper must not be returned to the client.
func NewColumnDefault(schemaName string, tableName string, columnName string) ColumnDefault {
	if len(schemaName) == 0 && len(tableName) == 0 && len(columnName) == 0 {
		return NullColumnDefault
	}
	return ColumnDefault(NewId(Section_ColumnDefault, schemaName, tableName, columnName))
}

// NewDatabase returns a new Database. This wrapper must not be returned to the client.
func NewDatabase(dbName string) Database {
	if len(dbName) == 0 {
		return NullDatabase
	}
	return Database(NewId(Section_Database, dbName))
}

// NewEnumLabel returns a new EnumLabel. This wrapper must not be returned to the client.
func NewEnumLabel(parent Type, label string) EnumLabel {
	if len(parent) == 0 && len(label) == 0 {
		return NullEnumLabel
	}
	return EnumLabel(NewId(Section_EnumLabel, string(parent), label))
}

// NewForeignKey returns a new ForeignKey. This wrapper must not be returned to the client.
func NewForeignKey(schemaName string, tableName string, fkName string) ForeignKey {
	if len(schemaName) == 0 && len(tableName) == 0 && len(fkName) == 0 {
		return NullForeignKey
	}
	return ForeignKey(NewId(Section_ForeignKey, schemaName, tableName, fkName))
}

// NewFunction returns a new Function. This wrapper must not be returned to the client.
func NewFunction(schemaName string, funcName string, params ...Type) Function {
	if len(schemaName) == 0 && len(funcName) == 0 && len(params) == 0 {
		return NullFunction
	}
	data := make([]string, len(params)+2)
	data[0] = schemaName
	data[1] = funcName
	for i := range params {
		data[2+i] = string(params[i])
	}
	return Function(NewId(Section_Function, data...))
}

// NewIndex returns a new Index. This wrapper must not be returned to the client.
func NewIndex(schemaName string, tableName string, indexName string) Index {
	if len(schemaName) == 0 && len(tableName) == 0 && len(indexName) == 0 {
		return NullIndex
	}
	return Index(NewId(Section_Index, schemaName, tableName, indexName))
}

// NewNamespace returns a new Namespace. This wrapper must not be returned to the client.
func NewNamespace(schemaName string) Namespace {
	if len(schemaName) == 0 {
		return NullNamespace
	}
	return Namespace(NewId(Section_Namespace, schemaName))
}

// NewOID returns a new Oid. This wrapper must not be returned to the client.
func NewOID(val uint32) Oid {
	return Oid(NewId(Section_OID, strconv.FormatUint(uint64(val), 10)))
}

// NewSequence returns a new Sequence. This wrapper must not be returned to the client.
func NewSequence(schemaName string, sequenceName string) Sequence {
	if len(schemaName) == 0 && len(sequenceName) == 0 {
		return NullSequence
	}
	return Sequence(NewId(Section_Sequence, schemaName, sequenceName))
}

// NewTable returns a new Table. This wrapper must not be returned to the client.
func NewTable(schemaName string, tableName string) Table {
	if len(schemaName) == 0 && len(tableName) == 0 {
		return NullTable
	}
	return Table(NewId(Section_Table, schemaName, tableName))
}

// NewType returns a new Type. This wrapper must not be returned to the client.
func NewType(schemaName string, typeName string) Type {
	if len(schemaName) == 0 && len(typeName) == 0 {
		return NullType
	}
	return Type(NewId(Section_Type, schemaName, typeName))
}

// NewView returns a new View. This wrapper must not be returned to the client.
func NewView(schemaName string, viewName string) View {
	if len(schemaName) == 0 && len(viewName) == 0 {
		return NullView
	}
	return View(NewId(Section_View, schemaName, viewName))
}

// MethodName returns the method's name.
func (id AccessMethod) MethodName() string {
	return Id(id).Segment(0)
}

// CheckName returns the check's name.
func (id Check) CheckName() string {
	return Id(id).Segment(2)
}

// SchemaName returns the schema name of the check.
func (id Check) SchemaName() string {
	return Id(id).Segment(0)
}

// TableName returns the name of the table that the check belongs to.
func (id Check) TableName() string {
	return Id(id).Segment(1)
}

// CollationName returns the collation's name.
func (id Collation) CollationName() string {
	return Id(id).Segment(1)
}

// SchemaName returns the schema name of the collation.
func (id Collation) SchemaName() string {
	return Id(id).Segment(0)
}

// ColumnName returns the column's name that the default belongs to.
func (id ColumnDefault) ColumnName() string {
	return Id(id).Segment(2)
}

// SchemaName returns the schema name of the column default.
func (id ColumnDefault) SchemaName() string {
	return Id(id).Segment(0)
}

// TableName returns the name of the table that the column belongs to.
func (id ColumnDefault) TableName() string {
	return Id(id).Segment(1)
}

// DatabaseName returns the database's name.
func (id Database) DatabaseName() string {
	return Id(id).Segment(0)
}

// Parent returns the parent ENUM for the label.
func (id EnumLabel) Parent() Type {
	return Type(Id(id).Segment(0))
}

// Label returns the name of the label.
func (id EnumLabel) Label() string {
	return Id(id).Segment(1)
}

// ForeignKeyName returns the foreign key's name.
func (id ForeignKey) ForeignKeyName() string {
	return Id(id).Segment(2)
}

// SchemaName returns the schema name of the foreign key.
func (id ForeignKey) SchemaName() string {
	return Id(id).Segment(0)
}

// TableName returns the name of the table that the foreign key belongs to.
func (id ForeignKey) TableName() string {
	return Id(id).Segment(1)
}

// FunctionName returns the function's name.
func (id Function) FunctionName() string {
	return Id(id).Segment(1)
}

// Parameters returns the function's name.
func (id Function) Parameters() []Type {
	data := Id(id).Data()[2:]
	params := make([]Type, len(data))
	for i := range data {
		params[i] = Type(data[i])
	}
	return params
}

// ParameterCount returns the function's name.
func (id Function) ParameterCount() int {
	return Id(id).SegmentCount() - 2
}

// SchemaName returns the schema name of the function.
func (id Function) SchemaName() string {
	return Id(id).Segment(0)
}

// IndexName returns the index's name.
func (id Index) IndexName() string {
	return Id(id).Segment(2)
}

// SchemaName returns the schema name of the index.
func (id Index) SchemaName() string {
	return Id(id).Segment(0)
}

// TableName returns the name of the table that the index belongs to.
func (id Index) TableName() string {
	return Id(id).Segment(1)
}

// SchemaName returns the schema name.
func (id Namespace) SchemaName() string {
	return Id(id).Segment(0)
}

// OID returns the contained uint32 value.
func (id Oid) OID() uint32 {
	val, _ := strconv.ParseUint(Id(id).Segment(0), 10, 32)
	return uint32(val)
}

// SchemaName returns the schema name of the sequence.
func (id Sequence) SchemaName() string {
	return Id(id).Segment(0)
}

// SequenceName returns the name of the sequence.
func (id Sequence) SequenceName() string {
	return Id(id).Segment(1)
}

// SchemaName returns the schema name of the table.
func (id Table) SchemaName() string {
	return Id(id).Segment(0)
}

// TableName returns the table's name.
func (id Table) TableName() string {
	return Id(id).Segment(1)
}

// SchemaName returns the schema name of the type.
func (id Type) SchemaName() string {
	return Id(id).Segment(0)
}

// TypeName returns the type's name.
func (id Type) TypeName() string {
	return Id(id).Segment(1)
}

// SchemaName returns the schema name of the view.
func (id View) SchemaName() string {
	return Id(id).Segment(0)
}

// ViewName returns the view's name.
func (id View) ViewName() string {
	return Id(id).Segment(1)
}

// IsValid returns whether the ID is valid.
func (id AccessMethod) IsValid() bool { return Id(id).IsValid() }

// IsValid returns whether the ID is valid.
func (id Check) IsValid() bool { return Id(id).IsValid() }

// IsValid returns whether the ID is valid.
func (id Collation) IsValid() bool { return Id(id).IsValid() }

// IsValid returns whether the ID is valid.
func (id ColumnDefault) IsValid() bool { return Id(id).IsValid() }

// IsValid returns whether the ID is valid.
func (id Database) IsValid() bool { return Id(id).IsValid() }

// IsValid returns whether the ID is valid.
func (id EnumLabel) IsValid() bool { return Id(id).IsValid() }

// IsValid returns whether the ID is valid.
func (id ForeignKey) IsValid() bool { return Id(id).IsValid() }

// IsValid returns whether the ID is valid.
func (id Function) IsValid() bool { return Id(id).IsValid() }

// IsValid returns whether the ID is valid.
func (id Index) IsValid() bool { return Id(id).IsValid() }

// IsValid returns whether the ID is valid.
func (id Namespace) IsValid() bool { return Id(id).IsValid() }

// IsValid returns whether the ID is valid.
func (id Oid) IsValid() bool { return Id(id).IsValid() }

// IsValid returns whether the ID is valid.
func (id Sequence) IsValid() bool { return Id(id).IsValid() }

// IsValid returns whether the ID is valid.
func (id Table) IsValid() bool { return Id(id).IsValid() }

// IsValid returns whether the ID is valid.
func (id Type) IsValid() bool { return Id(id).IsValid() }

// IsValid returns whether the ID is valid.
func (id View) IsValid() bool { return Id(id).IsValid() }

// AsId returns the unwrapped ID.
func (id AccessMethod) AsId() Id { return Id(id) }

// AsId returns the unwrapped ID.
func (id Check) AsId() Id { return Id(id) }

// AsId returns the unwrapped ID.
func (id Collation) AsId() Id { return Id(id) }

// AsId returns the unwrapped ID.
func (id ColumnDefault) AsId() Id { return Id(id) }

// AsId returns the unwrapped ID.
func (id Database) AsId() Id { return Id(id) }

// AsId returns the unwrapped ID.
func (id EnumLabel) AsId() Id { return Id(id) }

// AsId returns the unwrapped ID.
func (id ForeignKey) AsId() Id { return Id(id) }

// AsId returns the unwrapped ID.
func (id Function) AsId() Id { return Id(id) }

// AsId returns the unwrapped ID.
func (id Index) AsId() Id { return Id(id) }

// AsId returns the unwrapped ID.
func (id Namespace) AsId() Id { return Id(id) }

// AsId returns the unwrapped ID.
func (id Oid) AsId() Id { return Id(id) }

// AsId returns the unwrapped ID.
func (id Sequence) AsId() Id { return Id(id) }

// AsId returns the unwrapped ID.
func (id Table) AsId() Id { return Id(id) }

// AsId returns the unwrapped ID.
func (id Type) AsId() Id { return Id(id) }

// AsId returns the unwrapped ID.
func (id View) AsId() Id { return Id(id) }
