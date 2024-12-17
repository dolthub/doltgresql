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

// InternalCharLength will always be 1.
const InternalCharLength = 1

// InternalChar is a single-byte internal type. In Postgres, it's displayed as `"char"`.
var InternalChar = &DoltgresType{
	ID:            toInternal("char"),
	TypLength:     int16(InternalCharLength),
	PassedByVal:   true,
	TypType:       TypeType_Base,
	TypCategory:   TypeCategory_InternalUseTypes,
	IsPreferred:   false,
	IsDefined:     true,
	Delimiter:     ",",
	RelID:         id.Null,
	SubscriptFunc: toFuncID("-"),
	Elem:          id.Null,
	Array:         toInternal("_char"),
	InputFunc:     toFuncID("charin", toInternal("cstring")),
	OutputFunc:    toFuncID("charout", toInternal("char")),
	ReceiveFunc:   toFuncID("charrecv", toInternal("internal")),
	SendFunc:      toFuncID("charsend", toInternal("char")),
	ModInFunc:     toFuncID("-"),
	ModOutFunc:    toFuncID("-"),
	AnalyzeFunc:   toFuncID("-"),
	Align:         TypeAlignment_Char,
	Storage:       TypeStorage_Plain,
	NotNull:       false,
	BaseTypeID:    id.Null,
	TypMod:        -1,
	NDims:         0,
	TypCollation:  id.Null,
	DefaulBin:     "",
	Default:       "",
	Acl:           nil,
	Checks:        nil,
	attTypMod:     -1,
	CompareFunc:   toFuncID("btcharcmp", toInternal("char"), toInternal("char")),
	InternalName:  `"char"`,
}
