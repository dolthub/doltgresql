// Copyright 2025 Dolthub, Inc.
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
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
	"github.com/dolthub/go-mysql-server/sql"
)

// initGenerateSeries registers the functions to the catalog.
func initGenerateSubscripts() {
	framework.RegisterFunction(generate_subscripts)
}

// generate_series_int32_int32 represents the PostgreSQL function of the same name, taking the same parameters.
var generate_subscripts = framework.Function2{
	Name:       "generate_subscripts",
	Return:     pgtypes.Row,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.AnyArray, pgtypes.Int32},
	Strict:     true,
	SRF:        true,
	Callable: func(ctx *sql.Context, t [3]*pgtypes.DoltgresType, val1, val2 any) (any, error) {
		arr := val1.([]any)
		dimension := val2.(int32)
		
		if dimension != 1 {
			return nil, sql.ErrUnsupportedFeature.New("generate_subscripts only supports 1-dimensional arrays")
		}

		rows := make([]any, len(arr))
		for i := range arr {
			rows[i] = int32(i + 1)
		}
		
		return nil, nil // pgtypes.NewRowValues(rows, pgtypes.Int32, int64(len(rows))), nil
	},
}
