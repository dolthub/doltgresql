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
	"context"
	"fmt"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
	"github.com/dolthub/doltgresql/server/types"
)

// SubscriptAssignment constructs the new array value for a subscript UPDATE target.
// The stored value is never mutated, including when it contains nested arrays.
type SubscriptAssignment struct {
	Subscript
	Value sql.Expression
}

// Type implements sql.Expression.
func (s SubscriptAssignment) Type(ctx *sql.Context) sql.Type {
	return s.Child.Type(ctx)
}

// Resolved implements sql.Expression.
func (s SubscriptAssignment) Resolved() bool {
	return s.Subscript.Resolved() && s.Value.Resolved()
}

// String implements sql.Expression.
func (s SubscriptAssignment) String() string {
	return fmt.Sprintf("%s = %s", s.Subscript.String(), s.Value)
}

// Children implements sql.Expression.
func (s SubscriptAssignment) Children() []sql.Expression {
	return append(s.Subscript.Children(), s.Value)
}

// WithChildren implements sql.Expression.
func (s SubscriptAssignment) WithChildren(ctx *sql.Context, children ...sql.Expression) (sql.Expression, error) {
	if len(children) < 3 {
		return nil, sql.ErrInvalidChildrenNumber.New(s, len(children), 3)
	}

	sub, err := s.Subscript.WithChildren(ctx, children[:len(children)-1]...)
	if err != nil {
		return nil, err
	}

	return &SubscriptAssignment{
		Subscript: *sub.(*Subscript),
		Value:     children[len(children)-1],
	}, nil
}

// WithResolvedChildren implements vitess.Injectable.
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

// Eval implements sql.Expression.
func (s SubscriptAssignment) Eval(ctx *sql.Context, row sql.Row) (any, error) {
	// Evaluate the original array first, then cast the replacement to its element or slice type.
	value, err := s.Child.Eval(ctx, row)
	if err != nil {
		return nil, err
	}

	dt, ok := s.childType(ctx)
	if !ok || !dt.IsArrayCategory() {
		return nil, pgerror.New(pgcode.DatatypeMismatch, "subscripted object is not an array")
	}

	replacement, err := s.evalReplacement(ctx, row, dt)
	if err != nil {
		return nil, err
	}

	var vals []any
	if value != nil {
		vals = value.([]any)
	}

	// Determine the array rank and evaluate inclusive bounds, filling omitted slice bounds
	// from the existing dimensions; empty arrays require explicit bounds.
	dims, err := s.assignmentDimensions(vals, dt.BaseType())
	if err != nil {
		return nil, err
	}

	lower, upper, err := s.evalAssignmentBounds(ctx, row, dims, len(vals) == 0)
	if err != nil {
		return nil, err
	}

	// A NULL slice replacement leaves the array unchanged, but its subscript expressions
	// must still be evaluated and checked for NULL before taking this shortcut.
	if s.Slice && replacement == nil {
		return value, nil
	}

	// Range checks apply only when the assignment writes values.
	if err = validateArrayAssignmentBounds(dims, lower, upper, len(vals) == 0); err != nil {
		return nil, err
	}

	// Flatten the replacement in storage order; scalar assignments consume one value,
	// including NULL, while slice assignments consume the requested rectangular region.
	var replacements []any
	if s.Slice {
		replacements = types.FlattenArray(replacement.([]any), dt.BaseType())
	} else {
		replacements = []any{replacement}
	}

	// Build a fresh result so neither successful nor failed assignments mutate the stored array.
	return replaceArrayRegion(vals, replacements, dt.BaseType(), dims, lower, upper)
}

// evalReplacement returns the replacement cast to the array or element type, or an evaluation or cast error.
func (s SubscriptAssignment) evalReplacement(ctx *sql.Context, row sql.Row, arrayType *types.DoltgresType) (any, error) {
	target := arrayType.BaseType()
	if s.Slice {
		target = arrayType
	}

	sourceType, ok := s.Value.Type(ctx).(*types.DoltgresType)
	if !ok {
		return nil, pgerror.New(pgcode.DatatypeMismatch, "invalid array assignment type")
	}

	return NewAssignmentCast(s.Value, sourceType, target).Eval(ctx, row)
}

