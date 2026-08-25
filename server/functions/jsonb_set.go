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
	"fmt"
	"strconv"
	"strings"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/types"

	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initJsonbSet registers the jsonb_set overloads. PostgreSQL defines the fourth
// argument with a default of true; built-in defaults are represented here by a
// three-argument overload that delegates to the same implementation.
func initJsonbSet() {
	framework.RegisterFunction(jsonb_set_jsonb_text_array_jsonb)
	framework.RegisterFunction(jsonb_set_jsonb_text_array_jsonb_bool)
}

var jsonb_set_jsonb_text_array_jsonb = framework.Function3{
	Name:       "jsonb_set",
	Return:     pgtypes.JsonB,
	Parameters: [3]*pgtypes.DoltgresType{pgtypes.JsonB, pgtypes.TextArray, pgtypes.JsonB},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [4]*pgtypes.DoltgresType, target, path, newValue any) (any, error) {
		return jsonbSet(ctx, target, path, newValue, true)
	},
}

var jsonb_set_jsonb_text_array_jsonb_bool = framework.Function4{
	Name:       "jsonb_set",
	Return:     pgtypes.JsonB,
	Parameters: [4]*pgtypes.DoltgresType{pgtypes.JsonB, pgtypes.TextArray, pgtypes.JsonB, pgtypes.Bool},
	Strict:     true,
	Callable: func(ctx *sql.Context, _ [5]*pgtypes.DoltgresType, target, path, newValue, createIfMissing any) (any, error) {
		return jsonbSet(ctx, target, path, newValue, createIfMissing.(bool))
	},
}

func jsonbSet(ctx *sql.Context, target, path, newValue any, createIfMissing bool) (any, error) {
	targetValue, err := jsonbValueToInterface(ctx, target)
	if err != nil {
		return nil, err
	}
	newInterface, err := jsonbValueToInterface(ctx, newValue)
	if err != nil {
		return nil, err
	}
	pathElements, err := jsonbSetPath(ctx, path)
	if err != nil {
		return nil, err
	}

	// PostgreSQL rejects scalar targets even when the path is empty.
	switch targetValue.(type) {
	case map[string]any, []any:
	default:
		return nil, pgerror.WithCandidateCode(fmt.Errorf("cannot set path in scalar"), pgcode.InvalidParameterValue)
	}

	if len(pathElements) == 0 {
		return types.JSONDocument{Val: targetValue}, nil
	}
	result, _, err := jsonbSetAtPath(targetValue, pathElements, newInterface, createIfMissing, 0)
	if err != nil {
		return nil, err
	}
	return types.JSONDocument{Val: result}, nil
}

func jsonbValueToInterface(ctx *sql.Context, value any) (any, error) {
	unwrapped, err := sql.UnwrapAny(ctx, value)
	if err != nil {
		return nil, err
	}
	if bytesValue, ok := unwrapped.(types.JSONBytes); ok {
		bytes, err := bytesValue.GetBytes(ctx)
		if err != nil {
			return nil, err
		}
		return pgtypes.DecodeJSONValue(bytes)
	}
	wrapper, ok := unwrapped.(sql.JSONWrapper)
	if !ok {
		return nil, fmt.Errorf("expected JSON wrapper, got %T", unwrapped)
	}
	return wrapper.ToInterface(ctx)
}

func jsonbSetPath(ctx *sql.Context, value any) ([]string, error) {
	unwrapped, err := sql.UnwrapAny(ctx, value)
	if err != nil {
		return nil, err
	}
	values, ok := unwrapped.([]any)
	if !ok {
		return nil, fmt.Errorf("expected text array, got %T", unwrapped)
	}
	path := make([]string, len(values))
	for i, value := range values {
		if value == nil {
			return nil, pgerror.WithCandidateCode(
				fmt.Errorf("path element at position %d is null", i+1),
				pgcode.NullValueNotAllowed,
			)
		}
		path[i], err = framework.UnwrapString(ctx, value)
		if err != nil {
			return nil, err
		}
	}
	return path, nil
}

// jsonbSetAtPath returns the updated value and whether the requested path was
// found or created. It copies only containers along the modified path, leaving
// the input JSON document untouched.
func jsonbSetAtPath(current any, path []string, newValue any, createIfMissing bool, pathOffset int) (any, bool, error) {
	key := path[0]
	last := len(path) == 1

	switch current := current.(type) {
	case map[string]any:
		existing, found := current[key]
		if !found {
			if !last || !createIfMissing {
				return current, false, nil
			}
			result := cloneJSONObject(current)
			result[key] = newValue
			return result, true, nil
		}

		if last {
			result := cloneJSONObject(current)
			result[key] = newValue
			return result, true, nil
		}
		updated, changed, err := jsonbSetAtPath(existing, path[1:], newValue, createIfMissing, pathOffset+1)
		if err != nil || !changed {
			return current, false, err
		}
		result := cloneJSONObject(current)
		result[key] = updated
		return result, true, nil

	case []any:
		// PostgreSQL parses JSON array path indexes as signed 32-bit integers.
		// Its integer parser accepts leading C-locale whitespace, but deliberately
		// rejects trailing whitespace. Keep key unchanged for the error message.
		indexText := strings.TrimLeft(key, " \t\n\r\f\v")
		parsedIndex, err := strconv.ParseInt(indexText, 10, 32)
		if err != nil {
			return nil, false, pgerror.WithCandidateCode(
				fmt.Errorf("path element at position %d is not an integer: %q", pathOffset+1, key),
				pgcode.InvalidTextRepresentation,
			)
		}
		index := parsedIndex
		if index < 0 {
			index += int64(len(current))
		}
		if index >= 0 && index < int64(len(current)) {
			if last {
				result := append([]any(nil), current...)
				result[index] = newValue
				return result, true, nil
			}
			updated, changed, err := jsonbSetAtPath(current[index], path[1:], newValue, createIfMissing, pathOffset+1)
			if err != nil || !changed {
				return current, false, err
			}
			result := append([]any(nil), current...)
			result[index] = updated
			return result, true, nil
		}
		if !last || !createIfMissing {
			return current, false, nil
		}
		if parsedIndex < 0 {
			return append([]any{newValue}, current...), true, nil
		}
		return append(append([]any(nil), current...), newValue), true, nil

	default:
		// PostgreSQL does not create missing intermediate containers and leaves
		// the document unchanged when traversal reaches a scalar.
		return current, false, nil
	}
}

func cloneJSONObject(value map[string]any) map[string]any {
	result := make(map[string]any, len(value)+1)
	for key, item := range value {
		result[key] = item
	}
	return result
}
