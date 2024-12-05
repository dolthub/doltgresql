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

// Interval is the interval type.
var Interval = &DoltgresType{
	OID:           uint32(oid.T_interval),
	Name:          "interval",
	Schema:        "pg_catalog",
	TypLength:     int16(16),
	PassedByVal:   false,
	TypType:       TypeType_Base,
	TypCategory:   TypeCategory_TimespanTypes,
	IsPreferred:   true,
	IsDefined:     true,
	Delimiter:     ",",
	RelID:         0,
	SubscriptFunc: toFuncID("-"),
	Elem:          0,
	Array:         uint32(oid.T__interval),
	InputFunc:     toFuncID("interval_in", oid.T_cstring, oid.T_oid, oid.T_int4),
	OutputFunc:    toFuncID("interval_out", oid.T_interval),
	ReceiveFunc:   toFuncID("interval_recv", oid.T_internal, oid.T_oid, oid.T_int4),
	SendFunc:      toFuncID("interval_send", oid.T_interval),
	ModInFunc:     toFuncID("intervaltypmodin", oid.T__cstring),
	ModOutFunc:    toFuncID("intervaltypmodout", oid.T_int4),
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
	CompareFunc:   toFuncID("interval_cmp", oid.T_interval, oid.T_interval),
}
