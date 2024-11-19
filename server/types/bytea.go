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

// Bytea is the byte string type.
var Bytea = DoltgresType{
	OID:           uint32(oid.T_bytea),
	Name:          "bytea",
	Schema:        "pg_catalog",
	TypLength:     int16(-1),
	PassedByVal:   false,
	TypType:       TypeType_Base,
	TypCategory:   TypeCategory_UserDefinedTypes,
	IsPreferred:   false,
	IsDefined:     true,
	Delimiter:     ",",
	RelID:         0,
	SubscriptFunc: "-",
	Elem:          0,
	Array:         uint32(oid.T__bytea),
	InputFunc:     "byteain",
	OutputFunc:    "byteaout",
	ReceiveFunc:   "bytearecv",
	SendFunc:      "byteasend",
	ModInFunc:     "-",
	ModOutFunc:    "-",
	AnalyzeFunc:   "-",
	Align:         TypeAlignment_Int,
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
	AttTypMod:     -1,
	CompareFunc:   "byteacmp",
}
