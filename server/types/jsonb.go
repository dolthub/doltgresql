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

// JsonB is the deserialized and structured version of JSON that deals with JsonDocument.
var JsonB = &DoltgresType{
	ID:            toInternal("jsonb"),
	TypLength:     int16(-1),
	PassedByVal:   false,
	TypType:       TypeType_Base,
	TypCategory:   TypeCategory_UserDefinedTypes,
	IsPreferred:   false,
	IsDefined:     true,
	Delimiter:     ",",
	RelID:         id.Null,
	SubscriptFunc: toFuncID("jsonb_subscript_handler", toInternal("internal")),
	Elem:          id.NullType,
	Array:         toInternal("_jsonb"),
	InputFunc:     toFuncID("jsonb_in", toInternal("cstring")),
	OutputFunc:    toFuncID("jsonb_out", toInternal("jsonb")),
	ReceiveFunc:   toFuncID("jsonb_recv", toInternal("internal")),
	SendFunc:      toFuncID("jsonb_send", toInternal("jsonb")),
	ModInFunc:     toFuncID("-"),
	ModOutFunc:    toFuncID("-"),
	AnalyzeFunc:   toFuncID("-"),
	Align:         TypeAlignment_Int,
	Storage:       TypeStorage_Extended,
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
	CompareFunc:   toFuncID("jsonb_cmp", toInternal("jsonb"), toInternal("jsonb")),
}
