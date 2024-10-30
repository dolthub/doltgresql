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
	Parameters: [3]pgtypes.DoltgresType{pgtypes.Text, pgtypes.Oid, pgtypes.Int32}, // cstring
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [4]pgtypes.DoltgresType, val1, val2, val3 any) (any, error) {
		input := val1.(string)
		val, err := decimal.NewFromString(strings.TrimSpace(input))
		if err != nil {
			return nil, pgtypes.ErrInvalidSyntaxForType.New("numeric", input)
		}
		return val, nil
	},
}

// numeric_out represents the PostgreSQL function of numeric type IO output.
var numeric_out = framework.Function1{
	Name:       "numeric_out",
	Return:     pgtypes.Text, // cstring
	Parameters: [1]pgtypes.DoltgresType{pgtypes.Numeric},
	Strict:     true,
	Callable: func(ctx *sql.Context, t [2]pgtypes.DoltgresType, val any) (any, error) {
		dec := val.(decimal.Decimal)
		//scale := b.Scale
		//if scale == -1 {
		//	scale = dec.Exponent() * -1
		//}
		return dec.StringFixed(dec.Exponent() * -1), nil
	},
}

// numeric_recv represents the PostgreSQL function of numeric type IO receive.
var numeric_recv = framework.Function3{
	Name:       "numeric_recv",
	Return:     pgtypes.Numeric,
	Parameters: [3]pgtypes.DoltgresType{pgtypes.Internal, pgtypes.Oid, pgtypes.Int32},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [4]pgtypes.DoltgresType, val1, val2, val3 any) (any, error) {
		// TODO: should the value be converted here according to typmod?
		switch v := val1.(type) {
		case decimal.Decimal:
			return v, nil
		default:
			return nil, pgtypes.ErrUnhandledType.New("numeric", v)
		}
	},
}

// numeric_send represents the PostgreSQL function of numeric type IO send.
var numeric_send = framework.Function1{
	Name:       "numeric_send",
	Return:     pgtypes.Bytea,
	Parameters: [1]pgtypes.DoltgresType{pgtypes.Numeric},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, val any) (any, error) {
		dec := val.(decimal.Decimal)
		return []byte(dec.StringFixed(dec.Exponent() * -1)), nil
	},
}

// numerictypmodin represents the PostgreSQL function of numeric type IO typmod input.
var numerictypmodin = framework.Function1{
	Name:       "numerictypmodin",
	Return:     pgtypes.Int32,
	Parameters: [1]pgtypes.DoltgresType{pgtypes.TextArray}, // cstring[]
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, val any) (any, error) {
		// TODO: typmod=(precision<<16)∣scale
		return nil, nil
	},
}

// numerictypmodout represents the PostgreSQL function of numeric type IO typmod output.
var numerictypmodout = framework.Function1{
	Name:       "numerictypmodout",
	Return:     pgtypes.Text, // cstring
	Parameters: [1]pgtypes.DoltgresType{pgtypes.Int32},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, val any) (any, error) {
		// TODO
		// Precision = typmod & 0xFFFF
		// Scale = (typmod >> 16) & 0xFFFF
		return nil, nil
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
