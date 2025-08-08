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
	"cmp"
	"context"
	"fmt"
	"math"
	"reflect"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/types"
	"github.com/dolthub/vitess/go/sqltypes"
	"github.com/dolthub/vitess/go/vt/proto/query"
	"github.com/shopspring/decimal"

	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/postgres/parser/duration"
	"github.com/dolthub/doltgresql/postgres/parser/uuid"
	"github.com/dolthub/doltgresql/utils"
)

// DoltgresType represents a single type. Many of these fields map directly to the type definitions in the pg_types
// system table. See https://www.postgresql.org/docs/current/catalog-pg-type.html for more information.
//
// TODO: the serialization logic always serializes every field for built-in types, which is kind of silly. They are
// effectively hard-coded. We could serialize much more cheaply by only serializing values which can't be derived
// (for custom types) and hard-coding everything else.
type DoltgresType struct {
	ID          id.Type
	TypType     TypeType
	TypCategory TypeCategory
	TypLength   int16
	PassedByVal bool
	IsPreferred bool
	IsDefined   bool
	Delimiter   string

	RelID         id.Id // for Composite types
	SubscriptFunc uint32
	Elem          id.Type
	Array         id.Type
	InputFunc     uint32 // for deserializing a text representation
	OutputFunc    uint32 // for serializing a text representation
	ReceiveFunc   uint32 // for deserializing a binary representation
	SendFunc      uint32 // for serializing a binary representation
	ModInFunc     uint32
	ModOutFunc    uint32
	AnalyzeFunc   uint32
	Align         TypeAlignment
	Storage       TypeStorage

	NotNull      bool    // for Domain types
	BaseTypeID   id.Type // for Domain types
	TypMod       int32   // for Domain types
	NDims        int32   // for Domain types
	TypCollation id.Collation
	DefaulBin    string // for Domain types
	Default      string
	Acl          []string // TODO: list of privileges

	// Below are not part of pg_type fields
	Checks         []*sql.CheckDefinition // TODO: should be in `pg_constraint` for Domain types
	attTypMod      int32                  // TODO: should be in `pg_attribute.atttypmod`
	CompareFunc    uint32                 // TODO: should be in `pg_amproc`
	InternalName   string                 // Name() and InternalName differ for some types. e.g.: "int2" vs "smallint"
	EnumLabels     map[string]EnumLabel   // TODO: should be in `pg_enum`
	CompositeAttrs []CompositeAttribute   // TODO: should be in `pg_attribute`

	// Below are not stored
	IsSerial            bool    // used for serial types only (e.g.: smallserial)
	IsUnresolved        bool    // used internally to know if a type has been resolved
	BaseTypeForInternal id.Type // used for INTERNAL type only
}

var _ types.ExtendedType = &DoltgresType{}
var _ sql.NullType = &DoltgresType{}
var _ sql.StringType = &DoltgresType{}

// NewUnresolvedDoltgresType returns DoltgresType that is not resolved.
// The type will have the schema and name defined with given values, with IsUnresolved == true.
func NewUnresolvedDoltgresType(sch, name string) *DoltgresType {
	return &DoltgresType{
		ID:           id.NewType(sch, name),
		IsUnresolved: true,
	}
}

// AnalyzeFuncName returns the name that would be displayed in pg_type for the `typanalyze` field.
func (t *DoltgresType) AnalyzeFuncName() string {
	return globalFunctionRegistry.GetString(t.AnalyzeFunc)
}

// ArrayBaseType returns a base type of given array type.
// If this type is not an array type, it returns itself.
func (t *DoltgresType) ArrayBaseType() *DoltgresType {
	if !t.IsArrayType() {
		return t
	}

	var elem *DoltgresType
	var ok bool

	elem, ok = IDToBuiltInDoltgresType[t.Elem]
	if !ok {
		// Some array types have no declared element type for pg_catalog compatibilty, but still have a logical type
		// we return for analysis
		elem, ok = LogicalArrayElementTypes[t.ID]
		if !ok {
			panic(fmt.Sprintf("cannot get base type from: %s", t.Name()))
		}
	}

	newElem := *elem.WithAttTypMod(t.attTypMod)
	return &newElem
}

