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

package aggregate

import (
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initArrayAgg registers the functions to the catalog.
func initArrayAgg() {
	framework.RegisterAggregateFunction(array_agg)
}

// array_agg represents the PostgreSQL array_agg function.
var array_agg = framework.Func1Aggregate{
	Function1: framework.Function1{
		Name:   "array_agg",
		Return: pgtypes.AnyArray,
		Parameters: [1]*pgtypes.DoltgresType{
			pgtypes.AnyElement,
		},
		Callable: func(ctx *sql.Context, paramsAndReturn [2]*pgtypes.DoltgresType, val1 any) (any, error) {
			return nil, nil
		},
	},
	NewAggBuffer: newArrayAggBuffer,
}

type arrayAggBuffer struct {
	elements []any
}

func newArrayAggBuffer() (sql.AggregationBuffer, error) {
	return &arrayAggBuffer{
		elements: make([]any, 0),
	}, nil
}

func (a *arrayAggBuffer) Dispose() {}

func (a *arrayAggBuffer) Eval(context *sql.Context) (interface{}, error) {
	if len(a.elements) == 0 {
		return nil, nil
	}
	return a.elements, nil
}

func (a *arrayAggBuffer) Update(ctx *sql.Context, row sql.Row) error {
	a.elements = append(a.elements, row[0])
	return nil
}
