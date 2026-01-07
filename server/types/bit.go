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

// ErrWrongLengthBit is returned when a value with the incorrect length is inserted into a Bit column.
var ErrWrongLengthBit = errors.NewKind(`bit string length %d does not match type bit(%d)`)

// Bit is a fixed-length bit string.
var Bit = &DoltgresType{
	ID:            toInternal("bit"),
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
	Array:         toInternal("_bit"),
	InputFunc:     toFuncID("bit_in", toInternal("cstring"), toInternal("oid"), toInternal("int4")),
	OutputFunc:    toFuncID("bit_out", toInternal("bit")),
	ReceiveFunc:   toFuncID("bit_recv", toInternal("internal"), toInternal("oid"), toInternal("int4")),
	SendFunc:      toFuncID("bit_send", toInternal("bit")),
	ModInFunc:     toFuncID("bittypmodin", toInternal("_cstring")),
	ModOutFunc:    toFuncID("bittypmodout", toInternal("int4")),
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

// NewBitType returns a Bit type with type modifier set
// representing the number of bits in the string.
func NewBitType(width int32) (*DoltgresType, error) {
	typmod, err := GetTypModFromCharLength("bit", width)
	if err != nil {
		return nil, err
	}
	newType := *Bit.WithAttTypMod(typmod)
	return &newType, nil
}
