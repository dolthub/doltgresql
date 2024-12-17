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
	"github.com/dolthub/doltgresql/core/id"
)

// Timestamp is the timestamp without a time zone. Precision is unbounded.
var Timestamp = &DoltgresType{
	ID:            toInternal("timestamp"),
	TypLength:     int16(8),
	PassedByVal:   true,
	TypType:       TypeType_Base,
	TypCategory:   TypeCategory_DateTimeTypes,
	IsPreferred:   false,
	IsDefined:     true,
	Delimiter:     ",",
	RelID:         id.Null,
	SubscriptFunc: toFuncID("-"),
	Elem:          id.Null,
	Array:         toInternal("_timestamp"),
	InputFunc:     toFuncID("timestamp_in", toInternal("cstring"), toInternal("oid"), toInternal("int4")),
	OutputFunc:    toFuncID("timestamp_out", toInternal("timestamp")),
	ReceiveFunc:   toFuncID("timestamp_recv", toInternal("internal"), toInternal("oid"), toInternal("int4")),
	SendFunc:      toFuncID("timestamp_send", toInternal("timestamp")),
	ModInFunc:     toFuncID("timestamptypmodin", toInternal("_cstring")),
	ModOutFunc:    toFuncID("timestamptypmodout", toInternal("int4")),
	AnalyzeFunc:   toFuncID("-"),
	Align:         TypeAlignment_Double,
	Storage:       TypeStorage_Plain,
	NotNull:       false,
	BaseTypeID:    id.Null,
	TypMod:        -1,
	NDims:         0,
	TypCollation:  id.Null,
	DefaulBin:     "",
	Default:       "",
	Acl:           nil,
	Checks:        nil,
	attTypMod:     -1,
	CompareFunc:   toFuncID("timestamp_cmp", toInternal("timestamp"), toInternal("timestamp")),
}

// NewTimestampType returns Timestamp type with typmod set. // TODO: implement precision
func NewTimestampType(precision int32) (*DoltgresType, error) {
	typmod, err := GetTypmodFromTimePrecision(precision)
	if err != nil {
		return nil, err
	}
	newType := *Timestamp.WithAttTypMod(typmod)
	return &newType, nil
}
