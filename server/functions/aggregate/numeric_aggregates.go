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
	"github.com/dolthub/go-mysql-server/sql/expression/function/aggregation"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// windowFramerState holds the sql.WindowFramer setup shared by every native window-function implementation
// in this package, regardless of accumulator type: bind a framer from the window's explicit frame clause if
// one was given, otherwise fall back to Postgres's default (unbounded preceding to current row). None of
// this touches the per-row hot path (StartPartition/DefaultFramer/Dispose all run once per partition, not
// once per row), so unlike the accumulator itself, it's free to share via embedding across every T.
type windowFramerState struct {
	framer sql.WindowFramer
}

// bindFramer builds and stores this window's framer, if it declared an explicit frame clause; with no
// explicit frame, DefaultFramer's fallback applies instead.
func (s *windowFramerState) bindFramer(window *sql.WindowDefinition) error {
	if window == nil || window.Frame == nil {
		return nil
	}
	framer, err := window.Frame.NewFramer(window)
	if err != nil {
		return err
	}
	s.framer = framer
	return nil
}

// StartPartition implements the sql.WindowFunction interface.
func (s *windowFramerState) StartPartition(ctx *sql.Context, interval sql.WindowInterval, buf sql.WindowBuffer) error {
	return nil
}

// DefaultFramer implements the sql.WindowFunction interface; with no explicit frame, this supplies
// Postgres's default (unbounded preceding to current row).
func (s *windowFramerState) DefaultFramer() sql.WindowFramer {
	if s.framer != nil {
		return s.framer
	}
	return aggregation.NewUnboundedPrecedingToCurrentRowFramer()
}

// Dispose implements the sql.WindowFunction interface.
func (s *windowFramerState) Dispose(ctx *sql.Context) {}

// int64ToDecimal converts an int64 sum to *apd.Decimal, for use as the decimalConvert of a decimalSumBuffer/
// decimalAvgBuffer instantiated over int64 (i.e. sum(int8)/avg(int8), whose accumulator needs to be decimal
// since a running sum of bigints can itself overflow a bigint).
func int64ToDecimal(v int64) *apd.Decimal { return apd.New(v, 0) }

// decimalIdentity is the decimalConvert for a decimalSumBuffer/decimalAvgBuffer instantiated over
// *apd.Decimal (i.e. sum(numeric)/avg(numeric)), whose input values are already decimal.
func decimalIdentity(v *apd.Decimal) *apd.Decimal { return v }

// initNumericAggs registers the functions to the catalog.
//
// Note that overload resolution for aggregates does not insert implicit widening casts the way it does for
// scalar functions (CompiledAggregateFunction hands NewBuffer/NewWindowFunc the raw, uncast Arguments) - so
// every distinct argument type sum/avg can be called on needs its own overload registered below; there's no
// way to cover e.g. both int2 and int4 with a single numeric-ish overload. This mirrors how other multi-type
// functions in this package (e.g. abs.go) already register one overload per Postgres type.
func initNumericAggs() {
	framework.RegisterAggregateFunction(sumOverload("sum", pgtypes.Int16, pgtypes.Int64, newIntSumBuffer[int16], newIntSumWindowFunction[int16]))
	framework.RegisterAggregateFunction(sumOverload("sum", pgtypes.Int32, pgtypes.Int64, newIntSumBuffer[int32], newIntSumWindowFunction[int32]))
	framework.RegisterAggregateFunction(sumOverload("sum", pgtypes.Int64, pgtypes.Numeric, newDecimalSumBuffer(int64ToDecimal), newDecimalSumWindowFunction(int64ToDecimal)))
	framework.RegisterAggregateFunction(sumOverload("sum", pgtypes.Numeric, pgtypes.Numeric, newDecimalSumBuffer(decimalIdentity), newDecimalSumWindowFunction(decimalIdentity)))
	framework.RegisterAggregateFunction(sumOverload("sum", pgtypes.Float32, pgtypes.Float32, newFloatSumBuffer[float32], newFloatSumWindowFunction[float32]))
	framework.RegisterAggregateFunction(sumOverload("sum", pgtypes.Float64, pgtypes.Float64, newFloatSumBuffer[float64], newFloatSumWindowFunction[float64]))
}

