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
	"github.com/dolthub/doltgresql/core/id"
)

// NameLength is the constant length of Name in Postgres 15. Represents (NAMEDATALEN-1)
const NameLength = 63

// Name is a 63-byte internal type for object names.
var Name = &DoltgresType{
	ID:            toInternal("name"),
	TypLength:     int16(64),
	PassedByVal:   false,
	TypType:       TypeType_Base,
	TypCategory:   TypeCategory_StringTypes,
	IsPreferred:   false,
	IsDefined:     true,
	Delimiter:     ",",
	RelID:         id.Null,
	SubscriptFunc: toFuncID("raw_array_subscript_handler", toInternal("internal")),
	Elem:          toInternal("char"),
	Array:         toInternal("_name"),
	InputFunc:     toFuncID("namein", toInternal("cstring")),
	OutputFunc:    toFuncID("nameout", toInternal("name")),
	ReceiveFunc:   toFuncID("namerecv", toInternal("internal")),
	SendFunc:      toFuncID("namesend", toInternal("name")),
	ModInFunc:     toFuncID("-"),
	ModOutFunc:    toFuncID("-"),
	AnalyzeFunc:   toFuncID("-"),
	Align:         TypeAlignment_Char,
	Storage:       TypeStorage_Plain,
	NotNull:       false,
	BaseTypeID:    id.Null,
	TypMod:        -1,
	NDims:         0,
	TypCollation:  id.NewInternal(id.Section_Collation, "pg_catalog", "C"),
	DefaulBin:     "",
	Default:       "",
	Acl:           nil,
	Checks:        nil,
	attTypMod:     -1,
	CompareFunc:   toFuncID("btnamecmp", toInternal("name"), toInternal("name")),
}
