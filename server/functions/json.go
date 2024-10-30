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
	"unsafe"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/goccy/go-json"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

func initJson() {
	framework.RegisterFunction(json_in)
	framework.RegisterFunction(json_out)
	framework.RegisterFunction(json_recv)
	framework.RegisterFunction(json_send)
}

// json_in represents the PostgreSQL function of json type IO input.
var json_in = framework.Function1{
	Name:       "json_in",
	Return:     pgtypes.Json,
	Parameters: [1]pgtypes.DoltgresType{pgtypes.Text}, // cstring
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, val any) (any, error) {
		input := val.(string)
		if json.Valid(unsafe.Slice(unsafe.StringData(input), len(input))) {
			return input, nil
		}
		return nil, pgtypes.ErrInvalidSyntaxForType.New("json", input[:10]+"...")
	},
}

// json_out represents the PostgreSQL function of json type IO output.
var json_out = framework.Function1{
	Name:       "json_out",
	Return:     pgtypes.Text, // cstring
	Parameters: [1]pgtypes.DoltgresType{pgtypes.Json},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, val any) (any, error) {
		return val.(string), nil
	},
}

// json_recv represents the PostgreSQL function of json type IO receive.
var json_recv = framework.Function1{
	Name:       "json_recv",
	Return:     pgtypes.Json,
	Parameters: [1]pgtypes.DoltgresType{pgtypes.Internal},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, val any) (any, error) {
		switch val := val.(type) {
		case string:
			return val, nil
		default:
			return nil, pgtypes.ErrUnhandledType.New("json", val)
		}
	},
}

// json_send represents the PostgreSQL function of json type IO send.
var json_send = framework.Function1{
	Name:       "json_send",
	Return:     pgtypes.Bytea,
	Parameters: [1]pgtypes.DoltgresType{pgtypes.Json},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]pgtypes.DoltgresType, val any) (any, error) {
		return []byte(val.(string)), nil
	},
}
