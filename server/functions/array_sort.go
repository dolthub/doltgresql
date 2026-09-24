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
	"sort"
)

func initArraySort() {
	framework.RegisterFunction(array_sort_one)
	framework.RegisterFunction(array_sort_two)
	framework.RegisterFunction(array_sort_three)
}

var array_sort_one = framework.Function1{
	Name: "array_sort", Return: pgtypes.AnyArray, Parameters: [1]*pgtypes.DoltgresType{pgtypes.AnyArray}, Strict: true,
	Callable: func(ctx *sql.Context, t [2]*pgtypes.DoltgresType, val any) (any, error) {
		return sortArray(ctx, t[0].ArrayBaseType(), val, false, false)
	},
}
var array_sort_two = framework.Function2{
	Name: "array_sort", Return: pgtypes.AnyArray, Parameters: [2]*pgtypes.DoltgresType{pgtypes.AnyArray, pgtypes.Bool}, Strict: true,
	Callable: func(ctx *sql.Context, t [3]*pgtypes.DoltgresType, val, desc any) (any, error) {
		return sortArray(ctx, t[0].ArrayBaseType(), val, desc.(bool), desc.(bool))
	},
}
var array_sort_three = framework.Function3{
	Name: "array_sort", Return: pgtypes.AnyArray, Parameters: [3]*pgtypes.DoltgresType{pgtypes.AnyArray, pgtypes.Bool, pgtypes.Bool}, Strict: true,
	Callable: func(ctx *sql.Context, t [4]*pgtypes.DoltgresType, val, desc, nullsFirst any) (any, error) {
		return sortArray(ctx, t[0].ArrayBaseType(), val, desc.(bool), nullsFirst.(bool))
	},
}

// sortArray sorts a copy of the first axis, leaving the input and inner axes unchanged.
func sortArray(ctx *sql.Context, base *pgtypes.DoltgresType, val any, descending, nullsFirst bool) (any, error) {
	vals := val.([]any)
	result := append([]any{}, vals...)
	dims := pgtypes.ArrayDims(vals, base)
	var compareErr error
	sort.SliceStable(result, func(i, j int) bool {
		if compareErr != nil {
			return false
		}
		a, b := result[i], result[j]
		if a == nil || b == nil {
			return a == nil && b != nil && nullsFirst || a != nil && b == nil && !nullsFirst
		}
		cmp, err := compareArraySortValues(ctx, base, a, b, len(dims) > 1)
		if err != nil {
			compareErr = err
			return false
		}
		if descending {
			return cmp > 0
		}
		return cmp < 0
	})
	if compareErr != nil {
		return nil, compareErr
	}
	return result, nil
}

// Within a row, null elements sort after non-null elements, as in PostgreSQL array comparisons.
func compareArraySortValues(ctx *sql.Context, base *pgtypes.DoltgresType, a, b any, nested bool) (int, error) {
	if a == nil {
		if b == nil {
			return 0, nil
		}
		return 1, nil
	}
	if b == nil {
		return -1, nil
	}
	if !nested {
		return base.Compare(ctx, a, b)
	}
	av := pgtypes.FlattenArray(a.([]any), base)
	bv := pgtypes.FlattenArray(b.([]any), base)
	for i := range av {
		cmp, err := compareArraySortValues(ctx, base, av[i], bv[i], false)
		if err != nil || cmp != 0 {
			return cmp, err
		}
	}
	return 0, nil
}
