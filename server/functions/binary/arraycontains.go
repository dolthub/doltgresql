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

package binary

import (
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
	"github.com/dolthub/go-mysql-server/sql"
)

func initArrayContains() {
	framework.RegisterBinaryFunction(framework.Operator_BinaryJSONContainsRight, arraycontains)
}

var arraycontains = framework.Function2{
	Name: "arraycontains", Return: pgtypes.Bool, Parameters: [2]*pgtypes.DoltgresType{pgtypes.AnyArray, pgtypes.AnyArray}, Strict: true,
	Callable: func(ctx *sql.Context, t [3]*pgtypes.DoltgresType, left, right any) (any, error) {
		return arrayContains(ctx, t[0].ArrayBaseType(), left.([]any), right.([]any))
	},
}

// arrayContains treats arrays as collections of scalar elements. Shape and duplicate counts
// do not affect containment, and a NULL element does not equal another NULL element.
func arrayContains(ctx *sql.Context, base *pgtypes.DoltgresType, left, right []any) (bool, error) {
	left = pgtypes.FlattenArray(left, base)
	right = pgtypes.FlattenArray(right, base)
	for _, r := range right {
		found := false
		for _, l := range left {
			if l == nil || r == nil {
				continue
			}
			cmp, err := base.Compare(ctx, l, r)
			if err != nil {
				return false, err
			}
			if cmp == 0 {
				found = true
				break
			}
		}
		if !found {
			return false, nil
		}
	}
	return true, nil
}