// CharacterSet implements the sql.StringType interface.
func (t *DoltgresType) CharacterSet() sql.CharacterSetID {
	switch t.ID.TypeName() {
	case "varchar", "text", "name":
		return sql.CharacterSet_binary
	default:
		return sql.CharacterSet_Unspecified
	}
}

// Collation implements the sql.StringType interface.
func (t *DoltgresType) Collation() sql.CollationID {
	switch t.ID.TypeName() {
	case "varchar", "text", "name":
		return sql.Collation_Default
	default:
		return sql.Collation_Unspecified
	}
}

// CollationCoercibility implements the types.ExtendedType interface.
func (t *DoltgresType) CollationCoercibility(ctx *sql.Context) (collation sql.CollationID, coercibility byte) {
	return sql.Collation_binary, 5
}

// Compare implements the types.ExtendedType interface.
func (t *DoltgresType) Compare(ctx context.Context, v1 interface{}, v2 interface{}) (int, error) {
	var err error
	v1, err = sql.UnwrapAny(ctx, v1)
	if err != nil {
		return 0, err
	}
	v2, err = sql.UnwrapAny(ctx, v2)
	if err != nil {
		return 0, err
	}

	// TODO: use IoCompare
	if v1 == nil && v2 == nil {
		return 0, nil
	} else if v1 != nil && v2 == nil {
		return 1, nil
	} else if v1 == nil && v2 != nil {
		return -1, nil
	}

	if t.TypType == TypeType_Enum {
		// TODO: temporary solution to getting the enum type (which has label info) into the 'enum_cmp' function
		qf := globalFunctionRegistry.GetFunction(t.CompareFunc)
		resTypes := qf.ResolvedTypes()
		newFunc := qf.WithResolvedTypes([]*DoltgresType{t, t, resTypes[len(resTypes)-1]})
		i, err := newFunc.(QuickFunction).CallVariadic(nil, v1, v2)
		if err != nil {
			return 0, err
		}
		return int(i.(int32)), nil
	}

	switch ab := v1.(type) {
	case bool:
		bb := v2.(bool)
		if ab == bb {
			return 0, nil
		} else if !ab {
			return -1, nil
		} else {
			return 1, nil
		}
	case float32:
		bb := v2.(float32)
		if ab == bb {
			return 0, nil
		} else if ab < bb {
			return -1, nil
		} else {
			return 1, nil
		}
	case float64:
		bb := v2.(float64)
		if ab == bb {
			return 0, nil
		} else if ab < bb {
			return -1, nil
		} else {
			return 1, nil
		}
	case int16:
		bb := v2.(int16)
		if ab == bb {
			return 0, nil
		} else if ab < bb {
			return -1, nil
		} else {
			return 1, nil
		}
	case int32:
		bb := v2.(int32)
		if ab == bb {
			return 0, nil
		} else if ab < bb {
			return -1, nil
		} else {
			return 1, nil
		}
	case int64:
		bb := v2.(int64)
		if ab == bb {
			return 0, nil
		} else if ab < bb {
			return -1, nil
		} else {
			return 1, nil
		}
	case uint32:
		bb := v2.(uint32)
		if ab == bb {
			return 0, nil
		} else if ab < bb {
			return -1, nil
		} else {
			return 1, nil
		}
	case string:
		bb := v2.(string)
		if ab == bb {
			return 0, nil
		} else if ab < bb {
			return -1, nil
		} else {
			return 1, nil
		}
	case []byte:
		bb := v2.([]byte)
		return bytes.Compare(ab, bb), nil
	case time.Time:
		bb := v2.(time.Time)
		return ab.Compare(bb), nil
	case duration.Duration:
		bb := v2.(duration.Duration)
		return ab.Compare(bb), nil
	case JsonDocument:
		bb := v2.(JsonDocument)
		return JsonValueCompare(ab.Value, bb.Value), nil
	case decimal.Decimal:
		bb := v2.(decimal.Decimal)
		return ab.Cmp(bb), nil
	case uuid.UUID:
		bb := v2.(uuid.UUID)
		return bytes.Compare(ab.GetBytesMut(), bb.GetBytesMut()), nil
	case id.Id:
		return cmp.Compare(id.Cache().ToOID(ab), id.Cache().ToOID(v2.(id.Id))), nil
	case []any:
		if !t.IsArrayType() {
			return 0, errors.Errorf("array value received in Compare for non array type")
		}
		bb := v2.([]any)
		minLength := utils.Min(len(ab), len(bb))
		for i := 0; i < minLength; i++ {
			res, err := t.ArrayBaseType().Compare(ctx, ab[i], bb[i])
			if err != nil {
				return 0, err
			}
			if res != 0 {
				return res, nil
			}
		}
		if len(ab) == len(bb) {
			return 0, nil
		} else if len(ab) < len(bb) {
			return -1, nil
		} else {
			return 1, nil
		}
	default:
		return 0, errors.Errorf("unhandled type %T in Compare", v1)
	}
}

