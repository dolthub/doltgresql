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
	"github.com/lib/pq/oid"
)

// TimestampTZ is the timestamp with a time zone. Precision is unbounded.
var TimestampTZ = &DoltgresType{
	OID:           uint32(oid.T_timestamptz),
	Name:          "timestamptz",
	Schema:        "pg_catalog",
	TypLength:     int16(8),
	PassedByVal:   true,
	TypType:       TypeType_Base,
	TypCategory:   TypeCategory_DateTimeTypes,
	IsPreferred:   true,
	IsDefined:     true,
	Delimiter:     ",",
	RelID:         0,
	SubscriptFunc: toFuncID("-"),
	Elem:          0,
	Array:         uint32(oid.T__timestamptz),
	InputFunc:     toFuncID("timestamptz_in", oid.T_cstring, oid.T_oid, oid.T_int4),
	OutputFunc:    toFuncID("timestamptz_out", oid.T_timestamptz),
	ReceiveFunc:   toFuncID("timestamptz_recv", oid.T_internal, oid.T_oid, oid.T_int4),
	SendFunc:      toFuncID("timestamptz_send", oid.T_timestamptz),
	ModInFunc:     toFuncID("timestamptztypmodin", oid.T__cstring),
	ModOutFunc:    toFuncID("timestamptztypmodout", oid.T_int4),
	AnalyzeFunc:   toFuncID("-"),
	Align:         TypeAlignment_Double,
	Storage:       TypeStorage_Plain,
	NotNull:       false,
	BaseTypeOID:   0,
	TypMod:        -1,
	NDims:         0,
	TypCollation:  0,
	DefaulBin:     "",
	Default:       "",
	Acl:           nil,
	Checks:        nil,
	AttTypMod:     -1,
	CompareFunc:   toFuncID("timestamptz_cmp", oid.T_timestamptz, oid.T_timestamptz),
}

// NewTimestampTZType returns TimestampTZ type with typmod set. // TODO: implement precision
func NewTimestampTZType(precision int32) (*DoltgresType, error) {
	newType := *TimestampTZ
	typmod, err := GetTypmodFromTimePrecision(precision)
	if err != nil {
		return nil, err
	}
	newType.AttTypMod = typmod
	return &newType, nil
}
