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

// Uuid is the UUID type.
var Uuid = DoltgresType{
	OID:           uint32(oid.T_uuid),
	Name:          "uuid",
	Schema:        "pg_catalog",
	Owner:         "doltgres", // TODO
	Length:        int16(16),
	PassedByVal:   false,
	TypType:       TypeType_Base,
	TypCategory:   TypeCategory_UserDefinedTypes,
	IsPreferred:   false,
	IsDefined:     true,
	Delimiter:     ",",
	RelID:         0,
	SubscriptFunc: "-",
	Elem:          0,
	Array:         uint32(oid.T__uuid),
	InputFunc:     "uuid_in",
	OutputFunc:    "uuid_out",
	ReceiveFunc:   "uuid_recv",
	SendFunc:      "uuid_send",
	ModInFunc:     "-",
	ModOutFunc:    "-",
	AnalyzeFunc:   "-",
	Align:         TypeAlignment_Char,
	Storage:       TypeStorage_Plain,
	NotNull:       false,
	BaseTypeOID:   0,
	TypMod:        -1,
	NDims:         0,
	Collation:     0,
	DefaulBin:     "",
	Default:       "",
	Acl:           "",
	Checks:        nil,
}
