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
	"github.com/lib/pq/oid"
)

// BpChar is a char that has an unbounded length.
var BpChar = &DoltgresType{
	OID:           uint32(oid.T_bpchar),
	Name:          "bpchar",
	Schema:        "pg_catalog",
	TypLength:     int16(-1),
	PassedByVal:   false,
	TypType:       TypeType_Base,
	TypCategory:   TypeCategory_StringTypes,
	IsPreferred:   false,
	IsDefined:     true,
	Delimiter:     ",",
	RelID:         0,
	SubscriptFunc: toFuncID("-"),
	Elem:          0,
	Array:         uint32(oid.T__bpchar),
	InputFunc:     toFuncID("bpcharin", oid.T_cstring, oid.T_oid, oid.T_int4),
	OutputFunc:    toFuncID("bpcharout", oid.T_bpchar),
	ReceiveFunc:   toFuncID("bpcharrecv", oid.T_internal, oid.T_oid, oid.T_int4),
	SendFunc:      toFuncID("bpcharsend", oid.T_bpchar),
	ModInFunc:     toFuncID("bpchartypmodin", oid.T__cstring),
	ModOutFunc:    toFuncID("bpchartypmodout", oid.T_int4),
	AnalyzeFunc:   toFuncID("-"),
	Align:         TypeAlignment_Int,
	Storage:       TypeStorage_Extended,
	NotNull:       false,
	BaseTypeOID:   0,
	TypMod:        -1,
	NDims:         0,
	TypCollation:  100,
	DefaulBin:     "",
	Default:       "",
	Acl:           nil,
	Checks:        nil,
	attTypMod:     -1,
	CompareFunc:   toFuncID("bpcharcmp", oid.T_bpchar, oid.T_bpchar),
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
