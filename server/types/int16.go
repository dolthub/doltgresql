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
	"encoding/binary"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/core/id"
)

// Int16 is an int16.
var Int16 = &DoltgresType{
	ID:                  toInternal("int2"),
	TypLength:           int16(2),
	PassedByVal:         true,
	TypType:             TypeType_Base,
	TypCategory:         TypeCategory_NumericTypes,
	IsPreferred:         false,
	IsDefined:           true,
	Delimiter:           ",",
	RelID:               id.Null,
	SubscriptFunc:       toFuncID("-"),
	Elem:                id.NullType,
	Array:               toInternal("_int2"),
	InputFunc:           toFuncID("int2in", toInternal("cstring")),
	OutputFunc:          toFuncID("int2out", toInternal("int2")),
	ReceiveFunc:         toFuncID("int2recv", toInternal("internal")),
	SendFunc:            toFuncID("int2send", toInternal("int2")),
	ModInFunc:           toFuncID("-"),
	ModOutFunc:          toFuncID("-"),
	AnalyzeFunc:         toFuncID("-"),
	Align:               TypeAlignment_Short,
	Storage:             TypeStorage_Plain,
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
	CompareFunc:         toFuncID("btint2cmp", toInternal("int2"), toInternal("int2")),
	InternalName:        "smallint",
	SerializationFunc:   serializeTypeInt16,
	DeserializationFunc: deserializeTypeInt16,
}

// serializeTypeInt16 handles serialization from the standard representation to our serialized representation that is
// written in Dolt.
func serializeTypeInt16(ctx *sql.Context, t *DoltgresType, val any) ([]byte, error) {
	retVal := make([]byte, 2)
	binary.BigEndian.PutUint16(retVal, uint16(val.(int16))+(1<<15))
	return retVal, nil
}

// deserializeTypeInt16 handles deserialization from the Dolt serialized format to our standard representation used by
// expressions and nodes.
func deserializeTypeInt16(ctx *sql.Context, t *DoltgresType, data []byte) (any, error) {
	if len(data) == 0 {
		return nil, nil
	}
	return int16(binary.BigEndian.Uint16(data) - (1 << 15)), nil
}
