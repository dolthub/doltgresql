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
	"math"
	"reflect"
	"time"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/types"
	"github.com/dolthub/vitess/go/sqltypes"
	"github.com/dolthub/vitess/go/vt/proto/query"
	"github.com/lib/pq/oid"
	"github.com/shopspring/decimal"
	"gopkg.in/src-d/go-errors.v1"

	"github.com/dolthub/doltgresql/postgres/parser/duration"
	"github.com/dolthub/doltgresql/postgres/parser/uuid"
	"github.com/dolthub/doltgresql/utils"
)

var ErrTypeAlreadyExists = errors.NewKind(`type "%s" already exists`)
var ErrTypeDoesNotExist = errors.NewKind(`type "%s" does not exist`)
var ErrUnhandledType = errors.NewKind(`%s: unhandled type: %T`)
var ErrInvalidSyntaxForType = errors.NewKind(`invalid input syntax for type %s: %q`)
var ErrValueIsOutOfRangeForType = errors.NewKind(`value %q is out of range for type %s`)
var ErrTypmodArrayMustBe1D = errors.NewKind(`typmod array must be one-dimensional`)
var ErrInvalidTypeModifier = errors.NewKind(`invalid %s type modifier`)

// DoltgresType represents a single type.
type DoltgresType struct {
	OID           uint32
	Name          string
	Schema        string // TODO: should be `uint32`.
	Owner         string // TODO: should be `uint32`.
	TypLength     int16
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
	TypCollation  uint32
	DefaulBin     string // for Domain types
	Default       string
	Acl           []string               // TODO: list of privileges
	Checks        []*sql.CheckDefinition // TODO: this is not part of `pg_type` instead `pg_constraint` for Domain types.
	AttTypMod     int32                  // TODO: should be stored in pg_attribute.atttypmod
	internalName  string                 // TODO: Name and internalName differ for some types. e.g.: "int2" vs "smallint"

	// These are for internal use
	isSerial            bool // TODO: to replace serial types
	isUnresolved        bool
	BaseTypeForInternal uint32 // used for INTERNAL type only
}

var IoOutput func(ctx *sql.Context, t DoltgresType, val any) (string, error)
var IoReceive func(ctx *sql.Context, t DoltgresType, val any) (any, error)
var IoSend func(ctx *sql.Context, t DoltgresType, val any) ([]byte, error)
var IoCompare func(ctx *sql.Context, t DoltgresType, v1, v2 any) (int, error)
var SQL func(ctx *sql.Context, t DoltgresType, val any) (string, error)

var _ types.ExtendedType = DoltgresType{}

func NewUnresolvedDoltgresType(sch, name string) DoltgresType {
	return DoltgresType{
		Name:         name,
		Schema:       sch,
		isUnresolved: true,
	}
}

// ArrayBaseType returns a base type of this array type if it exists.
// If this type is not an array type, it returns false.
func (t DoltgresType) ArrayBaseType() (DoltgresType, bool) {
	if !t.IsArrayType() {
		return DoltgresType{}, false
	}
	elem, ok := OidToBuildInDoltgresType[t.Elem]
	elem.AttTypMod = t.AttTypMod
	return elem, ok
}

// CharacterSet implements the sql.StringType interface.
func (t DoltgresType) CharacterSet() sql.CharacterSetID {
	// TODO: only varchar has charset info.
	if t.OID == uint32(oid.T_varchar) {
		return sql.CharacterSet_binary // TODO
	} else {
		return sql.CharacterSet_Unspecified
	}
}

// Collation implements the sql.StringType interface.
func (t DoltgresType) Collation() sql.CollationID {
	// TODO: only varchar has collation info.
	if t.OID == uint32(oid.T_varchar) {
		return sql.Collation_Default // TODO
	} else {
		return sql.Collation_Unspecified
	}
}

// CollationCoercibility implements the types.ExtendedType interface.
func (t DoltgresType) CollationCoercibility(ctx *sql.Context) (collation sql.CollationID, coercibility byte) {
	return sql.Collation_binary, 5
}

// Compare implements the types.ExtendedType interface.
func (t DoltgresType) Compare(v1 interface{}, v2 interface{}) (int, error) {
	return IoCompare(sql.NewEmptyContext(), t, v1, v2)
}

