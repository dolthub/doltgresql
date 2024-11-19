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

// Bool is the bool type.
var Bool = DoltgresType{
	OID:           uint32(oid.T_bool),
	Name:          "bool",
	Schema:        "pg_catalog",
	Owner:         "doltgres",
	TypLength:     int16(1),
	PassedByVal:   true,
	TypType:       TypeType_Base,
	TypCategory:   TypeCategory_BooleanTypes,
	IsPreferred:   true,
	IsDefined:     true,
	Delimiter:     ",",
	RelID:         0,
	SubscriptFunc: "-",
	Elem:          0,
	Array:         uint32(oid.T__bool),
	InputFunc:     "boolin",
	OutputFunc:    "boolout",
	ReceiveFunc:   "boolrecv",
	SendFunc:      "boolsend",
	ModInFunc:     "-",
	ModOutFunc:    "-",
	AnalyzeFunc:   "-",
	Align:         TypeAlignment_Char,
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
	CompareFunc:   "btboolcmp",
	InternalName:  "boolean",
}
