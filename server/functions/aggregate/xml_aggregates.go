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
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
	"github.com/dolthub/doltgresql/server/xml"
)

// initXmlAggs registers the functions to the catalog.
func initXmlAggs() {
	framework.RegisterAggregateFunction(xmlAgg)
}

// xmlAgg represents the PostgreSQL xmlagg function.
var xmlAgg = framework.Func1Aggregate{
	Function1: framework.Function1{
		Name:       "xmlagg",
		Return:     pgtypes.Xml,
		Parameters: [1]*pgtypes.DoltgresType{pgtypes.Xml},
		Callable: func(ctx *sql.Context, paramsAndReturn [2]*pgtypes.DoltgresType, val1 any) (any, error) {
			return nil, nil
		},
	},
	NewAggBuffer:     newXmlAggBuffer,
	NewAggWindowFunc: newXmlAggWindowFunction,
}

// xmlAggBuffer concatenates the non-null xml values of the input rows.
type xmlAggBuffer struct {
	expr  sql.Expression
	value *string
}

var _ sql.AggregationBuffer = (*xmlAggBuffer)(nil)

// newXmlAggBuffer creates an aggregation buffer for xmlagg.
func newXmlAggBuffer(exprs []sql.Expression) (sql.AggregationBuffer, error) {
	return &xmlAggBuffer{expr: exprs[0]}, nil
}

// Dispose implements sql.AggregationBuffer.
func (b *xmlAggBuffer) Dispose(ctx *sql.Context) {}

// Eval implements sql.AggregationBuffer.
func (b *xmlAggBuffer) Eval(ctx *sql.Context) (interface{}, error) {
	if b.value == nil {
		return nil, nil
	}
	return *b.value, nil
}

// Update implements sql.AggregationBuffer.
func (b *xmlAggBuffer) Update(ctx *sql.Context, row sql.Row) error {
	value, include, err := framework.EvalAggregateArgument(ctx, b.expr, row)
	if err != nil || !include {
		return err
	}
	b.value, err = xmlAggConcat(ctx, b.value, value)
	return err
}

// xmlAggWindowFunction computes xmlagg over a window frame.
type xmlAggWindowFunction struct {
	framework.WindowFramerState
	expr sql.Expression
}

var _ sql.WindowFunction = (*xmlAggWindowFunction)(nil)

// newXmlAggWindowFunction creates a window-function implementation of xmlagg.
func newXmlAggWindowFunction(exprs []sql.Expression, window *sql.WindowDefinition) (sql.WindowFunction, error) {
	wf := &xmlAggWindowFunction{expr: exprs[0]}
	if err := wf.BindFramer(window); err != nil {
		return nil, err
	}
	return wf, nil
}

// Compute implements sql.WindowFunction.
func (w *xmlAggWindowFunction) Compute(ctx *sql.Context, interval sql.WindowInterval, buffer sql.WindowBuffer) (interface{}, error) {
	var result *string
	for i := interval.Start; i < interval.End; i++ {
		value, err := w.expr.Eval(ctx, buffer[i])
		if err != nil {
			return nil, err
		}
		if result, err = xmlAggConcat(ctx, result, value); err != nil {
			return nil, err
		}
	}
	if result == nil {
		return nil, nil
	}
	return *result, nil
}

// xmlAggConcat appends the xml `value` to the running `result`, which is nil until a non-null value is seen.
func xmlAggConcat(ctx *sql.Context, result *string, value any) (*string, error) {
	if value == nil {
		return result, nil
	}
	str, err := framework.UnwrapString(ctx, value)
	if err != nil {
		return nil, err
	}
	if result != nil {
		str = xml.Concat(*result, str)
	}
	return &str, nil
}
