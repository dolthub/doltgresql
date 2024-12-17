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

// Xid is a data type used for internal transaction IDs. It is implemented as an unsigned 32 bit integer.
var Xid = &DoltgresType{
	ID:            toInternal("xid"),
	TypLength:     int16(4),
	PassedByVal:   true,
	TypType:       TypeType_Base,
	TypCategory:   TypeCategory_UserDefinedTypes,
	IsPreferred:   false,
	IsDefined:     true,
	Delimiter:     ",",
	RelID:         id.Null,
	SubscriptFunc: toFuncID("-"),
	Elem:          id.Null,
	Array:         toInternal("_xid"),
	InputFunc:     toFuncID("xidin", toInternal("cstring")),
	OutputFunc:    toFuncID("xidout", toInternal("xid")),
	ReceiveFunc:   toFuncID("xidrecv", toInternal("internal")),
	SendFunc:      toFuncID("xidsend", toInternal("xid")),
	ModInFunc:     toFuncID("-"),
	ModOutFunc:    toFuncID("-"),
	AnalyzeFunc:   toFuncID("-"),
	Align:         TypeAlignment_Int,
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
	CompareFunc:   toFuncID("-"),
}