// Convert implements the types.ExtendedType interface.
func (t *DoltgresType) Convert(ctx context.Context, v interface{}) (interface{}, sql.ConvertInRange, error) {
	if v == nil {
		return nil, sql.InRange, nil
	}
	switch t.ID.TypeName() {
	case "bool":
		if _, ok := v.(bool); ok {
			return v, sql.InRange, nil
		}
	case "bytea":
		if _, ok := v.([]byte); ok {
			return v, sql.InRange, nil
		}
	case "bpchar", "char", "json", "name", "text", "unknown", "varchar":
		_, ok, err := sql.Unwrap[string](ctx, v)
		if err != nil {
			return nil, sql.OutOfRange, err
		}
		if ok {
			return v, sql.InRange, nil
		}
	case "date", "time", "timestamp", "timestamptz", "timetz":
		if _, ok := v.(time.Time); ok {
			return v, sql.InRange, nil
		}
	case "float4":
		if _, ok := v.(float32); ok {
			return v, sql.InRange, nil
		}
	case "float8":
		if _, ok := v.(float64); ok {
			return v, sql.InRange, nil
		}
	case "int2":
		if _, ok := v.(int16); ok {
			return v, sql.InRange, nil
		}
	case "int4":
		if _, ok := v.(int32); ok {
			return v, sql.InRange, nil
		}
	case "int8":
		if _, ok := v.(int64); ok {
			return v, sql.InRange, nil
		}
	case "interval":
		if _, ok := v.(duration.Duration); ok {
			return v, sql.InRange, nil
		}
	case "jsonb":
		if _, ok := v.(JsonDocument); ok {
			return v, sql.InRange, nil
		}
	case "oid", "regclass", "regproc", "regtype":
		if _, ok := v.(id.Id); ok {
			return v, sql.InRange, nil
		}
	case "xid":
		if _, ok := v.(uint32); ok {
			return v, sql.InRange, nil
		}
	case "uuid":
		if _, ok := v.(uuid.UUID); ok {
			return v, sql.InRange, nil
		}
	default:
		return v, sql.InRange, nil
	}
	return nil, sql.OutOfRange, ErrUnhandledType.New(t.String(), v)
}

// GetImplicitCast is a reference to the implicit cast logic in the functions/framework package, which we can't use
// here due to import cycles
var GetImplicitCast func(fromType *DoltgresType, toType *DoltgresType) TypeCastFunction

// GetAssignmentCast is a reference to the assignment cast logic in the functions/framework package, which we can't use
// here due to import cycles
var GetAssignmentCast func(fromType *DoltgresType, toType *DoltgresType) TypeCastFunction

// GetExplicitCast is a reference to the explicit cast logic in the functions/framework package, which we can't use
// here due to import cycles
var GetExplicitCast func(fromType *DoltgresType, toType *DoltgresType) TypeCastFunction

// ConvertToType implements the types.ExtendedType interface.
func (t *DoltgresType) ConvertToType(ctx *sql.Context, typ types.ExtendedType, val any) (any, error) {
	dt, ok := typ.(*DoltgresType)
	if !ok {
		return nil, errors.Errorf("expected DoltgresType, got %T", typ)
	}

	castFn := GetAssignmentCast(dt, t)
	if castFn == nil {
		return nil, errors.Errorf("no assignment cast from %s to %s", dt.Name(), t.Name())
	}

	return castFn(ctx, val, t)
}

