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

// InternalCharLength will always be 1.
const InternalCharLength = 1

// InternalChar is a single-byte internal type. In Postgres, it's displayed as "char".
var InternalChar = &DoltgresType{
	OID:           uint32(oid.T_char),
	Name:          "char",
	Schema:        "pg_catalog",
	TypLength:     int16(InternalCharLength),
	PassedByVal:   true,
	TypType:       TypeType_Base,
	TypCategory:   TypeCategory_InternalUseTypes,
	IsPreferred:   false,
	IsDefined:     true,
	Delimiter:     ",",
	RelID:         0,
	SubscriptFunc: toFuncID("-"),
	Elem:          0,
	Array:         uint32(oid.T__char),
	InputFunc:     toFuncID("charin", oid.T_cstring),
	OutputFunc:    toFuncID("charout", oid.T_char),
	ReceiveFunc:   toFuncID("charrecv", oid.T_internal),
	SendFunc:      toFuncID("charsend", oid.T_char),
	ModInFunc:     toFuncID("-"),
	ModOutFunc:    toFuncID("-"),
	AnalyzeFunc:   toFuncID("-"),
	Align:         TypeAlignment_Char,
	Storage:       TypeStorage_Plain,
	NotNull:       false,
	BaseTypeOID:   0,
	TypMod:        -1,
	NDims:         0,
	TypCollation:  0,
	DefaulBin:     "",
	Default:       "",
	Acl:           nil,
	Checks:        nil,
	attTypMod:     -1,
	CompareFunc:   toFuncID("btcharcmp", oid.T_char, oid.T_char),
	InternalName:  `"char"`,
}
