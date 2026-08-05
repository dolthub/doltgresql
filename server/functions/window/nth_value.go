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

package window

import (
	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// nthValue represents the PostgreSQL nth_value(value, n) window function. Its return type is
// AnyElement, which CompiledFunction.Type resolves back to value's actual argument type.
var nthValue = framework.Func2Window{
	Function2: framework.Function2{
		Name:   "nth_value",
		Return: pgtypes.AnyElement,
		Parameters: [2]*pgtypes.DoltgresType{
			pgtypes.AnyElement,
			pgtypes.Int32,
		},
		Callable: func(ctx *sql.Context, paramsAndReturn [3]*pgtypes.DoltgresType, val1, val2 any) (any, error) {
			return nil, nil
		},
	},
	NewWinFunc: newNthValueWindowFunction,
}

// nthValueWindowFunction is the sql.WindowFunction used for nth_value() within an OVER(...) clause.
// Unlike the ranking functions, nth_value respects the window's explicit frame clause (default RANGE
// UNBOUNDED PRECEDING to CURRENT ROW), so it embeds windowFramerState the same way the native sum/avg
// window functions do.
type nthValueWindowFunction struct {
	windowFramerState
	valueExpr sql.Expression
	nExpr     sql.Expression
}

var _ sql.WindowFunction = (*nthValueWindowFunction)(nil)

// newNthValueWindowFunction creates the sql.WindowFunction for nth_value().
func newNthValueWindowFunction(exprs []sql.Expression, window *sql.WindowDefinition) (sql.WindowFunction, error) {
	wf := &nthValueWindowFunction{valueExpr: exprs[0], nExpr: exprs[1]}
	if err := wf.bindFramer(window); err != nil {
		return nil, err
	}
	return wf, nil
}

// Compute implements the sql.WindowFunction interface.
func (w *nthValueWindowFunction) Compute(ctx *sql.Context, interval sql.WindowInterval, buf sql.WindowBuffer) (interface{}, error) {
	if interval.End <= interval.Start {
		return nil, nil
	}
	nVal, err := w.nExpr.Eval(ctx, buf[interval.Start])
	if err != nil {
		return nil, err
	}
	if nVal == nil {
		return nil, nil
	}
	n, ok := nVal.(int32)
	if !ok {
		return nil, errors.Errorf("nth_value: expected int32 offset, got %T", nVal)
	}
	if n <= 0 {
		return nil, sql.ErrInvalidArgument.New("NTH_VALUE")
	}

	idx := interval.Start + int(n) - 1
	if idx >= interval.End {
		return nil, nil
	}
	return w.valueExpr.Eval(ctx, buf[idx])
}
