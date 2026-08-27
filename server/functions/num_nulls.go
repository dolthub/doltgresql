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

// initNumNulls registers the functions to the catalog.
func initNumNulls() {
	framework.RegisterFunction(num_nonnulls_any)
	framework.RegisterFunction(num_nulls_any)
}

// num_nonnulls_any represents the PostgreSQL function of the same name, taking the same parameters.
var num_nonnulls_any = framework.Function1N{
	Name:       "num_nonnulls",
	Return:     pgtypes.Int32,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Any},
	Strict:     false,
	Callable: func(ctx *sql.Context, t []*pgtypes.DoltgresType, val1 any, vals []any) (any, error) {
		count := int32(0)
		if val1 != nil {
			count++
		}
		for _, val := range vals {
			if val != nil {
				count++
			}
		}
		return count, nil
	},
}

// num_nulls_any represents the PostgreSQL function of the same name, taking the same parameters.
var num_nulls_any = framework.Function1N{
	Name:       "num_nulls",
	Return:     pgtypes.Int32,
	Parameters: [1]*pgtypes.DoltgresType{pgtypes.Any},
	Strict:     false,
	Callable: func(ctx *sql.Context, t []*pgtypes.DoltgresType, val1 any, vals []any) (any, error) {
		count := int32(0)
		if val1 == nil {
			count++
		}
		for _, val := range vals {
			if val == nil {
				count++
			}
		}
		return count, nil
	},
}
