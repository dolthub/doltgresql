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
	"fmt"

	"github.com/lib/pq/oid"
)

// Time is the time without a time zone. Precision is unbounded.
var Time = DoltgresType{
	OID:           uint32(oid.T_time),
	Name:          "time",
	Schema:        "pg_catalog",
	TypLength:     int16(8),
	PassedByVal:   true,
	TypType:       TypeType_Base,
	TypCategory:   TypeCategory_DateTimeTypes,
	IsPreferred:   false,
	IsDefined:     true,
	Delimiter:     ",",
	RelID:         0,
	SubscriptFunc: toFuncID("-"),
	Elem:          0,
	Array:         uint32(oid.T__time),
	InputFunc:     toFuncID("time_in", oid.T_cstring, oid.T_oid, oid.T_int4),
	OutputFunc:    toFuncID("time_out", oid.T_time),
	ReceiveFunc:   toFuncID("time_recv", oid.T_internal, oid.T_oid, oid.T_int4),
	SendFunc:      toFuncID("time_send", oid.T_time),
	ModInFunc:     toFuncID("timetypmodin", oid.T__cstring),
	ModOutFunc:    toFuncID("timetypmodout", oid.T_int4),
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
	CompareFunc:   toFuncID("time_cmp", oid.T_time, oid.T_time),
}

// NewTimeType returns Time type with typmod set. // TODO: implement precision
func NewTimeType(precision int32) (DoltgresType, error) {
	newType := Time
	typmod, err := GetTypmodFromTimePrecision(precision)
	if err != nil {
		return DoltgresType{}, err
	}
	newType.AttTypMod = typmod
	return newType, nil
}

// GetTypmodFromTimePrecision takes Time type precision and returns the type modifier value.
func GetTypmodFromTimePrecision(precision int32) (int32, error) {
	if precision < 0 {
		// TIME(-1) precision must not be negative
		return 0, fmt.Errorf("TIME(%v) precision must be not be negative", precision)
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