// Convert implements the types.ExtendedType interface.
func (t DoltgresType) Convert(v interface{}) (interface{}, sql.ConvertInRange, error) {
	if v == nil {
		return nil, sql.InRange, nil
	}
	// TODO: should assignment cast, but need info on 'from type'
	switch oid.Oid(t.OID) {
	case oid.T_bool:
		if _, ok := v.(bool); ok {
			return v, sql.InRange, nil
		}
	case oid.T_bytea:
		if _, ok := v.([]byte); ok {
			return v, sql.InRange, nil
		}
	case oid.T_bpchar, oid.T_char, oid.T_json, oid.T_name, oid.T_text, oid.T_unknown, oid.T_varchar:
		if _, ok := v.(string); ok {
			return v, sql.InRange, nil
		}
	case oid.T_date, oid.T_time, oid.T_timestamp, oid.T_timestamptz, oid.T_timetz:
		if _, ok := v.(time.Time); ok {
			return v, sql.InRange, nil
		}
	case oid.T_float4:
		if _, ok := v.(float32); ok {
			return v, sql.InRange, nil
		}
	case oid.T_float8:
		if _, ok := v.(float64); ok {
			return v, sql.InRange, nil
		}
	case oid.T_int2:
		if _, ok := v.(int16); ok {
			return v, sql.InRange, nil
		}
	case oid.T_int4:
		if _, ok := v.(int32); ok {
			return v, sql.InRange, nil
		}
	case oid.T_int8:
		if _, ok := v.(int64); ok {
			return v, sql.InRange, nil
		}
	case oid.T_interval:
		if _, ok := v.(duration.Duration); ok {
			return v, sql.InRange, nil
		}
	case oid.T_jsonb:
		if _, ok := v.(JsonDocument); ok {
			return v, sql.InRange, nil
		}
	case oid.T_oid, oid.T_regclass, oid.T_regproc, oid.T_regtype, oid.T_xid:
		if _, ok := v.(uint32); ok {
			return v, sql.InRange, nil
		}
	case oid.T_uuid:
		if _, ok := v.(uuid.UUID); ok {
			return v, sql.InRange, nil
		}
	default:
		return v, sql.InRange, nil
	}
	return nil, sql.OutOfRange, fmt.Errorf("%s: unhandled type: %T", t.String(), v)
}

// DomainUnderlyingBaseType returns an underlying base type of this domain type.
// It can be a nested domain type, so it recursively searches for a valid base type.
func (t DoltgresType) DomainUnderlyingBaseType() DoltgresType {
	// TODO: handle user-defined type
	bt, ok := OidToBuildInDoltgresType[t.BaseTypeOID]
	if !ok {
		panic(fmt.Sprintf("unable to get DoltgresType from OID: %v", t.BaseTypeOID))
	}
	if bt.TypType == TypeType_Domain {
		return bt.DomainUnderlyingBaseType()
	} else {
		return bt
	}
}

// Equals implements the types.ExtendedType interface.
func (t DoltgresType) Equals(otherType sql.Type) bool {
	if otherExtendedType, ok := otherType.(DoltgresType); ok {
		return bytes.Equal(t.Serialize(), otherExtendedType.Serialize())
	}
	return false
}

// FormatValue implements the types.ExtendedType interface.
func (t DoltgresType) FormatValue(val any) (string, error) {
	if val == nil {
		return "", nil
	}
	return IoOutput(sql.NewEmptyContext(), t, val)
}

// IsArrayType returns true if the type is of 'array' category
func (t DoltgresType) IsArrayType() bool {
	return t.TypCategory == TypeCategory_ArrayTypes && t.Elem != 0
}

// IsEmptyType returns true if the type has no valid OID or Name.
func (t DoltgresType) IsEmptyType() bool {
	return t.OID == 0 && t.Name == ""
}

// IsPolymorphicType types are special built-in pseudo-types
// that are used during function resolution to allow a function
// to handle multiple types from a single definition.
// All polymorphic types have "any" as a prefix.
// The exception is the "any" type, which is not a polymorphic type.
func (t DoltgresType) IsPolymorphicType() bool {
	return t.TypType == TypeType_Pseudo
}

