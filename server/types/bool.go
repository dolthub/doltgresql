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
	"reflect"
	"strings"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/types"
	"github.com/dolthub/vitess/go/sqltypes"
	"github.com/lib/pq/oid"
)

var Bool = Type{
	Oid:           uint32(oid.T_bool),
	Name:          "bool",
	Schema:        "pg_catalog",
	Owner:         "doltgres", // TODO
	Length:        int16(1),
	PassedByVal:   true,
	Typ:           TypeType_Base,
	TypCategory:   TypeCategory_BooleanTypes,
	IsPreferred:   true,
	IsDefined:     true,
	Delimiter:     ",",
	RelID:         0,
	SubscriptFunc: "-",
	Elem:          0,
	Array:         uint32(oid.T__bool),
	InputFunc:     "boolin",
	OutputFunc:    "boolout",
	ReceiveFunc:   "boolrecv",
	SendFunc:      "boolsend",
	ModInFunc:     "-",
	ModOutFunc:    "-",
	AnalyzeFunc:   "-",
	Align:         TypeAlignment_Char,
	Storage:       TypeStorage_Plain,
	NotNull:       false,
	BaseTypeOID:   0,
	TypMod:        -1,
	NDims:         0,
	Collation:     0,
	DefaulBin:     "",
	Default:       "",
	Acl:           "",
	Checks:        nil,

	baseID:          DoltgresTypeBaseID_Bool,
	serializationID: SerializationID_Bool,
	compareFunc: func(converted1 interface{}, converted2 interface{}) (int, error) {
		ab := converted1.(bool)
		bb := converted2.(bool)
		if ab == bb {
			return 0, nil
		} else if !ab {
			return -1, nil
		} else {
			return 1, nil
		}
	},
	convertFunc: func(v interface{}) (interface{}, sql.ConvertInRange, error) {
		switch v := v.(type) {
		case bool:
			return v, sql.InRange, nil
		case nil:
			return nil, sql.InRange, nil
		default:
			return nil, sql.OutOfRange, ErrUnhandledType.New("boolean", v)
		}
	},
	ioInputFunc: func(ctx *sql.Context, input string) (any, error) {
		input = strings.TrimSpace(strings.ToLower(input))
		if input == "true" || input == "t" || input == "yes" || input == "on" || input == "1" {
			return true, nil
		} else if input == "false" || input == "f" || input == "no" || input == "off" || input == "0" {
			return false, nil
		} else {
			return nil, ErrInvalidSyntaxForType.New("boolean", input)
		}
	},
	ioOutputFunc: func(ctx *sql.Context, converted any) (string, error) {
		if converted.(bool) {
			return "true", nil
		} else {
			return "false", nil
		}
	},
	isUnbounded:               false,
	maxSerializedWidth:        types.ExtendedTypeSerializedWidth_64K,
	maxTextResponseByteLength: 1,
	serializedCompareFunc: func(v1 []byte, v2 []byte) (int, error) {
		if len(v1) == 0 && len(v2) == 0 {
			return 0, nil
		} else if len(v1) > 0 && len(v2) == 0 {
			return 1, nil
		} else if len(v1) == 0 && len(v2) > 0 {
			return -1, nil
		}

		if v1[0] == v2[0] {
			return 0, nil
		} else if v1[0] == 0 {
			return -1, nil
		} else {
			return 1, nil
		}
	},
	sqlFunc: func(ioOutputStr string) (sqltypes.Value, error) {
		valBytes := types.AppendAndSliceBytes(nil, []byte{ioOutputStr[0]})
		return sqltypes.MakeTrusted(sqltypes.Text, valBytes), nil
	},
	stringName: "boolean",
	toArrayTypeFunc: func() DoltgresArrayType {
		return nil
	},
	queryType:           sqltypes.Text,
	valueType:           reflect.TypeOf(bool(false)),
	zero:                false,
	serializeTypeFunc:   nil,
	deserializeTypeFunc: nil,
	serializeValueFunc: func(converted any) ([]byte, error) {
		if converted.(bool) {
			return []byte{1}, nil
		} else {
			return []byte{0}, nil
		}
	},
	deserializeValueFunc: func(val []byte) (any, error) {
		if len(val) == 0 {
			return nil, nil
		}
		return val[0] != 0, nil
	},
}
