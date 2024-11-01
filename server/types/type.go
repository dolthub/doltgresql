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
	"time"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/types"
	"github.com/dolthub/vitess/go/sqltypes"
	"github.com/dolthub/vitess/go/vt/proto/query"
	"gopkg.in/src-d/go-errors.v1"

	"github.com/dolthub/doltgresql/postgres/parser/duration"
	"github.com/dolthub/doltgresql/utils"
)

var ErrTypeAlreadyExists = errors.NewKind(`type "%s" already exists`)
var ErrTypeDoesNotExist = errors.NewKind(`type "%s" does not exist`)

var ErrUnhandledType = errors.NewKind(`%s: unhandled type: %T`)
var ErrInvalidSyntaxForType = errors.NewKind(`invalid input syntax for type %s: %q`)
var ErrValueIsOutOfRangeForType = errors.NewKind(`value %q is out of range for type %s`)

// DoltgresType represents a single type.
type DoltgresType struct {
	OID           uint32
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
	isSerial     bool // TODO: to replace serial types
	isUnresolved bool
}

var _ types.ExtendedType = DoltgresType{}

func NewUnresolvedDoltgresType(sch, name string) DoltgresType {
	return DoltgresType{
		Name:         name,
		Schema:       sch,
		isUnresolved: true,
	}
}

func (t DoltgresType) Resolved() bool {
	return !t.isUnresolved
}

func (t DoltgresType) ArrayBaseType() (DoltgresType, bool) {
	if t.Elem == 0 {
		return DoltgresType{}, false
	}
	elem, ok := OidToBuildInDoltgresType[t.Elem]
	return elem, ok
}

// IsArrayType returns true if the type is of 'array' category
func (t DoltgresType) IsArrayType() bool {
	return t.TypCategory == TypeCategory_ArrayTypes
}