// assignmentDimensions returns validated dimensions, using zero-sized axes when the assignment creates an array.
func (s SubscriptAssignment) assignmentDimensions(vals []any, baseType *types.DoltgresType) ([]int32, error) {
	dims := types.ArrayDims(vals, baseType)
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

	return dims, nil
}

// evalAssignmentBounds returns inclusive lower and upper bounds, or an error for invalid subscript expressions.
func (s SubscriptAssignment) evalAssignmentBounds(ctx *sql.Context, row sql.Row, dims []int32, empty bool) ([]int, []int, error) {
	// Unspecified axes and omitted slice bounds initially cover the whole existing dimension.
	lower := make([]int, len(dims))
	upper := make([]int, len(dims))
	for i, d := range dims {
		lower[i] = 1
		upper[i] = int(d)
	}

	for i, expr := range s.Indexes {
		if s.Slice && s.Omitted[i] {
			if empty {
				return nil, nil, pgerror.New(pgcode.ArraySubscript, "array slice subscript must provide both boundaries")
			}

			continue
		}

		v, err := expr.Eval(ctx, row)
		if err != nil {
			return nil, nil, err
		}

		if v == nil {
			return nil, nil, pgerror.New(pgcode.NullValueNotAllowed, "array subscript in assignment must not be null")
		}

		v, _, err = types.Int32.Convert(ctx, v)
		if err != nil {
			return nil, nil, err
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

	return lower, upper, nil
}

// validateArrayAssignmentBounds returns an error for reversed ranges, multidimensional growth, or unsupported lower bounds.
func validateArrayAssignmentBounds(dims []int32, lower, upper []int, empty bool) error {
	for i := range dims {
		if lower[i] > upper[i] {
			return pgerror.New(pgcode.ArraySubscript, "upper bound cannot be less than lower bound")
		}

		if !empty && len(dims) > 1 && (lower[i] < 1 || upper[i] > int(dims[i])) {
			return pgerror.New(pgcode.ArraySubscript, "array subscript out of range")
		}

		if lower[i] < 1 || empty && lower[i] != 1 {
			return pgerror.New(pgcode.FeatureNotSupported, "non-default array lower bounds are not yet supported")
		}
	}

	return nil
}

// arrayAssignmentSize returns the result dimensions and element count, or an error for insufficient replacement values or excessive size.
func arrayAssignmentSize(dims []int32, lower, upper []int, replacementCount int) ([]int32, int, error) {
	// The source must cover the entire assigned region; excess replacement elements are ignored.
	count := int64(1)
	for i := range dims {
		count *= int64(upper[i] - lower[i] + 1)
		if count > 134217727 {
			return nil, 0, pgerror.New(pgcode.ProgramLimitExceeded, "array size exceeds the maximum allowed (134217727)")
		}
	}

	if int64(replacementCount) < count {
		return nil, 0, pgerror.New(pgcode.ArraySubscript, "source array too small")
	}

	// Bounds were validated before sizing, so only one-dimensional or empty arrays can grow.
	resultDims := make([]int32, len(dims))
	for i := range dims {
		resultDims[i] = max(dims[i], int32(upper[i]))
	}

	total := int64(1)
	for _, d := range resultDims {
		total *= int64(d)
		if total > 134217727 {
			return nil, 0, pgerror.New(pgcode.ProgramLimitExceeded, "array size exceeds the maximum allowed (134217727)")
		}
	}

	return resultDims, int(total), nil
}

// replaceArrayRegion returns a new array with the selected region replaced, or an error if the result cannot be sized safely.
func replaceArrayRegion(vals, replacements []any, baseType *types.DoltgresType, dims []int32, lower, upper []int) ([]any, error) {
	dims, total, err := arrayAssignmentSize(dims, lower, upper, len(replacements))
	if err != nil {
		return nil, err
	}

	// Copy existing values into a new buffer, leaving any gap from one-dimensional growth as NULL.
	flat := make([]any, total)
	copy(flat, types.FlattenArray(vals, baseType))

	// Convert each flat offset to one-based coordinates, replacing only values inside every bound.
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
