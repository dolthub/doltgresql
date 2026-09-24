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

	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

func initArrayRemove() { framework.RegisterFunction(array_remove) }

var array_remove = framework.Function2{
	Name:       "array_remove",
	Return:     pgtypes.AnyArray,
	Parameters: [2]*pgtypes.DoltgresType{pgtypes.AnyArray, pgtypes.AnyElement},
	Callable: func(ctx *sql.Context, t [3]*pgtypes.DoltgresType, val, search any) (any, error) {
		if val == nil {
			return nil, nil
		}
		vals := val.([]any)
		base := t[0].ArrayBaseType()
		if len(pgtypes.ArrayDims(vals, base)) > 1 {
			return nil, pgerror.New(pgcode.FeatureNotSupported, "removing elements from multidimensional arrays is not supported")
		}
		result := make([]any, 0, len(vals))
		for _, v := range vals {
			equal := v == nil && search == nil
			if v != nil && search != nil {
				cmp, err := base.Compare(ctx, v, search)
				if err != nil {
					return nil, err
				}
				equal = cmp == 0
			}
			if !equal {
				result = append(result, v)
			}
		}
		return result, nil
	},
}
