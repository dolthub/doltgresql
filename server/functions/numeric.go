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

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/jackc/pgtype"

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
		val, err := pgtypes.GetNumericFromString(input, typmod)
		if err != nil {
			return nil, pgtypes.ErrInvalidSyntaxForType.New("numeric", input)
		}
		return val, nil
	},
}

// numeric_out represents the PostgreSQL function of numeric type IO output.
var numeric_out = framework.Function1{
	Name:       "numeric_out",
	Return:     pgtypes.Cstring,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Numeric},
	Strict:     true,
	Callable: func(ctx *sql.Context, t [2]*pgtypes.DoltgresType, val any) (any, error) {
		num := val.(pgtype.Numeric)
		str := pgtypes.NumericToStringRepresentation(num, -1)
		return str, nil
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
		typmod := val3.(int32)
		var out pgtype.Numeric
		err := out.DecodeBinary(nil, data)
		if err != nil {
			return nil, err
		}
		val, err := pgtypes.GetNumericFromString(pgtypes.NumericToStringRepresentation(out, -1), typmod)
		if err != nil {
			return nil, pgtypes.ErrInvalidSyntaxForType.New("numeric", val)
		}
		return val, nil
	},
}

// numeric_send represents the PostgreSQL function of numeric type IO send.
var numeric_send = framework.Function1{
	Name:       "numeric_send",
	Return:     pgtypes.Bytea,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Numeric},
	Strict:     true,
	Callable: func(ctx *sql.Context, t [2]*pgtypes.DoltgresType, val any) (any, error) {
		num := val.(pgtype.Numeric)
		typmod := t[0].GetAttTypMod()
		writer := utils.NewWireWriter()
		if num.NaN || num.InfinityModifier == pgtype.Infinity || num.InfinityModifier == pgtype.NegativeInfinity {
			var buf []byte
			buf, err := num.EncodeBinary(nil, buf)
			if err != nil {
				return nil, err
			}
			if typmod == -1 {
				binary.BigEndian.PutUint16(buf[6:], uint16(32))
			}
			writer.WriteBytes(buf)
			return writer.BufferData(), nil
		}

		// Short-circuit if this is the zero value
		if num.Int != nil && num.Int.Sign() == 0 {
			writer.WriteBytes([]byte{0, 0, 0, 0, 0, 0, 0, 0})
			return writer.BufferData(), nil
		}

		// There's a way to do this more efficiently, but we can do that work once this becomes a performance issue.
		// This is based on the terminology used in Postgres' `numeric.c` file
		decStr := pgtypes.NumericToStringRepresentation(num, -1)
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
		return pgtypes.NumericCompare(val1.(pgtype.Numeric), val2.(pgtype.Numeric)), nil
	},
}
