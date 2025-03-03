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
	"fmt"
	"unsafe"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/goccy/go-json"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initJson registers the functions to the catalog.
func initJson() {
	framework.RegisterFunction(json_in)
	framework.RegisterFunction(json_out)
	framework.RegisterFunction(json_recv)
	framework.RegisterFunction(json_send)
	framework.RegisterFunction(json_build_array)
	framework.RegisterFunction(json_build_object)
}

// json_in represents the PostgreSQL function of json type IO input.
var json_in = framework.Function1{
	Name:       "json_in",
	Return:     pgtypes.Json,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Cstring},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
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
	Return:     pgtypes.Cstring,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Json},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		return val.(string), nil
	},
}

// json_recv represents the PostgreSQL function of json type IO receive.
var json_recv = framework.Function1{
	Name:       "json_recv",
	Return:     pgtypes.Json,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Internal},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		data := val.([]byte)
		if len(data) == 0 {
			return nil, nil
		}
		return string(data), nil
	},
}

// json_send represents the PostgreSQL function of json type IO send.
var json_send = framework.Function1{
	Name:       "json_send",
	Return:     pgtypes.Bytea,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Json},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		return []byte(val.(string)), nil
	},
}

// json_build_array represents the PostgreSQL function json_build_array.
var json_build_array = framework.Function1{
	Name:       "json_build_array",
	Return:     pgtypes.Json,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.AnyArray},
	Variadic:   true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val1 any) (any, error) {
		inputArray := val1.([]any)
		json, err := json.Marshal(inputArray)
		return string(json), err
	},
}

// json_build_object represents the PostgreSQL function json_build_object.
var json_build_object = framework.Function1{
	Name:       "json_build_object",
	Return:     pgtypes.Json,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.AnyArray},
	Variadic:   true,
	Callable: func(ctx *sql.Context, argTypes [2]*pgtypes.DoltgresType, val1 any) (any, error) {
		json, err := buildJsonObject("json_build_object", argTypes, val1)
		if err != nil {
			return nil, err
		}
		return string(json), nil
	},
}

// buildJsonObject constructs a json object from the input array provided, which are alternating keys and values.
func buildJsonObject(fnName string, _ [2]*pgtypes.DoltgresType, val1 any) ([]byte, error) {
	inputArray := val1.([]any)
	if len(inputArray)%2 != 0 {
		return nil, sql.ErrInvalidArgumentNumber.New(fnName, "even number of arguments", len(inputArray))
	}
	jsonObject := make(map[string]any)
	var key string
	for i, e := range inputArray {
		if i%2 == 0 {
			var ok bool
			key, ok = e.(string)
			if !ok {
				// TODO: This isn't correct for every type we might use as a value. To get better type info to transform
				//  every value into its string format, we need to pass detailed arg type info for the vararg params (the
				//  unused param in the function call).
				key = fmt.Sprintf("%v", e)
			}
		} else {
			jsonObject[key] = e
			key = ""
		}
	}

	return json.Marshal(jsonObject)
}
