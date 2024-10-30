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

	"github.com/shopspring/decimal"
)

const (
	MaxUint32 = 4294967295  // MaxUint32 is the largest possible value of Uint32
	MinInt32  = -2147483648 // MinInt32 is the smallest possible value of Int32
)

var (
	NumericValueMaxInt16  = decimal.NewFromInt(32767)                // NumericValueMaxInt16 is the max Int16 value for NUMERIC types
	NumericValueMaxInt32  = decimal.NewFromInt(2147483647)           // NumericValueMaxInt32 is the max Int32 value for NUMERIC types
	NumericValueMaxInt64  = decimal.NewFromInt(9223372036854775807)  // NumericValueMaxInt64 is the max Int64 value for NUMERIC types
	NumericValueMinInt16  = decimal.NewFromInt(-32768)               // NumericValueMinInt16 is the min Int16 value for NUMERIC types
	NumericValueMinInt32  = decimal.NewFromInt(MinInt32)             // NumericValueMinInt32 is the min Int32 value for NUMERIC types
	NumericValueMinInt64  = decimal.NewFromInt(-9223372036854775808) // NumericValueMinInt64 is the min Int64 value for NUMERIC types
	NumericValueMaxUint32 = decimal.NewFromInt(MaxUint32)            // NumericValueMaxUint32 is the max Uint32 value for NUMERIC types
)

// Numeric is a precise and unbounded decimal value.
var Numeric = DoltgresType{
	OID:           uint32(oid.T_numeric),
	Name:          "numeric",
	Schema:        "pg_catalog",
	Owner:         "doltgres", // TODO
	Length:        int16(-1),
	PassedByVal:   false,
	TypType:       TypeType_Base,
	TypCategory:   TypeCategory_NumericTypes,
	IsPreferred:   false,
	IsDefined:     true,
	Delimiter:     ",",
	RelID:         0,
	SubscriptFunc: "-",
	Elem:          0,
	Array:         uint32(oid.T__numeric),
	InputFunc:     "numeric_in",
	OutputFunc:    "numeric_out",
	ReceiveFunc:   "numeric_recv",
	SendFunc:      "numeric_send",
	ModInFunc:     "numerictypmodin",
	ModOutFunc:    "numerictypmodout",
	AnalyzeFunc:   "-",
	Align:         TypeAlignment_Int,
	Storage:       TypeStorage_Main,
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

func NewNumericType(precision, scale int32) DoltgresType {
	// TODO: implement precision and scale
	return Numeric
}
