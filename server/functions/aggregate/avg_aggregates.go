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
	"github.com/cockroachdb/apd/v3"
	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// avgGuardDigits is added on top of the dividend's own digit count when dividing for an AVG result, to get a
// meaningful number of fractional digits out of Quo (an integer dividend alone would otherwise round to an
// integer quotient). 16 matches Postgres's own displayed scale for avg() of integer types.
const avgGuardDigits = 16

// quoAvg divides dividend/divisor for use as an AVG result, then reduces the result to strip the trailing
// zeros Quo pads on to fill out the requested precision (e.g. 30/2 would otherwise come back as
// 15.00000...0). The precision is computed dynamically from the dividend's digit count, the same convention
// used elsewhere in this codebase for apd division/rounding (see div.go, round.go, sqrt.go, ln.go), rather
// than a single fixed precision for every call: sql.DecimalCtx (apd.BaseContext) has Precision: 0, which
// disables rounding entirely and Quo requires a nonzero value, but a fixed precision that's too low doesn't
// just lose fractional digits - if the dividend needs more significant digits than that to represent its
// *integer* part (e.g. avg() of a single huge bigint/numeric sum), apd rounds the integer part too, silently
// producing a wrong answer, not just an imprecise one.
func quoAvg(dividend, divisor *apd.Decimal) (*apd.Decimal, error) {
	p := dividend.NumDigits()
	if dividend.Exponent > 0 {
		p += int64(dividend.Exponent)
	}
	p += avgGuardDigits
	result := new(apd.Decimal)
	if _, err := sql.DecimalCtx.WithPrecision(uint32(p)).Quo(result, dividend, divisor); err != nil {
		return nil, err
	}
	result.Reduce(result)
	return result, nil
}

// initAvgAggs registers the functions to the catalog. See the comment on initNumericAggs for why avg needs a
// separate overload per input type rather than one generic/numeric-ish overload.
func initAvgAggs() {
	framework.RegisterAggregateFunction(avgOverload("avg", pgtypes.Int16, pgtypes.Numeric, newIntAvgBuffer[int16], newIntAvgWindowFunction[int16]))
	framework.RegisterAggregateFunction(avgOverload("avg", pgtypes.Int32, pgtypes.Numeric, newIntAvgBuffer[int32], newIntAvgWindowFunction[int32]))
	framework.RegisterAggregateFunction(avgOverload("avg", pgtypes.Int64, pgtypes.Numeric, newDecimalAvgBuffer(int64ToDecimal), newDecimalAvgWindowFunction(int64ToDecimal)))
	framework.RegisterAggregateFunction(avgOverload("avg", pgtypes.Numeric, pgtypes.Numeric, newDecimalAvgBuffer(decimalIdentity), newDecimalAvgWindowFunction(decimalIdentity)))
	framework.RegisterAggregateFunction(avgOverload("avg", pgtypes.Float32, pgtypes.Float64, newFloatAvgBuffer[float32], newFloatAvgWindowFunction[float32]))
	framework.RegisterAggregateFunction(avgOverload("avg", pgtypes.Float64, pgtypes.Float64, newFloatAvgBuffer[float64], newFloatAvgWindowFunction[float64]))
}

// avgOverload builds a single avg(...) overload; see sumOverload, which this mirrors.
func avgOverload(name string, paramType, returnType *pgtypes.DoltgresType, newBuffer framework.NewBufferFn, newWindowFunc framework.NewWindowFunctionFn) framework.Func1Aggregate {
	return framework.Func1Aggregate{
		Function1: framework.Function1{
			Name:   name,
			Return: returnType,
			Parameters: [1]*pgtypes.DoltgresType{
				paramType,
			},
			Callable: func(ctx *sql.Context, paramsAndReturn [2]*pgtypes.DoltgresType, val1 any) (any, error) {
				return nil, nil
			},
		},
		NewAggBuffer:     newBuffer,
		NewAggWindowFunc: newWindowFunc,
	}
}

// intAvgBuffer is the GROUP BY buffer for avg(int2)/avg(int4), both of which promote to numeric. The running
// sum fits safely in an int64 accumulator; only the final division (in Eval) needs decimal arithmetic.
type intAvgBuffer[T int16 | int32] struct {
	expr  sql.Expression
	sum   int64
	count int64
}

