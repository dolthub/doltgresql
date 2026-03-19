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
	"math"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/core/id"
)

// Float32 is a float32.
var Float32 = &DoltgresType{
	ID:                  toInternal("float4"),
	TypLength:           int16(4),
	PassedByVal:         true,
	TypType:             TypeType_Base,
	TypCategory:         TypeCategory_NumericTypes,
	IsPreferred:         false,
	IsDefined:           true,
	Delimiter:           ",",
	RelID:               id.Null,
	SubscriptFunc:       toFuncID("-"),
	Elem:                id.NullType,
	Array:               toInternal("_float4"),
	InputFunc:           toFuncID("float4in", toInternal("cstring")),
	OutputFunc:          toFuncID("float4out", toInternal("float4")),
	ReceiveFunc:         toFuncID("float4recv", toInternal("internal")),
	SendFunc:            toFuncID("float4send", toInternal("float4")),
	ModInFunc:           toFuncID("-"),
	ModOutFunc:          toFuncID("-"),
	AnalyzeFunc:         toFuncID("-"),
	Align:               TypeAlignment_Int,
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
	CompareFunc:         toFuncID("btfloat4cmp", toInternal("float4"), toInternal("float4")),
	InternalName:        "real",
	SerializationFunc:   serializeTypeFloat32,
	DeserializationFunc: deserializeTypeFloat32,
}

// serializeTypeFloat32 handles serialization from the standard representation to our serialized representation that is
// written in Dolt.
func serializeTypeFloat32(ctx *sql.Context, t *DoltgresType, val any) ([]byte, error) {
	f32 := val.(float32)
	retVal := make([]byte, 4)
	// Make the serialized form trivially comparable using bytes.Compare: https://stackoverflow.com/a/54557561
	unsignedBits := math.Float32bits(f32)
	if f32 >= 0 {
		unsignedBits ^= 1 << 31
	} else {
		unsignedBits = ^unsignedBits
	}
	binary.BigEndian.PutUint32(retVal, unsignedBits)
	return retVal, nil
}

// deserializeTypeFloat32 handles deserialization from the Dolt serialized format to our standard representation used by
// expressions and nodes.
func deserializeTypeFloat32(ctx *sql.Context, t *DoltgresType, data []byte) (any, error) {
	unsignedBits := binary.BigEndian.Uint32(data)
	if unsignedBits&(1<<31) != 0 {
		unsignedBits ^= 1 << 31
	} else {
		unsignedBits = ^unsignedBits
	}
	return math.Float32frombits(unsignedBits), nil
}
