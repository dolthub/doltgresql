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

import "github.com/dolthub/doltgresql/core/id"

// Internal is an internal type, which means `external binary` type.
var Internal = &DoltgresType{
	ID:            toInternal("internal"),
	TypLength:     int16(8),
	PassedByVal:   true,
	TypType:       TypeType_Pseudo,
	TypCategory:   TypeCategory_PseudoTypes,
	IsPreferred:   false,
	IsDefined:     true,
	Delimiter:     ",",
	RelID:         id.Null,
	SubscriptFunc: toFuncID("-"),
	Elem:          id.Null,
	Array:         id.Null,
	InputFunc:     toFuncID("internal_in", toInternal("cstring")),
	OutputFunc:    toFuncID("internal_out", toInternal("internal")),
	ReceiveFunc:   toFuncID("-"),
	SendFunc:      toFuncID("-"),
	ModInFunc:     toFuncID("-"),
	ModOutFunc:    toFuncID("-"),
	AnalyzeFunc:   toFuncID("-"),
	Align:         TypeAlignment_Double,
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

// NewInternalTypeWithBaseType returns Internal type with
// internal base type set with given type.
func NewInternalTypeWithBaseType(internalID id.Internal) *DoltgresType {
	it := Internal
	it.BaseTypeForInternal = internalID
	return it
}
