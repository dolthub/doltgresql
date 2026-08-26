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
	"time"

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
	return arrayElementsToJsonRaw(ctx, arrType.ArrayBaseType(), arr)
}

// arrayElementsToJsonRaw converts array elements to a JSON byte slice. PostgreSQL
// arrays carry one element type regardless of their number of dimensions.
func arrayElementsToJsonRaw(ctx *sql.Context, elemType *pgtypes.DoltgresType, arr []any) (json.RawMessage, error) {
	elements := make([]json.RawMessage, len(arr))
	for i, el := range arr {
		var raw json.RawMessage
		var err error
		if nested, ok := el.([]any); ok {
			raw, err = arrayElementsToJsonRaw(ctx, elemType, nested)
		} else {
			raw, err = valueToJsonRaw(ctx, elemType, el)
		}
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
	if v, ok := val.(pgtypes.JsonDocument); ok {
		sb := strings.Builder{}
		pgtypes.JsonValueFormatter(&sb, v.Value)
		return json.RawMessage(sb.String()), nil
	}
	conversionType := elemType
	if conversionType != nil && conversionType.TypType == pgtypes.TypeType_Domain {
		conversionType = conversionType.DomainUnderlyingBaseType()
	}
	if conversionType != nil && (conversionType.ID.TypeName() == "json" || conversionType.ID.TypeName() == "jsonb") {
		output, err := elemType.IoOutput(ctx, val)
		return json.RawMessage(output), err
	}
	if conversionType != nil {
		switch conversionType.ID.TypeName() {
		case "date":
			return json.Marshal(FormatDateTimeWithBC(val.(time.Time), dateStyleFormatDateOnly_ISO, false))
		case "timestamp":
			return json.Marshal(FormatDateTimeWithBC(val.(time.Time), "2006-01-02T15:04:05.999999", false))
		case "timestamptz":
			location, err := GetServerLocation(ctx)
			if err != nil {
				return nil, err
			}
			t := val.(time.Time).In(location)
			formatted := FormatDateTimeWithBC(t, "2006-01-02T15:04:05.999999", false)
			if strings.HasSuffix(formatted, " BC") {
				formatted = strings.TrimSuffix(formatted, " BC") + t.Format("-07:00") + " BC"
			} else {
				formatted += t.Format("-07:00")
			}
			return json.Marshal(formatted)
		case "bool":
			return json.Marshal(val)
		case "int2", "int4", "int8", "float4", "float8", "numeric":
			return marshalJsonNumber(ctx, elemType, val)
		}
	}
	switch v := val.(type) {
	case types.JSONDocument:
		return json.Marshal(v.Val)
	case sql.JSONWrapper:
		jsonVal, err := v.ToInterface(ctx)
		if err != nil {
			return nil, err
		}
		return json.Marshal(jsonVal)
	case []pgtypes.RecordValue:
		result, err := rowToJson(ctx, conversionType, v, false)
		return json.RawMessage(result), err
	case []any:
		if conversionType != nil && conversionType.IsArrayType() {
			return arrayToJsonRaw(ctx, conversionType, v)
		}
		return arrayElementsToJsonRaw(ctx, nil, v)
	default:
		return marshalJsonTypeOutput(ctx, elemType, val)
	}
}

// marshalJsonNumber formats PostgreSQL numeric types through their output
// functions. Non-finite float and numeric values are JSON strings because they
// are not valid JSON number tokens.
func marshalJsonNumber(ctx *sql.Context, typ *pgtypes.DoltgresType, val any) (json.RawMessage, error) {
	output, err := typ.IoOutput(ctx, val)
	if err != nil {
		return nil, err
	}
	switch output {
	case "NaN":
		return json.Marshal("NaN")
	case "Infinity", "+Inf":
		return json.Marshal("Infinity")
	case "-Infinity", "-Inf":
		return json.Marshal("-Infinity")
	default:
		return json.RawMessage(output), nil
	}
}

// marshalJsonTypeOutput quotes a value's PostgreSQL text representation as a
// JSON string. PostgreSQL uses type output functions for scalar types that are
// not JSON booleans or ordinary finite JSON numbers.
func marshalJsonTypeOutput(ctx *sql.Context, typ *pgtypes.DoltgresType, val any) (json.RawMessage, error) {
	if typ == nil {
		return json.Marshal(val)
	}
	output, err := typ.IoOutput(ctx, val)
	if err != nil {
		return nil, err
	}
	return json.Marshal(output)
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
