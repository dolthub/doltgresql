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
	"time"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/core/id"
)

// TimestampTZ is the timestamp with a time zone. Precision is unbounded.
var TimestampTZ = &DoltgresType{
	ID:                  toInternal("timestamptz"),
	TypLength:           int16(8),
	PassedByVal:         true,
	TypType:             TypeType_Base,
	TypCategory:         TypeCategory_DateTimeTypes,
	IsPreferred:         true,
	IsDefined:           true,
	Delimiter:           ",",
	RelID:               id.Null,
	SubscriptFunc:       toFuncID("-"),
	Elem:                id.NullType,
	Array:               toInternal("_timestamptz"),
	InputFunc:           toFuncID("timestamptz_in", toInternal("cstring"), toInternal("oid"), toInternal("int4")),
	OutputFunc:          toFuncID("timestamptz_out", toInternal("timestamptz")),
	ReceiveFunc:         toFuncID("timestamptz_recv", toInternal("internal"), toInternal("oid"), toInternal("int4")),
	SendFunc:            toFuncID("timestamptz_send", toInternal("timestamptz")),
	ModInFunc:           toFuncID("timestamptztypmodin", toInternal("_cstring")),
	ModOutFunc:          toFuncID("timestamptztypmodout", toInternal("int4")),
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
	CompareFunc:         toFuncID("timestamptz_cmp", toInternal("timestamptz"), toInternal("timestamptz")),
	SerializationFunc:   serializeTypeTimestampTZ,
	DeserializationFunc: deserializeTypeTimestampTZ,
}

// NewTimestampTZType returns TimestampTZ type with typmod set. // TODO: implement precision
func NewTimestampTZType(precision int32) (*DoltgresType, error) {
	typmod, err := GetTypmodFromTimePrecision(precision)
	if err != nil {
		return nil, err
	}
	newType := *TimestampTZ.WithAttTypMod(typmod)
	return &newType, nil
}

// serializeTypeTimestampTZ handles serialization from the standard representation to our serialized representation that is
// written in Dolt.
func serializeTypeTimestampTZ(ctx *sql.Context, t *DoltgresType, val any) ([]byte, error) {
	return val.(time.Time).MarshalBinary()
}

// deserializeTypeTimestampTZ handles deserialization from the Dolt serialized format to our standard representation used by
// expressions and nodes.
func deserializeTypeTimestampTZ(ctx *sql.Context, _ *DoltgresType, data []byte) (any, error) {
	if len(data) == 0 {
		return nil, nil
	}
	t := time.Time{}
	if err := t.UnmarshalBinary(data); err != nil {
		return nil, err
	}
	return t, nil
}
