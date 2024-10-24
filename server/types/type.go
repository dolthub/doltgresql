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
	"fmt"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/types"
	"github.com/dolthub/vitess/go/sqltypes"
	"github.com/dolthub/vitess/go/vt/proto/query"
	"gopkg.in/src-d/go-errors.v1"
	"reflect"
)

var ErrTypeAlreadyExists = errors.NewKind(`type "%s" already exists`)
var ErrTypeDoesNotExist = errors.NewKind(`type "%s" does not exist`)

var ErrUnhandledType = errors.NewKind(`%s: unhandled type: %T`)
var ErrInvalidSyntaxForType = errors.NewKind(`invalid input syntax for type %s: %q`)

// Type represents a single type.
type Type struct {
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
	compareFunc               TypeCompareFunc
	convertFunc               TypeConvertFunc
	serializationID           SerializationID
	ioInputFunc               IoInputFunc
	ioOutputFunc              IoOutputFunc
	isUnbounded               bool
	maxSerializedWidth        types.ExtendedTypeSerializedWidth
	maxTextResponseByteLength uint32
	serializedCompareFunc     SerializedCompareFunc
	sqlFunc                   SQLFunc
	stringName                string
	toArrayTypeFunc           ToArrayTypeFunc
	queryType                 query.Type
	valueType                 reflect.Type
	zero                      any
	serializeTypeFunc         SerializeTypeFunc
	deserializeTypeFunc       DeserializeTypeFunc
	serializeValueFunc        SerializeValueFunc
	deserializeValueFunc      DeserializeValueFunc
}

type TypeCompareFunc func(converted1 interface{}, converted2 interface{}) (int, error)
type TypeConvertFunc func(v interface{}) (interface{}, sql.ConvertInRange, error)
type IoInputFunc func(ctx *sql.Context, input string) (any, error)
type IoOutputFunc func(ctx *sql.Context, converted any) (string, error)
type SerializedCompareFunc func(v1 []byte, v2 []byte) (int, error)
type SQLFunc func(ioOutputStr string) (sqltypes.Value, error)
type ToArrayTypeFunc func() DoltgresArrayType
type SerializeTypeFunc func() ([]byte, error)
type DeserializeTypeFunc func(version uint16, metadata []byte) (DoltgresType, error)
type SerializeValueFunc func(val any) ([]byte, error)
type DeserializeValueFunc func(val []byte) (any, error)

var _ DoltgresType = Type{}

// Alignment implements the DoltgresType interface.
func (t Type) Alignment() TypeAlignment {
	return t.Align
}

// BaseID implements the DoltgresType interface.
func (t Type) BaseID() DoltgresTypeBaseID {
	return t.baseID
}

// BaseName implements the DoltgresType interface.
func (t Type) BaseName() string {
	return t.Name
}

// Category implements the DoltgresType interface.
func (t Type) Category() TypeCategory {
	return t.TypCategory
}

// CollationCoercibility implements the DoltgresType interface.
func (t Type) CollationCoercibility(ctx *sql.Context) (collation sql.CollationID, coercibility byte) {
	// TODO: seems all types are the same??
	return sql.Collation_binary, 5
}

// Compare implements the DoltgresType interface.
func (t Type) Compare(v1 interface{}, v2 interface{}) (int, error) {
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

// Convert implements the DoltgresType interface.
func (t Type) Convert(v interface{}) (interface{}, sql.ConvertInRange, error) {
	return t.convertFunc(v)
}

// Equals implements the DoltgresType interface.
func (t Type) Equals(otherType sql.Type) bool {
	// TODO: pseudo types should be true?
	if otherExtendedType, ok := otherType.(types.ExtendedType); ok {
		return bytes.Equal(MustSerializeType(t), MustSerializeType(otherExtendedType))
	}
	return false
}

// FormatValue implements the types.ExtendedType interface.
func (t Type) FormatValue(val any) (string, error) {
	if val == nil {
		return "", nil
	}
	return t.IoOutput(sql.NewEmptyContext(), val)
}

// GetSerializationID implements the DoltgresType interface.
func (t Type) GetSerializationID() SerializationID {
	return t.serializationID
}

// IoInput implements the DoltgresType interface.
func (t Type) IoInput(ctx *sql.Context, input string) (any, error) {
	return t.ioInputFunc(ctx, input)
}

// IoOutput implements the DoltgresType interface.
func (t Type) IoOutput(ctx *sql.Context, output any) (string, error) {
	converted, _, err := t.Convert(output)
	if err != nil {
		return "", err
	}
	return t.ioOutputFunc(ctx, converted)
}

// IsPreferredType implements the DoltgresType interface.
func (t Type) IsPreferredType() bool {
	return t.IsPreferred
}

// IsUnbounded implements the DoltgresType interface.
func (t Type) IsUnbounded() bool {
	return t.isUnbounded
}

// MaxSerializedWidth implements the types.ExtendedType interface.
func (t Type) MaxSerializedWidth() types.ExtendedTypeSerializedWidth {
	return t.maxSerializedWidth
}

// MaxTextResponseByteLength implements the DoltgresType interface.
func (t Type) MaxTextResponseByteLength(ctx *sql.Context) uint32 {
	return t.maxTextResponseByteLength
}

// OID implements the DoltgresType interface.
func (t Type) OID() uint32 {
	//TODO: generate unique oid
	return t.Oid
}

// Promote implements the DoltgresType interface.
func (t Type) Promote() sql.Type {
	return t
}

// SerializedCompare implements the DoltgresType interface.
func (t Type) SerializedCompare(v1 []byte, v2 []byte) (int, error) {
	return t.serializedCompareFunc(v1, v2)
}

// SQL implements the DoltgresType interface.
func (t Type) SQL(ctx *sql.Context, dest []byte, v interface{}) (sqltypes.Value, error) {
	if v == nil {
		return sqltypes.NULL, nil
	}
	value, err := t.ioOutputFunc(ctx, v)
	if err != nil {
		return sqltypes.Value{}, err
	}
	return t.sqlFunc(value)
}

// String implements the DoltgresType interface.
func (t Type) String() string {
	return t.stringName
}

// ToArrayType implements the DoltgresType interface.
func (t Type) ToArrayType() DoltgresArrayType {
	return t.toArrayTypeFunc()
}

// Type implements the DoltgresType interface.
func (t Type) Type() query.Type {
	return t.queryType
}

// ValueType implements the DoltgresType interface.
func (t Type) ValueType() reflect.Type {
	return t.valueType
}

// Zero implements the DoltgresType interface.
func (t Type) Zero() interface{} {
	return t.zero
}

// SerializeType implements the DoltgresType interface.
func (t Type) SerializeType() ([]byte, error) {
	if t.serializeTypeFunc == nil {
		return t.serializationID.ToByteSlice(0), nil
	}
	return t.serializeTypeFunc()
}

// deserializeType implements the DoltgresType interface.
func (t Type) deserializeType(version uint16, metadata []byte) (DoltgresType, error) {
	if t.deserializeTypeFunc == nil {
		switch version {
		case 0:
			return t, nil
		default:
			return nil, fmt.Errorf("version %d is not yet supported for %s", version, t.String())
		}
	}
	return t.deserializeTypeFunc(version, metadata)
}

// SerializeValue implements the types.ExtendedType interface.
func (t Type) SerializeValue(val any) ([]byte, error) {
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
func (t Type) DeserializeValue(val []byte) (any, error) {
	return t.deserializeValueFunc(val)
}
