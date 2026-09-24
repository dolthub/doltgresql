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
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
	"github.com/dolthub/go-mysql-server/sql"
)

func initArrayReplace() { framework.RegisterFunction(array_replace) }

var array_replace = framework.Function3{
	Name: "array_replace", Return: pgtypes.AnyArray, Parameters: [3]*pgtypes.DoltgresType{pgtypes.AnyArray, pgtypes.AnyElement, pgtypes.AnyElement},
	Callable: func(ctx *sql.Context, t [4]*pgtypes.DoltgresType, val, search, replacement any) (any, error) {
		if val == nil {
			return nil, nil
		}
		vals := val.([]any)
		base := t[0].ArrayBaseType()
		dims := pgtypes.ArrayDims(vals, base)
		flat := pgtypes.FlattenArray(vals, base)
		result := make([]any, len(flat))
		for i, v := range flat {
			equal := v == nil && search == nil
			if v != nil && search != nil {
				cmp, err := base.Compare(ctx, v, search)
				if err != nil {
					return nil, err
				}
				equal = cmp == 0
			}
			result[i] = v
			if equal {
				result[i] = replacement
			}
		}
		return pgtypes.InflateArray(result, dims), nil
	},
}
