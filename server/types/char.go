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

// BpChar is a char that has an unbounded length.
var BpChar = &DoltgresType{
	ID:            toInternal("bpchar"),
	TypLength:     int16(-1),
	PassedByVal:   false,
	TypType:       TypeType_Base,
	TypCategory:   TypeCategory_StringTypes,
	IsPreferred:   false,
	IsDefined:     true,
	Delimiter:     ",",
	RelID:         id.Null,
	SubscriptFunc: toFuncID("-"),
	Elem:          id.NullType,
	Array:         toInternal("_bpchar"),
	InputFunc:     toFuncID("bpcharin", toInternal("cstring"), toInternal("oid"), toInternal("int4")),
	OutputFunc:    toFuncID("bpcharout", toInternal("bpchar")),
	ReceiveFunc:   toFuncID("bpcharrecv", toInternal("internal"), toInternal("oid"), toInternal("int4")),
	SendFunc:      toFuncID("bpcharsend", toInternal("bpchar")),
	ModInFunc:     toFuncID("bpchartypmodin", toInternal("_cstring")),
	ModOutFunc:    toFuncID("bpchartypmodout", toInternal("int4")),
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
	CompareFunc:   toFuncID("bpcharcmp", toInternal("bpchar"), toInternal("bpchar")),
}

// NewCharType returns BpChar type with typmod set.
func NewCharType(length int32) (*DoltgresType, error) {
	typmod, err := GetTypModFromCharLength("char", length)
	if err != nil {
		return nil, err
	}
	newType := *BpChar.WithAttTypMod(typmod)
	return &newType, nil
}
