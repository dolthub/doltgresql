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
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/shopspring/decimal"

	"github.com/dolthub/doltgresql/core/id"
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
var Numeric = &DoltgresType{
	ID:            toInternal("numeric"),
	TypLength:     int16(-1),
	PassedByVal:   false,
	TypType:       TypeType_Base,
	TypCategory:   TypeCategory_NumericTypes,
	IsPreferred:   false,
	IsDefined:     true,
	Delimiter:     ",",
	RelID:         id.Null,
	SubscriptFunc: toFuncID("-"),
	Elem:          id.NullType,
	Array:         toInternal("_numeric"),
	InputFunc:     toFuncID("numeric_in", toInternal("cstring"), toInternal("oid"), toInternal("int4")),
	OutputFunc:    toFuncID("numeric_out", toInternal("numeric")),
	ReceiveFunc:   toFuncID("numeric_recv", toInternal("internal"), toInternal("oid"), toInternal("int4")),
	SendFunc:      toFuncID("numeric_send", toInternal("numeric")),
	ModInFunc:     toFuncID("numerictypmodin", toInternal("_cstring")),
	ModOutFunc:    toFuncID("numerictypmodout", toInternal("int4")),
	AnalyzeFunc:   toFuncID("-"),
	Align:         TypeAlignment_Int,
	Storage:       TypeStorage_Main,
	NotNull:       false,
	BaseTypeID:    id.NullType,
	TypMod:        -1,
	NDims:         0,
	TypCollation:  id.NullCollation,
	DefaulBin:     "",
	Default:       "",
	Acl:           nil,
	Checks:        nil,
	attTypMod:     -1,
	CompareFunc:   toFuncID("numeric_cmp", toInternal("numeric"), toInternal("numeric")),
}

// NewNumericTypeWithPrecisionAndScale returns Numeric type with typmod set.
func NewNumericTypeWithPrecisionAndScale(precision, scale int32) (*DoltgresType, error) {
	typmod, err := GetTypmodFromNumericPrecisionAndScale(precision, scale)
	if err != nil {
		return nil, err
	}
	newType := *Numeric.WithAttTypMod(typmod)
	return &newType, nil
}

// GetTypmodFromNumericPrecisionAndScale takes Numeric type precision and scale and returns the type modifier value.
func GetTypmodFromNumericPrecisionAndScale(precision, scale int32) (int32, error) {
	if precision < 1 || precision > 1000 {
		return 0, errors.Errorf("NUMERIC precision %v must be between 1 and 1000", precision)
	}
	if scale < -1000 || scale > 1000 {
		return 0, errors.Errorf("NUMERIC scale 20000 must be between -1000 and 1000")
	}
	return (precision << 16) | scale, nil
}

// GetPrecisionAndScaleFromTypmod takes Numeric type modifier and returns precision and scale values.
func GetPrecisionAndScaleFromTypmod(typmod int32) (int32, int32) {
	scale := typmod & 0xFFFF
	precision := (typmod >> 16) & 0xFFFF
	return precision, scale
}

// GetNumericValueWithTypmod returns either given numeric value or truncated or error
// depending on the precision and scale decoded from given type modifier value.
func GetNumericValueWithTypmod(val decimal.Decimal, typmod int32) (decimal.Decimal, error) {
	if typmod == -1 {
		return val, nil
	}
	precision, scale := GetPrecisionAndScaleFromTypmod(typmod)
	str := val.StringFixed(scale)
	parts := strings.Split(str, ".")
	if int32(len(parts[0])) > precision-scale && val.IntPart() != 0 {
		// TODO: split error message to ERROR and DETAIL
		return decimal.Decimal{}, errors.Errorf("numeric field overflow - A field with precision %v, scale %v must round to an absolute value less than 10^%v", precision, scale, precision-scale)
	}
	return decimal.NewFromString(str)
}
