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

// Cid is a data type used by the hidden `cmin` and `cmax` columns. It is implemented as an unsigned 32 bit integer.
var Cid = &DoltgresType{
	ID:                  toInternal("cid"),
	TypLength:           int16(4),
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
	InputFunc:           toFuncID("cidin", toInternal("cstring")),
	OutputFunc:          toFuncID("cidout", toInternal("cid")),
	ReceiveFunc:         toFuncID("cidrecv", toInternal("internal")),
	SendFunc:            toFuncID("cidsend", toInternal("cid")),
	ModInFunc:           toFuncID("-"),
	ModOutFunc:          toFuncID("-"),
	AnalyzeFunc:         toFuncID("-"),
	Align:               TypeAlignment_Int,
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
	CompareFunc:         toFuncID("-"),
	SerializationFunc:   serializeTypeCid,
	DeserializationFunc: deserializeTypeCid,
}

// serializeTypeCid handles serialization from the standard representation to our serialized representation that is
// written in Dolt.
func serializeTypeCid(ctx *sql.Context, t *DoltgresType, val any) ([]byte, error) {
	retVal := make([]byte, 4)
	binary.BigEndian.PutUint32(retVal, val.(uint32))
	return retVal, nil
}

// deserializeTypeCid handles deserialization from the Dolt serialized format to our standard representation used by
// expressions and nodes.
func deserializeTypeCid(ctx *sql.Context, t *DoltgresType, data []byte) (any, error) {
	if len(data) == 0 {
		return nil, nil
	}
	return binary.BigEndian.Uint32(data), nil
}
