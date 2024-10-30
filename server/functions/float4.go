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

func initFloat4() {
	framework.RegisterFunction(float4in)
	framework.RegisterFunction(float4out)
	framework.RegisterFunction(float4recv)
	framework.RegisterFunction(float4send)
	framework.RegisterFunction(btfloat4cmp)
	framework.RegisterFunction(btfloat48cmp)
}

// float4in represents the PostgreSQL function of float4 type IO input.
var float4in = framework.Function1{
	Name:       "float4in",
	Return:     pgtypes.Float32,
	Parameters: [1]pgtypes.DoltgresType{pgtypes.Text}, // cstring
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, val any) (any, error) {
		input := val.(string)
		fVal, err := strconv.ParseFloat(strings.TrimSpace(input), 32)
		if err != nil {
			return nil, pgtypes.ErrInvalidSyntaxForType.New("float4", input)
		}
		return float32(fVal), nil
	},
}

// float4out represents the PostgreSQL function of float4 type IO output.
var float4out = framework.Function1{
	Name:       "float4out",
	Return:     pgtypes.Text, // cstring
	Parameters: [1]pgtypes.DoltgresType{pgtypes.Float32},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, val any) (any, error) {
		return strconv.FormatFloat(float64(val.(float32)), 'f', -1, 32), nil
	},
}

// float4recv represents the PostgreSQL function of float4 type IO receive.
var float4recv = framework.Function1{
	Name:       "float4recv",
	Return:     pgtypes.Float32,
	Parameters: [1]pgtypes.DoltgresType{pgtypes.Internal},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, val any) (any, error) {
		switch val := val.(type) {
		case float32:
			return val, nil
		default:
			return nil, pgtypes.ErrUnhandledType.New("float4", val)
		}
	},
}

// float4send represents the PostgreSQL function of float4 type IO send.
var float4send = framework.Function1{
	Name:       "float4send",
	Return:     pgtypes.Bytea,
	Parameters: [1]pgtypes.DoltgresType{pgtypes.Float32},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, val any) (any, error) {
		return []byte(strconv.FormatFloat(float64(val.(float32)), 'g', -1, 32)), nil
	},
}

// btfloat4cmp represents the PostgreSQL function of float4 type compare.
var btfloat4cmp = framework.Function2{
	Name:       "btfloat4cmp",
	Return:     pgtypes.Int32,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Float32, pgtypes.Float32},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1, val2 any) (any, error) {
		ab := val1.(float32)
		bb := val2.(float32)
		if ab == bb {
			return int32(0), nil
		} else if ab < bb {
			return int32(-1), nil
		} else {
			return int32(1), nil
		}
	},
}

// btfloat48cmp represents the PostgreSQL function of float4 type compare with float8.
var btfloat48cmp = framework.Function2{
	Name:       "btfloat48cmp",
	Return:     pgtypes.Int32,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.Float32, pgtypes.Float64},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1, val2 any) (any, error) {
		ab := float64(val1.(float32))
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
