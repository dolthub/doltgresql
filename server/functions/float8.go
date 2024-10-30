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
	"strconv"
	"strings"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

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
	Parameters: [1]pgtypes.DoltgresType{pgtypes.Text}, // cstring
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, val any) (any, error) {
		input := val.(string)
		fVal, err := strconv.ParseFloat(strings.TrimSpace(input), 64)
		if err != nil {
			return nil, pgtypes.ErrInvalidSyntaxForType.New("float8", input)
		}
		return float32(fVal), nil
	},
}

// float8out represents the PostgreSQL function of float8 type IO output.
var float8out = framework.Function1{
	Name:       "float8out",
	Return:     pgtypes.Text, // cstring
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
		switch val := val.(type) {
		case float32:
			return val, nil
		default:
			return nil, pgtypes.ErrUnhandledType.New("float8", val)
		}
	},
}

// float8send represents the PostgreSQL function of float8 type IO send.
var float8send = framework.Function1{
	Name:       "float8send",
	Return:     pgtypes.Bytea,
	Parameters: [1]pgtypes.DoltgresType{pgtypes.Float64},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, val any) (any, error) {
		return []byte(strconv.FormatFloat(val.(float64), 'g', -1, 64)), nil
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
