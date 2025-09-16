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
	"unsafe"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/goccy/go-json"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
	"github.com/dolthub/doltgresql/utils"
)

// initJsonB registers the functions to the catalog.
func initJsonB() {
	framework.RegisterFunction(jsonb_in)
	framework.RegisterFunction(jsonb_out)
	framework.RegisterFunction(jsonb_recv)
	framework.RegisterFunction(jsonb_send)
	framework.RegisterFunction(jsonb_cmp)
	framework.RegisterFunction(jsonb_build_array)
	framework.RegisterFunction(jsonb_build_object)

}

// jsonb_in represents the PostgreSQL function of jsonb type IO input.
var jsonb_in = framework.Function1{
	Name:       "jsonb_in",
	Return:     pgtypes.JsonB,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Cstring},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		input := val.(string)
		inputBytes := unsafe.Slice(unsafe.StringData(input), len(input))
		if json.Valid(inputBytes) {
			doc, err := pgtypes.UnmarshalToJsonDocument(inputBytes)
			return doc, err
		}
		return nil, pgtypes.ErrInvalidSyntaxForType.New("jsonb", input[:10]+"...")
	},
}

// jsonb_out represents the PostgreSQL function of jsonb type IO output.
var jsonb_out = framework.Function1{
	Name:       "jsonb_out",
	Return:     pgtypes.Cstring,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.JsonB},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		sb := strings.Builder{}
		sb.Grow(256)
		pgtypes.JsonValueFormatter(&sb, val.(pgtypes.JsonDocument).Value)
		return sb.String(), nil
	},
}

// jsonb_recv represents the PostgreSQL function of jsonb type IO receive.
var jsonb_recv = framework.Function1{
	Name:       "jsonb_recv",
	Return:     pgtypes.JsonB,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Internal},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		data := val.([]byte)
		if len(data) == 0 {
			return nil, nil
		}
		reader := utils.NewReader(data)
		jsonValue, err := pgtypes.JsonValueDeserialize(reader)
		return pgtypes.JsonDocument{Value: jsonValue}, err
	},
}

// jsonb_send represents the PostgreSQL function of jsonb type IO send.
var jsonb_send = framework.Function1{
	Name:       "jsonb_send",
	Return:     pgtypes.Bytea,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.JsonB},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		res, err := sql.UnwrapAny(ctx, val)
		if err != nil {
			return nil, err
		}
		writer := utils.NewWriter(256)
		pgtypes.JsonValueSerialize(writer, res.(pgtypes.JsonDocument).Value)
		return writer.Data(), nil
	},
}

// jsonb_cmp represents the PostgreSQL function of jsonb type compare.
var jsonb_cmp = framework.Function2{
	Name:       "jsonb_cmp",
	Return:     pgtypes.Int32,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.JsonB, pgtypes.JsonB},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1, val2 any) (any, error) {
		ab := val1.(pgtypes.JsonDocument)
		bb := val2.(pgtypes.JsonDocument)
		return int32(pgtypes.JsonValueCompare(ab.Value, bb.Value)), nil
	},
}

// jsonb_build_array represents the PostgreSQL function jsonb_build_array.
var jsonb_build_array = framework.Function1{
	Name:       "jsonb_build_array",
	Return:     pgtypes.JsonB,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.AnyArray},
	Variadic:   true,
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val1 any) (any, error) {
		inputArray := val1.([]any)
		json, err := json.Marshal(inputArray)
		if err != nil {
			return nil, err
		}

		jsonDoc, err := pgtypes.UnmarshalToJsonDocument(json)
		if err != nil {
			return nil, err
		}

		return jsonDoc, nil
	},
}

// jsonb_build_object represents the PostgreSQL function jsonb_build_object.
var jsonb_build_object = framework.Function1{
	Name:       "jsonb_build_object",
	Return:     pgtypes.JsonB,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.AnyArray},
	Variadic:   true,
	Callable: func(ctx *sql.Context, argTypes [2]*pgtypes.DoltgresType, val1 any) (any, error) {
		json, err := buildJsonObject("jsonb_build_object", argTypes, val1)
		if err != nil {
			return nil, err
		}

		jsonDoc, err := pgtypes.UnmarshalToJsonDocument(json)
		if err != nil {
			return nil, err
		}

		return jsonDoc, nil
	},
}
