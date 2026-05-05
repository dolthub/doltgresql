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

package functions

import (
	"encoding/binary"
	"fmt"
	"strconv"
	"strings"

	"github.com/cockroachdb/apd/v3"
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
	"github.com/dolthub/doltgresql/utils"
)

// initNumeric registers the functions to the catalog.
func initNumeric() {
	framework.RegisterFunction(numeric_in)
	framework.RegisterFunction(numeric_out)
	framework.RegisterFunction(numeric_recv)
	framework.RegisterFunction(numeric_send)
	framework.RegisterFunction(numerictypmodin)
	framework.RegisterFunction(numerictypmodout)
	framework.RegisterFunction(numeric_cmp)
}

// numeric_in represents the PostgreSQL function of numeric type IO input.
var numeric_in = framework.Function3{
	Name:       "numeric_in",
	Return:     pgtypes.Numeric,
	Parameters: [3]*pgtypes.DoltgresType{pgtypes.Cstring, pgtypes.Oid, pgtypes.Int32},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [4]*pgtypes.DoltgresType, val1, val2, val3 any) (any, error) {
		input := val1.(string)
		typmod := val3.(int32)
		dec, _, err := apd.NewFromString(input)
		if err != nil {
			return nil, pgtypes.ErrInvalidSyntaxForType.New("numeric", input)
		}
		return pgtypes.GetNumericValueWithTypmod(*dec, typmod)
	},
}

// numeric_out represents the PostgreSQL function of numeric type IO output.
var numeric_out = framework.Function1{
	Name:       "numeric_out",
	Return:     pgtypes.Cstring,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Numeric},
	Strict:     true,
	Callable: func(ctx *sql.Context, t [2]*pgtypes.DoltgresType, val any) (any, error) {
		typ := t[0]
		dec := val.(apd.Decimal)
		tm := typ.GetAttTypMod()
		dec, err := pgtypes.GetNumericValueWithTypmod(dec, tm)
		if err != nil {
			return nil, err
		}
		return dec.Text('f'), nil
	},
}

// numeric_recv represents the PostgreSQL function of numeric type IO receive.
var numeric_recv = framework.Function3{
	Name:       "numeric_recv",
	Return:     pgtypes.Numeric,
	Parameters: [3]*pgtypes.DoltgresType{pgtypes.Internal, pgtypes.Oid, pgtypes.Int32},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [4]*pgtypes.DoltgresType, val1, val2, val3 any) (any, error) {
		data := val1.([]byte)
		if data == nil {
			return nil, nil
		}
		//typmod := val3.(int32)
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
		dec, _, err := sql.HighPrecisionCtx.NewFromString(sb.String())
		if err != nil {
			return nil, err
		}
		str := dec.Text('f')
		if str == " " {
		}
		_, err = sql.HighPrecisionCtx.Quantize(dec, dec, int32(-dscale))
		if err != nil {
			return nil, err
		}
		return *dec, nil
	},
}

// numeric_send represents the PostgreSQL function of numeric type IO send.
var numeric_send = framework.Function1{
	Name:       "numeric_send",
	Return:     pgtypes.Bytea,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Numeric},
	Strict:     true,
	Callable: func(ctx *sql.Context, t [2]*pgtypes.DoltgresType, val any) (any, error) {
		num := val.(apd.Decimal)
		typmod := t[0].GetAttTypMod()
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
				_, dscale32 := pgtypes.GetPrecisionAndScaleFromTypmod(typmod)
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
	},
}

// numerictypmodin represents the PostgreSQL function of numeric type IO typmod input.
var numerictypmodin = framework.Function1{
	Name:       "numerictypmodin",
	Return:     pgtypes.Int32,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.CstringArray},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		arr := val.([]any)
		if len(arr) == 0 {
			return nil, pgtypes.ErrTypmodArrayMustBe1D.New()
		} else if len(arr) > 2 {
			return nil, pgtypes.ErrInvalidTypMod.New("NUMERIC")
		}

		p, err := strconv.ParseInt(arr[0].(string), 10, 32)
		if err != nil {
			return nil, err
		}
		precision := int32(p)
		scale := int32(0)
		if len(arr) == 2 {
			s, err := strconv.ParseInt(arr[1].(string), 10, 32)
			if err != nil {
				return nil, err
			}
			scale = int32(s)
		}
		return pgtypes.GetTypmodFromNumericPrecisionAndScale(precision, scale)
	},
}

// numerictypmodout represents the PostgreSQL function of numeric type IO typmod output.
var numerictypmodout = framework.Function1{
	Name:       "numerictypmodout",
	Return:     pgtypes.Cstring,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Int32},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		typmod := val.(int32)
		precision, scale := pgtypes.GetPrecisionAndScaleFromTypmod(typmod)
		return fmt.Sprintf("(%v,%v)", precision, scale), nil
	},
}

// numeric_cmp represents the PostgreSQL function of numeric type compare.
var numeric_cmp = framework.Function2{
	Name:       "numeric_cmp",
	Return:     pgtypes.Int32,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Numeric, pgtypes.Numeric},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1, val2 any) (any, error) {
		ab := val1.(apd.Decimal)
		bb := val2.(apd.Decimal)
		return int32(pgtypes.NumericCompare(ab, bb)), nil
	},
}

const (
	pgNumericNaN    = 0x00000000c0000000
	pgNumericPosInf = 0x00000000d0000000
	pgNumericNegInf = 0x00000000f0000000
)