// IsResolvedType whether the type is resolved and has complete information.
// This is used to resolve types during analyzing when non-built-in type is used.
func (t DoltgresType) IsResolvedType() bool {
	return !t.isUnresolved
}

// IsSerialType returns whether the type is serial type.
// This is true for int16serial, int32serial and int64serial types.
func (t DoltgresType) IsSerialType() bool {
	return t.isSerial
}

// IsValidForPolymorphicType returns whether the given type is valid for the calling polymorphic type.
func (t DoltgresType) IsValidForPolymorphicType(target DoltgresType) bool {
	if !t.IsPolymorphicType() {
		return false
	}
	switch oid.Oid(t.OID) {
	case oid.T_anyarray:
		return target.TypCategory == TypeCategory_ArrayTypes
	case oid.T_anynonarray:
		return target.TypCategory != TypeCategory_ArrayTypes
	case oid.T_anyelement, oid.T_any, oid.T_internal:
		return true
	default:
		return false
	}
}

// Length implements the sql.StringType interface.
func (t DoltgresType) Length() int64 {
	if t.OID == uint32(oid.T_varchar) {
		if t.AttTypMod == -1 {
			return StringUnbounded
		} else {
			return int64(GetMaxCharsFromTypmod(t.AttTypMod))
		}
	}
	return int64(0)
}

// MaxByteLength implements the sql.StringType interface.
func (t DoltgresType) MaxByteLength() int64 {
	if t.OID == uint32(oid.T_varchar) {
		return t.Length() * 4
	} else if t.TypLength == -1 {
		return StringUnbounded
	} else {
		return int64(t.TypLength) * 4
	}
}

// MaxCharacterLength implements the sql.StringType interface.
func (t DoltgresType) MaxCharacterLength() int64 {
	if t.OID == uint32(oid.T_varchar) {
		return t.Length()
	} else if t.TypLength == -1 {
		return StringUnbounded
	} else {
		return int64(t.TypLength)
	}
}

// MaxSerializedWidth implements the types.ExtendedType interface.
func (t DoltgresType) MaxSerializedWidth() types.ExtendedTypeSerializedWidth {
	// TODO: need better way to get accurate result
	switch t.TypCategory {
	case TypeCategory_ArrayTypes:
		return types.ExtendedTypeSerializedWidth_Unbounded
	case TypeCategory_BooleanTypes:
		return types.ExtendedTypeSerializedWidth_64K
	case TypeCategory_CompositeTypes, TypeCategory_EnumTypes, TypeCategory_GeometricTypes, TypeCategory_NetworkAddressTypes,
		TypeCategory_RangeTypes, TypeCategory_PseudoTypes, TypeCategory_UserDefinedTypes, TypeCategory_BitStringTypes,
		TypeCategory_InternalUseTypes:
		return types.ExtendedTypeSerializedWidth_Unbounded
	case TypeCategory_DateTimeTypes:
		return types.ExtendedTypeSerializedWidth_64K
	case TypeCategory_NumericTypes:
		return types.ExtendedTypeSerializedWidth_64K
	case TypeCategory_StringTypes, TypeCategory_UnknownTypes:
		if t.OID == uint32(oid.T_varchar) {
			l := t.Length()
			if l != StringUnbounded && l <= stringInline {
				return types.ExtendedTypeSerializedWidth_64K
			}
		}
		return types.ExtendedTypeSerializedWidth_Unbounded
	case TypeCategory_TimespanTypes:
		return types.ExtendedTypeSerializedWidth_64K
	default:
		// shouldn't happen
		return types.ExtendedTypeSerializedWidth_Unbounded
	}
}

// MaxTextResponseByteLength implements the types.ExtendedType interface.
func (t DoltgresType) MaxTextResponseByteLength(ctx *sql.Context) uint32 {
	if t.OID == uint32(oid.T_varchar) {
		l := t.Length()
		if l == StringUnbounded {
			return math.MaxUint32
		} else {
			return uint32(l * 4)
		}
	} else if t.TypLength == -1 {
		return math.MaxUint32
	} else {
		return uint32(t.TypLength)
	}
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

	if t.TypCategory == TypeCategory_StringTypes {
		return serializedStringCompare(v1, v2), nil
	}

	return bytes.Compare(v1, v2), nil
}

