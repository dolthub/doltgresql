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
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/postgres/parser/duration"
	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initInterval registers the functions to the catalog.
func initInterval() {
	framework.RegisterFunction(interval_in)
	framework.RegisterFunction(interval_out)
	framework.RegisterFunction(interval_recv)
	framework.RegisterFunction(interval_send)
	framework.RegisterFunction(intervaltypmodin)
	framework.RegisterFunction(intervaltypmodout)
	framework.RegisterFunction(interval_cmp)
}

// interval_in represents the PostgreSQL function of interval type IO input.
var interval_in = framework.Function3{
	Name:       "interval_in",
	Return:     pgtypes.Interval,
	Parameters: [3]pgtypes.DoltgresType{pgtypes.Text, pgtypes.Oid, pgtypes.Int32}, // cstring
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [4]pgtypes.DoltgresType, val1, val2, val3 any) (any, error) {
		input := val1.(string)
		//oid := val2.(uint32)
		//typmod := val3.(int32) // precision?
		dInterval, err := tree.ParseDInterval(input)
		if err != nil {
			return nil, err
		}
		return dInterval.Duration, nil
	},
}

// interval_out represents the PostgreSQL function of interval type IO output.
var interval_out = framework.Function1{
	Name:       "byteaout",
	Return:     pgtypes.Text, // cstring
	Parameters: [1]pgtypes.DoltgresType{pgtypes.Interval},
	Strict:     true,
	Callable: func(ctx *sql.Context, t [2]pgtypes.DoltgresType, val any) (any, error) {
		return val.(duration.Duration).String(), nil
	},
}

// interval_recv represents the PostgreSQL function of interval type IO receive.
var interval_recv = framework.Function3{
	Name:       "bytearecv",
	Return:     pgtypes.Interval,
	Parameters: [3]pgtypes.DoltgresType{pgtypes.Internal, pgtypes.Oid, pgtypes.Int32},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [4]pgtypes.DoltgresType, val1, val2, val3 any) (any, error) {
		//oid := val2.(uint32)
		//typmod := val3.(int32) // precision?
		switch v := val1.(type) {
		case duration.Duration:
			return v, nil
		default:
			return nil, pgtypes.ErrUnhandledType.New("interval", v)
		}
	},
}

// interval_send represents the PostgreSQL function of interval type IO send.
var interval_send = framework.Function1{
	Name:       "byteasend",
	Return:     pgtypes.Bytea,
	Parameters: [1]pgtypes.DoltgresType{pgtypes.Interval},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, val any) (any, error) {
		return []byte(val.(duration.Duration).String()), nil
	},
}

// intervaltypmodin represents the PostgreSQL function of interval type IO typmod input.
var intervaltypmodin = framework.Function1{
	Name:       "intervaltypmodin",
	Return:     pgtypes.Int32,
	Parameters: [1]pgtypes.DoltgresType{pgtypes.TextArray}, // cstring[]
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, val any) (any, error) {
		// TODO
		return nil, nil
	},
}

// intervaltypmodout represents the PostgreSQL function of interval type IO typmod output.
var intervaltypmodout = framework.Function1{
	Name:       "intervaltypmodout",
	Return:     pgtypes.Text, // cstring
	Parameters: [1]pgtypes.DoltgresType{pgtypes.Int32},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, val any) (any, error) {
		// TODO
		return nil, nil
	},
}

// interval_cmp represents the PostgreSQL function of interval type compare.
var interval_cmp = framework.Function2{
	Name:       "interval_cmp",
	Return:     pgtypes.Int32,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Interval, pgtypes.Interval},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1, val2 any) (any, error) {
		ab := val1.(duration.Duration)
		bb := val2.(duration.Duration)
		return int32(ab.Compare(bb)), nil
	},
}
