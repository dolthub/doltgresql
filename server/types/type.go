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
	"bytes"
	"reflect"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/types"
	"github.com/dolthub/vitess/go/sqltypes"
	"github.com/dolthub/vitess/go/vt/proto/query"
	"gopkg.in/src-d/go-errors.v1"
)

var ErrTypeAlreadyExists = errors.NewKind(`type "%s" already exists`)
var ErrTypeDoesNotExist = errors.NewKind(`type "%s" does not exist`)

var ErrUnhandledType = errors.NewKind(`%s: unhandled type: %T`)
var ErrInvalidSyntaxForType = errors.NewKind(`invalid input syntax for type %s: %q`)

// DoltgresType represents a single type.
type DoltgresType struct {
	Oid           uint32
	Name          string
	Schema        string // TODO: should be `uint32`.
	Owner         string // TODO: should be `uint32`.
	Length        int16
	PassedByVal   bool
	TypType       TypeType
	TypCategory   TypeCategory
	IsPreferred   bool
	IsDefined     bool
	Delimiter     string
	RelID         uint32 // for Composite types
	SubscriptFunc string
	Elem          uint32
	Array         uint32
	InputFunc     string
	OutputFunc    string
	ReceiveFunc   string
	SendFunc      string
	ModInFunc     string
	ModOutFunc    string
	AnalyzeFunc   string
	Align         TypeAlignment
	Storage       TypeStorage
	NotNull       bool   // for Domain types
	BaseTypeOID   uint32 // for Domain types
	TypMod        int32  // for Domain types
	NDims         int32  // for Domain types
	Collation     uint32
	DefaulBin     string // for Domain types
	Default       string
	Acl           string                 // TODO: list of privileges
	Checks        []*sql.CheckDefinition // TODO: this is not part of `pg_type` instead `pg_constraint` for Domain types.

	// These are for internal use
	baseID                    DoltgresTypeBaseID
	serializationID           SerializationID
	isUnbounded               bool
	maxSerializedWidth        types.ExtendedTypeSerializedWidth
	maxTextResponseByteLength uint32
	stringName                string
	queryType                 query.Type
	valueType                 reflect.Type
	zero                      any
}

var _ types.ExtendedType = DoltgresType{}

func (t DoltgresType) ArrayBaseType() (DoltgresType, bool) {
	if t.Elem == 0 {
		return DoltgresType{}, false
	}
	elem, ok := OidToBuildInDoltgresType[t.Elem]
	return elem, ok
}

// BaseID implements the DoltgresTypeInterface interface.
func (t DoltgresType) BaseID() DoltgresTypeBaseID {
	return t.baseID
}

// IsArrayType returns true if the type is of 'array' category
func (t DoltgresType) IsArrayType() bool {
	return t.TypCategory == TypeCategory_ArrayTypes
}

func (t DoltgresType) EmptyType() bool {
	// TODO
	return t.Name == ""
}

func (t DoltgresType) DomainUnderlyingBaseType() DoltgresType {
	// TODO: account for user-defined type
	bt, ok := OidToBuildInDoltgresType[t.BaseTypeOID]
	if !ok {
		// TODO
	}
	if bt.TypType == TypeType_Domain {
		return bt.DomainUnderlyingBaseType()
	} else {
		return bt
	}
}

// IsPolymorphicType These types are special built-in pseudo-types
// that are used during function resolution to allow a function
// to handle multiple types from a single definition.
// All polymorphic types have "any" as a prefix.
// The exception is the "any" type, which is not a polymorphic type.
func (t DoltgresType) IsPolymorphicType() bool {
	return t.TypType == TypeType_Pseudo
}

// IsValidForPolymorphicType returns whether the given type is valid for the calling polymorphic type.
func (t DoltgresType) IsValidForPolymorphicType(target DoltgresType) bool {
	// TODO: check for other pseudo types?
	if t.TypType != TypeType_Pseudo {
		return false
	}
	if t.Name == "anyarray" {
		return target.TypCategory == TypeCategory_ArrayTypes
	} else if t.Name == "anynonarray" {
		return target.TypCategory != TypeCategory_ArrayTypes
	} else if t.Name == "anyelement" {
		return true
	} else {
		return false
	}
}

