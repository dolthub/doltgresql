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
	"sort"
	"strings"

	"github.com/dolthub/doltgresql/server/types"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/expression"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"
)

type ArrayAgg struct {
	selectExprs []sql.Expression
	orderBy     sql.SortFields
	id          sql.ColumnId
}

var _ sql.Aggregation = (*ArrayAgg)(nil)
var _ vitess.Injectable = (*ArrayAgg)(nil)

// WithResolvedChildren returns a new ArrayAgg with the provided children as its select expressions.
// The last child is expected to be the order by expressions.
func (a *ArrayAgg) WithResolvedChildren(children []any) (any, error) {
	a.selectExprs = make([]sql.Expression, len(children)-1)
	for i := 0; i < len(children) - 1; i++ {
		a.selectExprs[i] = children[i].(sql.Expression)
	}
	
	a.orderBy = children[len(children)-1].(sql.SortFields)
	return a, nil
}

func (a *ArrayAgg) Resolved() bool {
	return expression.ExpressionsResolved(a.selectExprs...) && expression.ExpressionsResolved(a.orderBy.ToExpressions()...)
}

func (a *ArrayAgg) String() string {
	sb := strings.Builder{}
	sb.WriteString("group_concat(")

	if a.selectExprs != nil {
		var exprs = make([]string, len(a.selectExprs))
		for i, expr := range a.selectExprs {
			exprs[i] = expr.String()
		}

		sb.WriteString(strings.Join(exprs, ", "))
	}

	if len(a.orderBy) > 0 {
		sb.WriteString(" order by ")
		for i, ob := range a.orderBy {
			if i > 0 {
				sb.WriteString(", ")
			}
			sb.WriteString(ob.String())
		}
	}

	sb.WriteString(")")
	return sb.String()
}

func (a *ArrayAgg) Type() sql.Type {
	dt := a.selectExprs[0].Type().(*types.DoltgresType)
	return dt.ToArrayType()
}

func (a *ArrayAgg) IsNullable() bool {
	return true
}

func (a *ArrayAgg) Eval(ctx *sql.Context, row sql.Row) (interface{}, error) {
	panic("eval should never be called on an aggregation function")
}

func (a *ArrayAgg) Children() []sql.Expression {
	return append(a.selectExprs, a.orderBy.ToExpressions()...)
}

func (a ArrayAgg) WithChildren(children ...sql.Expression) (sql.Expression, error) {
	if len(children) != len(a.selectExprs)+len(a.orderBy) {
		return nil, sql.ErrInvalidChildrenNumber.New(a, len(children), len(a.selectExprs)+len(a.orderBy))
	}

	a.selectExprs = children[:len(a.selectExprs)]
	a.orderBy = a.orderBy.FromExpressions(children[len(a.selectExprs):]...)
	return &a, nil
}

func (a *ArrayAgg) Id() sql.ColumnId {
	return a.id
}

func (a ArrayAgg) WithId(id sql.ColumnId) sql.IdExpression {
	a.id = id
	return &a
}

func (a *ArrayAgg) NewWindowFunction() (sql.WindowFunction, error) {
	panic("window functions not yet supported for array_agg")
}

func (a *ArrayAgg) WithWindow(window *sql.WindowDefinition) sql.WindowAdaptableExpression {
	panic("window functions not yet supported for array_agg")
}

func (a *ArrayAgg) Window() *sql.WindowDefinition {
	return nil
}

func (a *ArrayAgg) NewBuffer() (sql.AggregationBuffer, error) {
	return &arrayAggBuffer{
		elements: make([]sql.Row, 0),
		a: a,
	}, nil
}

type arrayAggBuffer struct {
	elements []sql.Row
	a        *ArrayAgg
}

func newArrayAggBuffer(a *ArrayAgg) (sql.AggregationBuffer, error) {
	return &arrayAggBuffer{
		elements: make([]sql.Row, 0),
		a:        a,
	}, nil
}

func (a *arrayAggBuffer) Dispose() {}

func (a *arrayAggBuffer) Eval(ctx *sql.Context) (interface{}, error) {
	if len(a.elements) == 0 {
		return nil, nil
	}
	
	if a.a.orderBy != nil {
		sorter := &expression.Sorter{
			SortFields: a.a.orderBy,
			Rows:       a.elements,
			Ctx:        ctx,
		}

		sort.Stable(sorter)
		if sorter.LastError != nil {
			return nil, sorter.LastError
		}
	}
	
	// convert to []interface for return. The last element in each row is the one we want to return, the rest are sort fields.
	result := make([]interface{}, len(a.elements))
	for i, row := range a.elements {
		result[i] = row[(len(row) - 1)]
	}
	
	return result, nil
}

func (a *arrayAggBuffer) Update(ctx *sql.Context, row sql.Row) error {
	evalRow, err := evalExprs(ctx, a.a.selectExprs, row)
	if err != nil {
		return err
	}

	// TODO: unwrap values as necessary
	// Append the current value to the end of the row. We want to preserve the row's original structure
	// for sort ordering in the final step.
	a.elements = append(a.elements, append(row, nil, evalRow[0]))
	return nil
}

func evalExprs(ctx *sql.Context, exprs []sql.Expression, row sql.Row) (sql.Row, error) {
	result := make(sql.Row, len(exprs))
	for i, expr := range exprs {
		var err error
		result[i], err = expr.Eval(ctx, row)
		if err != nil {
			return nil, err
		}
	}

	return result, nil
}
