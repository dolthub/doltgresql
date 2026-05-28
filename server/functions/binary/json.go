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

package binary

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/types"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// makeObjectKeyPath builds a JSON path that selects the given top-level
// object key, quoting and escaping the key so any characters are allowed.
func makeObjectKeyPath(key string) string {
	return `$."` + strings.ReplaceAll(key, `"`, `\"`) + `"`
}

// makeArrayIndexPath builds a JSON path that selects the given top-level
// array element by non-negative index.
func makeArrayIndexPath(idx int) string {
	return "$[" + strconv.Itoa(idx) + "]"
}

// These functions can be gathered using the following query from a Postgres 15 instance:
// SELECT * FROM pg_operator o WHERE o.oprname = <OPERATOR> ORDER BY o.oprcode::varchar;
// Replace <OPERATOR> with the desired JSON operator

// initJSON registers the functions to the catalog.
func initJSON() {
	framework.RegisterBinaryFunction(framework.Operator_BinaryJSONExtractJson, json_array_element)
	framework.RegisterBinaryFunction(framework.Operator_BinaryJSONExtractJson, jsonb_array_element)
	framework.RegisterBinaryFunction(framework.Operator_BinaryJSONExtractJson, json_object_field)
	framework.RegisterBinaryFunction(framework.Operator_BinaryJSONExtractJson, jsonb_object_field)
	framework.RegisterBinaryFunction(framework.Operator_BinaryJSONExtractText, json_array_element_text)
	framework.RegisterBinaryFunction(framework.Operator_BinaryJSONExtractText, jsonb_array_element_text)
	framework.RegisterBinaryFunction(framework.Operator_BinaryJSONExtractText, json_object_field_text)
	framework.RegisterBinaryFunction(framework.Operator_BinaryJSONExtractText, jsonb_object_field_text)
	framework.RegisterBinaryFunction(framework.Operator_BinaryJSONExtractPathJson, json_extract_path)
	framework.RegisterBinaryFunction(framework.Operator_BinaryJSONExtractPathJson, jsonb_extract_path)
	framework.RegisterBinaryFunction(framework.Operator_BinaryJSONExtractPathText, json_extract_path_text)
	framework.RegisterBinaryFunction(framework.Operator_BinaryJSONExtractPathText, jsonb_extract_path_text)
	framework.RegisterBinaryFunction(framework.Operator_BinaryJSONContainsRight, jsonb_contains)
	framework.RegisterBinaryFunction(framework.Operator_BinaryJSONContainsLeft, jsonb_contained)
	framework.RegisterBinaryFunction(framework.Operator_BinaryJSONTopLevel, jsonb_exists)
	framework.RegisterBinaryFunction(framework.Operator_BinaryJSONTopLevelAny, jsonb_exists_any)
	framework.RegisterBinaryFunction(framework.Operator_BinaryJSONTopLevelAll, jsonb_exists_all)
	framework.RegisterBinaryFunction(framework.Operator_BinaryMinus, jsonb_delete_text)
	framework.RegisterBinaryFunction(framework.Operator_BinaryMinus, jsonb_delete_text_array)
	framework.RegisterBinaryFunction(framework.Operator_BinaryMinus, jsonb_delete_int32)
}

// toJSONWrapper converts a JSON value (either string or sql.JSONWrapper) to a sql.JSONWrapper.
func toJSONWrapper(ctx *sql.Context, val any) (sql.JSONWrapper, error) {
	switch v := val.(type) {
	case sql.JSONWrapper:
		return v, nil
	case sql.StringWrapper:
		s, err := v.Unwrap(ctx)
		if err != nil {
			return nil, err
		}

		doc, err := pgtypes.JsonB.IoInput(ctx, s)
		if err != nil {
			return nil, err
		}
		if doc == nil {
			return nil, nil
		}
		w, ok := doc.(sql.JSONWrapper)
		if !ok {
			return nil, fmt.Errorf("unexpected type from IoInput: %T", doc)
		}
		return w, nil
	case string:
		doc, err := pgtypes.JsonB.IoInput(ctx, v)
		if err != nil {
			return nil, err
		}
		if doc == nil {
			return nil, nil
		}
		w, ok := doc.(sql.JSONWrapper)
		if !ok {
			return nil, fmt.Errorf("unexpected type from IoInput: %T", doc)
		}
		return w, nil
	default:
		return nil, fmt.Errorf("unexpected type for JSON operation: %T", val)
	}
}