// SQL implements the types.ExtendedType interface.
func (t DoltgresType) SQL(ctx *sql.Context, dest []byte, v interface{}) (sqltypes.Value, error) {
	if v == nil {
		return sqltypes.NULL, nil
	}
	value, err := SQL(ctx, t, v)
	if err != nil {
		return sqltypes.Value{}, err
	}

	// TODO: check type
	return sqltypes.MakeTrusted(sqltypes.Text, types.AppendAndSliceBytes(dest, []byte(value))), nil
}

// String implements the types.ExtendedType interface.
func (t DoltgresType) String() string {
	if t.internalName == "" {
		return t.Name
	}
	return t.internalName
}

// ToArrayType returns an array type and whether it exists.
// For array types, ToArrayType causes them to return themselves.
func (t DoltgresType) ToArrayType() (DoltgresType, bool) {
	if t.IsArrayType() {
		return t, true
	}
	if t.Array == 0 {
		return DoltgresType{}, false
	}
	arr, ok := OidToBuildInDoltgresType[t.Array]
	arr.AttTypMod = t.AttTypMod
	return arr, ok
}

// Type implements the types.ExtendedType interface.
func (t DoltgresType) Type() query.Type {
	// TODO: need better way to get accurate result
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
		switch oid.Oid(t.OID) {
		case oid.T_float4:
			return sqltypes.Float32
		case oid.T_float8:
			return sqltypes.Float64
		case oid.T_int2:
			return sqltypes.Int16
		case oid.T_int4:
			return sqltypes.Int32
		case oid.T_int8:
			return sqltypes.Int64
		case oid.T_numeric:
			return sqltypes.Decimal
		case oid.T_oid:
			return sqltypes.Uint32
		case oid.T_regclass, oid.T_regproc, oid.T_regtype:
			return sqltypes.Text
		default:
			// TODO
			return sqltypes.Int64
		}
	case TypeCategory_StringTypes, TypeCategory_UnknownTypes:
		if t.OID == uint32(oid.T_varchar) {
			return sqltypes.VarChar
		}
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
	// TODO: need better way to get accurate result
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
		switch oid.Oid(t.OID) {
		case oid.T_float4:
			return float32(0)
		case oid.T_float8:
			return float64(0)
		case oid.T_int2:
			return int16(0)
		case oid.T_int4:
			return int32(0)
		case oid.T_int8:
			return int64(0)
		case oid.T_numeric:
			return decimal.Zero
		case oid.T_oid, oid.T_regclass, oid.T_regproc, oid.T_regtype:
			return uint32(0)
		default:
			// TODO
			return int64(0)
		}
	case TypeCategory_StringTypes, TypeCategory_UnknownTypes:
		return ""
	case TypeCategory_TimespanTypes:
		return duration.MakeDuration(0, 0, 0)
	default:
		// shouldn't happen
		return any(nil)
	}
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
	// TODO: use converted value or not needed?
	return IoSend(sql.NewEmptyContext(), t, converted)
}

// DeserializeValue implements the types.ExtendedType interface.
func (t DoltgresType) DeserializeValue(val []byte) (any, error) {
	if len(val) == 0 {
		return nil, nil
	}
	return IoReceive(sql.NewEmptyContext(), t, val)
}

// serializedStringCompare handles the efficient comparison of two strings that have been serialized using utils.Writer.
// The writer writes the string by prepending the string length, which prevents direct comparison of the byte slices. We
// thus read the string length manually, and extract the byte slices without converting to a string. This function
// assumes that neither byte slice is nil nor empty.
func serializedStringCompare(v1 []byte, v2 []byte) int {
	readerV1 := utils.NewReader(v1)
	readerV2 := utils.NewReader(v2)
	v1Bytes := utils.AdvanceReader(readerV1, readerV1.VariableUint())
	v2Bytes := utils.AdvanceReader(readerV2, readerV2.VariableUint())
	return bytes.Compare(v1Bytes, v2Bytes)
}
