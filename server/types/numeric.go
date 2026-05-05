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
	"math"

	"github.com/cockroachdb/apd/v3"
	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/types"

	"github.com/dolthub/doltgresql/core/id"
)

var (
	NumericValueMaxInt16 = types.DecimalFromInt64(math.MaxInt16) // NumericValueMaxInt16 is the max Int16 value for NUMERIC types
	NumericValueMaxInt32 = types.DecimalFromInt64(math.MaxInt32) // NumericValueMaxInt32 is the max Int32 value for NUMERIC types
	NumericValueMaxInt64 = types.DecimalFromInt64(math.MaxInt64) // NumericValueMaxInt64 is the max Int64 value for NUMERIC types
	NumericValueMinInt16 = types.DecimalFromInt64(math.MinInt16) // NumericValueMinInt16 is the min Int16 value for NUMERIC types
	NumericValueMinInt32 = types.DecimalFromInt64(math.MinInt32) // NumericValueMinInt32 is the min Int32 value for NUMERIC types
	NumericValueMinInt64 = types.DecimalFromInt64(math.MinInt64) // NumericValueMinInt64 is the min Int64 value for NUMERIC types
	NumericNaN           = apd.Decimal{Form: apd.NaN}
	NumericInf           = apd.Decimal{Form: apd.Infinite}
	NumericNegInf        = apd.Decimal{Form: apd.Infinite, Negative: true}
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
	_, err := sql.DecimalCtx.WithPrecision(uint32(precision)).Quantize(res, &val, -scale)
	if err != nil {
		return apd.Decimal{}, errors.Errorf("numeric field overflow - A field with precision %v, scale %v must round to an absolute value less than 10^%v", precision, scale, precision-scale)
	}
	return *res, nil
}

// GetNumericValueFromStringWithTypmod returns either given numeric value or truncated or error
// depending on the precision and scale decoded from given type modifier value.
func GetNumericValueFromStringWithTypmod(val string, typmod int32) (apd.Decimal, error) {
	dec, cond, err := sql.HighPrecisionCtx.NewFromString(val)
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
	d := val.(apd.Decimal)
	return d.MarshalText()
}

// deserializeTypeNumeric handles deserialization from the Dolt serialized format to our standard representation used by
// expressions and nodes.
func deserializeTypeNumeric(ctx *sql.Context, t *DoltgresType, data []byte) (any, error) {
	if len(data) == 0 {
		return nil, nil
	}
	retVal := *apd.New(0, 0)
	err := retVal.UnmarshalText(data)
	return retVal, err
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
