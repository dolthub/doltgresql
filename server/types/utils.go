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
	"github.com/dolthub/vitess/go/vt/proto/query"
	"gopkg.in/src-d/go-errors.v1"

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

// IoOutput is the implementation for IoOutput that is being set from another package to avoid circular dependencies.
var IoOutput func(ctx *sql.Context, t DoltgresType, val any) (string, error)

// IoReceive is the implementation for IoOutput that is being set from another package to avoid circular dependencies.
var IoReceive func(ctx *sql.Context, t DoltgresType, val any) (any, error)

// IoSend is the implementation for IoOutput that is being set from another package to avoid circular dependencies.
var IoSend func(ctx *sql.Context, t DoltgresType, val any) ([]byte, error)

// TypModOut is the implementation for IoOutput that is being set from another package to avoid circular dependencies.
var TypModOut func(ctx *sql.Context, t DoltgresType, val int32) (string, error)

// IoCompare is the implementation for IoOutput that is being set from another package to avoid circular dependencies.
var IoCompare func(ctx *sql.Context, t DoltgresType, v1, v2 any) (int32, error)

// SQL is the implementation for IoOutput that is being set from another package to avoid circular dependencies.
var SQL func(ctx *sql.Context, t DoltgresType, val any) (string, error)

// FromGmsType returns a DoltgresType that is most similar to the given GMS type.
// It returns UNKNOWN type for GMS types that are not handled.
func FromGmsType(typ sql.Type) DoltgresType {
	dt, err := FromGmsTypeToDoltgresType(typ)
	if err != nil {
		return Unknown
	}
	return dt
}

// FromGmsTypeToDoltgresType returns a DoltgresType that is most similar to the given GMS type.
// It errors if GMS type is not handled.
func FromGmsTypeToDoltgresType(typ sql.Type) (DoltgresType, error) {
	switch typ.Type() {
	case query.Type_INT8, query.Type_INT16:
		// Special treatment for boolean types when we can detect them
		if typ == types.Boolean {
			return Bool, nil
		}
		return Int16, nil
	case query.Type_INT24, query.Type_INT32:
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
		return DoltgresType{}, fmt.Errorf("encountered a GMS type that cannot be handled")
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
