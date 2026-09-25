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

	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/expression"
	"github.com/dolthub/go-mysql-server/sql/plan"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// RowComparison represents a comparison between a row constructor and either another row constructor or the rows of a
// subquery, which compares the rows field by field.
type RowComparison struct {
	operator   framework.Operator
	quantifier string // Empty, ANY, or ALL
	left       sql.Expression
	right      sql.Expression
	// leftFields and rightFields hold the pair of field values that is currently being compared.
	leftFields  []*expression.Literal
	rightFields []*expression.Literal
	// fieldOperators applies the operator to each pair of fields, except that `<=` and `>=` apply `<` and `>`.
	fieldOperators []sql.Expression
	// fieldEqualities applies `=` to each pair of fields, which orderings use to find the first pair that differs.
	fieldEqualities []sql.Expression
}

var _ vitess.Injectable = (*RowComparison)(nil)
var _ sql.Expression = (*RowComparison)(nil)

// NewRowComparison returns a new *RowComparison.
func NewRowComparison(operator framework.Operator, quantifier string) *RowComparison {
	return &RowComparison{operator: operator, quantifier: quantifier}
}

// Children implements the sql.Expression interface.
func (r *RowComparison) Children() []sql.Expression {
	return []sql.Expression{r.left, r.right}
}

// Eval implements the sql.Expression interface.
func (r *RowComparison) Eval(ctx *sql.Context, row sql.Row) (any, error) {
	if len(r.fieldOperators) == 0 {
		return nil, errors.Errorf("%T: cannot Eval as it has not been fully resolved", r)
	}
	left, err := r.left.Eval(ctx, row)
	if err != nil {
		return nil, err
	}
	leftValues := recordFieldValues(left.([]pgtypes.RecordValue))
	sub, ok := r.right.(*plan.Subquery)
	if !ok {
		right, err := r.right.Eval(ctx, row)
		if err != nil {
			return nil, err
		}
		return r.compareRow(ctx, leftValues, recordFieldValues(right.([]pgtypes.RecordValue)))
	}
	if r.quantifier == "" {
		right, err := sub.Eval(ctx, row)
		if err != nil || right == nil {
			return nil, err
		}
		return r.compareRow(ctx, leftValues, r.subqueryRow(right))
	}
	rights, err := sub.EvalMultiple(ctx, row)
	if err != nil {
		return nil, err
	}
	isAny := r.quantifier == "ANY"
	sawNull := false
	for _, right := range rights {
		result, err := r.compareRow(ctx, leftValues, r.subqueryRow(right))
		if err != nil {
			return nil, err
		}
		if result == nil {
			sawNull = true
		} else if result.(bool) == isAny {
			return isAny, nil
		}
	}
	if sawNull {
		return nil, nil
	}
	return !isAny, nil
}

// IsNullable implements the sql.Expression interface.
func (r *RowComparison) IsNullable(ctx *sql.Context) bool {
	return true
}

// Resolved implements the sql.Expression interface.
func (r *RowComparison) Resolved() bool {
	return r.left != nil && r.left.Resolved() && r.right != nil && r.right.Resolved() && len(r.fieldOperators) > 0
}

// String implements the sql.Expression interface.
func (r *RowComparison) String() string {
	if r.left == nil || r.right == nil {
		return fmt.Sprintf("? %s ?", r.operator.String())
	}
	right := r.right.String()
	if _, ok := r.right.(*plan.Subquery); ok {
		right = "(" + right + ")"
	}
	if r.quantifier != "" {
		return fmt.Sprintf("%s %s %s %s", r.left.String(), r.operator.String(), r.quantifier, right)
	}
	return fmt.Sprintf("%s %s %s", r.left.String(), r.operator.String(), right)
}

// Type implements the sql.Expression interface.
func (r *RowComparison) Type(ctx *sql.Context) sql.Type {
	return pgtypes.Bool
}

// WithChildren implements the sql.Expression interface.
func (r *RowComparison) WithChildren(ctx *sql.Context, children ...sql.Expression) (sql.Expression, error) {
	if len(children) != 2 {
		return nil, sql.ErrInvalidChildrenNumber.New(r, len(children), 2)
	}
	newComparison := *r
	newComparison.left = children[0]
	newComparison.right = children[1]
	leftTypes, rightTypes := rowFieldTypes(ctx, children[0]), rowFieldTypes(ctx, children[1])
	if leftTypes == nil || rightTypes == nil {
		return &newComparison, nil
	}
	if len(leftTypes) < len(rightTypes) {
		return nil, pgerror.New(pgcode.Syntax, "subquery has too many columns")
	} else if len(leftTypes) > len(rightTypes) {
		return nil, pgerror.New(pgcode.Syntax, "subquery has too few columns")
	}
	fieldOperator := r.operator
	switch r.operator {
	case framework.Operator_BinaryLessOrEqual:
		fieldOperator = framework.Operator_BinaryLessThan
	case framework.Operator_BinaryGreaterOrEqual:
		fieldOperator = framework.Operator_BinaryGreaterThan
	}
	newComparison.leftFields = make([]*expression.Literal, len(leftTypes))
	newComparison.rightFields = make([]*expression.Literal, len(leftTypes))
	newComparison.fieldOperators = make([]sql.Expression, len(leftTypes))
	if r.isOrdering() {
		newComparison.fieldEqualities = make([]sql.Expression, len(leftTypes))
	}
	for i := range leftTypes {
		leftField, rightField := expression.NewLiteral(nil, leftTypes[i]), expression.NewLiteral(nil, rightTypes[i])
		newComparison.leftFields[i], newComparison.rightFields[i] = leftField, rightField
		var err error
		if newComparison.fieldOperators[i], err = newFieldComparison(ctx, fieldOperator, leftField, rightField); err != nil {
			return nil, err
		}
		if r.isOrdering() {
			if newComparison.fieldEqualities[i], err = newFieldComparison(ctx, framework.Operator_BinaryEqual, leftField, rightField); err != nil {
				return nil, err
			}
		}
	}
	return &newComparison, nil
}

