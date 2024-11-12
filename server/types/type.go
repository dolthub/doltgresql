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

	"github.com/dolthub/doltgresql/postgres/parser/duration"
	"github.com/dolthub/doltgresql/postgres/parser/uuid"
)

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
	Acl           []string // TODO: list of privileges

	// Below are not part of pg_type fields
	Checks       []*sql.CheckDefinition // TODO: should be in `pg_constraint` for Domain types
	AttTypMod    int32                  // TODO: should be in `pg_attribute.atttypmod`
	CompareFunc  string                 // TODO: should be in `pg_amproc`
	InternalName string                 // Name and InternalName differ for some types. e.g.: "int2" vs "smallint"

	// Below are not stored
	IsSerial            bool   // used for serial types only (e.g.: smallserial)
	BaseTypeForInternal uint32 // used for INTERNAL type only
}

var _ types.ExtendedType = DoltgresType{}

// NewUnresolvedDoltgresType returns DoltgresType that is not resolved.
// The type will have 0 as OID and the schema and name defined with given values.
func NewUnresolvedDoltgresType(sch, name string) DoltgresType {
	return DoltgresType{
		OID:    0,
		Name:   name,
		Schema: sch,
	}
}

// ArrayBaseType returns a base type of given array type.
// If this type is not an array type, it returns itself.
func (t DoltgresType) ArrayBaseType() DoltgresType {
	if !t.IsArrayType() {
		return t
	}
	elem, ok := OidToBuildInDoltgresType[t.Elem]
	if !ok {
		panic(fmt.Sprintf("cannot get base type from: %s", t.Name))
	}
	elem.AttTypMod = t.AttTypMod
	return elem
}

// CharacterSet implements the sql.StringType interface.
func (t DoltgresType) CharacterSet() sql.CharacterSetID {
	switch oid.Oid(t.OID) {
	case oid.T_varchar, oid.T_text, oid.T_name:
		return sql.CharacterSet_binary
	default:
		return sql.CharacterSet_Unspecified
	}
}

// Collation implements the sql.StringType interface.
func (t DoltgresType) Collation() sql.CollationID {
	switch oid.Oid(t.OID) {
	case oid.T_varchar, oid.T_text, oid.T_name:
		return sql.Collation_Default
	default:
		return sql.Collation_Unspecified
	}
}

// CollationCoercibility implements the types.ExtendedType interface.
func (t DoltgresType) CollationCoercibility(ctx *sql.Context) (collation sql.CollationID, coercibility byte) {
	return sql.Collation_binary, 5
}

// Compare implements the types.ExtendedType interface.
func (t DoltgresType) Compare(v1 interface{}, v2 interface{}) (int, error) {
	res, err := IoCompare(sql.NewEmptyContext(), t, v1, v2)
	return int(res), err
}

// Convert implements the types.ExtendedType interface.
func (t DoltgresType) Convert(v interface{}) (interface{}, sql.ConvertInRange, error) {
	if v == nil {
		return nil, sql.InRange, nil
	}
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
	return nil, sql.OutOfRange, ErrUnhandledType.New(t.String(), v)
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
	switch oid.Oid(t.OID) {
	case oid.T_anyelement, oid.T_anyarray, oid.T_anynonarray:
		// TODO: add other polymorphic types
		// https://www.postgresql.org/docs/15/extend-type-system.html#EXTEND-TYPES-POLYMORPHIC-TABLE
		return true
	default:
		return false
	}
}

// IsResolvedType whether the type is resolved and has complete information.
// This is used to resolve types during analyzing when non-built-in type is used.
func (t DoltgresType) IsResolvedType() bool {
	// temporary serial types have 0 OID but are resolved.
	return t.OID != 0 || t.IsSerial
}

// IsValidForPolymorphicType returns whether the given type is valid for the calling polymorphic type.
func (t DoltgresType) IsValidForPolymorphicType(target DoltgresType) bool {
	switch oid.Oid(t.OID) {
	case oid.T_anyelement:
		return true
	case oid.T_anyarray:
		return target.TypCategory == TypeCategory_ArrayTypes
	case oid.T_anynonarray:
		return target.TypCategory != TypeCategory_ArrayTypes
	default:
		// TODO: add other polymorphic types
		// https://www.postgresql.org/docs/15/extend-type-system.html#EXTEND-TYPES-POLYMORPHIC-TABLE
		return false
	}
}

// Length implements the sql.StringType interface.
func (t DoltgresType) Length() int64 {
	switch oid.Oid(t.OID) {
	case oid.T_varchar:
		if t.AttTypMod == -1 {
			return StringUnbounded
		} else {
			return int64(GetCharLengthFromTypmod(t.AttTypMod))
		}
	case oid.T_text:
		return StringUnbounded
	case oid.T_name:
		return int64(t.TypLength)
	default:
		return int64(0)
	}
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

// ReceiveFuncExists returns whether IO receive function exists for this type.
func (t DoltgresType) ReceiveFuncExists() bool {
	return t.ReceiveFunc != "-"
}

// SendFuncExists returns whether IO send function exists for this type.
func (t DoltgresType) SendFuncExists() bool {
	return t.SendFunc != "-"
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
	str := t.InternalName
	if t.InternalName == "" {
		str = t.Name
	}
	if t.AttTypMod != -1 {
		if l, err := TypModOut(sql.NewEmptyContext(), t, t.AttTypMod); err == nil {
			str = fmt.Sprintf("%s%s", str, l)
		}
	}
	return str
}

// ToArrayType returns an array type of given base type.
// For array types, ToArrayType causes them to return themselves.
func (t DoltgresType) ToArrayType() DoltgresType {
	if t.IsArrayType() {
		return t
	}
	arr, ok := OidToBuildInDoltgresType[t.Array]
	if !ok {
		panic(fmt.Sprintf("cannot get array type from: %s", t.Name))
	}
	arr.AttTypMod = t.AttTypMod
	return arr
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
	return IoSend(sql.NewEmptyContext(), t, val)
}

// DeserializeValue implements the types.ExtendedType interface.
func (t DoltgresType) DeserializeValue(val []byte) (any, error) {
	if len(val) == 0 {
		return nil, nil
	}
	return IoReceive(sql.NewEmptyContext(), t, val)
}
