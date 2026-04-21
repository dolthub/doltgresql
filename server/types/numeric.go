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
	"math/big"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/jackc/pgtype"
	"github.com/shopspring/decimal"

	"github.com/dolthub/doltgresql/core/id"
)

const (
	MaxUint32 = 4294967295  // MaxUint32 is the largest possible value of Uint32
	MinInt32  = -2147483648 // MinInt32 is the smallest possible value of Int32
)

var (
	NumericNaN              = pgtype.Numeric{Status: pgtype.Present, NaN: true}
	NumericInfinite         = pgtype.Numeric{Status: pgtype.Present, InfinityModifier: pgtype.Infinity}
	NumericNegativeInfinite = pgtype.Numeric{Status: pgtype.Present, InfinityModifier: pgtype.NegativeInfinity}
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

func SetTypmod(str string, typmod int32) (string, error) {
	dec, err := decimal.NewFromString(str)
	if err != nil {
		return "", err
	}
	precision, scale := GetPrecisionAndScaleFromTypmod(typmod)
	str = dec.StringFixed(scale)
	parts := strings.Split(str, ".")
	if int32(len(parts[0])) > precision-scale && dec.IntPart() != 0 {
		// TODO: split error message to ERROR and DETAIL
		return "", errors.Errorf("numeric field overflow - A field with precision %v, scale %v must round to an absolute value less than 10^%v", precision, scale, precision-scale)
	}
	return str, nil
}

// GetNumericValueWithTypmod returns either given value converted into pgtype.Numeric
// with updated type modifier value if applicable(typmod == -1).
func GetNumericValueWithTypmod(val any, typmod int32) (pgtype.Numeric, error) {
	if val == nil {
		// TODO: should I return nil?
		val = nil
	}
	num, ok := val.(pgtype.Numeric)
	if !ok {
		err := num.Set(val)
		if err != nil {
			return pgtype.Numeric{}, err
		}
	}
	if num.NaN {
		return num, nil
	}
	if num.InfinityModifier == pgtype.Infinity || num.InfinityModifier == pgtype.NegativeInfinity {
		if typmod != -1 {
			return pgtype.Numeric{}, errors.Errorf(`numeric field overflow`)
		}
	}
	if typmod != -1 {
		dec := decimal.NewFromBigInt(num.Int, num.Exp)
		str, err := SetTypmod(dec.String(), typmod)
		if err != nil {
			return pgtype.Numeric{}, err
		}
		// TODO : or decode text???
		err = num.Set(str)
		if err != nil {
			return pgtype.Numeric{}, err
		}
	}
	return num, nil
}

// GetNumeric returns either given value converted into pgtype.Numeric.
func GetNumeric(val any) (pgtype.Numeric, error) {
	return GetNumericValueWithTypmod(val, -1)
}

// GetNumericFromString returns given string converted to pgtype.Numeric value.
// It handles the special values of numeric types including NaN, Infinity, -Infinity, case-insensitive.
func GetNumericFromString(val string, typmod int32) (pgtype.Numeric, error) {
	input := strings.TrimSpace(val)
	switch strings.ToLower(input) {
	case "nan":
		return NumericNaN, nil
	case "inf", "infinity":
		return NumericInfinite, nil
	case "-inf", "-infinity":
		return NumericNegativeInfinite, nil
	}

	if typmod != -1 {
		dec, err := decimal.NewFromString(val)
		if err != nil {
			return pgtype.Numeric{}, err
		}
		precision, scale := GetPrecisionAndScaleFromTypmod(typmod)
		input = dec.StringFixed(scale)
		parts := strings.Split(input, ".")
		if int32(len(parts[0])) > precision-scale && dec.IntPart() != 0 {
			// TODO: split error message to ERROR and DETAIL
			return pgtype.Numeric{}, errors.Errorf("numeric field overflow - A field with precision %v, scale %v must round to an absolute value less than 10^%v", precision, scale, precision-scale)
		}
	}

	var out pgtype.Numeric
	err := out.DecodeText(nil, []byte(input))
	if err != nil {
		return out, err
	}
	return out, nil
}

// NumericCompare compares two pgtype.Numeric values handling NaN, Infinity and -Infinity.
// It uses shopspring/decimal Cmp function logic for all other values.
func NumericCompare(num1 pgtype.Numeric, num2 pgtype.Numeric) int {
	if (num1.NaN && num2.NaN) ||
		(num1.InfinityModifier == pgtype.Infinity && num2.InfinityModifier == pgtype.Infinity) ||
		(num1.InfinityModifier == pgtype.NegativeInfinity && num2.InfinityModifier == pgtype.NegativeInfinity) {
		return 0
	}
	if num1.NaN {
		return 1
	}
	if num2.NaN {
		return -1
	}
	if num1.InfinityModifier == pgtype.Infinity || num2.InfinityModifier == pgtype.NegativeInfinity {
		return 1
	}
	if num1.InfinityModifier == pgtype.NegativeInfinity || num2.InfinityModifier == pgtype.Infinity {
		return -1
	}

	return NumericToDecimal(num1).Cmp(NumericToDecimal(num2))
}

// NumericZeroo converts a pgtype.Numeric to a shopspring decimal.Decimal.
// NOTE: NaN, Infinity, -Infinity values needs to be handled before using this function.
func NumericZeroo() pgtype.Numeric {
	// TODO:
	var num pgtype.Numeric
	_ = num.Set(0)
	return num
}

// NumericToDecimal converts a pgtype.Numeric to a shopspring decimal.Decimal.
// NOTE: NaN, Infinity, -Infinity values needs to be handled before using this function.
func NumericToDecimal(num pgtype.Numeric) decimal.Decimal {
	return decimal.NewFromBigInt(num.Int, num.Exp)
}

func NumericToStringRepresentation(num pgtype.Numeric, typmod int32) string {
	if num.NaN {
		return "NaN"
	}
	if num.InfinityModifier == pgtype.Infinity {
		return "Infinity"
	}
	if num.InfinityModifier == pgtype.NegativeInfinity {
		return "-Infinity"
	}

	dec := decimal.NewFromBigInt(num.Int, num.Exp)
	if typmod == -1 {
		return dec.StringFixed(dec.Exponent() * -1)
	} else {
		_, s := GetPrecisionAndScaleFromTypmod(typmod)
		return dec.StringFixed(s)
	}
}

// DecimalToNumeric converts a shopspring decimal.Decimal to a pgtype.Numeric.
func DecimalToNumeric(d decimal.Decimal) pgtype.Numeric {
	// TODO: check this
	coeff := new(big.Int).Set(d.Coefficient())
	return pgtype.Numeric{Int: coeff, Exp: d.Exponent(), Status: pgtype.Present}
}

// AnyToNumeric attempts to convert an any value to pgtype.Numeric
// including shopspring/decimal.Decimal.
func AnyToNumeric(v any) (pgtype.Numeric, error) {
	var num pgtype.Numeric
	d, ok := v.(decimal.Decimal)
	if !ok {
		err := num.Set(v)
		if err != nil {
			return pgtype.Numeric{}, err
		}
	} else {
		coeff := new(big.Int).Set(d.Coefficient())
		num = pgtype.Numeric{Int: coeff, Exp: d.Exponent(), Status: pgtype.Present}
	}
	return num, nil
}

// serializeTypeNumeric handles serialization from the standard representation to our serialized representation that is
// written in Dolt.
func serializeTypeNumeric(ctx *sql.Context, t *DoltgresType, val any) ([]byte, error) {
	v := val.(pgtype.Numeric)
	var buf []byte
	return v.EncodeBinary(nil, buf)
}

// deserializeTypeNumeric handles deserialization from the Dolt serialized format to our standard representation used by
// expressions and nodes.
func deserializeTypeNumeric(ctx *sql.Context, t *DoltgresType, data []byte) (any, error) {
	if len(data) == 0 {
		return nil, nil
	}
	var out pgtype.Numeric
	err := out.DecodeBinary(nil, data)
	if err != nil {
		return nil, err
	}
	return out, nil
}
