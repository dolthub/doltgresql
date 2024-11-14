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
	"fmt"
	"strconv"
	"strings"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/shopspring/decimal"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
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
	Parameters: [3]pgtypes.DoltgresType{pgtypes.Cstring, pgtypes.Oid, pgtypes.Int32},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [4]pgtypes.DoltgresType, val1, val2, val3 any) (any, error) {
		input := val1.(string)
		val, err := decimal.NewFromString(strings.TrimSpace(input))
		if err != nil {
			return nil, pgtypes.ErrInvalidSyntaxForType.New("numeric", input)
		}
		typmod := val3.(int32)
		return pgtypes.GetNumericValueWithTypmod(val, typmod)
	},
}

// numeric_out represents the PostgreSQL function of numeric type IO output.
var numeric_out = framework.Function1{
	Name:       "numeric_out",
	Return:     pgtypes.Cstring,
	Parameters: [1]pgtypes.DoltgresType{pgtypes.Numeric},
	Strict:     true,
	Callable: func(ctx *sql.Context, t [2]pgtypes.DoltgresType, val any) (any, error) {
		typ := t[0]
		dec := val.(decimal.Decimal)
		if typ.AttTypMod == -1 {
			return dec.StringFixed(dec.Exponent() * -1), nil
		} else {
			_, s := pgtypes.GetPrecisionAndScaleFromTypmod(typ.AttTypMod)
			return dec.StringFixed(s), nil
		}
	},
}

// numeric_recv represents the PostgreSQL function of numeric type IO receive.
var numeric_recv = framework.Function3{
	Name:       "numeric_recv",
	Return:     pgtypes.Numeric,
	Parameters: [3]pgtypes.DoltgresType{pgtypes.Internal, pgtypes.Oid, pgtypes.Int32},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [4]pgtypes.DoltgresType, val1, val2, val3 any) (any, error) {
		data := val1.([]byte)
		//typmod := val3.(int32)
		//precision, scale := getPrecisionAndScaleFromTypmod(typmod)
		if len(data) == 0 {
			return nil, nil
		}
		retVal := decimal.NewFromInt(0)
		err := retVal.UnmarshalBinary(data)
		return retVal, err
	},
}

// numeric_send represents the PostgreSQL function of numeric type IO send.
var numeric_send = framework.Function1{
	Name:       "numeric_send",
	Return:     pgtypes.Bytea,
	Parameters: [1]pgtypes.DoltgresType{pgtypes.Numeric},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, val any) (any, error) {
		return val.(decimal.Decimal).MarshalBinary()
	},
}

// numerictypmodin represents the PostgreSQL function of numeric type IO typmod input.
var numerictypmodin = framework.Function1{
	Name:       "numerictypmodin",
	Return:     pgtypes.Int32,
	Parameters: [1]pgtypes.DoltgresType{pgtypes.CstringArray},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, val any) (any, error) {
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
	Parameters: [1]pgtypes.DoltgresType{pgtypes.Int32},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, val any) (any, error) {
		typmod := val.(int32)
		precision, scale := pgtypes.GetPrecisionAndScaleFromTypmod(typmod)
		return fmt.Sprintf("(%v,%v)", precision, scale), nil
	},
}

// numeric_cmp represents the PostgreSQL function of numeric type compare.
var numeric_cmp = framework.Function2{
	Name:       "numeric_cmp",
	Return:     pgtypes.Int32,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Numeric, pgtypes.Numeric},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1, val2 any) (any, error) {
		ab := val1.(decimal.Decimal)
		bb := val2.(decimal.Decimal)
		return int32(ab.Cmp(bb)), nil
	},
}
