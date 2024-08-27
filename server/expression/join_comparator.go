// Copyright 2024 Dolthub, Inc.
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
	"fmt"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/expression"

	"github.com/dolthub/doltgresql/server/functions/framework"
)

// JoinComparator is specifically for handling how GMS implements joins by implementing expression.Comparer over binary
// operators.
type JoinComparator struct {
	eq   *BinaryOperator
	less *BinaryOperator
}

var _ sql.Expression = (*JoinComparator)(nil)
var _ expression.BinaryExpression = (*JoinComparator)(nil)
var _ expression.Equality = (*JoinComparator)(nil)
var _ expression.Comparer = (*JoinComparator)(nil)

// NewJoinComparator returns a new *JoinComparator.
func NewJoinComparator(eq *BinaryOperator) (*JoinComparator, error) {
	if eq.operator != framework.Operator_BinaryEqual {
		return nil, fmt.Errorf("join comparator may only be created from equality operators")
	}
	less, err := (&BinaryOperator{operator: framework.Operator_BinaryLessThan}).WithResolvedChildren([]any{eq.Left(), eq.Right()})
	if err != nil {
		return nil, err
	}
	return &JoinComparator{
		eq:   eq,
		less: less.(*BinaryOperator),
	}, nil
}

// Children implements the sql.Expression interface.
func (j *JoinComparator) Children() []sql.Expression {
	return []sql.Expression{j.eq}
}

// Compare implements the expression.Comparer interface.
func (j *JoinComparator) Compare(ctx *sql.Context, row sql.Row) (int, error) {
	// Check equals
	res, err := j.eq.Eval(ctx, row)
	if err != nil {
		return 0, err
	}
	if res.(bool) {
		return 0, nil
	}
	// Check less than
	res, err = j.less.Eval(ctx, row)
	if err != nil {
		return 0, err
	}
	if res.(bool) {
		return -1, nil
	}
	// We'll assume it's greater
	return 1, nil
}

// Eval implements the sql.Expression interface.
func (j *JoinComparator) Eval(ctx *sql.Context, row sql.Row) (any, error) {
	// We only care about less in Compare.
	return j.eq.Eval(ctx, row)
}

// IsNullable implements the sql.Expression interface.
func (j *JoinComparator) IsNullable() bool {
	return j.eq.IsNullable()
}

// RepresentsEquality implements the expression.Equality interface.
func (j *JoinComparator) RepresentsEquality() bool {
	return j.eq.RepresentsEquality()
}

// Resolved implements the sql.Expression interface.
func (j *JoinComparator) Resolved() bool {
	return j.eq.Resolved() && j.less.Resolved()
}

// String implements the sql.Expression interface.
func (j *JoinComparator) String() string {
	return j.eq.String()
}

// SwapParameters implements the expression.Equality interface.
func (j *JoinComparator) SwapParameters(ctx *sql.Context) (expression.Equality, error) {
	newOper, err := j.eq.SwapParameters(ctx)
	if err != nil {
		return nil, err
	}
	return NewJoinComparator(newOper.(*BinaryOperator))
}

// ToComparer implements the expression.Equality interface.
func (j *JoinComparator) ToComparer() (expression.Comparer, error) {
	return j, nil
}

// Type implements the sql.Expression interface.
func (j *JoinComparator) Type() sql.Type {
	return j.eq.Type()
}

// WithChildren implements the sql.Expression interface.
func (j *JoinComparator) WithChildren(children ...sql.Expression) (sql.Expression, error) {
	if len(children) != 1 {
		return nil, sql.ErrInvalidChildrenNumber.New(j, len(children), 1)
	}
	binaryOper, ok := children[0].(*BinaryOperator)
	if !ok {
		return nil, fmt.Errorf("expected join comparator child to be a binary operator but has type `%T`", children[0])
	}
	return NewJoinComparator(binaryOper)
}

// Left implements the expression.BinaryExpression interface.
func (j *JoinComparator) Left() sql.Expression {
	return j.eq.Left()
}

// Right implements the expression.BinaryExpression interface.
func (j *JoinComparator) Right() sql.Expression {
	return j.eq.Right()
}
