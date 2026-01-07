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
	"gopkg.in/src-d/go-errors.v1"

	"github.com/dolthub/doltgresql/core/id"
)

// ErrVarBitLengthExceeded is returned when a varbit value exceeds the defined length.
var ErrVarBitLengthExceeded = errors.NewKind(`bit string too long for type bit varying(%d)`)

// VarBit is a varying-length bit string.
var VarBit = &DoltgresType{
	ID:            toInternal("varbit"),
	TypLength:     int16(-1),
	PassedByVal:   false,
	TypType:       TypeType_Base,
	TypCategory:   TypeCategory_BitStringTypes,
	IsPreferred:   false,
	IsDefined:     true,
	Delimiter:     ",",
	RelID:         id.Null,
	SubscriptFunc: toFuncID("-"),
	Elem:          id.NullType,
	Array:         toInternal("_varbit"),
	InputFunc:     toFuncID("varbit_in", toInternal("cstring"), toInternal("oid"), toInternal("int4")),
	OutputFunc:    toFuncID("varbit_out", toInternal("varbit")),
	ReceiveFunc:   toFuncID("varbit_recv", toInternal("internal"), toInternal("oid"), toInternal("int4")),
	SendFunc:      toFuncID("varbit_send", toInternal("varbit")),
	ModInFunc:     toFuncID("varbittypmodin", toInternal("_cstring")),
	ModOutFunc:    toFuncID("varbittypmodout", toInternal("int4")),
	AnalyzeFunc:   toFuncID("-"),
	Align:         TypeAlignment_Int,
	Storage:       TypeStorage_Extended,
	NotNull:       false,
	BaseTypeID:    id.NullType,
	TypMod:        -1,
	NDims:         0,
	TypCollation:  id.NewCollation("pg_catalog", "default"),
	DefaulBin:     "",
	Default:       "",
	Acl:           nil,
	Checks:        nil,
	attTypMod:     -1,
	CompareFunc:   toFuncID("bttextcmp", toInternal("text"), toInternal("text")),
}

// NewVarBitType returns a VarBit type with type modifier set
// representing the max number of bits in the string.
func NewVarBitType(width int32) (*DoltgresType, error) {
	typmod, err := GetTypModFromCharLength("bit", width)
	if err != nil {
		return nil, err
	}
	newType := *VarBit.WithAttTypMod(typmod)
	return &newType, nil
}