// WithResolvedChildren implements the vitess.Injectable interface.
func (r *RowComparison) WithResolvedChildren(ctx context.Context, children []any) (any, error) {
	if len(children) != 2 {
		return nil, errors.Errorf("invalid vitess child count, expected `2` but got `%d`", len(children))
	}
	left, ok := children[0].(sql.Expression)
	if !ok {
		return nil, errors.Errorf("expected vitess child to be an expression but has type `%T`", children[0])
	}
	right, ok := children[1].(sql.Expression)
	if !ok {
		return nil, errors.Errorf("expected vitess child to be an expression but has type `%T`", children[1])
	}
	return r.WithChildren(ctx.(*sql.Context), left, right)
}

// Operator returns the operator that is used.
func (r *RowComparison) Operator() framework.Operator {
	return r.operator
}

// compareRow compares the field values of two rows, returning NULL when the result depends on a NULL field.
func (r *RowComparison) compareRow(ctx *sql.Context, left []any, right []any) (any, error) {
	if r.isOrdering() {
		return r.compareOrdering(ctx, left, right)
	}
	return r.compareEquality(ctx, left, right)
}

// compareOrdering implements `<`, `<=`, `>`, and `>=` by walking the fields from left to right and skipping each pair
// that is equal. The first pair that is not equal decides the result with `<` or `>`, while a NULL in that pair makes
// the result NULL. When every pair is equal, only `<=` and `>=` are true.
func (r *RowComparison) compareOrdering(ctx *sql.Context, left []any, right []any) (any, error) {
	for i, fieldOperator := range r.fieldOperators {
		if left[i] == nil || right[i] == nil {
			return nil, nil
		}
		r.leftFields[i].Val = left[i]
		r.rightFields[i].Val = right[i]
		equal, err := r.fieldEqualities[i].Eval(ctx, nil)
		if err != nil || equal == nil {
			return nil, err
		}
		if !equal.(bool) {
			return fieldOperator.Eval(ctx, nil)
		}
	}
	return r.operator == framework.Operator_BinaryLessOrEqual || r.operator == framework.Operator_BinaryGreaterOrEqual, nil
}

// compareEquality implements `=` and `<>` by applying the operator to every pair of fields. Any pair that differs
// decides the result, even when another pair is NULL. Otherwise a NULL pair makes the result NULL, and rows whose pairs
// are all equal are equal.
func (r *RowComparison) compareEquality(ctx *sql.Context, left []any, right []any) (any, error) {
	isEqual := r.operator == framework.Operator_BinaryEqual
	sawNull := false
	for i, fieldOperator := range r.fieldOperators {
		if left[i] == nil || right[i] == nil {
			sawNull = true
			continue
		}
		r.leftFields[i].Val = left[i]
		r.rightFields[i].Val = right[i]
		result, err := fieldOperator.Eval(ctx, nil)
		if err != nil {
			return nil, err
		}
		if result == nil {
			sawNull = true
		} else if result.(bool) != isEqual {
			return result, nil
		}
	}
	if sawNull {
		return nil, nil
	}
	return isEqual, nil
}

// isOrdering returns whether the operator is `<`, `<=`, `>`, or `>=`, which order the rows rather than test them for
// equality.
func (r *RowComparison) isOrdering() bool {
	switch r.operator {
	case framework.Operator_BinaryLessThan, framework.Operator_BinaryLessOrEqual,
		framework.Operator_BinaryGreaterThan, framework.Operator_BinaryGreaterOrEqual:
		return true
	default:
		return false
	}
}

// subqueryRow returns the field values of a row returned by the subquery.
func (r *RowComparison) subqueryRow(row any) []any {
	if len(r.fieldOperators) == 1 {
		return []any{row}
	}
	return row.([]any)
}

// newFieldComparison returns the comparison between two fields, or an error when the operator does not exist for them.
func newFieldComparison(ctx *sql.Context, operator framework.Operator, left sql.Expression, right sql.Expression) (sql.Expression, error) {
	comparison, err := NewBinaryOperator(operator).WithResolvedChildren(ctx, []any{left, right})
	if err != nil {
		return nil, err
	}
	binaryOperator := comparison.(*BinaryOperator)
	if compiledFunc, ok := binaryOperator.compiledFunc.(*framework.CompiledFunction); ok && compiledFunc.StashedError() != nil {
		return nil, pgerror.WithCandidateCode(compiledFunc.StashedError(), pgcode.UndefinedFunction)
	}
	return binaryOperator, nil
}

// recordFieldValues returns the value of each field in the record.
func recordFieldValues(record []pgtypes.RecordValue) []any {
	values := make([]any, len(record))
	for i, field := range record {
		values[i] = field.Value
	}
	return values
}

// rowFieldTypes returns the type of each field of a row constructor or subquery, or nil for any other expression.
func rowFieldTypes(ctx *sql.Context, expr sql.Expression) []sql.Type {
	var types []sql.Type
	switch expr := expr.(type) {
	case *RecordExpr:
		for _, field := range expr.exprs {
			types = append(types, field.Type(ctx))
		}
	case *plan.Subquery:
		for _, column := range expr.Query.Schema(ctx) {
			types = append(types, column.Type)
		}
	}
	return types
}