// sumOverload builds a single sum(...) overload: paramType is the Postgres type of the aggregated column,
// returnType is sum's result type for that input (e.g. sum(int4) promotes to bigint), and newBuffer/
// newWindowFunc construct the GROUP BY and OVER(...) implementations respectively.
func sumOverload(name string, paramType, returnType *pgtypes.DoltgresType, newBuffer framework.NewBufferFn, newWindowFunc framework.NewWindowFunctionFn) framework.Func1Aggregate {
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

// intSumBuffer is the GROUP BY buffer for sum(int2)/sum(int4), which both promote to bigint. Their sum fits
// safely in an int64 accumulator (the widest int2/int4 value is nowhere near int64's range), so this is a
// simple integer sum, unlike sum(int8)/sum(numeric) (see decimalSumBuffer) which need decimal accumulation.
type intSumBuffer[T int16 | int32] struct {
	expr   sql.Expression
	sum    int64
	sawOne bool
}

var _ sql.AggregationBuffer = (*intSumBuffer[int32])(nil)

func newIntSumBuffer[T int16 | int32](exprs []sql.Expression) (sql.AggregationBuffer, error) {
	return &intSumBuffer[T]{expr: exprs[0]}, nil
}

func (b *intSumBuffer[T]) Dispose(ctx *sql.Context) {}

func (b *intSumBuffer[T]) Eval(ctx *sql.Context) (interface{}, error) {
	if !b.sawOne {
		return nil, nil
	}
	return b.sum, nil
}

func (b *intSumBuffer[T]) Update(ctx *sql.Context, row sql.Row) error {
	v, err := b.expr.Eval(ctx, row)
	if err != nil {
		return err
	}
	if v == nil {
		return nil
	}
	i, ok := v.(T)
	if !ok {
		return errors.Errorf("sum: expected %T, got %T", i, v)
	}
	b.sum += int64(i)
	b.sawOne = true
	return nil
}

// intSumWindowFunction is the sql.WindowFunction used for sum(int2)/sum(int4) within an OVER(...) clause.
type intSumWindowFunction[T int16 | int32] struct {
	windowFramerState
	expr sql.Expression
}

var _ sql.WindowFunction = (*intSumWindowFunction[int32])(nil)

func newIntSumWindowFunction[T int16 | int32](exprs []sql.Expression, window *sql.WindowDefinition) (sql.WindowFunction, error) {
	wf := &intSumWindowFunction[T]{expr: exprs[0]}
	if err := wf.bindFramer(window); err != nil {
		return nil, err
	}
	return wf, nil
}

func (w *intSumWindowFunction[T]) Compute(ctx *sql.Context, interval sql.WindowInterval, buf sql.WindowBuffer) (interface{}, error) {
	if interval.End <= interval.Start {
		return nil, nil
	}
	var sum int64
	var sawOne bool
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
			return nil, errors.Errorf("sum: expected %T, got %T", iv, v)
		}
		sum += int64(iv)
		sawOne = true
	}
	if !sawOne {
		return nil, nil
	}
	return sum, nil
}

// decimalSumBuffer is the GROUP BY buffer for sum(int8)/sum(numeric). sum(int8) promotes to numeric because
// a running sum of bigints can itself overflow a bigint; sum(numeric) stays numeric. convert adapts the
// buffer to either input type: int64 values are boxed via int64ToDecimal, while numeric values (already
// *apd.Decimal) pass through decimalIdentity unchanged.
type decimalSumBuffer[T int64 | *apd.Decimal] struct {
	expr    sql.Expression
	sum     apd.Decimal
	sawOne  bool
	convert func(T) *apd.Decimal
}

var _ sql.AggregationBuffer = (*decimalSumBuffer[int64])(nil)

func newDecimalSumBuffer[T int64 | *apd.Decimal](convert func(T) *apd.Decimal) framework.NewBufferFn {
	return func(exprs []sql.Expression) (sql.AggregationBuffer, error) {
		return &decimalSumBuffer[T]{expr: exprs[0], convert: convert}, nil
	}
}

func (b *decimalSumBuffer[T]) Dispose(ctx *sql.Context) {}

func (b *decimalSumBuffer[T]) Eval(ctx *sql.Context) (interface{}, error) {
	if !b.sawOne {
		return nil, nil
	}
	result := b.sum
	return &result, nil
}