// jsonWrapperElementToText converts a JSON element (sql.JSONWrapper) to its text representation.
// For string values, it returns the raw string without quotes. For other types, it returns the JSON representation.
func jsonWrapperElementToText(ctx *sql.Context, wrapper sql.JSONWrapper) (string, error) {
	v, err := wrapper.ToInterface(ctx)
	if err != nil {
		return "", err
	}
	if v == nil {
		return "", nil
	}
	if s, ok := v.(string); ok {
		return s, nil
	}
	return types.JSONDocument{Val: v}.JSONString()
}

// json_array_element represents the PostgreSQL function of the same name, taking the same parameters.
var json_array_element = framework.Function2{
	Name:       "json_array_element",
	Return:     pgtypes.Json,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Json, pgtypes.Int32},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		newVal, err := toJSONWrapper(ctx, val1)
		if err != nil {
			return nil, err
		}
		if newVal == nil {
			return nil, nil
		}
		var unusedTypes [3]*pgtypes.DoltgresType
		retVal, err := jsonb_array_element_callable(ctx, unusedTypes, newVal, val2)
		if err != nil {
			return nil, err
		}
		if retVal == nil {
			return nil, nil
		}
		return pgtypes.JsonB.IoOutput(ctx, retVal)
	},
}

// jsonb_array_element represents the PostgreSQL function of the same name, taking the same parameters.
var jsonb_array_element = framework.Function2{
	Name:       "jsonb_array_element",
	Return:     pgtypes.JsonB,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.JsonB, pgtypes.Int32},
	Strict:     true,
	Callable:   jsonb_array_element_callable,
}

func jsonb_array_element_callable(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
	wrapper, ok := val1.(sql.JSONWrapper)
	if !ok {
		return nil, nil
	}
	idx := int(val2.(int32))
	// Fast path: for a ComparableJSON wrapper backed by an indexed JSON
	// array, use Lookup to fetch the element without materializing the
	// entire array. Negative indices need the array length, so they fall
	// through to the materialized path even on indexed values.
	if comparable, ok := wrapper.(types.ComparableJSON); ok {
		jt, err := comparable.JsonType(ctx)
		if err != nil {
			return nil, err
		}
		if jt != "ARRAY" {
			return nil, nil
		}
		if idx >= 0 {
			result, err := types.LookupJSONValue(ctx, wrapper, makeArrayIndexPath(idx))
			if err != nil {
				return nil, err
			}
			if result == nil {
				return nil, nil
			}
			return result, nil
		}
	}
	// Materialized fallback: covers wrappers that don't implement
	// ComparableJSON (e.g. literal jsonb values) and the negative-index
	// case on ComparableJSON arrays.
	v, err := wrapper.ToInterface(ctx)
	if err != nil {
		return nil, err
	}
	array, ok := v.([]interface{})
	if !ok {
		return nil, nil
	}
	if idx < 0 {
		idx += len(array)
	}
	if idx < 0 || idx >= len(array) {
		return nil, nil
	}
	return types.JSONDocument{Val: array[idx]}, nil
}

// json_object_field represents the PostgreSQL function of the same name, taking the same parameters.
var json_object_field = framework.Function2{
	Name:       "json_object_field",
	Return:     pgtypes.Json,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Json, pgtypes.Text},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		// TODO: make a bespoke implementation that preserves whitespace
		newVal, err := toJSONWrapper(ctx, val1)
		if err != nil {
			return nil, err
		}
		if newVal == nil {
			return nil, nil
		}
		var unusedTypes [3]*pgtypes.DoltgresType
		retVal, err := jsonb_object_field.Callable(ctx, unusedTypes, newVal, val2)
		if err != nil {
			return nil, err
		}
		if retVal == nil {
			return nil, nil
		}
		return pgtypes.JsonB.IoOutput(ctx, retVal)
	},
}

