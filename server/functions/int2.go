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

// initInt2 registers the functions to the catalog.
func initInt2() {
	framework.RegisterFunction(int2in)
	framework.RegisterFunction(int2out)
	framework.RegisterFunction(int2recv)
	framework.RegisterFunction(int2send)
	framework.RegisterFunction(btint2cmp)
	framework.RegisterFunction(btint24cmp)
	framework.RegisterFunction(btint28cmp)
}

// int2in represents the PostgreSQL function of int2 type IO input.
var int2in = framework.Function1{
	Name:       "int2in",
	Return:     pgtypes.Int16,
	Parameters: [1]pgtypes.DoltgresType{pgtypes.Text}, // cstring
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, val any) (any, error) {
		input := val.(string)
		iVal, err := strconv.ParseInt(strings.TrimSpace(input), 10, 16)
		if err != nil {
			return nil, pgtypes.ErrInvalidSyntaxForType.New("int2", input)
		}
		if iVal > 32767 || iVal < -32768 {
			return nil, pgtypes.ErrValueIsOutOfRangeForType.New(input, "int2")
		}
		return int16(iVal), nil
	},
}

// int2out represents the PostgreSQL function of int2 type IO output.
var int2out = framework.Function1{
	Name:       "int2out",
	Return:     pgtypes.Text, // cstring
	Parameters: [1]pgtypes.DoltgresType{pgtypes.Int16},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, val any) (any, error) {
		return strconv.FormatInt(int64(val.(int16)), 10), nil
	},
}

// int2recv represents the PostgreSQL function of int2 type IO receive.
var int2recv = framework.Function1{
	Name:       "int2recv",
	Return:     pgtypes.Int16,
	Parameters: [1]pgtypes.DoltgresType{pgtypes.Internal},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, val any) (any, error) {
		data := val.([]byte)
		if len(data) == 0 {
			return nil, nil
		}
		return int16(binary.BigEndian.Uint16(data) - (1 << 15)), nil
	},
}

// int2send represents the PostgreSQL function of int2 type IO send.
var int2send = framework.Function1{
	Name:       "int2send",
	Return:     pgtypes.Bytea,
	Parameters: [1]pgtypes.DoltgresType{pgtypes.Int16},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, val any) (any, error) {
		retVal := make([]byte, 2)
		binary.BigEndian.PutUint16(retVal, uint16(val.(int16))+(1<<15))
		return retVal, nil
	},
}

// btint2cmp represents the PostgreSQL function of int2 type compare.
var btint2cmp = framework.Function2{
	Name:       "btint2cmp",
	Return:     pgtypes.Int32,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Int16, pgtypes.Int16},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1, val2 any) (any, error) {
		ab := val1.(int16)
		bb := val2.(int16)
		if ab == bb {
			return int32(0), nil
		} else if ab < bb {
			return int32(-1), nil
		} else {
			return int32(1), nil
		}
	},
}

// btint24cmp represents the PostgreSQL function of int2 type compare with int4.
var btint24cmp = framework.Function2{
	Name:       "btint24cmp",
	Return:     pgtypes.Int32,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Int16, pgtypes.Int32},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1, val2 any) (any, error) {
		ab := int32(val1.(int16))
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

// btint28cmp represents the PostgreSQL function of int2 type compare with int8.
var btint28cmp = framework.Function2{
	Name:       "btint28cmp",
	Return:     pgtypes.Int32,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Int16, pgtypes.Int64},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1, val2 any) (any, error) {
		ab := int64(val1.(int16))
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