// DomainUnderlyingBaseType returns an underlying base type of this domain type.
// It can be a nested domain type, so it recursively searches for a valid base type.
func (t *DoltgresType) DomainUnderlyingBaseType() *DoltgresType {
	// TODO: handle user-defined type
	bt, ok := IDToBuiltInDoltgresType[t.BaseTypeID]
	if !ok {
		panic(fmt.Sprintf("unable to get DoltgresType from ID: %s", t.BaseTypeID.AsId().String()))
	}
	if bt.TypType == TypeType_Domain {
		return bt.DomainUnderlyingBaseType()
	} else {
		return bt
	}
}

// Equals implements the types.ExtendedType interface.
func (t *DoltgresType) Equals(otherType sql.Type) bool {
	if otherExtendedType, ok := otherType.(*DoltgresType); ok {
		return bytes.Equal(t.Serialize(), otherExtendedType.Serialize())
	}
	return false
}

// FormatValue implements the types.ExtendedType interface.
func (t *DoltgresType) FormatValue(val any) (string, error) {
	if val == nil {
		return "", nil
	}
	// TODO: need valid sql.Context. This panics on certain types, regclass
	//  https://github.com/dolthub/doltgresql/issues/1142
	return t.IoOutput(nil, val)
}

// GetAttTypMod returns the attTypMod field of the type.
func (t *DoltgresType) GetAttTypMod() int32 {
	return t.attTypMod
}

// InputFuncName returns the name that would be displayed in pg_type for the `typinput` field.
func (t *DoltgresType) InputFuncName() string {
	return globalFunctionRegistry.GetString(t.InputFunc)
}

// IoInput converts input string value to given type value.
func (t *DoltgresType) IoInput(ctx *sql.Context, input string) (any, error) {
	if t.TypType == TypeType_Domain {
		return globalFunctionRegistry.GetFunction(t.InputFunc).CallVariadic(ctx, input, t.BaseTypeID.AsId(), t.attTypMod)
	} else if t.ModInFunc != 0 || t.IsArrayType() {
		if t.Elem != id.NullType {
			return globalFunctionRegistry.GetFunction(t.InputFunc).CallVariadic(ctx, input, t.Elem.AsId(), t.attTypMod)
		} else {
			return globalFunctionRegistry.GetFunction(t.InputFunc).CallVariadic(ctx, input, t.ID.AsId(), t.attTypMod)
		}
	} else if t.TypType == TypeType_Enum {
		return globalFunctionRegistry.GetFunction(t.InputFunc).CallVariadic(ctx, input, t.ID.AsId())
	} else {
		return globalFunctionRegistry.GetFunction(t.InputFunc).CallVariadic(ctx, input)
	}
}

// IoOutput converts given type value to output string.
func (t *DoltgresType) IoOutput(ctx *sql.Context, val any) (string, error) {
	var o any
	var err error
	if t.ModInFunc != 0 || t.IsArrayType() || t.IsCompositeType() {
		send := globalFunctionRegistry.GetFunction(t.OutputFunc)
		resolvedTypes := send.ResolvedTypes()
		resolvedTypes[0] = t
		o, err = send.WithResolvedTypes(resolvedTypes).(QuickFunction).CallVariadic(ctx, val)
	} else {
		o, err = globalFunctionRegistry.GetFunction(t.OutputFunc).CallVariadic(ctx, val)
	}
	if err != nil {
		return "", err
	}
	var ok bool
	os, ok, err := sql.Unwrap[string](ctx, o)
	if !ok {
		return "", errors.Errorf("unexpected type for io output, expected string, got %T", val)
	}
	return os, err
}

// IsArrayType returns true if the type is of 'array' category
func (t *DoltgresType) IsArrayType() bool {
	return (t.TypCategory == TypeCategory_ArrayTypes && t.Elem != id.NullType) ||
		(t.TypCategory == TypeCategory_PseudoTypes && t.ID.TypeName() == "anyarray")
}

// IsCompositeType returns true if the type is a composite type, such as an anonymous record, or a
// user-created composite type.
func (t *DoltgresType) IsCompositeType() bool {
	return t.ID.TypeName() == "record" || t.TypType == TypeType_Composite
}

