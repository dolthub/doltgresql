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
	"reflect"

	"github.com/cockroachdb/apd/v3"
	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/goccy/go-json"
	"github.com/shopspring/decimal"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initJsonTypeof registers the functions to the catalog.
func initJsonTypeof() {
	framework.RegisterFunction(json_typeof_json)
	framework.RegisterFunction(jsonb_typeof_jsonb)
}

// json_typeof_json represents the PostgreSQL function of the same name, taking the same parameters.
var json_typeof_json = framework.Function1{
	Name:       "json_typeof",
	Return:     pgtypes.Text,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Json},
	Strict:     true,
	Callable:   json_typeof_callable,
}

// jsonb_typeof_jsonb represents the PostgreSQL function of the same name, taking the same parameters.
var jsonb_typeof_jsonb = framework.Function1{
	Name:       "jsonb_typeof",
	Return:     pgtypes.Text,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.JsonB},
	Strict:     true,
	Callable:   json_typeof_callable,
}

func json_typeof_callable(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
	v, err := jsonValueToInterface(ctx, val)
	if err != nil {
		return nil, err
	}
	return jsonTypeName(v)
}

// jsonTypeName returns the name that json_typeof and jsonb_typeof report for a JSON value: one of
// "object", "array", "string", "number", "boolean", or "null". Note that the JSON value `null` is
// reported as "null", which is distinct from a SQL NULL input (which yields a SQL NULL result).
func jsonTypeName(val any) (string, error) {
	switch val.(type) {
	case nil:
		return "null", nil
	case bool:
		return "boolean", nil
	case string:
		return "string", nil
	case map[string]any:
		return "object", nil
	case []any:
		return "array", nil
	case json.Number, decimal.Decimal, *decimal.Decimal, apd.Decimal, *apd.Decimal:
		return "number", nil
	}
	// JSON documents read back out of storage can hold numbers as any of Go's numeric types, and
	// nested containers as concretely-typed slices and maps, so fall back to the reflected kind.
	switch reflect.ValueOf(val).Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64:
		return "number", nil
	case reflect.Slice, reflect.Array:
		return "array", nil
	case reflect.Map:
		return "object", nil
	default:
		return "", errors.Errorf("unexpected type in json document: %T", val)
	}
}
