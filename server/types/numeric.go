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
	"encoding/binary"
	"strconv"
	"strings"

	"github.com/cockroachdb/apd/v3"
	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/utils"
)

const (
	MaxUint32    = 4294967295  // MaxUint32 is the largest possible value of Uint32
	MinInt32     = -2147483648 // MinInt32 is the smallest possible value of Int32
	MaxPrecision = uint32(100000)
)

const (
	pgNumericNaN     = 0x00000000c0000000
	pgNumericNaNSign = 0xc000

	pgNumericPosInf     = 0x00000000d0000000
	pgNumericPosInfSign = 0xd000

	pgNumericNegInf     = 0x00000000f0000000
	pgNumericNegInfSign = 0xf000
)

var (
	NumericValueMaxInt16  = apd.New(32767, 0)                // NumericValueMaxInt16 is the max Int16 value for NUMERIC types
	NumericValueMaxInt32  = apd.New(2147483647, 0)           // NumericValueMaxInt32 is the max Int32 value for NUMERIC types
	NumericValueMaxInt64  = apd.New(9223372036854775807, 0)  // NumericValueMaxInt64 is the max Int64 value for NUMERIC types
	NumericValueMinInt16  = apd.New(-32768, 0)               // NumericValueMinInt16 is the min Int16 value for NUMERIC types
	NumericValueMinInt32  = apd.New(MinInt32, 0)             // NumericValueMinInt32 is the min Int32 value for NUMERIC types
	NumericValueMinInt64  = apd.New(-9223372036854775808, 0) // NumericValueMinInt64 is the min Int64 value for NUMERIC types
	NumericValueMaxUint32 = apd.New(MaxUint32, 0)            // NumericValueMaxUint32 is the max Uint32 value for NUMERIC types
	NumericNaN            = apd.Decimal{Form: apd.NaN}
	NumericInf            = apd.Decimal{Form: apd.Infinite}
	NumericNegInf         = apd.Decimal{Form: apd.Infinite, Negative: true}
	BaseContext           = apd.BaseContext.WithPrecision(MaxPrecision)
)

// Numeric is a precise and unbounded decimal value.
var Numeric = &DoltgresType{
	ID:                  toInternal("numeric"),
	TypLength:           int16(-1),
	PassedByVal:         false,
	TypType:             TypeType_Base,
	TypCategory:         TypeCategory_NumericTypes,
	IsPreferred:         false,
	IsDefined:           true,
	Delimiter:           ",",
	RelID:               id.Null,
	SubscriptFunc:       toFuncID("-"),
	Elem:                id.NullType,
	Array:               toInternal("_numeric"),
	InputFunc:           toFuncID("numeric_in", toInternal("cstring"), toInternal("oid"), toInternal("int4")),
	OutputFunc:          toFuncID("numeric_out", toInternal("numeric")),
	ReceiveFunc:         toFuncID("numeric_recv", toInternal("internal"), toInternal("oid"), toInternal("int4")),
	SendFunc:            toFuncID("numeric_send", toInternal("numeric")),
	ModInFunc:           toFuncID("numerictypmodin", toInternal("_cstring")),
	ModOutFunc:          toFuncID("numerictypmodout", toInternal("int4")),
	AnalyzeFunc:         toFuncID("-"),
	Align:               TypeAlignment_Int,
	Storage:             TypeStorage_Main,
	NotNull:             false,
	BaseTypeID:          id.NullType,
	TypMod:              -1,
	NDims:               0,
	TypCollation:        id.NullCollation,
	DefaulBin:           "",
	Default:             "",
	Acl:                 nil,
	Checks:              nil,
	attTypMod:           -1,
	CompareFunc:         toFuncID("numeric_cmp", toInternal("numeric"), toInternal("numeric")),
	SerializationFunc:   serializeTypeNumeric,
	DeserializationFunc: deserializeTypeNumeric,
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
	return ((precision << 16) | scale) + 4, nil
}

// GetPrecisionAndScaleFromTypmod takes Numeric type modifier and returns precision and scale values.
func GetPrecisionAndScaleFromTypmod(typmod int32) (int32, int32) {
	typmod -= 4
	scale := typmod & 0xFFFF
	precision := (typmod >> 16) & 0xFFFF
	return precision, scale
}