// jsonb_object_field represents the PostgreSQL function of the same name, taking the same parameters.
var jsonb_object_field = framework.Function2{
	Name:       "jsonb_object_field",
	Return:     pgtypes.JsonB,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.JsonB, pgtypes.Text},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		wrapper, ok := val1.(sql.JSONWrapper)
		if !ok {
			return nil, nil
		}
		key := val2.(string)
		// Fast path: for a ComparableJSON wrapper backed by an indexed JSON
		// object, use Lookup to fetch the value without materializing the
		// entire document.
		if isObj, err := isComparableJsonObject(ctx, wrapper); err != nil {
			return nil, err
		} else if isObj {
			result, err := types.LookupJSONValue(ctx, wrapper, makeObjectKeyPath(key))
			if err != nil {
				return nil, err
			}
			if result == nil {
				return nil, nil
			}
			return result, nil
		}
		// Materialized fallback: covers wrappers that don't implement
		// ComparableJSON (e.g. literal jsonb values), where the embedded
		// jsonpath library has trouble with edge cases like keys that
		// contain escaped quotes, or array operands with text paths.
		v, err := wrapper.ToInterface(ctx)
		if err != nil {
			return nil, err
		}
		obj, ok := v.(map[string]any)
		if !ok {
			return nil, nil
		}
		val, ok := obj[key]
		if !ok {
			return nil, nil
		}
		return types.JSONDocument{Val: val}, nil
	},
}

// json_array_element_text represents the PostgreSQL function of the same name, taking the same parameters.
var json_array_element_text = framework.Function2{
	Name:       "json_array_element_text",
	Return:     pgtypes.Text,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Json, pgtypes.Int32},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		// TODO: make a bespoke implementation that preserves whitespace
		newVal, err := toJSONWrapper(ctx, val1)
		if err != nil {
			return nil, err
		}
		if newVal == nil {
			return nil, nil
		}
		var unusedTypes [3]*pgtypes.DoltgresType
		return jsonb_array_element_text.Callable(ctx, unusedTypes, newVal, val2)
	},
}

// jsonb_array_element_text represents the PostgreSQL function of the same name, taking the same parameters.
var jsonb_array_element_text = framework.Function2{
	Name:       "jsonb_array_element_text",
	Return:     pgtypes.Text,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.JsonB, pgtypes.Int32},
	Strict:     true,
	Callable: func(ctx *sql.Context, dt [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		elem, err := jsonb_array_element_callable(ctx, dt, val1, val2)
		if err != nil || elem == nil {
			return nil, err
		}
		wrapper, ok := elem.(sql.JSONWrapper)
		if !ok {
			return nil, nil
		}
		return jsonWrapperElementToText(ctx, wrapper)
	},
}

// json_object_field_text represents the PostgreSQL function of the same name, taking the same parameters.
var json_object_field_text = framework.Function2{
	Name:       "json_object_field_text",
	Return:     pgtypes.Text,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Json, pgtypes.Text},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		// TODO: make a bespoke implementation that preserves whitespace
		newVal, err := toJSONWrapper(ctx, val1)
		if err != nil {
			return nil, err
		}
		if newVal == nil {
			return nil, nil
		}
		var unusedTypes [3]*pgtypes.DoltgresType
		return jsonb_object_field_text.Callable(ctx, unusedTypes, newVal, val2)
	},
}

// jsonb_object_field_text represents the PostgreSQL function of the same name, taking the same parameters.
var jsonb_object_field_text = framework.Function2{
	Name:       "jsonb_object_field_text",
	Return:     pgtypes.Text,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.JsonB, pgtypes.Text},
	Strict:     true,
	Callable: func(ctx *sql.Context, dt [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		elem, err := jsonb_object_field.Callable(ctx, dt, val1, val2)
		if err != nil || elem == nil {
			return nil, err
		}
		wrapper, ok := elem.(sql.JSONWrapper)
		if !ok {
			return nil, nil
		}
		return jsonWrapperElementToText(ctx, wrapper)
	},
}

