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

// Float64 is an float64.
var Float64 = DoltgresType{
	OID:           uint32(oid.T_float8),
	Name:          "float8",
	Schema:        "pg_catalog",
	Owner:         "doltgres", // TODO
	Length:        int16(8),
	PassedByVal:   true,
	TypType:       TypeType_Base,
	TypCategory:   TypeCategory_NumericTypes,
	IsPreferred:   true,
	IsDefined:     true,
	Delimiter:     ",",
	RelID:         0,
	SubscriptFunc: "-",
	Elem:          0,
	Array:         uint32(oid.T__float8),
	InputFunc:     "float8in",
	OutputFunc:    "float8out",
	ReceiveFunc:   "float8recv",
	SendFunc:      "float8send",
	ModInFunc:     "-",
	ModOutFunc:    "-",
	AnalyzeFunc:   "-",
	Align:         TypeAlignment_Double,
	Storage:       TypeStorage_Plain,
	NotNull:       false,
	BaseTypeOID:   0,
	TypMod:        -1,
	NDims:         0,
	Collation:     0,
	DefaulBin:     "",
	Default:       "",
	Acl:           "",
	Checks:        nil,
}
