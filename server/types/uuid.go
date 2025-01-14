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

// Uuid is the UUID type.
var Uuid = &DoltgresType{
	ID:            toInternal("uuid"),
	TypLength:     int16(16),
	PassedByVal:   false,
	TypType:       TypeType_Base,
	TypCategory:   TypeCategory_UserDefinedTypes,
	IsPreferred:   false,
	IsDefined:     true,
	Delimiter:     ",",
	RelID:         id.Null,
	SubscriptFunc: toFuncID("-"),
	Elem:          id.NullType,
	Array:         toInternal("_uuid"),
	InputFunc:     toFuncID("uuid_in", toInternal("cstring")),
	OutputFunc:    toFuncID("uuid_out", toInternal("uuid")),
	ReceiveFunc:   toFuncID("uuid_recv", toInternal("internal")),
	SendFunc:      toFuncID("uuid_send", toInternal("uuid")),
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
	CompareFunc:   toFuncID("uuid_cmp", toInternal("uuid"), toInternal("uuid")),
}