// GetNumericValueWithTypmod returns either given numeric value or truncated or error
// depending on the precision and scale decoded from given type modifier value.
func GetNumericValueWithTypmod(val apd.Decimal, typmod int32) (apd.Decimal, error) {
	if typmod == -1 {
		return val, nil
	}
	res := new(apd.Decimal)
	precision, scale := GetPrecisionAndScaleFromTypmod(typmod)
	_, err := BaseContext.WithPrecision(uint32(precision)).Quantize(res, &val, -scale)
	if err != nil {
		return apd.Decimal{}, errors.Errorf("numeric field overflow - A field with precision %v, scale %v must round to an absolute value less than 10^%v", precision, scale, precision-scale)
	}
	return *res, nil
}

// GetNumericValueFromStringWithTypmod returns either given numeric value or truncated or error
// depending on the precision and scale decoded from given type modifier value.
func GetNumericValueFromStringWithTypmod(val string, typmod int32) (apd.Decimal, error) {
	dec, cond, err := BaseContext.WithPrecision(MaxPrecision).NewFromString(val)
	if err != nil {
		return apd.Decimal{}, err
	}
	if cond.Inexact() || cond.Rounded() {
		return apd.Decimal{}, errors.Errorf(`numeric precision was lost or truncated for %s`, val)
	}
	return GetNumericValueWithTypmod(*dec, typmod)
}

// serializeTypeNumeric handles serialization from the standard representation to our serialized representation that is
// written in Dolt.
func serializeTypeNumeric(ctx *sql.Context, t *DoltgresType, val any) ([]byte, error) {
	num := val.(apd.Decimal)
	typmod := t.GetAttTypMod()
	writer := utils.NewWireWriter()
	if num.Form == apd.Finite {
		// Short-circuit if this is the zero value
		if num.IsZero() {
			writer.WriteBytes([]byte{0, 0, 0, 0, 0, 0, 0, 0})
			return writer.BufferData(), nil
		}
		// There's a way to do this more efficiently, but we can do that work once this becomes a performance issue.
		// This is based on the terminology used in Postgres' `numeric.c` file
		decStr := num.Text('f')
		isNegative := false
		if strings.HasPrefix(decStr, "-") {
			isNegative = true
			decStr = decStr[1:]
		}
		// Split the integer and fractional parts
		var intPart string
		var fractPart string
		if idx := strings.Index(decStr, "."); idx != -1 {
			intPart = decStr[:idx]
			fractPart = decStr[idx+1:]
		} else {
			intPart = decStr
		}
		// Find the "dscale", which is the number of digits in the fractional part
		var dscale int16
		if typmod != -1 {
			_, dscale32 := GetPrecisionAndScaleFromTypmod(typmod)
			dscale = int16(dscale32)
		} else {
			dscale = int16(len(fractPart))
		}
		// Pad the integer and fractional parts so that we can take groups of 4 numbers
		if intPart == "0" {
			intPart = ""
		} else if len(intPart)%4 != 0 {
			intPart = strings.Repeat("0", 4-(len(intPart)%4)) + intPart
		}
		if len(fractPart)%4 != 0 {
			// remove trailing zeroes on right side before filling it.
			fractPart = strings.TrimRightFunc(fractPart, func(r rune) bool {
				return r == '0'
			})
			fractPart = fractPart + strings.Repeat("0", 4-(len(fractPart)%4))
		}
		// Write the "ndigits" first, or the number of base-10000 digits
		writer.WriteInt16(int16((len(intPart) / 4) + (len(fractPart) / 4)))
		// Write the "weight", which is the number of base-10000 digits in the integer part subtracted by 1
		writer.WriteInt16(int16((len(intPart) / 4) - 1))
		// Write the "sign"
		if isNegative {
			writer.WriteInt16(16384)
		} else {
			writer.WriteInt16(0)
		}
		// Write the "dscale"
		writer.WriteInt16(dscale)
		// Write all of the digits
		fullPart := intPart + fractPart
		for i := 0; i < len(fullPart); i += 4 {
			part, err := strconv.Atoi(fullPart[i : i+4])
			if err != nil {
				return nil, err
			}
			writer.WriteInt16(int16(part))
		}
	} else {
		var buf []byte
		wp := len(buf)
		buf = append(buf, 0, 0, 0, 0, 0, 0, 0, 0)
		if num.Form == apd.NaN {
			binary.BigEndian.PutUint64(buf[wp:], pgNumericNaN)
		} else if num.Form == apd.Infinite {
			if num.Negative {
				binary.BigEndian.PutUint64(buf[wp:], pgNumericNegInf)
			} else {
				binary.BigEndian.PutUint64(buf[wp:], pgNumericPosInf)
			}
		}
		if typmod == -1 {
			binary.BigEndian.PutUint16(buf[6:], uint16(32))
		}
		writer.WriteBytes(buf)
	}
	return writer.BufferData(), nil
}

