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

package analyzer

import (
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/analyzer"
	gms_expression "github.com/dolthub/go-mysql-server/sql/expression"
	"github.com/dolthub/go-mysql-server/sql/plan"
	"github.com/dolthub/go-mysql-server/sql/transform"

	"github.com/dolthub/doltgresql/server/expression"
	"github.com/dolthub/doltgresql/server/functions/framework"
)

// SplitRowComparisons splits each comparison between two row constructors in a filter into comparisons between their
// fields, so that an index on the columns can serve it.
func SplitRowComparisons(ctx *sql.Context, a *analyzer.Analyzer, node sql.Node, scope *plan.Scope, selector analyzer.RuleSelector, qFlags *sql.QueryFlags) (sql.Node, transform.TreeIdentity, error) {
	return transform.Node(ctx, node, func(ctx *sql.Context, node sql.Node) (sql.Node, transform.TreeIdentity, error) {
		filter, ok := node.(*plan.Filter)
		if !ok {
			return node, transform.SameTree, nil
		}
		newExpr, same, err := transform.Expr(ctx, filter.Expression, splitRowComparison)
		if err != nil || same {
			return node, transform.SameTree, err
		}
		newNode, err := filter.WithExpressions(ctx, newExpr)
		return newNode, transform.NewTree, err
	})
}

// splitRowComparison returns `e` as comparisons between fields when it compares two row constructors whose fields are
// all deterministic.
func splitRowComparison(ctx *sql.Context, e sql.Expression) (sql.Expression, transform.TreeIdentity, error) {
	comparison, ok := e.(*expression.RowComparison)
	if !ok {
		return e, transform.SameTree, nil
	}
	left, ok := comparison.Children()[0].(*expression.RecordExpr)
	if !ok {
		return e, transform.SameTree, nil
	}
	right, ok := comparison.Children()[1].(*expression.RecordExpr)
	if !ok {
		return e, transform.SameTree, nil
	}
	if transform.InspectExpr(ctx, comparison, func(ctx *sql.Context, e sql.Expression) bool {
		switch e := e.(type) {
		case *plan.Subquery:
			return true
		case sql.NonDeterministicExpression:
			return e.IsNonDeterministic()
		}
		return false
	}) {
		return e, transform.SameTree, nil
	}
	fieldComparison := func(operator framework.Operator, leftField sql.Expression, rightField sql.Expression) (sql.Expression, error) {
		expr, err := expression.NewBinaryOperator(operator).WithResolvedChildren(ctx, []any{leftField, rightField})
		if err != nil {
			return nil, err
		}
		return expr.(sql.Expression), nil
	}
	leftFields, rightFields := left.Children(), right.Children()
	last := len(leftFields) - 1
	expr, err := fieldComparison(comparison.Operator(), leftFields[last], rightFields[last])
	if err != nil {
		return nil, transform.SameTree, err
	}
	fieldOperator := comparison.Operator()
	switch fieldOperator {
	case framework.Operator_BinaryLessOrEqual:
		fieldOperator = framework.Operator_BinaryLessThan
	case framework.Operator_BinaryGreaterOrEqual:
		fieldOperator = framework.Operator_BinaryGreaterThan
	}
	for i := last - 1; i >= 0; i-- {
		strict, err := fieldComparison(fieldOperator, leftFields[i], rightFields[i])
		if err != nil {
			return nil, transform.SameTree, err
		}
		equal, err := fieldComparison(framework.Operator_BinaryEqual, leftFields[i], rightFields[i])
		if err != nil {
			return nil, transform.SameTree, err
		}
		expr = gms_expression.NewOr(strict, gms_expression.NewAnd(equal, expr))
	}
	return expr, transform.NewTree, nil
}