func (t DoltgresType) EmptyType() bool {
	// TODO
	return t.OID == 0 && t.Name == ""
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
	return t.TypCategory == TypeCategory_PseudoTypes
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

var IoCompare func(ctx *sql.Context, t DoltgresType, v1, v2 any) (int, error)

// Compare implements the types.ExtendedType interface.
func (t DoltgresType) Compare(v1 interface{}, v2 interface{}) (int, error) {
	return IoCompare(sql.NewEmptyContext(), t, v1, v2)
}

var IoReceive func(ctx *sql.Context, t DoltgresType, val any) (any, error)

// Convert implements the types.ExtendedType interface.
func (t DoltgresType) Convert(v interface{}) (interface{}, sql.ConvertInRange, error) {
	val, err := IoReceive(sql.NewEmptyContext(), t, v)
	if err != nil {
		return nil, false, err
	}
	return val, true, nil
}

// Equals implements the types.ExtendedType interface.
func (t DoltgresType) Equals(otherType sql.Type) bool {
	if otherExtendedType, ok := otherType.(DoltgresType); ok {
		return bytes.Equal(t.Serialize(), otherExtendedType.Serialize())
	}
	return false
}

var IoOutput func(ctx *sql.Context, t DoltgresType, val any) (string, error)

// FormatValue implements the types.ExtendedType interface.
func (t DoltgresType) FormatValue(val any) (string, error) {
	if val == nil {
		return "", nil
	}
	return IoOutput(sql.NewEmptyContext(), t, val)
}

// MaxSerializedWidth implements the types.ExtendedType interface.
func (t DoltgresType) MaxSerializedWidth() types.ExtendedTypeSerializedWidth {
	// TODO
	return types.ExtendedTypeSerializedWidth_64K
}

// MaxTextResponseByteLength implements the types.ExtendedType interface.
func (t DoltgresType) MaxTextResponseByteLength(ctx *sql.Context) uint32 {
	// TODO
	return 1
}

// Promote implements the types.ExtendedType interface.
func (t DoltgresType) Promote() sql.Type {
	return t
}

// SerializedCompare implements the types.ExtendedType interface.
func (t DoltgresType) SerializedCompare(v1 []byte, v2 []byte) (int, error) {
	if len(v1) == 0 && len(v2) == 0 {
		return 0, nil
	} else if len(v1) > 0 && len(v2) == 0 {
		return 1, nil
	} else if len(v1) == 0 && len(v2) > 0 {
		return -1, nil
	}
	return bytes.Compare(v1, v2), nil
}

// SQL implements the types.ExtendedType interface.
func (t DoltgresType) SQL(ctx *sql.Context, dest []byte, v interface{}) (sqltypes.Value, error) {
	if v == nil {
		return sqltypes.NULL, nil
	}
	value, err := IoOutput(ctx, t, v)
	if err != nil {
		return sqltypes.Value{}, err
	}

	// TODO: check type
	return sqltypes.MakeTrusted(sqltypes.Text, types.AppendAndSliceBytes(dest, []byte(value))), nil
}

// String implements the types.ExtendedType interface.
func (t DoltgresType) String() string {
	return t.Name
}

// Type implements the types.ExtendedType interface.
func (t DoltgresType) Type() query.Type {
	switch t.TypCategory {
	case TypeCategory_ArrayTypes:
		return sqltypes.Text
	case TypeCategory_BooleanTypes:
		return sqltypes.Text
	case TypeCategory_CompositeTypes, TypeCategory_EnumTypes, TypeCategory_GeometricTypes, TypeCategory_NetworkAddressTypes,
		TypeCategory_RangeTypes, TypeCategory_PseudoTypes, TypeCategory_UserDefinedTypes, TypeCategory_BitStringTypes,
		TypeCategory_InternalUseTypes:
		// TODO
		return sqltypes.Text
	case TypeCategory_DateTimeTypes:
		return sqltypes.Text
	case TypeCategory_NumericTypes:
		// decimal.Zero
		return sqltypes.Int64
	case TypeCategory_StringTypes, TypeCategory_UnknownTypes:
		return sqltypes.Text
	case TypeCategory_TimespanTypes:
		return sqltypes.Text
	default:
		// shouldn't happen
		return sqltypes.Text
	}
}

// ValueType implements the types.ExtendedType interface.
func (t DoltgresType) ValueType() reflect.Type {
	return reflect.TypeOf(t.Zero())
}

// Zero implements the types.ExtendedType interface.
func (t DoltgresType) Zero() interface{} {
	switch t.TypCategory {
	case TypeCategory_ArrayTypes:
		return []any{}
	case TypeCategory_BooleanTypes:
		return false
	case TypeCategory_CompositeTypes, TypeCategory_EnumTypes, TypeCategory_GeometricTypes, TypeCategory_NetworkAddressTypes,
		TypeCategory_RangeTypes, TypeCategory_PseudoTypes, TypeCategory_UserDefinedTypes, TypeCategory_BitStringTypes,
		TypeCategory_InternalUseTypes:
		// TODO
		return any(nil)
	case TypeCategory_DateTimeTypes:
		return time.Time{}
	case TypeCategory_NumericTypes:
		// decimal.Zero
		return 0
	case TypeCategory_StringTypes, TypeCategory_UnknownTypes:
		return ""
	case TypeCategory_TimespanTypes:
		return duration.MakeDuration(0, 0, 0)
	default:
		// shouldn't happen
		return any(nil)
	}
}

var IoSend func(ctx *sql.Context, t DoltgresType, val any) ([]byte, error)

// SerializeValue implements the types.ExtendedType interface.
func (t DoltgresType) SerializeValue(val any) ([]byte, error) {
	if val == nil {
		return nil, nil
	}
	converted, _, err := t.Convert(val)
	if err != nil {
		return nil, err
	}
	// TODO: use converted value or not needed?
	return IoSend(sql.NewEmptyContext(), t, converted)
}

// DeserializeValue implements the types.ExtendedType interface.
func (t DoltgresType) DeserializeValue(val []byte) (any, error) {
	// TODO: how to deserialize?
	if len(val) == 0 {
		return nil, nil
	}
	reader := utils.NewReader(val)
	return reader.String(), nil
}

// IsSerial returns whether the type is serial type.
// This is true for int16serial, int32serial and int64serial types.
func (t DoltgresType) IsSerial() bool {
	return t.isSerial
}
