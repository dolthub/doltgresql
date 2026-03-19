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
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/utils"
)

// Bytea is the byte string type.
var Bytea = &DoltgresType{
	ID:                  toInternal("bytea"),
	TypLength:           int16(-1),
	PassedByVal:         false,
	TypType:             TypeType_Base,
	TypCategory:         TypeCategory_UserDefinedTypes,
	IsPreferred:         false,
	IsDefined:           true,
	Delimiter:           ",",
	RelID:               id.Null,
	SubscriptFunc:       toFuncID("-"),
	Elem:                id.NullType,
	Array:               toInternal("_bytea"),
	InputFunc:           toFuncID("byteain", toInternal("cstring")),
	OutputFunc:          toFuncID("byteaout", toInternal("bytea")),
	ReceiveFunc:         toFuncID("bytearecv", toInternal("internal")),
	SendFunc:            toFuncID("byteasend", toInternal("bytea")),
	ModInFunc:           toFuncID("-"),
	ModOutFunc:          toFuncID("-"),
	AnalyzeFunc:         toFuncID("-"),
	Align:               TypeAlignment_Int,
	Storage:             TypeStorage_Extended,
	NotNull:             false,
	BaseTypeID:          id.NullType,
	TypMod:              -1,
	NDims:               0,
	TypCollation:        id.NullCollation,
	DefaulBin:           "",
	Default:             "",
	Acl:                 nil,
	Checks:              nil,
	attTypMod:           -1,
	CompareFunc:         toFuncID("byteacmp", toInternal("bytea"), toInternal("bytea")),
	SerializationFunc:   serializeTypeBytea,
	DeserializationFunc: deserializeTypeBytea,
}

// serializeTypeBytea handles serialization from the standard representation to our serialized representation that is
// written in Dolt.
func serializeTypeBytea(ctx *sql.Context, t *DoltgresType, val any) ([]byte, error) {
	res, err := sql.UnwrapAny(ctx, val)
	if err != nil {
		return nil, err
	}
	str := res.([]byte)
	writer := utils.NewWriter(uint64(len(str) + 4))
	writer.ByteSlice(str)
	return writer.Data(), nil
}

// deserializeTypeBytea handles deserialization from the Dolt serialized format to our standard representation used by
// expressions and nodes.
func deserializeTypeBytea(ctx *sql.Context, t *DoltgresType, data []byte) (any, error) {
	reader := utils.NewReader(data)
	return reader.ByteSlice(), nil
}