// json_extract_path represents the PostgreSQL function of the same name, taking the same parameters.
var json_extract_path = framework.Function2{
	Name:       "json_extract_path",
	Return:     pgtypes.Json,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Json, pgtypes.TextArray},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		// TODO: make a bespoke implementation that preserves whitespace
		newVal, err := toJSONWrapper(ctx, val1)
		if err != nil {
			return nil, err
		}
		if newVal == nil {
			return nil, nil
		}
		var unusedTypes [3]*pgtypes.DoltgresType
		retVal, err := jsonb_extract_path_callable(ctx, unusedTypes, newVal, val2)
		if err != nil {
			return nil, err
		}
		if retVal == nil {
			return nil, nil
		}
		return pgtypes.JsonB.IoOutput(ctx, retVal)
	},
}

// jsonb_extract_path represents the PostgreSQL function of the same name, taking the same parameters.
var jsonb_extract_path = framework.Function2{
	Name:       "jsonb_extract_path",
	Return:     pgtypes.JsonB,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.JsonB, pgtypes.TextArray},
	Strict:     true,
	Callable:   jsonb_extract_path_callable,
}

func jsonb_extract_path_callable(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
	// TODO: we could do this faster by creating a new single path and fetching it via sql.SearchableJSON.Lookup
	cur, ok := val1.(sql.JSONWrapper)
	if !ok {
		return nil, nil
	}
	paths := val2.([]interface{})
	for _, path := range paths {
		if cur == nil {
			return nil, nil
		}
		textPath, ok := path.(string)
		if !ok {
			return nil, nil
		}
		next, err := extractOneJsonPathStep(ctx, cur, textPath)
		if err != nil {
			return nil, err
		}
		if next == nil {
			return nil, nil
		}
		cur = next
	}
	return cur, nil
}

// extractOneJsonPathStep performs a single Postgres-style extract step against
// the given JSON wrapper. The text is treated as a key when the wrapper is an
// object and as an integer index when the wrapper is an array. Returns nil if
// the step cannot be resolved (missing key, out-of-range index, scalar
// wrapper, or non-integer text on an array).
func extractOneJsonPathStep(ctx *sql.Context, cur sql.JSONWrapper, textPath string) (sql.JSONWrapper, error) {
	// Fast path: use the ComparableJSON.JsonType / SearchableJSON.Lookup
	// interfaces to avoid materializing the entire document.
	if comparable, ok := cur.(types.ComparableJSON); ok {
		jt, err := comparable.JsonType(ctx)
		if err != nil {
			return nil, err
		}
		switch jt {
		case "OBJECT":
			return types.LookupJSONValue(ctx, cur, makeObjectKeyPath(textPath))
		case "ARRAY":
			idx, err := strconv.Atoi(textPath)
			if err != nil {
				return nil, nil
			}
			if idx >= 0 {
				return types.LookupJSONValue(ctx, cur, makeArrayIndexPath(idx))
			}
			// Negative indices count from the end and require knowing the
			// array length, so fall through to the materialized path below.
		default:
			return nil, nil
		}
	}
	// Materialized fallback for wrappers that don't implement ComparableJSON,
	// and for negative array indices on ones that do.
	v, err := cur.ToInterface(ctx)
	if err != nil {
		return nil, err
	}
	switch currentValue := v.(type) {
	case map[string]interface{}:
		next, ok := currentValue[textPath]
		if !ok {
			return nil, nil
		}
		return types.JSONDocument{Val: next}, nil
	case []interface{}:
		idx, err := strconv.Atoi(textPath)
		if err != nil {
			return nil, nil
		}
		if idx < 0 {
			idx += len(currentValue)
		}
		if idx < 0 || idx >= len(currentValue) {
			return nil, nil
		}
		return types.JSONDocument{Val: currentValue[idx]}, nil
	default:
		return nil, nil
	}
}

