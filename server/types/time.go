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

	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/core/id"
)

// Time is the time without a time zone. Precision is unbounded.
var Time = &DoltgresType{
	ID:                  toInternal("time"),
	TypLength:           int16(8),
	PassedByVal:         true,
	TypType:             TypeType_Base,
	TypCategory:         TypeCategory_DateTimeTypes,
	IsPreferred:         false,
	IsDefined:           true,
	Delimiter:           ",",
	RelID:               id.Null,
	SubscriptFunc:       toFuncID("-"),
	Elem:                id.NullType,
	Array:               toInternal("_time"),
	InputFunc:           toFuncID("time_in", toInternal("cstring"), toInternal("oid"), toInternal("int4")),
	OutputFunc:          toFuncID("time_out", toInternal("time")),
	ReceiveFunc:         toFuncID("time_recv", toInternal("internal"), toInternal("oid"), toInternal("int4")),
	SendFunc:            toFuncID("time_send", toInternal("time")),
	ModInFunc:           toFuncID("timetypmodin", toInternal("_cstring")),
	ModOutFunc:          toFuncID("timetypmodout", toInternal("int4")),
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
	CompareFunc:         toFuncID("time_cmp", toInternal("time"), toInternal("time")),
	SerializationFunc:   serializeTypeTime,
	DeserializationFunc: deserializeTypeTime,
}

// NewTimeType returns Time type with typmod set. // TODO: implement precision
func NewTimeType(precision int32) (*DoltgresType, error) {
	typmod, err := GetTypmodFromTimePrecision(precision)
	if err != nil {
		return nil, err
	}
	newType := *Time.WithAttTypMod(typmod)
	return &newType, nil
}

// GetTypmodFromTimePrecision takes Time type precision and returns the type modifier value.
func GetTypmodFromTimePrecision(precision int32) (int32, error) {
	if precision < 0 {
		// TIME(-1) precision must not be negative
		return 0, errors.Errorf("TIME(%v) precision must be not be negative", precision)
	}
	if precision > 6 {
		precision = 6
		//WARNING:  TIME(7) precision reduced to maximum allowed, 6
	}
	return precision, nil
}

// GetTimePrecisionFromTypMod takes Time type modifier and returns precision value.
func GetTimePrecisionFromTypMod(typmod int32) int32 {
	return typmod
}

// serializeTypeTime handles serialization from the standard representation to our serialized representation that is
// written in Dolt.
func serializeTypeTime(ctx *sql.Context, t *DoltgresType, val any) ([]byte, error) {
	return val.(time.Time).MarshalBinary()
}

// deserializeTypeTime handles deserialization from the Dolt serialized format to our standard representation used by
// expressions and nodes.
func deserializeTypeTime(ctx *sql.Context, _ *DoltgresType, data []byte) (any, error) {
	if len(data) == 0 {
		return nil, nil
	}
	t := time.Time{}
	if err := t.UnmarshalBinary(data); err != nil {
		return nil, err
	}
	return t, nil
}
