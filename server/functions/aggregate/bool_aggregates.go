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

// initBoolAggs registers the functions to the catalog.
func initBoolAggs() {
	framework.RegisterAggregateFunction(boolAnd)
	framework.RegisterAggregateFunction(boolOr)
}

// boolAnd represents the PostgreSQL bool_and function.
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
	NewAggBuffer: newBoolAndBuffer,
}

// boolAnd represents the PostgreSQL bool_or function.
var boolOr = framework.Func1Aggregate{
	Function1: framework.Function1{
		Name:   "bool_or",
		Return: pgtypes.Bool,
		Parameters: [1]*pgtypes.DoltgresType{
			pgtypes.Bool,
		},
		Callable: func(ctx *sql.Context, paramsAndReturn [2]*pgtypes.DoltgresType, val1 any) (any, error) {
			return nil, nil
		},
	},
	NewAggBuffer: newBoolOrBuffer,
}

type boolAggBuffer struct {
	expr   sql.Expression
	b      bool
	sawOne bool
	isAnd  bool
}

func newBoolAndBuffer(exprs []sql.Expression) (sql.AggregationBuffer, error) {
	return &boolAggBuffer{
		expr:  exprs[0],
		b:     true,
		isAnd: true,
	}, nil
}

func newBoolOrBuffer(exprs []sql.Expression) (sql.AggregationBuffer, error) {
	return &boolAggBuffer{
		expr: exprs[0],
	}, nil
}

func (a *boolAggBuffer) Dispose() {}

func (a *boolAggBuffer) Eval(context *sql.Context) (interface{}, error) {
	if !a.sawOne {
		return nil, nil
	}
	return a.b, nil
}

func (a *boolAggBuffer) Update(ctx *sql.Context, row sql.Row) error {
	eval, err := a.expr.Eval(ctx, row)
	if err != nil {
		return err
	}

	if eval == nil {
		return nil
	}

	a.sawOne = true
	if a.isAnd {
		a.b = a.b && eval.(bool)
	} else {
		a.b = a.b || eval.(bool)
	}
	return nil
}
