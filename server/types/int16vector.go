// Copyright 2026 Dolthub, Inc.
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

// Int16vector is the vector variant of Int16.
var Int16vector = &DoltgresType{
	ID:            toInternal("int2vector"),
	TypLength:     int16(-1),
	PassedByVal:   false,
	TypType:       TypeType_Base,
	TypCategory:   TypeCategory_ArrayTypes,
	IsPreferred:   false,
	IsDefined:     true,
	Delimiter:     ",",
	RelID:         id.Null,
	SubscriptFunc: toFuncID("array_subscript_handler", toInternal("internal")),
	Elem:          Int16.ID,
	Array:         toInternal("_int2vector"),
	InputFunc:     toFuncID("int2vectorin", toInternal("cstring")),
	OutputFunc:    toFuncID("int2vectorout", toInternal("int2vector")),
	ReceiveFunc:   toFuncID("int2vectorrecv", toInternal("internal")),
	SendFunc:      toFuncID("int2vectorsend", toInternal("int2vector")),
	ModInFunc:     toFuncID("-"),
	ModOutFunc:    toFuncID("-"),
	AnalyzeFunc:   toFuncID("-"),
	Align:         TypeAlignment_Int,
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
}
