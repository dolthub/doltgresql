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

// NameLength is the constant length of Name in Postgres 15. Represents (NAMEDATALEN-1)
const NameLength = 63

// Name is a 63-byte internal type for object names.
var Name = DoltgresType{
	OID:           uint32(oid.T_name),
	Name:          "name",
	Schema:        "pg_catalog",
	Owner:         "doltgres", // TODO
	TypLength:     int16(64),
	PassedByVal:   false,
	TypType:       TypeType_Base,
	TypCategory:   TypeCategory_StringTypes,
	IsPreferred:   false,
	IsDefined:     true,
	Delimiter:     ",",
	RelID:         0,
	SubscriptFunc: "raw_array_subscript_handler",
	Elem:          uint32(oid.T_char),
	Array:         uint32(oid.T__name),
	InputFunc:     "namein",
	OutputFunc:    "nameout",
	ReceiveFunc:   "namerecv",
	SendFunc:      "namesend",
	ModInFunc:     "-",
	ModOutFunc:    "-",
	AnalyzeFunc:   "-",
	Align:         TypeAlignment_Char,
	Storage:       TypeStorage_Plain,
	NotNull:       false,
	BaseTypeOID:   0,
	TypMod:        -1,
	NDims:         0,
	TypCollation:  950,
	DefaulBin:     "",
	Default:       "",
	Acl:           nil,
	Checks:        nil,
}
