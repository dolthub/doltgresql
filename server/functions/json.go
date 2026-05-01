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
	"github.com/dolthub/go-mysql-server/sql/types"
	"github.com/goccy/go-json"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
	"github.com/dolthub/doltgresql/utils"
)

// jsonWrapperToFormattedString converts a sql.JSONWrapper to a formatted JSON string with spaces (JSONB format).
func jsonWrapperToFormattedString(ctx *sql.Context, val sql.JSONWrapper) (string, error) {
	v, err := val.ToInterface(ctx)
	if err != nil {
		return "", err
	}
	return types.JSONDocument{Val: v}.JSONString()
}

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
	Callable:   json_in_callable,
}

func json_in_callable(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
	input := val.(string)
	var jsonVal any
	err := json.Unmarshal(unsafe.Slice(unsafe.StringData(input), len(input)), &jsonVal)
	if err != nil {
		return nil, pgtypes.ErrInvalidSyntaxForType.New("json", input[:10]+"...")
	}
	return types.JSONDocument{Val: jsonVal}, nil
}

// json_out represents the PostgreSQL function of json type IO output.
var json_out = framework.Function1{
	Name:       "json_out",
	Return:     pgtypes.Cstring,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Json},
	Strict:     true,
	Callable:   json_out_callable,
}

func json_out_callable(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
	switch v := val.(type) {
	case string:
		return v, nil
	case sql.JSONWrapper:
		// JSON type is stored as binary JSON (same as JSONB), so output is normalized with spaces
		return jsonWrapperToFormattedString(ctx, v)
	default:
		return nil, fmt.Errorf("unexpected type for json_out: %T", val)
	}
}

// jsonb_out_callable formats a JSONB value for output. JSONB normalizes JSON with spaces after ':' and ','.
func jsonb_out_callable(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
	switch v := val.(type) {
	case string:
		// Parse and reformat the string as proper JSONB (normalized with spaces)
		doc, err := pgtypes.JsonB.IoInput(ctx, v)
		if err != nil {
			return nil, err
		}
		if doc == nil {
			return nil, nil
		}
		return jsonWrapperToFormattedString(ctx, doc.(sql.JSONWrapper))
	case sql.JSONWrapper:
		return jsonWrapperToFormattedString(ctx, v)
	default:
		return nil, fmt.Errorf("unexpected type for jsonb_out: %T", val)
	}
}

// json_recv represents the PostgreSQL function of json type IO receive.
var json_recv = framework.Function1{
	Name:       "json_recv",
	Return:     pgtypes.Json,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Internal},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		data := val.([]byte)
		if data == nil {
			return nil, nil
		}
		return json_in_callable(ctx, [2]*pgtypes.DoltgresType{}, string(data))
	},
}

// json_send represents the PostgreSQL function of json type IO send.
var json_send = framework.Function1{
	Name:       "json_send",
	Return:     pgtypes.Bytea,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Json},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		if wrapper, ok := val.(sql.AnyWrapper); ok {
			var err error
			val, err = wrapper.UnwrapAny(ctx)
			if err != nil {
				return nil, err
			}
			if val == nil {
				return nil, nil
			}
		}
		writer := utils.NewWireWriter()
		var jsonStr string
		switch v := val.(type) {
		case string:
			jsonStr = v
		case sql.JSONWrapper:
			var err error
			jsonStr, err = jsonWrapperToFormattedString(ctx, v)
			if err != nil {
				return nil, err
			}
		default:
			return nil, fmt.Errorf("unexpected type for json_send: %T", val)
		}
		writer.WriteString(jsonStr)
		return writer.BufferData(), nil
	},
}

// json_build_array represents the PostgreSQL function json_build_array.
var json_build_array = framework.Function1{
	Name:       "json_build_array",
	Return:     pgtypes.Json,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.AnyArray},
	Variadic:   true,
	Callable:   json_build_array_callable,
}

func json_build_array_callable(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val1 any) (any, error) {
	inputArray := val1.([]any)
	return types.JSONDocument{Val: inputArray}, nil
}

// json_build_object represents the PostgreSQL function json_build_object.
var json_build_object = framework.Function1{
	Name:       "json_build_object",
	Return:     pgtypes.Json,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.AnyArray},
	Variadic:   true,
	Callable:   json_build_object_callable,
}

func json_build_object_callable(ctx *sql.Context, argTypes [2]*pgtypes.DoltgresType, val1 any) (any, error) {
	json, err := buildJsonObject("json_build_object", argTypes, val1)
	if err != nil {
		return nil, err
	}
	return json, nil
}

// buildJsonObject constructs a json object from the input array provided, which are alternating keys and values.
func buildJsonObject(fnName string, _ [2]*pgtypes.DoltgresType, val1 any) (types.JSONDocument, error) {
	inputArray := val1.([]any)
	if len(inputArray)%2 != 0 {
		return types.JSONDocument{}, sql.ErrInvalidArgumentNumber.New(fnName, "even number of arguments", len(inputArray))
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

	return types.JSONDocument{Val: jsonObject}, nil
}
