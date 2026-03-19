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
	"github.com/dolthub/doltgresql/postgres/parser/duration"
	"github.com/dolthub/doltgresql/utils"
)

// Interval is the interval type.
var Interval = &DoltgresType{
	ID:                  toInternal("interval"),
	TypLength:           int16(16),
	PassedByVal:         false,
	TypType:             TypeType_Base,
	TypCategory:         TypeCategory_TimespanTypes,
	IsPreferred:         true,
	IsDefined:           true,
	Delimiter:           ",",
	RelID:               id.Null,
	SubscriptFunc:       toFuncID("-"),
	Elem:                id.NullType,
	Array:               toInternal("_interval"),
	InputFunc:           toFuncID("interval_in", toInternal("cstring"), toInternal("oid"), toInternal("int4")),
	OutputFunc:          toFuncID("interval_out", toInternal("interval")),
	ReceiveFunc:         toFuncID("interval_recv", toInternal("internal"), toInternal("oid"), toInternal("int4")),
	SendFunc:            toFuncID("interval_send", toInternal("interval")),
	ModInFunc:           toFuncID("intervaltypmodin", toInternal("_cstring")),
	ModOutFunc:          toFuncID("intervaltypmodout", toInternal("int4")),
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
	CompareFunc:         toFuncID("interval_cmp", toInternal("interval"), toInternal("interval")),
	SerializationFunc:   serializeTypeInterval,
	DeserializationFunc: deserializeTypeInterval,
}

// serializeTypeInterval handles serialization from the standard representation to our serialized representation that is
// written in Dolt.
func serializeTypeInterval(ctx *sql.Context, t *DoltgresType, val any) ([]byte, error) {
	sortNanos, months, days, err := val.(duration.Duration).Encode()
	if err != nil {
		return nil, err
	}
	writer := utils.NewWriter(0)
	writer.Int64(sortNanos)
	writer.Int32(int32(months))
	writer.Int32(int32(days))
	return writer.Data(), nil
}

// deserializeTypeInterval handles deserialization from the Dolt serialized format to our standard representation used by
// expressions and nodes.
func deserializeTypeInterval(ctx *sql.Context, t *DoltgresType, data []byte) (any, error) {
	if len(data) == 0 {
		return nil, nil
	}
	reader := utils.NewReader(data)
	sortNanos := reader.Int64()
	months := reader.Int32()
	days := reader.Int32()
	return duration.Decode(sortNanos, int64(months), int64(days))
}
