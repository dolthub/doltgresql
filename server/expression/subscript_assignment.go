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

package expression

import (
	"context"
	"fmt"
	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
	"github.com/dolthub/doltgresql/server/types"
	"github.com/dolthub/go-mysql-server/sql"
)

// SubscriptAssignment constructs the new array value for a subscript UPDATE target.
// The stored value is never mutated, including when it contains nested arrays.
type SubscriptAssignment struct {
	Subscript
	Value sql.Expression
}

func (s SubscriptAssignment) Type(ctx *sql.Context) sql.Type { return s.Child.Type(ctx) }
func (s SubscriptAssignment) Resolved() bool                 { return s.Subscript.Resolved() && s.Value.Resolved() }
func (s SubscriptAssignment) String() string {
	return fmt.Sprintf("%s = %s", s.Subscript.String(), s.Value)
}
func (s SubscriptAssignment) Children() []sql.Expression {
	return append(s.Subscript.Children(), s.Value)
}
func (s SubscriptAssignment) WithChildren(ctx *sql.Context, children ...sql.Expression) (sql.Expression, error) {
	if len(children) < 3 {
		return nil, sql.ErrInvalidChildrenNumber.New(s, len(children), 3)
	}
	sub, err := s.Subscript.WithChildren(ctx, children[:len(children)-1]...)
	if err != nil {
		return nil, err
	}
	return &SubscriptAssignment{Subscript: *sub.(*Subscript), Value: children[len(children)-1]}, nil
}
func (s SubscriptAssignment) WithResolvedChildren(ctx context.Context, children []any) (any, error) {
	exprs := make([]sql.Expression, len(children))
	for i, c := range children {
		var ok bool
		exprs[i], ok = c.(sql.Expression)
		if !ok {
			return nil, fmt.Errorf("expected expression, got %T", c)
		}
	}
	return s.WithChildren(ctx.(*sql.Context), exprs...)
}
func (s SubscriptAssignment) Eval(ctx *sql.Context, row sql.Row) (any, error) {
	value, err := s.Child.Eval(ctx, row)
	if err != nil {
		return nil, err
	}
	dt, ok := s.childType(ctx)
	if !ok {
		return nil, pgerror.New(pgcode.DatatypeMismatch, "subscripted object is not an array")
	}
	target := dt.BaseType()
	if s.Slice {
		target = dt
	}
	sourceType, ok := s.Value.Type(ctx).(*types.DoltgresType)
	if !ok {
		return nil, pgerror.New(pgcode.DatatypeMismatch, "invalid array assignment type")
	}
	replacement, err := NewAssignmentCast(s.Value, sourceType, target).Eval(ctx, row)
	if err != nil {
		return nil, err
	}
	if s.Slice && replacement == nil {
		return value, nil
	}
	var vals []any
	if value != nil {
		vals = value.([]any)
	}
	dims := types.ArrayDims(vals, dt.BaseType())
	rank := len(s.Indexes)
	if s.Slice {
		rank /= 2
	}
	if rank > 6 {
		return nil, pgerror.Newf(pgcode.ProgramLimitExceeded, "number of array dimensions (%d) exceeds the maximum allowed (6)", rank)
	}
	if len(dims) > 0 && (!s.Slice && rank != len(dims) || s.Slice && rank > len(dims)) {
		return nil, pgerror.New(pgcode.ArraySubscript, "array subscript out of range")
	}
	if len(dims) == 0 {
		dims = make([]int32, rank)
	}
	lower := make([]int, len(dims))
	upper := make([]int, len(dims))
	for i, d := range dims {
		lower[i] = 1
		upper[i] = int(d)
	}
	for i, expr := range s.Indexes {
		if s.Slice && s.Omitted[i] {
			if len(vals) == 0 {
				return nil, pgerror.New(pgcode.ArraySubscript, "array slice subscript must provide both boundaries")
			}
			continue
		}
		v, err := expr.Eval(ctx, row)
		if err != nil {
			return nil, err
		}
		if v == nil {
			return nil, pgerror.New(pgcode.NullValueNotAllowed, "array subscript in assignment must not be null")
		}
		v, _, err = types.Int32.Convert(ctx, v)
		if err != nil {
			return nil, err
		}
		n := int(v.(int32))
		if s.Slice {
			if i%2 == 0 {
				lower[i/2] = n
			} else {
				upper[i/2] = n
			}
		} else {
			lower[i] = n
			upper[i] = n
		}
	}
	for i := range dims {
		if lower[i] > upper[i] {
			return nil, pgerror.New(pgcode.ArraySubscript, "upper bound cannot be less than lower bound")
		}
		if len(vals) > 0 && len(dims) > 1 && (lower[i] < 1 || upper[i] > int(dims[i])) {
			return nil, pgerror.New(pgcode.ArraySubscript, "array subscript out of range")
		}
		if lower[i] < 1 || len(vals) == 0 && lower[i] != 1 {
			return nil, pgerror.New(pgcode.FeatureNotSupported, "non-default array lower bounds are not yet supported")
		}
	}
	var replacements []any
	if s.Slice {
		replacements = types.FlattenArray(replacement.([]any), dt.BaseType())
	} else {
		replacements = []any{replacement}
	}
	count := int64(1)
	for i := range dims {
		count *= int64(upper[i] - lower[i] + 1)
		if count > 134217727 {
			return nil, pgerror.New(pgcode.ProgramLimitExceeded, "array size exceeds the maximum allowed (134217727)")
		}
	}
	if int64(len(replacements)) < count {
		return nil, pgerror.New(pgcode.ArraySubscript, "source array too small")
	}
	for i := range dims {
		dims[i] = max(dims[i], int32(upper[i]))
	}
	total := int64(1)
	for _, d := range dims {
		total *= int64(d)
		if total > 134217727 {
			return nil, pgerror.New(pgcode.ProgramLimitExceeded, "array size exceeds the maximum allowed (134217727)")
		}
	}
	flat := make([]any, int(total))
	copy(flat, types.FlattenArray(vals, dt.BaseType()))
	offset := 0
	for i := range flat {
		index := i
		inside := true
		for axis := len(dims) - 1; axis >= 0; axis-- {
			coord := index%int(dims[axis]) + 1
			index /= int(dims[axis])
			inside = inside && coord >= lower[axis] && coord <= upper[axis]
		}
		if inside {
			flat[i] = replacements[offset]
			offset++
		}
	}
	return types.InflateArray(flat, dims), nil
}
