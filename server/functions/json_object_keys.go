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
	"io"
	"sort"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initJsonObjectKeys registers the functions to the catalog.
func initJsonObjectKeys() {
	framework.RegisterFunction(json_object_keys_json)
	framework.RegisterFunction(jsonb_object_keys_jsonb)
}

// json_object_keys_json represents the PostgreSQL function of the same name, taking the same parameters.
var json_object_keys_json = framework.Function1{
	Name:       "json_object_keys",
	Return:     pgtypes.RowTypeWithReturnType(pgtypes.Text),
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Json},
	Strict:     true,
	SRF:        true,
	Callable: func(ctx *sql.Context, t [2]*pgtypes.DoltgresType, val any) (any, error) {
		return jsonObjectKeysIter(ctx, "json_object_keys", val)
	},
}

// jsonb_object_keys_jsonb represents the PostgreSQL function of the same name, taking the same parameters.
var jsonb_object_keys_jsonb = framework.Function1{
	Name:       "jsonb_object_keys",
	Return:     pgtypes.RowTypeWithReturnType(pgtypes.Text),
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.JsonB},
	Strict:     true,
	SRF:        true,
	Callable: func(ctx *sql.Context, t [2]*pgtypes.DoltgresType, val any) (any, error) {
		return jsonObjectKeysIter(ctx, "jsonb_object_keys", val)
	},
}

// jsonObjectKeysIter returns a row iterator over the top-level keys of the given JSON object,
// erroring if the value is not an object. fnName is only used to build that error message.
func jsonObjectKeysIter(ctx *sql.Context, fnName string, val any) (any, error) {
	v, err := jsonValueToInterface(ctx, val)
	if err != nil {
		return nil, err
	}
	obj, ok := v.(map[string]any)
	if !ok {
		// Postgres distinguishes an array from a scalar here, so match its wording for both.
		if _, isArray := v.([]any); isArray {
			return nil, pgerror.WithCandidateCode(errors.Errorf("cannot call %s on an array", fnName), pgcode.InvalidParameterValue)
		}
		return nil, pgerror.WithCandidateCode(errors.Errorf("cannot call %s on a scalar", fnName), pgcode.InvalidParameterValue)
	}
	keys := sortedJsonObjectKeys(obj)
	idx := 0
	return pgtypes.NewSetReturningFunctionRowIter(func(ctx *sql.Context) (sql.Row, error) {
		if idx >= len(keys) {
			return nil, io.EOF
		}
		defer func() { idx++ }()
		return sql.Row{keys[idx]}, nil
	}), nil
}

// sortedJsonObjectKeys returns the keys of a JSON object in jsonb order: shortest first, then
// bytewise. That is the order Postgres stores jsonb keys in, and the order our own json and jsonb
// output functions print them in, so the keys line up with the printed document.
//
// Postgres returns json (as opposed to jsonb) keys in document order, duplicates included, but we
// parse json into a map, so that ordering is not recoverable and we use the same order for both.
func sortedJsonObjectKeys(obj map[string]any) []string {
	keys := make([]string, 0, len(obj))
	for k := range obj {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool {
		if len(keys[i]) != len(keys[j]) {
			return len(keys[i]) < len(keys[j])
		}
		return keys[i] < keys[j]
	})
	return keys
}