var _ sql.AggregationBuffer = (*intAvgBuffer[int32])(nil)

func newIntAvgBuffer[T int16 | int32](exprs []sql.Expression) (sql.AggregationBuffer, error) {
	return &intAvgBuffer[T]{expr: exprs[0]}, nil
}

func (b *intAvgBuffer[T]) Dispose(ctx *sql.Context) {}

func (b *intAvgBuffer[T]) Eval(ctx *sql.Context) (interface{}, error) {
	if b.count == 0 {
		return nil, nil
	}
	return quoAvg(apd.New(b.sum, 0), apd.New(b.count, 0))
}

func (b *intAvgBuffer[T]) Update(ctx *sql.Context, row sql.Row) error {
	v, err := b.expr.Eval(ctx, row)
	if err != nil {
		return err
	}
	if v == nil {
		return nil
	}
	i, ok := v.(T)
	if !ok {
		return errors.Errorf("avg: expected %T, got %T", i, v)
	}
	b.sum += int64(i)
	b.count++
	return nil
}

// intAvgWindowFunction is the sql.WindowFunction used for avg(int2)/avg(int4) within an OVER(...) clause.
type intAvgWindowFunction[T int16 | int32] struct {
	windowFramerState
	expr sql.Expression
}

var _ sql.WindowFunction = (*intAvgWindowFunction[int32])(nil)

func newIntAvgWindowFunction[T int16 | int32](exprs []sql.Expression, window *sql.WindowDefinition) (sql.WindowFunction, error) {
	wf := &intAvgWindowFunction[T]{expr: exprs[0]}
	if err := wf.bindFramer(window); err != nil {
		return nil, err
	}
	return wf, nil
}

func (w *intAvgWindowFunction[T]) Compute(ctx *sql.Context, interval sql.WindowInterval, buf sql.WindowBuffer) (interface{}, error) {
	if interval.End <= interval.Start {
		return nil, nil
	}
	var sum, count int64
	for i := interval.Start; i < interval.End; i++ {
		v, err := w.expr.Eval(ctx, buf[i])
		if err != nil {
			return nil, err
		}
		if v == nil {
			continue
		}
		iv, ok := v.(T)
		if !ok {
			return nil, errors.Errorf("avg: expected %T, got %T", iv, v)
		}
		sum += int64(iv)
		count++
	}
	if count == 0 {
		return nil, nil
	}
	return quoAvg(apd.New(sum, 0), apd.New(count, 0))
}

// decimalAvgBuffer is the GROUP BY buffer for avg(int8)/avg(numeric), both of which stay numeric. convert
// adapts the buffer to either input type, same as decimalSumBuffer.
type decimalAvgBuffer[T int64 | *apd.Decimal] struct {
	expr    sql.Expression
	sum     apd.Decimal
	count   int64
	convert func(T) *apd.Decimal
}

var _ sql.AggregationBuffer = (*decimalAvgBuffer[int64])(nil)

func newDecimalAvgBuffer[T int64 | *apd.Decimal](convert func(T) *apd.Decimal) framework.NewBufferFn {
	return func(exprs []sql.Expression) (sql.AggregationBuffer, error) {
		return &decimalAvgBuffer[T]{expr: exprs[0], convert: convert}, nil
	}
}

func (b *decimalAvgBuffer[T]) Dispose(ctx *sql.Context) {}

func (b *decimalAvgBuffer[T]) Eval(ctx *sql.Context) (interface{}, error) {
	if b.count == 0 {
		return nil, nil
	}
	return quoAvg(&b.sum, apd.New(b.count, 0))
}

func (b *decimalAvgBuffer[T]) Update(ctx *sql.Context, row sql.Row) error {
	v, err := b.expr.Eval(ctx, row)
	if err != nil {
		return err
	}
	if v == nil {
		return nil
	}
	typedV, ok := v.(T)
	if !ok {
		return errors.Errorf("avg: expected %T, got %T", typedV, v)
	}
	_, err = sql.DecimalCtx.Add(&b.sum, &b.sum, b.convert(typedV))
	b.count++
	return err
}

// decimalAvgWindowFunction is the sql.WindowFunction used for avg(int8)/avg(numeric) within an OVER(...)
// clause.
type decimalAvgWindowFunction[T int64 | *apd.Decimal] struct {
	windowFramerState
	expr    sql.Expression
	convert func(T) *apd.Decimal
}

