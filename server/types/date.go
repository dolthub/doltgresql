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

// Date is the day, month, and year.
var Date = &DoltgresType{
	ID:                  toInternal("date"),
	TypLength:           int16(4),
	PassedByVal:         true,
	TypType:             TypeType_Base,
	TypCategory:         TypeCategory_DateTimeTypes,
	IsPreferred:         false,
	IsDefined:           true,
	Delimiter:           ",",
	RelID:               id.Null,
	SubscriptFunc:       toFuncID("-"),
	Elem:                id.NullType,
	Array:               toInternal("_date"),
	InputFunc:           toFuncID("date_in", toInternal("cstring")),
	OutputFunc:          toFuncID("date_out", toInternal("date")),
	ReceiveFunc:         toFuncID("date_recv", toInternal("internal")),
	SendFunc:            toFuncID("date_send", toInternal("date")),
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
	CompareFunc:         toFuncID("date_cmp", toInternal("date"), toInternal("date")),
	SerializationFunc:   serializeTypeDate,
	DeserializationFunc: deserializeTypeDate,
}

// serializeTypeDate handles serialization from the standard representation to our serialized representation that is
// written in Dolt.
func serializeTypeDate(ctx *sql.Context, t *DoltgresType, val any) ([]byte, error) {
	return val.(time.Time).MarshalBinary()
}

// deserializeTypeDate handles deserialization from the Dolt serialized format to our standard representation used by
// expressions and nodes.
func deserializeTypeDate(ctx *sql.Context, _ *DoltgresType, data []byte) (any, error) {
	if len(data) == 0 {
		return nil, nil
	}
	t := time.Time{}
	if err := t.UnmarshalBinary(data); err != nil {
		return nil, err
	}
	return t, nil
}
