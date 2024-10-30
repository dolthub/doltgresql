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

const (
	// StringMaxLength is the maximum number of characters (not bytes) that a Char, VarChar, or BpChar may contain.
	StringMaxLength = 10485760
	// stringInline is the maximum number of characters (not bytes) that are "guaranteed" to fit inline.
	stringInline = 16383
	// StringUnbounded is used to represent that a type does not define a limit on the strings that it accepts. Values
	// are still limited by the field size limit, but it won't be enforced by the type.
	StringUnbounded = 0
)

// VarChar is a varchar that has an unbounded length.
var VarChar = DoltgresType{
	OID:           uint32(oid.T_varchar),
	Name:          "varchar",
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
	Array:         uint32(oid.T__varchar),
	InputFunc:     "varcharin",
	OutputFunc:    "varcharout",
	ReceiveFunc:   "varcharrecv",
	SendFunc:      "varcharsend",
	ModInFunc:     "varchartypmodin",
	ModOutFunc:    "varchartypmodout",
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

func NewVarCharType(maxChars uint32) DoltgresType {
	// TODO: maxChars represents the maximum number of characters that the type may hold.
	//  When this is zero, we treat it as completely unbounded (which is still limited by the field size limit).
	return VarChar
}
