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

// initInt8 registers the functions to the catalog.
func initInt8() {
	framework.RegisterFunction(int8in)
	framework.RegisterFunction(int8out)
	framework.RegisterFunction(int8recv)
	framework.RegisterFunction(int8send)
	framework.RegisterFunction(btint8cmp)
	framework.RegisterFunction(btint82cmp)
	framework.RegisterFunction(btint84cmp)
}

// int8in represents the PostgreSQL function of int8 type IO input.
var int8in = framework.Function1{
	Name:       "int8in",
	Return:     pgtypes.Int64,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Cstring},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		input := val.(string)
		iVal, err := strconv.ParseInt(strings.TrimSpace(input), 10, 64)
		if err != nil {
			return nil, pgtypes.ErrInvalidSyntaxForType.New("int8", input)
		}
		return iVal, nil
	},
}

// int8out represents the PostgreSQL function of int8 type IO output.
var int8out = framework.Function1{
	Name:       "int8out",
	Return:     pgtypes.Cstring,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Int64},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		if val == nil {
			return nil, nil
		}
		return strconv.FormatInt(val.(int64), 10), nil
	},
}

// int8recv represents the PostgreSQL function of int8 type IO receive.
var int8recv = framework.Function1{
	Name:       "int8recv",
	Return:     pgtypes.Int64,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Internal},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		data := val.([]byte)
		if len(data) == 0 {
			return nil, nil
		}
		return int64(binary.BigEndian.Uint64(data) - (1 << 63)), nil
	},
}

// int8send represents the PostgreSQL function of int8 type IO send.
var int8send = framework.Function1{
	Name:       "int8send",
	Return:     pgtypes.Bytea,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Int64},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		retVal := make([]byte, 8)
		binary.BigEndian.PutUint64(retVal, uint64(val.(int64))+(1<<63))
		return retVal, nil
	},
}

// btint8cmp represents the PostgreSQL function of int8 type compare.
var btint8cmp = framework.Function2{
	Name:       "btint8cmp",
	Return:     pgtypes.Int32,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Int64, pgtypes.Int64},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1, val2 any) (any, error) {
		ab := val1.(int64)
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

// btint82cmp represents the PostgreSQL function of int8 type compare with int2.
var btint82cmp = framework.Function2{
	Name:       "btint82cmp",
	Return:     pgtypes.Int32,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Int64, pgtypes.Int16},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1, val2 any) (any, error) {
		ab := val1.(int64)
		bb := int64(val2.(int16))
		if ab == bb {
			return int32(0), nil
		} else if ab < bb {
			return int32(-1), nil
		} else {
			return int32(1), nil
		}
	},
}

// btint84cmp represents the PostgreSQL function of int8 type compare with int4.
var btint84cmp = framework.Function2{
	Name:       "btint84cmp",
	Return:     pgtypes.Int32,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Int64, pgtypes.Int32},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1, val2 any) (any, error) {
		ab := val1.(int64)
		bb := int64(val2.(int32))
		if ab == bb {
			return int32(0), nil
		} else if ab < bb {
			return int32(-1), nil
		} else {
			return int32(1), nil
		}
	},
}