// deserializeTypeNumeric handles deserialization from the Dolt serialized format to our standard representation used by
// expressions and nodes.
func deserializeTypeNumeric(ctx *sql.Context, t *DoltgresType, data []byte) (any, error) {
	if len(data) == 0 {
		return nil, nil
	}
	reader := utils.NewWireReader(data)
	var d apd.Decimal

	// 1. Read Header
	ndigits := reader.ReadInt16()
	weight := reader.ReadInt16()
	sign := reader.ReadInt16()
	dscale := reader.ReadInt16()

	// 2. Handle Special Values (NaN, Inf)
	// These usually manifest as specific bit patterns in the header
	switch uint16(sign) {
	case 0xC000: // pgNumericNaN
		d.Form = apd.NaN
		return d, nil
	case 0xD000: // pgNumericPosInf
		d.Form = apd.Infinite
		return d, nil
	case 0xF000: // pgNumericNegInf
		d.Form = apd.Infinite
		d.Negative = true
		return d, nil
	}

	// 3. Handle Finite Values
	if ndigits == 0 {
		d.SetInt64(0)
		return d, nil
	}

	// Read base-10000 digits
	digits := make([]int16, ndigits)
	for i := 0; i < int(ndigits); i++ {
		digits[i] = reader.ReadInt16()
	}

	// 4. Convert base-10000 to string for apd.Decimal
	// Each digit is exactly 4 characters wide (except potentially the first)
	var sb strings.Builder
	if sign == 16384 {
		sb.WriteByte('-')
	}

	for i, digit := range digits {
		// Calculate how many 10000-base digits are before the decimal
		// 'weight' is the index of the first digit, where 0 is 10^0 in base 10000
		if i == int(weight)+1 {
			sb.WriteByte('.')
		}

		sDigit := strconv.Itoa(int(digit))
		// Pad with leading zeros if not the very first digit
		if l := len(sDigit); l < 4 {
			padding := 4 - l
			for p := 0; p < padding; p++ {
				sb.WriteByte('0')
			}
		}
		sb.WriteString(sDigit)
	}

	// If weight is larger than digits, we need trailing zeros
	if int(weight) >= len(digits) {
		for i := 0; i < int(weight)-len(digits)+1; i++ {
			sb.WriteString("0000")
		}
	}

	// If weight is negative, we need leading zeros after decimal point
	if weight < 0 {
		// This logic can get complex; using apd.SetString is the safest path
		// but ensure the decimal point is placed correctly based on dscale.
	}

	dec, _, err := BaseContext.NewFromString(sb.String())
	if err != nil {
		return nil, err
	}
	_, _ = BaseContext.Quantize(dec, dec, int32(-dscale))
	return *dec, err
}

// NumericCompare compares two apd.Decimal values handling NaN separately.
func NumericCompare(ab, bb apd.Decimal) int {
	if (ab.Form == apd.NaN && bb.Form == apd.NaN) ||
		(ab.Form == apd.Infinite && bb.Form == apd.Infinite && ab.Negative == bb.Negative) {
		return 0
	}
	if ab.Form == apd.NaN {
		return 1
	}
	if bb.Form == apd.NaN {
		return -1
	}
	return ab.Cmp(&bb)
}
