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

	"github.com/cockroachdb/apd/v3"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/types"
	"github.com/goccy/go-json"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initArrayToJson registers the functions to the catalog.
func initArrayToJson() {
	framework.RegisterFunction(array_to_json_anyarray)
	framework.RegisterFunction(array_to_json_anyarray_bool)
}

// array_to_json_anyarray represents the PostgreSQL function of the same name, taking the same parameters.
var array_to_json_anyarray = framework.Function1{
	Name:               "array_to_json",
	Return:             pgtypes.Json,
	Parameters:         [1]*pgtypes.DoltgresType{pgtypes.AnyArray},
	IsNonDeterministic: true,
	Strict:             true,
	Callable: func(ctx *sql.Context, paramsAndReturn [2]*pgtypes.DoltgresType, val any) (any, error) {
		arr := val.([]any)
		raw, err := arrayToJsonRaw(ctx, paramsAndReturn[0], arr)
		if err != nil {
			return nil, err
		}
		return string(raw), nil
	},
}

// array_to_json_anyarray_bool represents the PostgreSQL function of the same name, taking the same parameters.
var array_to_json_anyarray_bool = framework.Function2{
	Name:               "array_to_json",
	Return:             pgtypes.Json,
	Parameters:         [2]*pgtypes.DoltgresType{pgtypes.AnyArray, pgtypes.Bool},
	IsNonDeterministic: true,
	Strict:             true,
	Callable: func(ctx *sql.Context, paramsAndReturn [3]*pgtypes.DoltgresType, val1, val2 any) (any, error) {
		arr := val1.([]any)
		pretty := val2.(bool)
		if !pretty {
			raw, err := arrayToJsonRaw(ctx, paramsAndReturn[0], arr)
			if err != nil {
				return nil, err
			}
			return string(raw), nil
		}
		return arrayToJsonPretty(ctx, paramsAndReturn[0], arr)
	},
}

// arrayToJsonRaw converts an anyarray to a JSON byte slice.
func arrayToJsonRaw(ctx *sql.Context, arrType *pgtypes.DoltgresType, arr []any) (json.RawMessage, error) {
	baseType := arrType.ArrayBaseType()
	elements := make([]json.RawMessage, len(arr))
	for i, el := range arr {
		raw, err := valueToJsonRaw(ctx, baseType, el)
		if err != nil {
			return nil, err
		}
		elements[i] = raw
	}
	return json.Marshal(elements)
}

// valueToJsonRaw converts a single value to a JSON byte slice.
func valueToJsonRaw(ctx *sql.Context, elemType *pgtypes.DoltgresType, val any) (json.RawMessage, error) {
	if val == nil {
		return json.RawMessage("null"), nil
	}
	switch v := val.(type) {
	case *apd.Decimal:
		return json.RawMessage(v.Text('f')), nil
	case pgtypes.JsonDocument:
		sb := strings.Builder{}
		pgtypes.JsonValueFormatter(&sb, v.Value)
		return json.RawMessage(sb.String()), nil
	case types.JSONDocument:
		return json.Marshal(v.Val)
	case sql.JSONWrapper:
		jsonVal, err := v.ToInterface(ctx)
		if err != nil {
			return nil, err
		}
		return json.Marshal(jsonVal)
	case []any:
		return arrayToJsonRaw(ctx, elemType, v)
	case string:
		// JSON-typed values are already valid JSON and should be embedded without re-quoting.
		if elemType != nil && elemType.ID.TypeName() == "json" {
			return json.RawMessage(v), nil
		}
		return json.Marshal(v)
	default:
		return json.Marshal(val)
	}
}

// arrayToJsonPretty produces a pretty-printed JSON array where dimension-1 elements are
// separated by a comma and newline.
func arrayToJsonPretty(ctx *sql.Context, arrType *pgtypes.DoltgresType, arr []any) (string, error) {
	baseType := arrType.ArrayBaseType()
	sb := strings.Builder{}
	sb.WriteRune('[')
	for i, el := range arr {
		if i > 0 {
			sb.WriteString(",\n ")
		}
		raw, err := valueToJsonRaw(ctx, baseType, el)
		if err != nil {
			return "", err
		}
		sb.Write(raw)
	}
	sb.WriteRune(']')
	return sb.String(), nil
}
