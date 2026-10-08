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

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initToJsonB registers the functions to the catalog.
func initToJsonB() {
	framework.RegisterFunction(to_jsonb_anyelement)
}

// to_jsonb_anyelement represents the PostgreSQL function of the same name, taking the same parameters.
var to_jsonb_anyelement = framework.Function1{
	Name:               "to_jsonb",
	Return:             pgtypes.JsonB,
	Parameters:         [1]*pgtypes.DoltgresType{pgtypes.AnyElement},
	IsNonDeterministic: true,
	Strict:             true,
	Callable: func(ctx *sql.Context, paramsAndReturn [2]*pgtypes.DoltgresType, val any) (any, error) {
		raw, err := valueToJsonRaw(ctx, paramsAndReturn[0], val)
		if err != nil {
			return nil, err
		}
		// Unlike to_json, which hands back the rendered text, jsonb is a parsed document: object
		// keys are deduplicated and reordered, and insignificant whitespace is dropped.
		return json_in_callable(ctx, [2]*pgtypes.DoltgresType{}, string(raw))
	},
}
