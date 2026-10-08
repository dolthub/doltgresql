// Copyright 2026 Dolthub, Inc.
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

// Tid is a data type used by the hidden `ctid` column. It uses the TidValue type.
var Tid = &DoltgresType{
	ID:                  toInternal("tid"),
	TypLength:           int16(6),
	PassedByVal:         true,
	TypType:             TypeType_Base,
	TypCategory:         TypeCategory_UserDefinedTypes,
	IsPreferred:         false,
	IsDefined:           true,
	Delimiter:           ",",
	RelID:               id.Null,
	SubscriptFunc:       toFuncID("-"),
	Elem:                internalNullType,
	Array:               internalNullType,
	InputFunc:           toFuncID("tidin", toInternal("cstring")),
	OutputFunc:          toFuncID("tidout", toInternal("tid")),
	ReceiveFunc:         toFuncID("tidrecv", toInternal("internal")),
	SendFunc:            toFuncID("tidsend", toInternal("tid")),
	ModInFunc:           toFuncID("-"),
	ModOutFunc:          toFuncID("-"),
	AnalyzeFunc:         toFuncID("-"),
	Align:               TypeAlignment_Short,
	Storage:             TypeStorage_Plain,
	NotNull:             false,
	BaseTypeType:        internalNullType,
	TypMod:              -1,
	NDims:               0,
	TypCollation:        id.NullCollation,
	DefaulBin:           "",
	Default:             "",
	Acl:                 nil,
	Checks:              nil,
	attTypMod:           -1,
	CompareFunc:         toFuncID("bttidcmp", toInternal("tid"), toInternal("tid")),
	SerializationFunc:   serializeTypeTid,
	DeserializationFunc: deserializeTypeTid,
}

// TidValue represents a value for the `Tid` type.
type TidValue struct {
	Block  uint32
	Offset uint16
}

// serializeTypeTid handles serialization from the standard representation to our serialized representation that is
// written in Dolt.
func serializeTypeTid(ctx *sql.Context, t *DoltgresType, val any) ([]byte, error) {
	retVal := make([]byte, 6)
	tidValue := val.(TidValue)
	binary.BigEndian.PutUint32(retVal[:4], tidValue.Block)
	binary.BigEndian.PutUint16(retVal[4:], tidValue.Offset)
	return retVal, nil
}

// deserializeTypeTid handles deserialization from the Dolt serialized format to our standard representation used by
// expressions and nodes.
func deserializeTypeTid(ctx *sql.Context, t *DoltgresType, data []byte) (any, error) {
	if len(data) == 0 {
		return nil, nil
	}
	return TidValue{
		Block:  binary.BigEndian.Uint32(data[:4]),
		Offset: binary.BigEndian.Uint16(data[4:]),
	}, nil
}
