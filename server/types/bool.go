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

// Bool is the bool type.
var Bool = &DoltgresType{
	ID:            toInternal("bool"),
	TypLength:     int16(1),
	PassedByVal:   true,
	TypType:       TypeType_Base,
	TypCategory:   TypeCategory_BooleanTypes,
	IsPreferred:   true,
	IsDefined:     true,
	Delimiter:     ",",
	RelID:         id.Null,
	SubscriptFunc: toFuncID("-"),
	Elem:          id.NullType,
	Array:         toInternal("_bool"),
	InputFunc:     toFuncID("boolin", toInternal("cstring")),
	OutputFunc:    toFuncID("boolout", toInternal("bool")),
	ReceiveFunc:   toFuncID("boolrecv", toInternal("internal")),
	SendFunc:      toFuncID("boolsend", toInternal("bool")),
	ModInFunc:     toFuncID("-"),
	ModOutFunc:    toFuncID("-"),
	AnalyzeFunc:   toFuncID("-"),
	Align:         TypeAlignment_Char,
	Storage:       TypeStorage_Plain,
	NotNull:       false,
	BaseTypeID:    id.NullType,
	TypMod:        -1,
	NDims:         0,
	TypCollation:  id.NullCollation,
	DefaulBin:     "",
	Default:       "",
	Acl:           nil,
	Checks:        nil,
	attTypMod:     -1,
	CompareFunc:   toFuncID("btboolcmp", toInternal("bool"), toInternal("bool")),
	InternalName:  "boolean",
}
