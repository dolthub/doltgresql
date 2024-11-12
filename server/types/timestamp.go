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

// Timestamp is the timestamp without a time zone. Precision is unbounded.
var Timestamp = DoltgresType{
	OID:           uint32(oid.T_timestamp),
	Name:          "timestamp",
	Schema:        "pg_catalog",
	Owner:         "doltgres", // TODO
	TypLength:     int16(8),
	PassedByVal:   true,
	TypType:       TypeType_Base,
	TypCategory:   TypeCategory_DateTimeTypes,
	IsPreferred:   false,
	IsDefined:     true,
	Delimiter:     ",",
	RelID:         0,
	SubscriptFunc: "-",
	Elem:          0,
	Array:         uint32(oid.T__timestamp),
	InputFunc:     "timestamp_in",
	OutputFunc:    "timestamp_out",
	ReceiveFunc:   "timestamp_recv",
	SendFunc:      "timestamp_send",
	ModInFunc:     "timestamptypmodin",
	ModOutFunc:    "timestamptypmodout",
	AnalyzeFunc:   "-",
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
	CompareFunc:   "timestamp_cmp",
}

// NewTimestampType returns Timestamp type with typmod set. // TODO: implement precision
func NewTimestampType(precision int32) (DoltgresType, error) {
	newType := Timestamp
	typmod, err := GetTypmodFromTimePrecision(precision)
	if err != nil {
		return DoltgresType{}, err
	}
	newType.AttTypMod = typmod
	return newType, nil
}
