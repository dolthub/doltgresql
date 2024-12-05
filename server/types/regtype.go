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
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/lib/pq/oid"
)

// Regtype is the OID type for finding items in pg_type.
var Regtype = &DoltgresType{
	OID:           uint32(oid.T_regtype),
	Name:          "regtype",
	Schema:        "pg_catalog",
	TypLength:     int16(4),
	PassedByVal:   true,
	TypType:       TypeType_Base,
	TypCategory:   TypeCategory_NumericTypes,
	IsPreferred:   false,
	IsDefined:     true,
	Delimiter:     ",",
	RelID:         0,
	SubscriptFunc: toFuncID("-"),
	Elem:          0,
	Array:         uint32(oid.T__regtype),
	InputFunc:     toFuncID("regtypein", oid.T_cstring),
	OutputFunc:    toFuncID("regtypeout", oid.T_regtype),
	ReceiveFunc:   toFuncID("regtyperecv", oid.T_internal),
	SendFunc:      toFuncID("regtypesend", oid.T_regtype),
	ModInFunc:     toFuncID("-"),
	ModOutFunc:    toFuncID("-"),
	AnalyzeFunc:   toFuncID("-"),
	Align:         TypeAlignment_Int,
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
	CompareFunc:   toFuncID("-"),
}

// Regtype_IoInput is the implementation for IoInput that is being set from another package to avoid circular dependencies.
var Regtype_IoInput func(ctx *sql.Context, input string) (uint32, error)

// Regtype_IoOutput is the implementation for IoOutput that is being set from another package to avoid circular dependencies.
var Regtype_IoOutput func(ctx *sql.Context, oid uint32) (string, error)
