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

// Internal is an internal type, which means `external binary` type.
var Internal = DoltgresType{
	OID:           uint32(oid.T_internal),
	Name:          "internal",
	Schema:        "pg_catalog",
	TypLength:     int16(8),
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
	InputFunc:     "internal_in",
	OutputFunc:    "internal_out",
	ReceiveFunc:   "-",
	SendFunc:      "-",
	ModInFunc:     "-",
	ModOutFunc:    "-",
	AnalyzeFunc:   "-",
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
	AttTypMod:     -1,
	CompareFunc:   "-",
}

// NewInternalTypeWithBaseType returns Internal type with
// internal base type set with given type.
func NewInternalTypeWithBaseType(t uint32) DoltgresType {
	it := Internal
	it.BaseTypeForInternal = t
	return it
}
