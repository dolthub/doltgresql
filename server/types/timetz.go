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

// TimeTZ is the time with a time zone. Precision is unbounded.
var TimeTZ = &DoltgresType{
	OID:           uint32(oid.T_timetz),
	Name:          "timetz",
	Schema:        "pg_catalog",
	TypLength:     int16(12),
	PassedByVal:   true,
	TypType:       TypeType_Base,
	TypCategory:   TypeCategory_DateTimeTypes,
	IsPreferred:   false,
	IsDefined:     true,
	Delimiter:     ",",
	RelID:         0,
	SubscriptFunc: toFuncID("-"),
	Elem:          0,
	Array:         uint32(oid.T__timetz),
	InputFunc:     toFuncID("timetz_in", oid.T_cstring, oid.T_oid, oid.T_int4),
	OutputFunc:    toFuncID("timetz_out", oid.T_timetz),
	ReceiveFunc:   toFuncID("timetz_recv", oid.T_internal, oid.T_oid, oid.T_int4),
	SendFunc:      toFuncID("timetz_send", oid.T_timetz),
	ModInFunc:     toFuncID("timetztypmodin", oid.T__cstring),
	ModOutFunc:    toFuncID("timetztypmodout", oid.T_int4),
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
	attTypMod:     -1,
	CompareFunc:   toFuncID("timetz_cmp", oid.T_timetz, oid.T_timetz),
}

// NewTimeTZType returns TimeTZ type with typmod set. // TODO: implement precision
func NewTimeTZType(precision int32) (*DoltgresType, error) {
	typmod, err := GetTypmodFromTimePrecision(precision)
	if err != nil {
		return nil, err
	}
	newType := *TimeTZ.WithAttTypMod(typmod)
	return &newType, nil
}