var _ sql.WindowFunction = (*decimalAvgWindowFunction[int64])(nil)

func newDecimalAvgWindowFunction[T int64 | *apd.Decimal](convert func(T) *apd.Decimal) framework.NewWindowFunctionFn {
	return func(exprs []sql.Expression, window *sql.WindowDefinition) (sql.WindowFunction, error) {
		wf := &decimalAvgWindowFunction[T]{expr: exprs[0], convert: convert}
		if err := wf.bindFramer(window); err != nil {
			return nil, err
		}
		return wf, nil
	}
}

func (w *decimalAvgWindowFunction[T]) Compute(ctx *sql.Context, interval sql.WindowInterval, buf sql.WindowBuffer) (interface{}, error) {
	if interval.End <= interval.Start {
		return nil, nil
	}
	var sum apd.Decimal
	var count int64
	for i := interval.Start; i < interval.End; i++ {
		v, err := w.expr.Eval(ctx, buf[i])
		if err != nil {
			return nil, err
		}
		if v == nil {
			continue
		}
		typedV, ok := v.(T)
		if !ok {
			return nil, errors.Errorf("avg: expected %T, got %T", typedV, v)
		}
		if _, err = sql.DecimalCtx.Add(&sum, &sum, w.convert(typedV)); err != nil {
			return nil, err
		}
		count++
	}
	if count == 0 {
		return nil, nil
	}
	return quoAvg(&sum, apd.New(count, 0))
}

// floatAvgBuffer is the GROUP BY buffer for avg(float4)/avg(float8), both of which promote to double
// precision (unlike sum's float overloads, which preserve their input type).
type floatAvgBuffer[T float32 | float64] struct {
	expr  sql.Expression
	sum   float64
	count int64
}

var _ sql.AggregationBuffer = (*floatAvgBuffer[float64])(nil)

func newFloatAvgBuffer[T float32 | float64](exprs []sql.Expression) (sql.AggregationBuffer, error) {
	return &floatAvgBuffer[T]{expr: exprs[0]}, nil
}

func (b *floatAvgBuffer[T]) Dispose(ctx *sql.Context) {}

func (b *floatAvgBuffer[T]) Eval(ctx *sql.Context) (interface{}, error) {
	if b.count == 0 {
		return nil, nil
	}
	return b.sum / float64(b.count), nil
}

func (b *floatAvgBuffer[T]) Update(ctx *sql.Context, row sql.Row) error {
	v, err := b.expr.Eval(ctx, row)
	if err != nil {
		return err
	}
	if v == nil {
		return nil
	}
	f, ok := v.(T)
	if !ok {
		return errors.Errorf("avg: expected %T, got %T", f, v)
	}
	b.sum += float64(f)
	b.count++
	return nil
}

// floatAvgWindowFunction is the sql.WindowFunction used for avg(float4)/avg(float8) within an OVER(...)
// clause.
type floatAvgWindowFunction[T float32 | float64] struct {
	windowFramerState
	expr sql.Expression
}

var _ sql.WindowFunction = (*floatAvgWindowFunction[float64])(nil)

func newFloatAvgWindowFunction[T float32 | float64](exprs []sql.Expression, window *sql.WindowDefinition) (sql.WindowFunction, error) {
	wf := &floatAvgWindowFunction[T]{expr: exprs[0]}
	if err := wf.bindFramer(window); err != nil {
		return nil, err
	}
	return wf, nil
}

func (w *floatAvgWindowFunction[T]) Compute(ctx *sql.Context, interval sql.WindowInterval, buf sql.WindowBuffer) (interface{}, error) {
	if interval.End <= interval.Start {
		return nil, nil
	}
	var sum float64
	var count int64
	for i := interval.Start; i < interval.End; i++ {
		v, err := w.expr.Eval(ctx, buf[i])
		if err != nil {
			return nil, err
		}
		if v == nil {
			continue
		}
		fv, ok := v.(T)
		if !ok {
			return nil, errors.Errorf("avg: expected %T, got %T", fv, v)
		}
		sum += float64(fv)
		count++
	}
	if count == 0 {
		return nil, nil
	}
	return sum / float64(count), nil
}
