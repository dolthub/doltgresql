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

package expression

import (
	"math"

	"github.com/cockroachdb/apd/v3"
	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"

	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// AggCast wraps a sql.Aggregation to override its declared return type and post-convert
// its result, whether reached through the GroupBy buffer path (NewBuffer/Eval) or the
// window function path (NewWindowFunction/Compute). It preserves both the sql.Aggregation
// and sql.WindowAdaptableExpression interfaces so GroupBy and window execution both work.
//
// GMS SUM and AVG over integer columns always accumulate as float64 internally, but
// Postgres specifies SUM(int2/int4/int8) → bigint and AVG(int2/int4/int8) → numeric.
// AggCast intercepts the float64 result and converts it to the target type so the
// correct wire type is used.
type AggCast struct {
	inner      sql.Aggregation
	targetType *pgtypes.DoltgresType
	convKind   aggConvKind
}

var _ sql.Expression = (*AggCast)(nil)
var _ sql.Aggregation = (*AggCast)(nil)
var _ sql.WindowFunction = (*aggCastWindowFunction)(nil)

// aggConvKind identifies which conversion convertAggResult should apply. It's derived from
// targetType once, at AggCast construction time (query-analysis time), rather than by
// calling DoltgresType.Equals on every row-group's Eval/Compute call (execution time,
// happening once per group per query execution): DoltgresType.Equals is not a cheap identity
// check — it compares two types by fully serializing both to byte buffers and comparing the
// bytes. Precomputing the target once and comparing a small int on the hot path avoids
// paying that serialization cost repeatedly.
type aggConvKind byte

const (
	aggConvInt64 aggConvKind = iota
	aggConvNumeric
	aggConvFloat32
	aggConvFloat64
)

func aggConvKindFor(targetType *pgtypes.DoltgresType) aggConvKind {
	switch {
	case targetType.Equals(pgtypes.Numeric):
		return aggConvNumeric
	case targetType.Equals(pgtypes.Float32):
		return aggConvFloat32
	case targetType.Equals(pgtypes.Float64):
		return aggConvFloat64
	default:
		return aggConvInt64
	}
}

// NewAggCast wraps inner so that its declared type is targetType and its buffer
// Eval result is converted from float64 to match targetType.
func NewAggCast(inner sql.Aggregation, targetType *pgtypes.DoltgresType) *AggCast {
	return &AggCast{inner: inner, targetType: targetType, convKind: aggConvKindFor(targetType)}
}

// Type overrides the inner aggregation's declared type.
func (a *AggCast) Type(ctx *sql.Context) sql.Type { return a.targetType }

// NewBuffer delegates to inner but wraps the result buffer.
func (a *AggCast) NewBuffer(ctx *sql.Context) (sql.AggregationBuffer, error) {
	buf, err := a.inner.NewBuffer(ctx)
	if err != nil {
		return nil, err
	}
	return &aggCastBuffer{inner: buf, convKind: a.convKind}, nil
}

// NewWindowFunction delegates to inner but wraps the result so Compute's float64 output is
// converted to match targetType, the same way NewBuffer wraps the AggregationBuffer path.
func (a *AggCast) NewWindowFunction(ctx *sql.Context) (sql.WindowFunction, error) {
	fn, err := a.inner.NewWindowFunction(ctx)
	if err != nil {
		return nil, err
	}
	return &aggCastWindowFunction{inner: fn, convKind: a.convKind}, nil
}

// WithWindow delegates to inner.
func (a *AggCast) WithWindow(ctx *sql.Context, w *sql.WindowDefinition) sql.WindowAdaptableExpression {
	return a.inner.WithWindow(ctx, w)
}

// Window delegates to inner.
func (a *AggCast) Window() *sql.WindowDefinition { return a.inner.Window() }

// Id delegates to inner.
func (a *AggCast) Id() sql.ColumnId {
	if ide, ok := a.inner.(sql.IdExpression); ok {
		return ide.Id()
	}
	return 0
}

// WithId delegates to inner and rewraps.
func (a *AggCast) WithId(id sql.ColumnId) sql.IdExpression {
	if ide, ok := a.inner.(sql.IdExpression); ok {
		if agg, ok := ide.WithId(id).(sql.Aggregation); ok {
			return NewAggCast(agg, a.targetType)
		}
	}
	return a
}

func (a *AggCast) Children() []sql.Expression { return a.inner.Children() }
func (a *AggCast) Eval(ctx *sql.Context, row sql.Row) (any, error) {
	return a.inner.Eval(ctx, row)
}
func (a *AggCast) IsNullable(ctx *sql.Context) bool { return a.inner.IsNullable(ctx) }
func (a *AggCast) Resolved() bool                   { return a.inner.Resolved() }
func (a *AggCast) String() string                   { return a.inner.String() }

func (a *AggCast) WithChildren(ctx *sql.Context, children ...sql.Expression) (sql.Expression, error) {
	inner, err := a.inner.WithChildren(ctx, children...)
	if err != nil {
		return nil, err
	}
	agg, ok := inner.(sql.Aggregation)
	if !ok {
		return nil, errors.New("AggCast.WithChildren: rebuilt expression does not implement sql.Aggregation")
	}
	return NewAggCast(agg, a.targetType), nil
}

// aggCastBuffer wraps an AggregationBuffer and post-converts its float64 Eval result to
// match convKind's target type (int64 for bigint, numeric, or float32 for real; float64 is
// a no-op).
type aggCastBuffer struct {
	inner    sql.AggregationBuffer
	convKind aggConvKind
}

func (b *aggCastBuffer) Update(ctx *sql.Context, row sql.Row) error {
	return b.inner.Update(ctx, row)
}

func (b *aggCastBuffer) Eval(ctx *sql.Context) (any, error) {
	v, err := b.inner.Eval(ctx)
	if err != nil || v == nil {
		return v, err
	}
	return convertAggResult(v, b.convKind)
}

func (b *aggCastBuffer) Dispose(ctx *sql.Context) {
	b.inner.Dispose(ctx)
}

// aggCastWindowFunction wraps a sql.WindowFunction and post-converts its float64 Compute
// result the same way aggCastBuffer does for the sql.AggregationBuffer (GroupBy) path.
type aggCastWindowFunction struct {
	inner    sql.WindowFunction
	convKind aggConvKind
}

func (w *aggCastWindowFunction) StartPartition(ctx *sql.Context, interval sql.WindowInterval, buffer sql.WindowBuffer) error {
	return w.inner.StartPartition(ctx, interval, buffer)
}

func (w *aggCastWindowFunction) DefaultFramer() sql.WindowFramer {
	return w.inner.DefaultFramer()
}

func (w *aggCastWindowFunction) Compute(ctx *sql.Context, interval sql.WindowInterval, buffer sql.WindowBuffer) (any, error) {
	v, err := w.inner.Compute(ctx, interval, buffer)
	if err != nil || v == nil {
		return v, err
	}
	return convertAggResult(v, w.convKind)
}

func (w *aggCastWindowFunction) Dispose(ctx *sql.Context) {
	w.inner.Dispose(ctx)
}

// convertAggResult converts a raw aggregation result (always float64 from GMS's SUM/AVG
// implementations, regardless of the input column's width) to match convKind: numeric,
// float32, float64 (a no-op), or int64 for everything else (bigint).
func convertAggResult(v any, convKind aggConvKind) (any, error) {
	f, ok := v.(float64)
	if !ok {
		return v, nil
	}
	switch convKind {
	case aggConvNumeric:
		d, err := new(apd.Decimal).SetFloat64(f)
		if err != nil {
			return nil, err
		}
		return d, nil
	case aggConvFloat32:
		return float32(f), nil
	case aggConvFloat64:
		return f, nil
	default:
		return int64(math.RoundToEven(f)), nil
	}
}
