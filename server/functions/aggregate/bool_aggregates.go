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
func initBoolAnd() {
	framework.RegisterAggregateFunction(boolAnd)
}

// boolAnd represents the PostgreSQL boolAnd function.
var boolAnd = framework.Func1Aggregate{
	Function1: framework.Function1{
		Name:   "bool_and",
		Return: pgtypes.Bool,
		Parameters: [1]*pgtypes.DoltgresType{
			pgtypes.Bool,
		},
		Callable: func(ctx *sql.Context, paramsAndReturn [2]*pgtypes.DoltgresType, val1 any) (any, error) {
			return nil, nil
		},
	},
	NewAggBuffer: newBoolBuffer,
}

type boolAndBuffer struct {
	expr sql.Expression
	b bool
}

func newBoolBuffer() (sql.AggregationBuffer, error) {
	return &boolAndBuffer{
		b: true,
	}, nil
}

func (a *boolAndBuffer) Dispose() {}

func (a *boolAndBuffer) Eval(context *sql.Context) (interface{}, error) {
	return a.b, nil
}

func (a *boolAndBuffer) Update(ctx *sql.Context, row sql.Row) error {
	a.b = a.b && row[0].(bool)
	return nil
}
