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
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/types"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initJsonStripNulls registers the functions to the catalog.
func initJsonStripNulls() {
	framework.RegisterFunction(json_strip_nulls_json)
	framework.RegisterFunction(jsonb_strip_nulls_jsonb)
}

// json_strip_nulls_json represents the PostgreSQL function of the same name, taking the same parameters.
var json_strip_nulls_json = framework.Function1{
	Name:       "json_strip_nulls",
	Return:     pgtypes.Json,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Json},
	Strict:     true,
	Callable:   json_strip_nulls_callable,
}

// jsonb_strip_nulls_jsonb represents the PostgreSQL function of the same name, taking the same parameters.
var jsonb_strip_nulls_jsonb = framework.Function1{
	Name:       "jsonb_strip_nulls",
	Return:     pgtypes.JsonB,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.JsonB},
	Strict:     true,
	Callable:   json_strip_nulls_callable,
}

func json_strip_nulls_callable(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
	v, err := jsonValueToInterface(ctx, val)
	if err != nil {
		return nil, err
	}
	return types.JSONDocument{Val: stripJsonNulls(v)}, nil
}

// stripJsonNulls recursively removes object fields whose value is JSON null. Null array elements
// and a top-level null are left alone, matching Postgres.
func stripJsonNulls(val any) any {
	switch v := val.(type) {
	case map[string]any:
		stripped := make(map[string]any, len(v))
		for key, elem := range v {
			if elem == nil {
				continue
			}
			stripped[key] = stripJsonNulls(elem)
		}
		return stripped
	case []any:
		stripped := make([]any, len(v))
		for i, elem := range v {
			stripped[i] = stripJsonNulls(elem)
		}
		return stripped
	default:
		return val
	}
}
