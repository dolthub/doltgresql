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

import "github.com/lib/pq/oid"

// Int16Serial is an int16 serial type.
var Int16Serial = DoltgresType{
	OID:           0, // doesn't have unique OID
	Name:          "smallserial",
	Schema:        "pg_catalog",
	Owner:         "doltgres", // TODO
	TypLength:     int16(2),
	PassedByVal:   true,
	TypType:       TypeType_Base,
	TypCategory:   TypeCategory_NumericTypes,
	IsPreferred:   false,
	IsDefined:     true,
	Delimiter:     ",",
	RelID:         0,
	SubscriptFunc: "-",
	Elem:          0,
	Array:         uint32(oid.T__int2),
	InputFunc:     "int2in",
	OutputFunc:    "int2out",
	ReceiveFunc:   "int2recv",
	SendFunc:      "int2send",
	ModInFunc:     "-",
	ModOutFunc:    "-",
	AnalyzeFunc:   "-",
	Align:         TypeAlignment_Short,
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
	CompareFunc:   "btint2cmp",
	IsSerial:      true,
}
