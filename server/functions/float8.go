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
	"math"
	"strconv"
	"strings"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initFloat8 registers the functions to the catalog.
func initFloat8() {
	framework.RegisterFunction(float8in)
	framework.RegisterFunction(float8out)
	framework.RegisterFunction(float8recv)
	framework.RegisterFunction(float8send)
	framework.RegisterFunction(btfloat8cmp)
	framework.RegisterFunction(btfloat84cmp)
}

// float8in represents the PostgreSQL function of float8 type IO input.
var float8in = framework.Function1{
	Name:       "float8in",
	Return:     pgtypes.Float64,
	Parameters: [1]pgtypes.DoltgresType{pgtypes.Cstring},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, val any) (any, error) {
		input := val.(string)
		fVal, err := strconv.ParseFloat(strings.TrimSpace(input), 64)
		if err != nil {
			return nil, pgtypes.ErrInvalidSyntaxForType.New("float8", input)
		}
		return fVal, nil
	},
}

// float8out represents the PostgreSQL function of float8 type IO output.
var float8out = framework.Function1{
	Name:       "float8out",
	Return:     pgtypes.Cstring,
	Parameters: [1]pgtypes.DoltgresType{pgtypes.Float64},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, val any) (any, error) {
		return strconv.FormatFloat(val.(float64), 'f', -1, 64), nil
	},
}

// float8recv represents the PostgreSQL function of float8 type IO receive.
var float8recv = framework.Function1{
	Name:       "float8recv",
	Return:     pgtypes.Float64,
	Parameters: [1]pgtypes.DoltgresType{pgtypes.Internal},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, val any) (any, error) {
		data := val.([]byte)
		if len(data) == 0 {
			return nil, nil
		}
		unsignedBits := binary.BigEndian.Uint64(data)
		if unsignedBits&(1<<63) != 0 {
			unsignedBits ^= 1 << 63
		} else {
			unsignedBits = ^unsignedBits
		}
		return math.Float64frombits(unsignedBits), nil
	},
}

// float8send represents the PostgreSQL function of float8 type IO send.
var float8send = framework.Function1{
	Name:       "float8send",
	Return:     pgtypes.Bytea,
	Parameters: [1]pgtypes.DoltgresType{pgtypes.Float64},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, val any) (any, error) {
		f64 := val.(float64)
		retVal := make([]byte, 8)
		// Make the serialized form trivially comparable using bytes.Compare: https://stackoverflow.com/a/54557561
		unsignedBits := math.Float64bits(f64)
		if f64 >= 0 {
			unsignedBits ^= 1 << 63
		} else {
			unsignedBits = ^unsignedBits
		}
		binary.BigEndian.PutUint64(retVal, unsignedBits)
		return retVal, nil
	},
}

// btfloat8cmp represents the PostgreSQL function of float8 type compare.
var btfloat8cmp = framework.Function2{
	Name:       "btfloat8cmp",
	Return:     pgtypes.Int32,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Float64, pgtypes.Float64},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1, val2 any) (any, error) {
		ab := val1.(float64)
		bb := val2.(float64)
		if ab == bb {
			return int32(0), nil
		} else if ab < bb {
			return int32(-1), nil
		} else {
			return int32(1), nil
		}
	},
}

// btfloat84cmp represents the PostgreSQL function of float8 type compare with float4.
var btfloat84cmp = framework.Function2{
	Name:       "btfloat84cmp",
	Return:     pgtypes.Int32,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Float64, pgtypes.Float32},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1, val2 any) (any, error) {
		ab := val1.(float64)
		bb := float64(val2.(float32))
		if ab == bb {
			return int32(0), nil
		} else if ab < bb {
			return int32(-1), nil
		} else {
			return int32(1), nil
		}
	},
}
