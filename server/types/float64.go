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

// Float64 is a float64.
var Float64 = &DoltgresType{
	ID:                  toInternal("float8"),
	TypLength:           int16(8),
	PassedByVal:         true,
	TypType:             TypeType_Base,
	TypCategory:         TypeCategory_NumericTypes,
	IsPreferred:         true,
	IsDefined:           true,
	Delimiter:           ",",
	RelID:               id.Null,
	SubscriptFunc:       toFuncID("-"),
	Elem:                id.NullType,
	Array:               toInternal("_float8"),
	InputFunc:           toFuncID("float8in", toInternal("cstring")),
	OutputFunc:          toFuncID("float8out", toInternal("float8")),
	ReceiveFunc:         toFuncID("float8recv", toInternal("internal")),
	SendFunc:            toFuncID("float8send", toInternal("float8")),
	ModInFunc:           toFuncID("-"),
	ModOutFunc:          toFuncID("-"),
	AnalyzeFunc:         toFuncID("-"),
	Align:               TypeAlignment_Double,
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
	CompareFunc:         toFuncID("btfloat8cmp", toInternal("float8"), toInternal("float8")),
	InternalName:        "double precision",
	SerializationFunc:   serializeTypeFloat64,
	DeserializationFunc: deserializeTypeFloat64,
}

// serializeTypeFloat64 handles serialization from the standard representation to our serialized representation that is
// written in Dolt.
func serializeTypeFloat64(ctx *sql.Context, t *DoltgresType, val any) ([]byte, error) {
	f64 := val.(float64)
	retVal := make([]byte, 8)
	// Make the serialized form trivially comparable using bytes.Compare: https://stackoverflow.com/a/54557561
	unsignedBits := math.Float64bits(f64)
	if f64 >= 0 {
		unsignedBits ^= 1 << 63
	} else {
		unsignedBits = ^unsignedBits
	}
	binary.BigEndian.PutUint64(retVal, unsignedBits)
	return retVal, nil
}

// deserializeTypeFloat64 handles deserialization from the Dolt serialized format to our standard representation used by
// expressions and nodes.
func deserializeTypeFloat64(ctx *sql.Context, t *DoltgresType, data []byte) (any, error) {
	if len(data) == 0 {
		return nil, nil
	}
	unsignedBits := binary.BigEndian.Uint64(data)
	if unsignedBits&(1<<63) != 0 {
		unsignedBits ^= 1 << 63
	} else {
		unsignedBits = ^unsignedBits
	}
	return math.Float64frombits(unsignedBits), nil
}
