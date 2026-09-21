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
	"strings"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/analyzer"
	"github.com/dolthub/go-mysql-server/sql/expression"
	"github.com/dolthub/go-mysql-server/sql/information_schema"
	"github.com/dolthub/go-mysql-server/sql/plan"
	"github.com/dolthub/go-mysql-server/sql/transform"

	pgexpression "github.com/dolthub/doltgresql/server/expression"
)

// ParenthesizeColumnDefaults wraps every operator nested within a column default or generated column expression in
// parentheses, so that the string form written to disk re-parses with the same precedence.
func ParenthesizeColumnDefaults(ctx *sql.Context, _ *analyzer.Analyzer, n sql.Node, _ *plan.Scope, _ analyzer.RuleSelector, _ *sql.QueryFlags) (sql.Node, transform.TreeIdentity, error) {
	span, ctx := ctx.Span("parenthesizeColumnDefaults")
	defer span.End()

	return transform.Node(ctx, n, func(ctx *sql.Context, n sql.Node) (sql.Node, transform.TreeIdentity, error) {
		switch node := n.(type) {
		case sql.SchemaTarget:
			expressioner, ok := node.(sql.Expressioner)
			if !ok {
				return n, transform.SameTree, nil
			}
			newExprs, same, err := parenthesizeColumnDefaults(ctx, expressioner.Expressions())
			if err != nil || same {
				return n, transform.SameTree, err
			}
			newNode, err := expressioner.WithExpressions(ctx, newExprs...)
			if err != nil {
				return nil, transform.SameTree, err
			}
			return newNode, transform.NewTree, nil
		case *plan.ResolvedTable:
			columnsTable, ok := node.Table.(*information_schema.ColumnsTable)
			if !ok {
				return node, transform.SameTree, nil
			}
			allColumns, err := columnsTable.AllColumns(ctx)
			if err != nil {
				return nil, transform.SameTree, err
			}
			allDefaults, same, err := parenthesizeColumnDefaults(ctx, transform.WrappedColumnDefaults(allColumns))
			if err != nil || same {
				return node, transform.SameTree, err
			}
			node.Table, err = columnsTable.WithColumnDefaults(allDefaults)
			if err != nil {
				return nil, transform.SameTree, err
			}
			return node, transform.NewTree, nil
		default:
			return node, transform.SameTree, nil
		}
	})
}

// parenthesizeColumnDefaults applies parenthesizeColumnDefault to each of the given expressions.
func parenthesizeColumnDefaults(ctx *sql.Context, exprs []sql.Expression) ([]sql.Expression, transform.TreeIdentity, error) {
	var newExprs []sql.Expression
	for i, e := range exprs {
		newExpr, same, err := parenthesizeColumnDefault(ctx, e)
		if err != nil {
			return nil, transform.SameTree, err
		}
		if same {
			continue
		}
		if newExprs == nil {
			newExprs = make([]sql.Expression, len(exprs))
			copy(newExprs, exprs)
		}
		newExprs[i] = newExpr
	}
	if newExprs == nil {
		return exprs, transform.SameTree, nil
	}
	return newExprs, transform.NewTree, nil
}

// parenthesizeColumnDefault parenthesizes the operators nested within a column default, which may be wrapped. Any other
// expression is returned unchanged.
func parenthesizeColumnDefault(ctx *sql.Context, e sql.Expression) (sql.Expression, transform.TreeIdentity, error) {
	switch e := e.(type) {
	case *expression.Wrapper:
		newInner, same, err := parenthesizeColumnDefault(ctx, e.Unwrap())
		if err != nil || same {
			return e, transform.SameTree, err
		}
		return expression.WrapExpression(newInner), transform.NewTree, nil
	case *sql.ColumnDefaultValue:
		if e == nil || e.Expr == nil {
			return e, transform.SameTree, nil
		}
		newExpr, same, err := transform.Expr(ctx, e.Expr, parenthesizeOperator)
		if err != nil || same {
			return e, transform.SameTree, err
		}
		if wrapper, ok := newExpr.(*expression.Wrapper); ok {
			newExpr = wrapper.Unwrap()
		}
		newDefault, err := e.WithChildren(ctx, newExpr)
		if err != nil {
			return nil, transform.SameTree, err
		}
		return newDefault, transform.NewTree, nil
	default:
		return e, transform.SameTree, nil
	}
}

// parenthesizeOperator wraps an operator expression so that its string form is enclosed in parentheses. A negative
// literal operand of a unary operator is wrapped as well, since `--1` would otherwise re-parse as a comment.
func parenthesizeOperator(ctx *sql.Context, e sql.Expression) (sql.Expression, transform.TreeIdentity, error) {
	switch e := e.(type) {
	case *pgexpression.UnaryOperator:
		operand := e.Children()[0]
		if len(operand.Children()) == 0 && strings.HasPrefix(operand.String(), "-") {
			newOperator, err := e.WithChildren(ctx, expression.WrapExpression(operand))
			if err != nil {
				return nil, transform.SameTree, err
			}
			return expression.WrapExpression(newOperator), transform.NewTree, nil
		}
		return expression.WrapExpression(e), transform.NewTree, nil
	case *pgexpression.BinaryOperator, *pgexpression.Not, *pgexpression.IsNull, *pgexpression.IsNotNull,
		*pgexpression.IsDistinctFrom, *pgexpression.IsNotDistinctFrom, *pgexpression.InTuple, *pgexpression.InSubquery,
		*pgexpression.AnyExpr, *expression.Like:
		return expression.WrapExpression(e), transform.NewTree, nil
	default:
		return e, transform.SameTree, nil
	}
}
