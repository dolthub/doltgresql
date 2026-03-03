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

// Oidvector is the vector variant of Oid.
var Oidvector = &DoltgresType{
	ID:            toInternal("oidvector"),
	TypLength:     int16(-1),
	PassedByVal:   false,
	TypType:       TypeType_Base,
	TypCategory:   TypeCategory_ArrayTypes,
	IsPreferred:   false,
	IsDefined:     true,
	Delimiter:     ",",
	RelID:         id.Null,
	SubscriptFunc: toFuncID("array_subscript_handler", toInternal("internal")),
	Elem:          Oid.ID,
	Array:         toInternal("_oidvector"),
	InputFunc:     toFuncID("oidvectorin", toInternal("cstring")),
	OutputFunc:    toFuncID("oidvectorout", toInternal("oidvector")),
	ReceiveFunc:   toFuncID("oidvectorrecv", toInternal("internal")),
	SendFunc:      toFuncID("oidvectorsend", toInternal("oidvector")),
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
	CompareFunc:   toFuncID("btoidvectorcmp", toInternal("oidvector"), toInternal("oidvector")),
}