func (b *decimalSumBuffer[T]) Update(ctx *sql.Context, row sql.Row) error {
	v, err := b.expr.Eval(ctx, row)
	if err != nil {
		return err
	}
	if v == nil {
		return nil
	}
	typedV, ok := v.(T)
	if !ok {
		return errors.Errorf("sum: expected %T, got %T", typedV, v)
	}
	d := b.convert(typedV)
	if !b.sawOne {
		b.sum.Set(d)
		b.sawOne = true
		return nil
	}
	_, err = sql.DecimalCtx.Add(&b.sum, &b.sum, d)
	return err
}

// decimalSumWindowFunction is the sql.WindowFunction used for sum(int8)/sum(numeric) within an OVER(...)
// clause.
type decimalSumWindowFunction[T int64 | *apd.Decimal] struct {
	windowFramerState
	expr    sql.Expression
	convert func(T) *apd.Decimal
}

var _ sql.WindowFunction = (*decimalSumWindowFunction[int64])(nil)

func newDecimalSumWindowFunction[T int64 | *apd.Decimal](convert func(T) *apd.Decimal) framework.NewWindowFunctionFn {
	return func(exprs []sql.Expression, window *sql.WindowDefinition) (sql.WindowFunction, error) {
		wf := &decimalSumWindowFunction[T]{expr: exprs[0], convert: convert}
		if err := wf.bindFramer(window); err != nil {
			return nil, err
		}
		return wf, nil
	}
}

func (w *decimalSumWindowFunction[T]) Compute(ctx *sql.Context, interval sql.WindowInterval, buf sql.WindowBuffer) (interface{}, error) {
	if interval.End <= interval.Start {
		return nil, nil
	}
	var sum apd.Decimal
	var sawOne bool
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
			return nil, errors.Errorf("sum: expected %T, got %T", typedV, v)
		}
		d := w.convert(typedV)
		if !sawOne {
			sum.Set(d)
			sawOne = true
			continue
		}
		if _, err = sql.DecimalCtx.Add(&sum, &sum, d); err != nil {
			return nil, err
		}
	}
	if !sawOne {
		return nil, nil
	}
	return &sum, nil
}

// floatSumBuffer is the GROUP BY buffer for sum(float4)/sum(float8), which (unlike the integer overloads)
// preserve their input type rather than promoting.
type floatSumBuffer[T float32 | float64] struct {
	expr   sql.Expression
	sum    T
	sawOne bool
}

var _ sql.AggregationBuffer = (*floatSumBuffer[float64])(nil)

func newFloatSumBuffer[T float32 | float64](exprs []sql.Expression) (sql.AggregationBuffer, error) {
	return &floatSumBuffer[T]{expr: exprs[0]}, nil
}

func (b *floatSumBuffer[T]) Dispose(ctx *sql.Context) {}

func (b *floatSumBuffer[T]) Eval(ctx *sql.Context) (interface{}, error) {
	if !b.sawOne {
		return nil, nil
	}
	return b.sum, nil
}

func (b *floatSumBuffer[T]) Update(ctx *sql.Context, row sql.Row) error {
	v, err := b.expr.Eval(ctx, row)
	if err != nil {
		return err
	}
	if v == nil {
		return nil
	}
	f, ok := v.(T)
	if !ok {
		return errors.Errorf("sum: expected %T, got %T", f, v)
	}
	b.sum += f
	b.sawOne = true
	return nil
}

// floatSumWindowFunction is the sql.WindowFunction used for sum(float4)/sum(float8) within an OVER(...)
// clause.
type floatSumWindowFunction[T float32 | float64] struct {
	windowFramerState
	expr sql.Expression
}

var _ sql.WindowFunction = (*floatSumWindowFunction[float64])(nil)

func newFloatSumWindowFunction[T float32 | float64](exprs []sql.Expression, window *sql.WindowDefinition) (sql.WindowFunction, error) {
	wf := &floatSumWindowFunction[T]{expr: exprs[0]}
	if err := wf.bindFramer(window); err != nil {
		return nil, err
	}
	return wf, nil
}

func (w *floatSumWindowFunction[T]) Compute(ctx *sql.Context, interval sql.WindowInterval, buf sql.WindowBuffer) (interface{}, error) {
	if interval.End <= interval.Start {
		return nil, nil
	}
	var sum T
	var sawOne bool
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
			return nil, errors.Errorf("sum: expected %T, got %T", fv, v)
		}
		sum += fv
		sawOne = true
	}
	if !sawOne {
		return nil, nil
	}
	return sum, nil
}
