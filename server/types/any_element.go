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

// AnyElement is a pseudo-type that can represent any type.
var AnyElement = DoltgresType{
	OID:           uint32(oid.T_anyelement),
	Name:          "anyelement",
	Schema:        "pg_catalog",
	Owner:         "doltgres", // TODO
	TypLength:     int16(4),
	PassedByVal:   true,
	TypType:       TypeType_Pseudo,
	TypCategory:   TypeCategory_PseudoTypes,
	IsPreferred:   false,
	IsDefined:     true,
	Delimiter:     ",",
	RelID:         0,
	SubscriptFunc: "-",
	Elem:          0,
	Array:         0,
	InputFunc:     "anyelement_in",
	OutputFunc:    "anyelement_out",
	ReceiveFunc:   "-",
	SendFunc:      "-",
	ModInFunc:     "-",
	ModOutFunc:    "-",
	AnalyzeFunc:   "-",
	Align:         TypeAlignment_Int,
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
	CompareFunc:   "-",
}
