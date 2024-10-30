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

// BpChar is a char that has an unbounded length.
var BpChar = DoltgresType{
	OID:           uint32(oid.T_bpchar),
	Name:          "bpchar",
	Schema:        "pg_catalog",
	Owner:         "doltgres", // TODO
	Length:        int16(-1),
	PassedByVal:   false,
	TypType:       TypeType_Base,
	TypCategory:   TypeCategory_StringTypes,
	IsPreferred:   false,
	IsDefined:     true,
	Delimiter:     ",",
	RelID:         0,
	SubscriptFunc: "-",
	Elem:          0,
	Array:         uint32(oid.T__bpchar),
	InputFunc:     "bpcharin",
	OutputFunc:    "bpcharout",
	ReceiveFunc:   "bpcharrecv",
	SendFunc:      "bpcharsend",
	ModInFunc:     "bpchartypmodin",
	ModOutFunc:    "bpchartypmodout",
	AnalyzeFunc:   "-",
	Align:         TypeAlignment_Int,
	Storage:       TypeStorage_Extended,
	NotNull:       false,
	BaseTypeOID:   0,
	TypMod:        -1,
	NDims:         0,
	Collation:     100,
	DefaulBin:     "",
	Default:       "",
	Acl:           "",
	Checks:        nil,
}

func NewCharType(length uint32) DoltgresType {
	// TODO: maxChars represents the maximum number of characters that the type may hold.
	//  When this is zero, we treat it as completely unbounded (which is still limited by the field size limit).
	// how would this be differentiated in casting when oids are use????
	bpChar := BpChar
	bpChar.Length = int16(length)
	return bpChar
}
