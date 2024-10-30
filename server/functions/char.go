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

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initChar registers the functions to the catalog.
func initChar() {
	framework.RegisterFunction(charin)
	framework.RegisterFunction(charout)
	framework.RegisterFunction(charrecv)
	framework.RegisterFunction(charsend)
	framework.RegisterFunction(btcharcmp)
}

// charin represents the PostgreSQL function of "char" type IO input.
var charin = framework.Function1{
	Name:       "charin",
	Return:     pgtypes.InternalChar,
	Parameters: [1]pgtypes.DoltgresType{pgtypes.Text}, // cstring
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, val any) (any, error) {
		input := val.(string)
		c := []byte(input)
		if uint32(len(c)) > pgtypes.InternalCharLength {
			return input[:pgtypes.InternalCharLength], nil
		}
		return input, nil
	},
}

// charout represents the PostgreSQL function of "char" type IO output.
var charout = framework.Function1{
	Name:       "charout",
	Return:     pgtypes.Text, // cstring
	Parameters: [1]pgtypes.DoltgresType{pgtypes.InternalChar},
	Strict:     true,
	Callable: func(ctx *sql.Context, t [2]pgtypes.DoltgresType, val any) (any, error) {
		str := val.(string)
		if uint32(len(str)) > pgtypes.InternalCharLength {
			return str[:pgtypes.InternalCharLength], nil
		}
		return str, nil
	},
}

// charrecv represents the PostgreSQL function of "char" type IO receive.
var charrecv = framework.Function1{
	Name:       "charrecv",
	Return:     pgtypes.InternalChar,
	Parameters: [1]pgtypes.DoltgresType{pgtypes.Internal},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, val any) (any, error) {
		switch v := val.(type) {
		case string:
			return v, nil
		default:
			return nil, pgtypes.ErrUnhandledType.New("char", v)
		}
	},
}

// charsend represents the PostgreSQL function of "char" type IO send.
var charsend = framework.Function1{
	Name:       "byteasend",
	Return:     pgtypes.Bytea,
	Parameters: [1]pgtypes.DoltgresType{pgtypes.InternalChar},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, val any) (any, error) {
		str := val.(string)
		if uint32(len(str)) > pgtypes.InternalCharLength {
			return str[:pgtypes.InternalCharLength], nil
		}
		return []byte(str), nil
	},
}

// btcharcmp represents the PostgreSQL function of "char" type compare.
var btcharcmp = framework.Function2{
	Name:       "charcmp",
	Return:     pgtypes.Int32,
	Parameters: [2]pgtypes.DoltgresType{pgtypes.InternalChar, pgtypes.InternalChar},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]pgtypes.DoltgresType, val1, val2 any) (any, error) {
		ab := strings.TrimRight(val1.(string), " ")
		bb := strings.TrimRight(val2.(string), " ")
		if ab == bb {
			return int32(0), nil
		} else if ab < bb {
			return int32(-1), nil
		} else {
			return int32(1), nil
		}
	},
}
