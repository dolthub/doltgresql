// Copyright 2026 Dolthub, Inc.
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

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initInt2vector registers the functions to the catalog.
func initInt2vector() {
	framework.RegisterFunction(int2vectorin)
	framework.RegisterFunction(int2vectorout)
	framework.RegisterFunction(int2vectorrecv)
	framework.RegisterFunction(int2vectorsend)
}

// int2vectorin represents the PostgreSQL function of int2vector type IO input.
var int2vectorin = framework.Function1{
	Name:       "int2vectorin",
	Return:     pgtypes.Int16vector,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Cstring},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		input := val.(string)
		strValues := strings.Split(input, " ")
		var values = make([]any, len(strValues))
		for i, strValue := range strValues {
			innerValue, err := pgtypes.Int16.IoInput(ctx, strValue)
			if err != nil {
				return nil, err
			}
			values[i] = innerValue.(int16)
		}
		return values, nil
	},
}

// int2vectorout represents the PostgreSQL function of int2vector type IO output.
var int2vectorout = framework.Function1{
	Name:       "int2vectorout",
	Return:     pgtypes.Cstring,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Int16vector},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		return pgtypes.VectorToString(ctx, val.([]any), pgtypes.Int16)
	},
}

// int2vectorrecv represents the PostgreSQL function of int2vector type IO receive.
var int2vectorrecv = framework.Function1{
	Name:       "int2vectorrecv",
	Return:     pgtypes.Int16vector,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Internal},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		data := val.([]byte)
		return deserializeArray(ctx, data, pgtypes.Int16)
	},
}

// int2vectorsend represents the PostgreSQL function of int2vector type IO send.
var int2vectorsend = framework.Function1{
	Name:       "int2vectorsend",
	Return:     pgtypes.Bytea,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Int16vector},
	Strict:     true,
	Callable: func(ctx *sql.Context, t [2]*pgtypes.DoltgresType, val any) (any, error) {
		return array_send.Callable(ctx, t, val)
	},
}