// json_extract_path_text represents the PostgreSQL function of the same name, taking the same parameters.
var json_extract_path_text = framework.Function2{
	Name:       "json_extract_path_text",
	Return:     pgtypes.Text,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Json, pgtypes.TextArray},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		// TODO: make a bespoke implementation that preserves whitespace
		newVal, err := toJSONWrapper(ctx, val1)
		if err != nil {
			return nil, err
		}
		if newVal == nil {
			return nil, nil
		}
		var unusedTypes [3]*pgtypes.DoltgresType
		return jsonb_extract_path_text.Callable(ctx, unusedTypes, newVal, val2)
	},
}

// jsonb_extract_path_text represents the PostgreSQL function of the same name, taking the same parameters.
var jsonb_extract_path_text = framework.Function2{
	Name:       "jsonb_extract_path_text",
	Return:     pgtypes.Text,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.JsonB, pgtypes.TextArray},
	Strict:     true,
	Callable: func(ctx *sql.Context, dt [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		elem, err := jsonb_extract_path_callable(ctx, dt, val1, val2)
		if err != nil || elem == nil {
			return nil, err
		}
		wrapper, ok := elem.(sql.JSONWrapper)
		if !ok {
			return nil, nil
		}
		return jsonWrapperElementToText(ctx, wrapper)
	},
}

// jsonb_contains represents the PostgreSQL function of the same name, taking the same parameters.
var jsonb_contains = framework.Function2{
	Name:       "jsonb_contains",
	Return:     pgtypes.Bool,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.JsonB, pgtypes.JsonB},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		return nil, errors.Errorf("JSON contains is not yet supported")
	},
}

// jsonb_contained represents the PostgreSQL function of the same name, taking the same parameters.
var jsonb_contained = framework.Function2{
	Name:       "jsonb_contained",
	Return:     pgtypes.Bool,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.JsonB, pgtypes.JsonB},
	Strict:     true,
	Callable: func(ctx *sql.Context, dt [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		return jsonb_contains.Callable(ctx, dt, val2, val1)
	},
}

// jsonb_exists represents the PostgreSQL function of the same name, taking the same parameters.
var jsonb_exists = framework.Function2{
	Name:       "jsonb_exists",
	Return:     pgtypes.Bool,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.JsonB, pgtypes.Text},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		wrapper, ok := val1.(sql.JSONWrapper)
		if !ok {
			return false, nil
		}
		key := val2.(string)
		// Fast path: for an indexed JSON object, use Lookup to test for the
		// key without materializing the document.
		if isObj, err := isComparableJsonObject(ctx, wrapper); err != nil {
			return nil, err
		} else if isObj {
			found, err := types.LookupJSONValue(ctx, wrapper, makeObjectKeyPath(key))
			if err != nil {
				return nil, err
			}
			return found != nil, nil
		}
		value, err := wrapper.ToInterface(ctx)
		if err != nil {
			return nil, err
		}
		switch v := value.(type) {
		case map[string]interface{}:
			_, ok := v[key]
			return ok, nil
		case []interface{}:
			for _, item := range v {
				if s, ok := item.(string); ok && s == key {
					return true, nil
				}
			}
			return false, nil
		case string:
			return v == key, nil
		default:
			return false, nil
		}
	},
}

// isComparableJsonObject reports whether the JSON wrapper is a JSON object and
// also implements types.ComparableJSON
func isComparableJsonObject(ctx *sql.Context, wrapper sql.JSONWrapper) (bool, error) {
	comparable, ok := wrapper.(types.ComparableJSON)
	if !ok {
		return false, nil
	}
	jt, err := comparable.JsonType(ctx)
	if err != nil {
		return false, err
	}
	return jt == "OBJECT", nil
}