// ToArrayType implements the types.ExtendedType interface.
func (t DoltgresType) ToArrayType() (DoltgresType, bool) {
	if t.Array == 0 {
		return DoltgresType{}, false
	}
	arr, ok := OidToBuildInDoltgresType[t.Array]
	return arr, ok
}

// CollationCoercibility implements the types.ExtendedType interface.
func (t DoltgresType) CollationCoercibility(ctx *sql.Context) (collation sql.CollationID, coercibility byte) {
	// TODO: seems all types are the same??
	return sql.Collation_binary, 5
}

// Compare implements the types.ExtendedType interface.
func (t DoltgresType) Compare(v1 interface{}, v2 interface{}) (int, error) {
	if v1 == nil && v2 == nil {
		return 0, nil
	} else if v1 != nil && v2 == nil {
		return 1, nil
	} else if v1 == nil && v2 != nil {
		return -1, nil
	}

	ac, _, err := t.Convert(v1)
	if err != nil {
		return 0, err
	}
	bc, _, err := t.Convert(v2)
	if err != nil {
		return 0, err
	}
	return t.compareFunc(ac, bc)
}

// Convert implements the types.ExtendedType interface.
func (t DoltgresType) Convert(v interface{}) (interface{}, sql.ConvertInRange, error) {
	return t.convertFunc(v)
}

// Equals implements the types.ExtendedType interface.
func (t DoltgresType) Equals(otherType sql.Type) bool {
	// TODO: pseudo types should be true?
	if otherExtendedType, ok := otherType.(types.ExtendedType); ok {
		return bytes.Equal(MustSerializeType(t), MustSerializeType(otherExtendedType))
	}
	return false
}

// FormatValue implements the types.ExtendedType interface.
func (t DoltgresType) FormatValue(val any) (string, error) {
	if val == nil {
		return "", nil
	}
	return t.IoOutput(sql.NewEmptyContext(), val)
}

// MaxSerializedWidth implements the types.ExtendedType interface.
func (t DoltgresType) MaxSerializedWidth() types.ExtendedTypeSerializedWidth {
	return t.maxSerializedWidth
}

// MaxTextResponseByteLength implements the types.ExtendedType interface.
func (t DoltgresType) MaxTextResponseByteLength(ctx *sql.Context) uint32 {
	return t.maxTextResponseByteLength
}

// Promote implements the types.ExtendedType interface.
func (t DoltgresType) Promote() sql.Type {
	return t
}

// SerializedCompare implements the types.ExtendedType interface.
func (t DoltgresType) SerializedCompare(v1 []byte, v2 []byte) (int, error) {
	return t.serializedCompareFunc(v1, v2)
}

// SQL implements the types.ExtendedType interface.
func (t DoltgresType) SQL(ctx *sql.Context, dest []byte, v interface{}) (sqltypes.Value, error) {
	if v == nil {
		return sqltypes.NULL, nil
	}
	value, err := t.ioOutputFunc(ctx, v)
	if err != nil {
		return sqltypes.Value{}, err
	}
	return t.sqlFunc(value)
}

// String implements the types.ExtendedType interface.
func (t DoltgresType) String() string {
	return t.stringName
}

// Type implements the types.ExtendedType interface.
func (t DoltgresType) Type() query.Type {
	return t.queryType
}

// ValueType implements the types.ExtendedType interface.
func (t DoltgresType) ValueType() reflect.Type {
	return t.valueType
}

// Zero implements the types.ExtendedType interface.
func (t DoltgresType) Zero() interface{} {
	return t.zero
}

// SerializeValue implements the types.ExtendedType interface.
func (t DoltgresType) SerializeValue(val any) ([]byte, error) {
	if val == nil {
		return nil, nil
	}
	converted, _, err := t.Convert(val)
	if err != nil {
		return nil, err
	}
	return t.serializeValueFunc(converted)
}

// DeserializeValue implements the types.ExtendedType interface.
func (t DoltgresType) DeserializeValue(val []byte) (any, error) {
	return t.deserializeValueFunc(val)
}
