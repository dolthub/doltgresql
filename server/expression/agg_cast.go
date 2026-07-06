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

	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"

	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// AggCast wraps a sql.Aggregation to override its declared return type and
// post-convert the buffer's Eval result. It preserves the sql.Aggregation
// interface so GroupBy aggregation machinery works correctly.
//
// GMS SUM over integer columns always accumulates as float64 internally, but
// Postgres specifies SUM(int2/int4/int8) → bigint. AggCast intercepts the
// float64 buffer result and converts it to int64 so the correct wire type is used.
type AggCast struct {
	inner      sql.Aggregation
	targetType *pgtypes.DoltgresType
}

var _ sql.Expression = (*AggCast)(nil)
var _ sql.Aggregation = (*AggCast)(nil)

// NewAggCast wraps inner so that its declared type is targetType and its buffer
// Eval result is converted from float64 to int64.
func NewAggCast(inner sql.Aggregation, targetType *pgtypes.DoltgresType) *AggCast {
	return &AggCast{inner: inner, targetType: targetType}
}

// Type overrides the inner aggregation's declared type.
func (a *AggCast) Type(ctx *sql.Context) sql.Type { return a.targetType }

// NewBuffer delegates to inner but wraps the result buffer.
func (a *AggCast) NewBuffer(ctx *sql.Context) (sql.AggregationBuffer, error) {
	buf, err := a.inner.NewBuffer(ctx)
	if err != nil {
		return nil, err
	}
	return &aggCastBuffer{inner: buf}, nil
}

// NewWindowFunction delegates to inner.
func (a *AggCast) NewWindowFunction(ctx *sql.Context) (sql.WindowFunction, error) {
	return a.inner.NewWindowFunction(ctx)
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

// aggCastBuffer wraps an AggregationBuffer and post-converts float64 Eval results to int64.
type aggCastBuffer struct {
	inner sql.AggregationBuffer
}

func (b *aggCastBuffer) Update(ctx *sql.Context, row sql.Row) error {
	return b.inner.Update(ctx, row)
}

func (b *aggCastBuffer) Eval(ctx *sql.Context) (any, error) {
	v, err := b.inner.Eval(ctx)
	if err != nil || v == nil {
		return v, err
	}
	if f, ok := v.(float64); ok {
		return int64(math.RoundToEven(f)), nil
	}
	return v, nil
}

func (b *aggCastBuffer) Dispose(ctx *sql.Context) {
	b.inner.Dispose(ctx)
}
