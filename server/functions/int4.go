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
	"strconv"
	"strings"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initInt4 registers the functions to the catalog.
func initInt4() {
	framework.RegisterFunction(int4in)
	framework.RegisterFunction(int4out)
	framework.RegisterFunction(int4recv)
	framework.RegisterFunction(int4send)
	framework.RegisterFunction(btint4cmp)
	framework.RegisterFunction(btint42cmp)
	framework.RegisterFunction(btint48cmp)
}

// int4in represents the PostgreSQL function of int4 type IO input.
var int4in = framework.Function1{
	Name:       "int4in",
	Return:     pgtypes.Int32,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Cstring},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		input := val.(string)
		iVal, err := strconv.ParseInt(strings.TrimSpace(input), 10, 32)
		if err != nil {
			return nil, pgtypes.ErrInvalidSyntaxForType.New("int4", input)
		}
		if iVal > 2147483647 || iVal < -2147483648 {
			return nil, pgtypes.ErrValueIsOutOfRangeForType.New(input, "int4")
		}
		return int32(iVal), nil
	},
}

// int4out represents the PostgreSQL function of int4 type IO output.
var int4out = framework.Function1{
	Name:       "int4out",
	Return:     pgtypes.Cstring,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Int32},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		return strconv.FormatInt(int64(val.(int32)), 10), nil
	},
}

// int4recv represents the PostgreSQL function of int4 type IO receive.
var int4recv = framework.Function1{
	Name:       "int4recv",
	Return:     pgtypes.Int32,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Internal},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		data := val.([]byte)
		if len(data) == 0 {
			return nil, nil
		}
		return int32(binary.BigEndian.Uint32(data) - (1 << 31)), nil
	},
}

// int4send represents the PostgreSQL function of int4 type IO send.
var int4send = framework.Function1{
	Name:       "int4send",
	Return:     pgtypes.Bytea,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Int32},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		retVal := make([]byte, 4)
		binary.BigEndian.PutUint32(retVal, uint32(val.(int32))+(1<<31))
		return retVal, nil
	},
}

// btint4cmp represents the PostgreSQL function of int4 type compare.
var btint4cmp = framework.Function2{
	Name:       "btint4cmp",
	Return:     pgtypes.Int32,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Int32, pgtypes.Int32},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1, val2 any) (any, error) {
		ab := val1.(int32)
		bb := val2.(int32)
		if ab == bb {
			return int32(0), nil
		} else if ab < bb {
			return int32(-1), nil
		} else {
			return int32(1), nil
		}
	},
}

// btint42cmp represents the PostgreSQL function of int4 type compare with int2.
var btint42cmp = framework.Function2{
	Name:       "btint42cmp",
	Return:     pgtypes.Int32,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Int32, pgtypes.Int16},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1, val2 any) (any, error) {
		ab := val1.(int32)
		bb := int32(val2.(int16))
		if ab == bb {
			return int32(0), nil
		} else if ab < bb {
			return int32(-1), nil
		} else {
			return int32(1), nil
		}
	},
}

// btint48cmp represents the PostgreSQL function of int4 type compare with int8.
var btint48cmp = framework.Function2{
	Name:       "btint48cmp",
	Return:     pgtypes.Int32,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Int32, pgtypes.Int64},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1, val2 any) (any, error) {
		ab := int64(val1.(int32))
		bb := val2.(int64)
		if ab == bb {
			return int32(0), nil
		} else if ab < bb {
			return int32(-1), nil
		} else {
			return int32(1), nil
		}
	},
}