// IsEmptyType returns true if the type is not valid.
func (t *DoltgresType) IsEmptyType() bool {
	return t == nil
}

// IsPolymorphicType types are special built-in pseudo-types
// that are used during function resolution to allow a function
// to handle multiple types from a single definition.
// All polymorphic types have "any" as a prefix.
// The exception is the "any" type, which is not a polymorphic type.
func (t *DoltgresType) IsPolymorphicType() bool {
	switch t.ID.TypeName() {
	case "anyelement", "anyarray", "anynonarray", "anyenum", "anyrange":
		// TODO: add other polymorphic types
		// https://www.postgresql.org/docs/15/extend-type-system.html#EXTEND-TYPES-POLYMORPHIC-TABLE
		return true
	default:
		return false
	}
}

// IsResolvedType whether the type is resolved and has complete information.
// This is used to resolve types during analyzing when non-built-in type is used.
func (t *DoltgresType) IsResolvedType() bool {
	return !t.IsUnresolved
}

// IsValidForPolymorphicType returns whether the given type is valid for the calling polymorphic type.
func (t *DoltgresType) IsValidForPolymorphicType(target *DoltgresType) bool {
	switch t.ID.TypeName() {
	case "anyelement":
		return true
	case "anyarray":
		return target.TypCategory == TypeCategory_ArrayTypes
	case "anynonarray":
		return target.TypCategory != TypeCategory_ArrayTypes
	case "anyenum":
		return target.TypCategory == TypeCategory_EnumTypes
	case "anyrange":
		return target.TypCategory == TypeCategory_RangeTypes
	default:
		// TODO: add other polymorphic types
		// https://www.postgresql.org/docs/15/extend-type-system.html#EXTEND-TYPES-POLYMORPHIC-TABLE
		return false
	}
}

// Length implements the sql.StringType interface.
func (t *DoltgresType) Length() int64 {
	switch t.ID.TypeName() {
	case "varchar", "bpchar":
		if t.attTypMod == -1 {
			return StringUnbounded
		} else {
			return int64(GetCharLengthFromTypmod(t.attTypMod))
		}
	case "text":
		return StringUnbounded
	case "name":
		return int64(t.TypLength)
	default:
		return int64(0)
	}
}

// MaxByteLength implements the sql.StringType interface.
func (t *DoltgresType) MaxByteLength() int64 {
	if t.ID == VarChar.ID {
		return t.Length() * 4
	} else if t.TypLength == -1 {
		return StringUnbounded
	} else {
		return int64(t.TypLength) * 4
	}
}

// MaxCharacterLength implements the sql.StringType interface.
func (t *DoltgresType) MaxCharacterLength() int64 {
	if t.ID == VarChar.ID {
		return t.Length()
	} else if t.TypLength == -1 {
		return StringUnbounded
	} else {
		return int64(t.TypLength)
	}
}

// MaxSerializedWidth implements the types.ExtendedType interface.
func (t *DoltgresType) MaxSerializedWidth() types.ExtendedTypeSerializedWidth {
	if t.TypLength < 0 {
		// Length will be 0 for any non-string type, as well as unbounded string types
		if t.Length() > 0 {
			return types.ExtendedTypeSerializedWidth_64K
		}
		return types.ExtendedTypeSerializedWidth_Unbounded
	}
	return types.ExtendedTypeSerializedWidth_64K
}

