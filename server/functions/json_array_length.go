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
	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initJsonArrayLength registers the functions to the catalog.
func initJsonArrayLength() {
	framework.RegisterFunction(json_array_length_json)
	framework.RegisterFunction(jsonb_array_length_jsonb)
}

// json_array_length_json represents the PostgreSQL function of the same name, taking the same parameters.
var json_array_length_json = framework.Function1{
	Name:       "json_array_length",
	Return:     pgtypes.Int32,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Json},
	Strict:     true,
	Callable:   json_array_length_callable,
}

// jsonb_array_length_jsonb represents the PostgreSQL function of the same name, taking the same parameters.
var jsonb_array_length_jsonb = framework.Function1{
	Name:       "jsonb_array_length",
	Return:     pgtypes.Int32,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.JsonB},
	Strict:     true,
	Callable:   json_array_length_callable,
}

func json_array_length_callable(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
	v, err := jsonValueToInterface(ctx, val)
	if err != nil {
		return nil, err
	}
	arr, ok := v.([]any)
	if !ok {
		// Postgres distinguishes an object from a scalar here, so match its wording for both.
		if _, isObject := v.(map[string]any); isObject {
			return nil, pgerror.WithCandidateCode(errors.New("cannot get array length of a non-array"), pgcode.InvalidParameterValue)
		}
		return nil, pgerror.WithCandidateCode(errors.New("cannot get array length of a scalar"), pgcode.InvalidParameterValue)
	}
	return int32(len(arr)), nil
}