// jsonb_exists_any represents the PostgreSQL function of the same name, taking the same parameters.
var jsonb_exists_any = framework.Function2{
	Name:       "jsonb_exists_any",
	Return:     pgtypes.Bool,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.JsonB, pgtypes.TextArray},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		wrapper, ok := val1.(sql.JSONWrapper)
		if !ok {
			return false, nil
		}
		keys := val2.([]interface{})
		// Fast path: for an indexed JSON object, test each key with Lookup
		// instead of materializing the document.
		if isObj, err := isComparableJsonObject(ctx, wrapper); err != nil {
			return nil, err
		} else if isObj {
			for _, key := range keys {
				found, err := types.LookupJSONValue(ctx, wrapper, makeObjectKeyPath(key.(string)))
				if err != nil {
					return nil, err
				}
				if found != nil {
					return true, nil
				}
			}
			return false, nil
		}
		value, err := wrapper.ToInterface(ctx)
		if err != nil {
			return nil, err
		}
		switch v := value.(type) {
		case map[string]interface{}:
			for _, key := range keys {
				if _, ok := v[key.(string)]; ok {
					return true, nil
				}
			}
			return false, nil
		case []interface{}:
			for _, key := range keys {
				for _, item := range v {
					if s, ok := item.(string); ok && s == key.(string) {
						return true, nil
					}
				}
			}
			return false, nil
		case string:
			for _, key := range keys {
				if v == key.(string) {
					return true, nil
				}
			}
			return false, nil
		default:
			return false, nil
		}
	},
}

// jsonb_exists_all represents the PostgreSQL function of the same name, taking the same parameters.
var jsonb_exists_all = framework.Function2{
	Name:       "jsonb_exists_all",
	Return:     pgtypes.Bool,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.JsonB, pgtypes.TextArray},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		wrapper, ok := val1.(sql.JSONWrapper)
		if !ok {
			return false, nil
		}
		keys := val2.([]interface{})
		// Fast path: for an indexed JSON object, test each key with Lookup
		// instead of materializing the document.
		if isObj, err := isComparableJsonObject(ctx, wrapper); err != nil {
			return nil, err
		} else if isObj {
			for _, key := range keys {
				found, err := types.LookupJSONValue(ctx, wrapper, makeObjectKeyPath(key.(string)))
				if err != nil {
					return nil, err
				}
				if found == nil {
					return false, nil
				}
			}
			return true, nil
		}
		value, err := wrapper.ToInterface(ctx)
		if err != nil {
			return nil, err
		}
		switch v := value.(type) {
		case map[string]interface{}:
			for _, key := range keys {
				if _, ok := v[key.(string)]; !ok {
					return false, nil
				}
			}
			return true, nil
		case []interface{}:
			for _, key := range keys {
				found := false
				for _, item := range v {
					if s, ok := item.(string); ok && s == key.(string) {
						found = true
						break
					}
				}
				if !found {
					return false, nil
				}
			}
			return true, nil
		case string:
			for _, key := range keys {
				if v != key.(string) {
					return false, nil
				}
			}
			return true, nil
		default:
			return false, nil
		}
	},
}

// jsonb_delete_text represents the PostgreSQL function of the same name, taking the same parameters.
var jsonb_delete_text = framework.Function2{
	Name:       "jsonb_delete",
	Return:     pgtypes.JsonB,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.JsonB, pgtypes.Text},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		return nil, errors.Errorf("JSON deletions are not yet supported")
	},
}

// jsonb_delete_text_array represents the PostgreSQL function of the same name, taking the same parameters.
var jsonb_delete_text_array = framework.Function2{
	Name:       "jsonb_delete",
	Return:     pgtypes.JsonB,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.JsonB, pgtypes.TextArray},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		return nil, errors.Errorf("JSON deletions are not yet supported")
	},
}

// jsonb_delete_int32 represents the PostgreSQL function of the same name, taking the same parameters.
var jsonb_delete_int32 = framework.Function2{
	Name:       "jsonb_delete",
	Return:     pgtypes.JsonB,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.JsonB, pgtypes.Int32},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [3]*pgtypes.DoltgresType, val1 any, val2 any) (any, error) {
		return nil, errors.Errorf("JSON deletions are not yet supported")
	},
}
