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

package aggregate

import (
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/server/functions"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// initJsonAggs registers the JSON aggregate functions to the catalog.
func initJsonAggs() {
	framework.RegisterAggregateFunction(jsonAgg)
}

// jsonAgg represents PostgreSQL's json_agg(anyelement) aggregate.
var jsonAgg = framework.Func1Aggregate{
	Function1: framework.Function1{
		Name:       "json_agg",
		Return:     pgtypes.Json,
		Parameters: [1]*pgtypes.DoltgresType{pgtypes.AnyElement},
		// json_agg is deliberately not strict: an input SQL NULL contributes a
		// JSON null element, while no input rows produce SQL NULL.
		Strict: false,
		Callable: func(ctx *sql.Context, paramsAndReturn [2]*pgtypes.DoltgresType, val any) (any, error) {
			return nil, nil
		},
	},
	NewAggBuffer:     newJsonAggBuffer,
	NewAggWindowFunc: newJsonAggWindowFunction,
}

// jsonAggBuffer accumulates the JSON representation of each input row.
type jsonAggBuffer struct {
	expr     sql.Expression
	elemType *pgtypes.DoltgresType
	elements []string
}

var _ sql.AggregationBuffer = (*jsonAggBuffer)(nil)

// newJsonAggBuffer creates an aggregation buffer for json_agg.
func newJsonAggBuffer(exprs []sql.Expression) (sql.AggregationBuffer, error) {
	return &jsonAggBuffer{expr: exprs[0]}, nil
}

// Dispose implements sql.AggregationBuffer.
func (b *jsonAggBuffer) Dispose(ctx *sql.Context) {}

// Eval implements sql.AggregationBuffer.
func (b *jsonAggBuffer) Eval(ctx *sql.Context) (interface{}, error) {
	if len(b.elements) == 0 {
		return nil, nil
	}
	return joinJsonAggregateElements(b.elemType, b.elements), nil
}

// Update implements sql.AggregationBuffer.
func (b *jsonAggBuffer) Update(ctx *sql.Context, row sql.Row) error {
	value, include, err := framework.EvalAggregateArgument(ctx, b.expr, row)
	if err != nil {
		return err
	}
	if !include {
		return nil
	}
	if b.elemType == nil {
		var ok bool
		b.elemType, ok = b.expr.Type(ctx).(*pgtypes.DoltgresType)
		if !ok {
			return errors.Errorf("json_agg: expected PostgreSQL argument type, got %T", b.expr.Type(ctx))
		}
	}
	raw, err := functions.ValueToJsonRaw(ctx, b.elemType, value)
	if err != nil {
		return err
	}
	b.elements = append(b.elements, string(raw))
	return nil
}

// jsonAggWindowFunction computes json_agg over a window frame.
type jsonAggWindowFunction struct {
	framework.WindowFramerState
	expr sql.Expression
}

var _ sql.WindowFunction = (*jsonAggWindowFunction)(nil)

// newJsonAggWindowFunction creates a window-function implementation of json_agg.
func newJsonAggWindowFunction(exprs []sql.Expression, window *sql.WindowDefinition) (sql.WindowFunction, error) {
	wf := &jsonAggWindowFunction{expr: exprs[0]}
	if err := wf.BindFramer(window); err != nil {
		return nil, err
	}
	return wf, nil
}

// Compute implements sql.WindowFunction.
func (w *jsonAggWindowFunction) Compute(ctx *sql.Context, interval sql.WindowInterval, buffer sql.WindowBuffer) (interface{}, error) {
	if interval.End <= interval.Start {
		return nil, nil
	}
	elements := make([]string, 0, interval.End-interval.Start)
	elemType, ok := w.expr.Type(ctx).(*pgtypes.DoltgresType)
	if !ok {
		return nil, errors.Errorf("json_agg: expected PostgreSQL argument type, got %T", w.expr.Type(ctx))
	}
	for i := interval.Start; i < interval.End; i++ {
		value, err := w.expr.Eval(ctx, buffer[i])
		if err != nil {
			return nil, err
		}
		raw, err := functions.ValueToJsonRaw(ctx, elemType, value)
		if err != nil {
			return nil, err
		}
		elements = append(elements, string(raw))
	}
	return joinJsonAggregateElements(elemType, elements), nil
}

// joinJsonAggregateElements formats collected JSON values using PostgreSQL's aggregate layout.
func joinJsonAggregateElements(elemType *pgtypes.DoltgresType, elements []string) string {
	separator := ", "
	if elemType != nil && (elemType.IsArrayType() || elemType.IsCompositeType() || elemType.ID.TypeName() == "record") {
		separator = ", \n "
	}
	return "[" + strings.Join(elements, separator) + "]"
}
