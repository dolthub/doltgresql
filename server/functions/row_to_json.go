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
	"strings"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/goccy/go-json"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initRowToJson registers the functions to the catalog.
func initRowToJson() {
	framework.RegisterFunction(row_to_json_record)
	framework.RegisterFunction(row_to_json_record_bool)
}

// row_to_json_record represents the PostgreSQL function of the same name, taking the same parameters.
var row_to_json_record = framework.Function1{
	Name:       "row_to_json",
	Return:     pgtypes.Json,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Record},
	Strict:     true,
	Callable: func(ctx *sql.Context, paramsAndReturn [2]*pgtypes.DoltgresType, val any) (any, error) {
		values := val.([]pgtypes.RecordValue)
		return rowToJson(ctx, paramsAndReturn[0], values, false)
	},
}

// row_to_json_record_bool represents the PostgreSQL function of the same name, taking the same parameters.
var row_to_json_record_bool = framework.Function2{
	Name:       "row_to_json",
	Return:     pgtypes.Json,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.Record, pgtypes.Bool},
	Strict:     true,
	Callable: func(ctx *sql.Context, paramsAndReturn [3]*pgtypes.DoltgresType, val1, val2 any) (any, error) {
		values := val1.([]pgtypes.RecordValue)
		pretty := val2.(bool)
		return rowToJson(ctx, paramsAndReturn[0], values, pretty)
	},
}

// rowToJson converts an array of RecordValue to a JSON object string.
func rowToJson(ctx *sql.Context, rowType *pgtypes.DoltgresType, values []pgtypes.RecordValue, pretty bool) (string, error) {
	keys := fieldNames(rowType, len(values))

	sb := strings.Builder{}
	sb.WriteRune('{')
	for i, rv := range values {
		if i > 0 {
			if pretty {
				sb.WriteString(",\n ")
			} else {
				sb.WriteRune(',')
			}
		}

		keyBytes, err := json.Marshal(keys[i])
		if err != nil {
			return "", err
		}
		sb.Write(keyBytes)
		sb.WriteRune(':')

		var elemType *pgtypes.DoltgresType
		if dgt, ok := rv.Type.(*pgtypes.DoltgresType); ok {
			elemType = dgt
		}
		raw, err := valueToJsonRaw(ctx, elemType, rv.Value)
		if err != nil {
			return "", err
		}
		sb.Write(raw)
	}
	sb.WriteRune('}')
	return sb.String(), nil
}

// fieldNames returns the JSON key names for each field. If the resolved type has
// CompositeAttrs (i.e. it's a named composite/table row type), those names are used.
// Otherwise, anonymous row field names are f1, f2, etc.
func fieldNames(rowType *pgtypes.DoltgresType, count int) []string {
	names := make([]string, count)
	if rowType != nil && len(rowType.CompositeAttrs) == count {
		for i, attr := range rowType.CompositeAttrs {
			names[i] = attr.Name
		}
	} else {
		for i := range names {
			names[i] = fmt.Sprintf("f%d", i+1)
		}
	}
	return names
}
