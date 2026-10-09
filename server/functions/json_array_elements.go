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

	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/types"

	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initJsonArrayElements registers the functions to the catalog.
func initJsonArrayElements() {
	framework.RegisterFunction(json_array_elements_json)
}

// json_array_elements_json represents the PostgreSQL function of the same name, taking the same parameters.
var json_array_elements_json = framework.Function1{
	Name:       "json_array_elements",
	Return:     pgtypes.RowTypeWithReturnType(pgtypes.Json),
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Json},
	Strict:     true,
	SRF:        true,
	OutParams: sql.Schema{
		{Name: "value", Type: pgtypes.Json, Nullable: true, Source: "json_array_elements"},
	},
	Callable: func(ctx *sql.Context, _ [2]*pgtypes.DoltgresType, val any) (any, error) {
		v, err := jsonValueToInterface(ctx, val)
		if err != nil {
			return nil, err
		}
		arr, ok := v.([]any)
		if !ok {
			if _, isObject := v.(map[string]any); isObject {
				return nil, pgerror.WithCandidateCode(errors.New("cannot call json_array_elements on a non-array"), pgcode.InvalidParameterValue)
			}
			return nil, pgerror.WithCandidateCode(errors.New("cannot call json_array_elements on a scalar"), pgcode.InvalidParameterValue)
		}

		idx := 0
		return pgtypes.NewSetReturningFunctionRowIter(func(ctx *sql.Context) (sql.Row, error) {
			if idx >= len(arr) {
				return nil, io.EOF
			}
			// Wrapping each element retains JSON typing, including JSON null as a non-NULL SQL value.
			row := sql.Row{types.JSONDocument{Val: arr[idx]}}
			idx++
			return row, nil
		}), nil
	},
}
