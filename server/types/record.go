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

// Record is the record type, similar to a row.
var Record = &DoltgresType{
	OID:           uint32(oid.T_record),
	Name:          "record",
	Schema:        "pg_catalog",
	TypLength:     -1,
	PassedByVal:   false,
	TypType:       TypeType_Pseudo,
	TypCategory:   TypeCategory_PseudoTypes,
	IsPreferred:   false,
	IsDefined:     true,
	Delimiter:     ",",
	RelID:         0,
	SubscriptFunc: toFuncID("-"),
	Elem:          0,
	Array:         uint32(oid.T__record),
	InputFunc:     toFuncID("record_in", oid.T_cstring, oid.T_oid, oid.T_int4),
	OutputFunc:    toFuncID("record_out", oid.T_record),
	ReceiveFunc:   toFuncID("record_recv", oid.T_internal, oid.T_oid, oid.T_int4),
	SendFunc:      toFuncID("record_send", oid.T_record),
	ModInFunc:     toFuncID("-"),
	ModOutFunc:    toFuncID("-"),
	AnalyzeFunc:   toFuncID("-"),
	Align:         TypeAlignment_Double,
	Storage:       TypeStorage_Extended,
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
	CompareFunc:   toFuncID("btrecordcmp", oid.T_record, oid.T_record),
}
