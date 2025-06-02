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
	"strings"

	cerrors "github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/types"
	"github.com/dolthub/vitess/go/vt/proto/query"
	"gopkg.in/src-d/go-errors.v1"

	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/utils"
)

// ErrTypeAlreadyExists is returned when creating given type when it already exists.
var ErrTypeAlreadyExists = errors.NewKind(`type "%s" already exists`)

// ErrTypeDoesNotExist is returned when using given type that does not exist.
var ErrTypeDoesNotExist = errors.NewKind(`type "%s" does not exist`)

// ErrUnhandledType is returned when the type of value does not match given type.
var ErrUnhandledType = errors.NewKind(`%s: unhandled type: %T`)

// ErrInvalidSyntaxForType is returned when the type of value is invalid for given type.
var ErrInvalidSyntaxForType = errors.NewKind(`invalid input syntax for type %s: %q`)

// ErrValueIsOutOfRangeForType is returned when the value is out-of-range for given type.
var ErrValueIsOutOfRangeForType = errors.NewKind(`value %q is out of range for type %s`)

// ErrTypmodArrayMustBe1D is returned when type modifier value is empty array.
var ErrTypmodArrayMustBe1D = errors.NewKind(`typmod array must be one-dimensional`)

// ErrInvalidTypMod is returned when given value is invalid for type modifier.
var ErrInvalidTypMod = errors.NewKind(`invalid %s type modifier`)

// ErrCannotDropSystemType is returned when given type is system/pg_catalog type.
var ErrCannotDropSystemType = errors.NewKind(`cannot drop type %s because it is required by the database system`)

// ErrCannotDropArrayType is returned when given type to drop is array type that is required for its base type.
var ErrCannotDropArrayType = errors.NewKind(`cannot drop type %s because type %s requires it`)

// FromGmsType returns a DoltgresType that is most similar to the given GMS type.
// It returns UNKNOWN type for GMS types that are not handled.
func FromGmsType(typ sql.Type) *DoltgresType {
	dt, err := FromGmsTypeToDoltgresType(typ)
	if err != nil {
		return Unknown
	}
	return dt
}

// FromGmsTypeToDoltgresType returns a DoltgresType that is most similar to the given GMS type.
// It errors if GMS type is not handled.
func FromGmsTypeToDoltgresType(typ sql.Type) (*DoltgresType, error) {
	switch typ.Type() {
	case query.Type_INT8:
		// Special treatment for boolean types when we can detect them
		if typ == types.Boolean {
			return Bool, nil
		}
		return Int32, nil
	case query.Type_INT16, query.Type_INT24, query.Type_INT32:
		return Int32, nil
	case query.Type_INT64:
		return Int64, nil
	case query.Type_UINT8, query.Type_UINT16, query.Type_UINT24, query.Type_UINT32, query.Type_UINT64:
		return Int64, nil
	case query.Type_YEAR:
		return Int16, nil
	case query.Type_FLOAT32:
		return Float32, nil
	case query.Type_FLOAT64:
		return Float64, nil
	case query.Type_DECIMAL:
		return Numeric, nil
	case query.Type_DATE:
		return Date, nil
	case query.Type_TIME:
		return Text, nil
	case query.Type_DATETIME, query.Type_TIMESTAMP:
		return Timestamp, nil
	case query.Type_CHAR, query.Type_VARCHAR, query.Type_TEXT, query.Type_BINARY, query.Type_VARBINARY, query.Type_BLOB:
		return Text, nil
	case query.Type_JSON:
		return Json, nil
	case query.Type_ENUM:
		return Int16, nil
	case query.Type_SET:
		return Int64, nil
	case query.Type_NULL_TYPE, query.Type_GEOMETRY:
		return Unknown, nil
	default:
		return nil, cerrors.Errorf("encountered a GMS type that cannot be handled")
	}
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

// sqlString converts given type value to output string. This is the same as IoOutput function
// with an exception to BOOLEAN type. It returns "t" instead of "true".
func sqlString(ctx *sql.Context, t *DoltgresType, val any) (string, error) {
	if t.IsArrayType() {
		baseType := t.ArrayBaseType()
		return ArrToString(ctx, val.([]any), baseType, true)
	}
	return t.IoOutput(ctx, val)
}

// ArrToString is used for array_out function. |trimBool| parameter allows replacing
// boolean result of "true" to "t" if the function is `Type.SQL()`.
func ArrToString(ctx *sql.Context, arr []any, baseType *DoltgresType, trimBool bool) (string, error) {
	sb := strings.Builder{}
	sb.WriteRune('{')
	for i, v := range arr {
		if i > 0 {
			sb.WriteString(",")
		}
		if v != nil {
			str, err := baseType.IoOutput(ctx, v)
			if err != nil {
				return "", err
			}
			if baseType.ID == Bool.ID && trimBool {
				str = string(str[0])
			}
			sb.WriteString(quoteString(str))
		} else {
			sb.WriteString("NULL")
		}
	}
	sb.WriteRune('}')
	return sb.String(), nil
}

// RecordToString is used for the record_out function, to serialize record values for wire transfer.
// |fields| contains the values to serialize.
func RecordToString(ctx *sql.Context, fields []types.TupleValue) (any, error) {
	sb := strings.Builder{}
	sb.WriteRune('(')
	for i, value := range fields {
		if i > 0 {
			sb.WriteString(",")
		}

		if value.Value == nil {
			continue
		}

		doltgresType, ok := value.Type.(*DoltgresType)
		if !ok {
			return nil, fmt.Errorf(`expected *DoltgresType but found: %T`, value.Type)
		}

		str, err := doltgresType.IoOutput(ctx, value.Value)
		if err != nil {
			return "", err
		}
		if doltgresType.ID == Bool.ID {
			str = string(str[0])
		}

		sb.WriteString(quoteString(str))
	}
	sb.WriteRune(')')

	return sb.String(), nil
}

// quoteString determines if |s| needs to be quoted, by looking for special characters like ' ' or ',',
// and if so, quotes the string and returns it. If quoting is not needed, then |s| is returned as is.
func quoteString(s string) string {
	shouldQuote := false
	for _, r := range s {
		switch r {
		case ' ', ',', '{', '}', '\\', '"':
			shouldQuote = true
		}
	}
	if shouldQuote || strings.EqualFold(s, "NULL") {
		return fmt.Sprintf(`"%s"`, strings.ReplaceAll(s, `"`, `\"`))
	} else {
		return s
	}
}

// toInternal returns an Internal ID for the given type. This is only used for the built-in types, since they all share
// the same schema (pg_catalog).
func toInternal(typeName string) id.Type {
	return id.NewType("pg_catalog", typeName)
}