// MaxTextResponseByteLength implements the types.ExtendedType interface.
func (t *DoltgresType) MaxTextResponseByteLength(ctx *sql.Context) uint32 {
	if t.ID == VarChar.ID {
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

// ModInFuncName returns the name that would be displayed in pg_type for the `typmodin` field.
func (t *DoltgresType) ModInFuncName() string {
	return globalFunctionRegistry.GetString(t.ModInFunc)
}

// ModOutFuncName returns the name that would be displayed in pg_type for the `typmodout` field.
func (t *DoltgresType) ModOutFuncName() string {
	return globalFunctionRegistry.GetString(t.ModOutFunc)
}

// Name returns the name of the type.
func (t *DoltgresType) Name() string {
	return t.ID.TypeName()
}

// OutputFuncName returns the name that would be displayed in pg_type for the `typoutput` field.
func (t *DoltgresType) OutputFuncName() string {
	return globalFunctionRegistry.GetString(t.OutputFunc)
}

// Promote implements the types.ExtendedType interface.
func (t *DoltgresType) Promote() sql.Type {
	return t
}

// ReceiveFuncName returns the name that would be displayed in pg_type for the `typreceive` field.
func (t *DoltgresType) ReceiveFuncName() string {
	return globalFunctionRegistry.GetString(t.ReceiveFunc)
}

// Schema returns the schema that the type is contained in.
func (t *DoltgresType) Schema() string {
	return t.ID.SchemaName()
}

// SendFuncName returns the name that would be displayed in pg_type for the `typsend` field.
func (t *DoltgresType) SendFuncName() string {
	return globalFunctionRegistry.GetString(t.SendFunc)
}

// SerializedCompare implements the types.ExtendedType interface.
func (t *DoltgresType) SerializedCompare(ctx context.Context, v1 []byte, v2 []byte) (int, error) {
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

// IsNullType implements the sql.NullType interface.
func (t *DoltgresType) IsNullType() bool {
	return t.ID.TypeName() == "unknown"
}

// SQL implements the types.ExtendedType interface.
func (t *DoltgresType) SQL(ctx *sql.Context, dest []byte, v interface{}) (sqltypes.Value, error) {
	if v == nil {
		return sqltypes.NULL, nil
	}
	value, err := sqlString(ctx, t, v)
	if err != nil {
		return sqltypes.Value{}, err
	}

	// TODO: check type
	return sqltypes.MakeTrusted(sqltypes.Text, types.AppendAndSliceString(dest, value)), nil
}

// String implements the types.ExtendedType interface.
func (t *DoltgresType) String() string {
	str := t.InternalName
	if t.InternalName == "" {
		str = t.Name()
	}
	if t.attTypMod != -1 {
		// TODO: need valid sql.Context
		if l, err := t.TypModOut(nil, t.attTypMod); err == nil {
			str = fmt.Sprintf("%s%s", str, l)
		}
	}
	return str
}

// SubscriptFuncName returns the name that would be displayed in pg_type for the `typsubscript` field.
func (t *DoltgresType) SubscriptFuncName() string {
	return globalFunctionRegistry.GetString(t.SubscriptFunc)
}

// ToArrayType returns an array type of given base type.
// For array types, ToArrayType causes them to return themselves.
func (t *DoltgresType) ToArrayType() *DoltgresType {
	if t.IsArrayType() {
		return t
	}
	arr, ok := IDToBuiltInDoltgresType[t.Array]
	if !ok {
		panic(fmt.Sprintf("cannot get array type from: %s", t.Name()))
	}
	newArr := *arr.WithAttTypMod(t.attTypMod)
	newArr.InternalName = fmt.Sprintf("%s[]", t.String())
	return &newArr
}

// Type implements the types.ExtendedType interface.
func (t *DoltgresType) Type() query.Type {
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
		switch t.ID.TypeName() {
		case "date":
			return sqltypes.Date
		case "time":
			return sqltypes.Time
		default:
			return sqltypes.Timestamp
		}
	case TypeCategory_NumericTypes:
		switch t.ID.TypeName() {
		case "float4":
			return sqltypes.Float32
		case "float8":
			return sqltypes.Float64
		case "int2":
			return sqltypes.Int16
		case "int4":
			return sqltypes.Int32
		case "int8":
			return sqltypes.Int64
		case "numeric":
			return sqltypes.Decimal
		case "oid":
			return sqltypes.VarChar
		case "regclass", "regproc", "regtype":
			return sqltypes.Text
		default:
			// TODO
			return sqltypes.Int64
		}
	case TypeCategory_StringTypes, TypeCategory_UnknownTypes:
		if t.ID.TypeName() == "varchar" {
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

// TypModIn encodes given text array value to type modifier in int32 format.
func (t *DoltgresType) TypModIn(ctx *sql.Context, val []any) (int32, error) {
	if t.ModInFunc == 0 {
		return 0, errors.Errorf("typmodin function for type '%s' doesn't exist", t.Name())
	}
	o, err := globalFunctionRegistry.GetFunction(t.ModInFunc).CallVariadic(ctx, val)
	if err != nil {
		return 0, err
	}
	output, ok := o.(int32)
	if !ok {
		return 0, errors.Errorf(`expected int32, got %T`, output)
	}
	return output, nil
}

// TypModOut decodes type modifier in int32 format to string representation of it.
func (t *DoltgresType) TypModOut(ctx *sql.Context, val int32) (string, error) {
	if t.ModOutFunc == 0 {
		return "", errors.Errorf("typmodout function for type '%s' doesn't exist", t.Name())
	}
	o, err := globalFunctionRegistry.GetFunction(t.ModOutFunc).CallVariadic(ctx, val)
	if err != nil {
		return "", err
	}
	output, ok := o.(string)
	if !ok {
		return "", errors.Errorf(`expected string, got %T`, output)
	}
	return output, nil
}

// ValueType implements the types.ExtendedType interface.
func (t *DoltgresType) ValueType() reflect.Type {
	return reflect.TypeOf(t.Zero())
}

// WithAttTypMod returns a copy of the type with attTypMod
// defined with given value. This function should be used
// to set attTypMod only, as it creates a copy of the type
// to avoid updating the original type.
func (t *DoltgresType) WithAttTypMod(tm int32) *DoltgresType {
	newDt := *t
	newDt.attTypMod = tm
	return &newDt
}

// Zero implements the types.ExtendedType interface.
func (t *DoltgresType) Zero() interface{} {
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
		switch t.ID.TypeName() {
		case "float4":
			return float32(0)
		case "float8":
			return float64(0)
		case "int2":
			return int16(0)
		case "int4":
			return int32(0)
		case "int8":
			return int64(0)
		case "numeric":
			return decimal.Zero
		case "oid", "regclass", "regproc", "regtype":
			return id.Null
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
func (t *DoltgresType) SerializeValue(ctx context.Context, val any) ([]byte, error) {
	if val == nil {
		return nil, nil
	}
	var o any
	var err error
	if t.ModInFunc != 0 || t.IsArrayType() {
		send := globalFunctionRegistry.GetFunction(t.SendFunc)
		resolvedTypes := send.ResolvedTypes()
		resolvedTypes[0] = t
		o, err = send.WithResolvedTypes(resolvedTypes).(QuickFunction).CallVariadic(nil, val)
	} else {
		o, err = globalFunctionRegistry.GetFunction(t.SendFunc).CallVariadic(nil, val)
	}
	if err != nil || o == nil {
		return nil, err
	}
	return o.([]byte), nil
}

// DeserializeValue implements the types.ExtendedType interface.
func (t *DoltgresType) DeserializeValue(ctx context.Context, val []byte) (any, error) {
	if len(val) == 0 {
		return nil, nil
	}
	if t.TypType == TypeType_Domain {
		return globalFunctionRegistry.GetFunction(t.ReceiveFunc).CallVariadic(nil, val, t.BaseTypeID.AsId(), t.attTypMod)
	} else if t.ModInFunc != 0 || t.IsArrayType() {
		if t.Elem != id.NullType {
			return globalFunctionRegistry.GetFunction(t.ReceiveFunc).CallVariadic(nil, val, t.Elem.AsId(), t.attTypMod)
		} else {
			return globalFunctionRegistry.GetFunction(t.ReceiveFunc).CallVariadic(nil, val, t.ID.AsId(), t.attTypMod)
		}
	} else if t.TypType == TypeType_Enum {
		return globalFunctionRegistry.GetFunction(t.ReceiveFunc).CallVariadic(nil, val, t.ID.AsId())
	} else {
		return globalFunctionRegistry.GetFunction(t.ReceiveFunc).CallVariadic(nil, val)
	}
}

// TypeCastFunction is a function that takes a value of a particular kind of type, and returns it as another kind of type.
// The targetType given should match the "To" type used to obtain the cast.
type TypeCastFunction func(ctx *sql.Context, val any, targetType *DoltgresType) (any, error)
