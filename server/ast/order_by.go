// Copyright 2023 Dolthub, Inc.
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

package ast

import (
	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql/expression"

	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
	pgexprs "github.com/dolthub/doltgresql/server/expression"
)

// nodeOrderBy handles *tree.OrderBy nodes.
func nodeOrderBy(ctx *Context, node tree.OrderBy) (vitess.OrderBy, error) {
	if len(node) == 0 {
		return nil, nil
	}
	orderBys := make([]*vitess.Order, len(node))
	for i := range node {
		order, err := nodeOrder(ctx, node[i])
		if err != nil {
			return nil, err
		}
		// GMS order by is hardcoded to expect vitess.SQLVal for expressions such as `ORDER BY 1`.
		// In addition, there is the requirement that columns in the order by also need to be referenced somewhere in
		// the query, which is not a requirement for Postgres. Whenever we add that functionality, we also need to
		// remove the dependency on vitess.SQLVal. For now, we'll just convert our literals to a vitess.SQLVal.
		if literal, ok := orderLiteral(order); ok {
			order.Expr = pgexprs.ToVitessLiteral(literal)
		}
		orderBys[i] = order
	}
	return orderBys, nil
}

// nodeAggregateOrderBy handles the *tree.OrderBy of an aggregate function call, such as `array_agg(x ORDER BY y)`.
// Unlike a query's ORDER BY, a literal here is a constant sort key rather than a column ordinal, so it never affects
// the order and is dropped. GMS would otherwise interpret an integer literal as an ordinal.
func nodeAggregateOrderBy(ctx *Context, node tree.OrderBy) (vitess.OrderBy, error) {
	var orderBys vitess.OrderBy
	for i := range node {
		order, err := nodeOrder(ctx, node[i])
		if err != nil {
			return nil, err
		}
		if _, ok := orderLiteral(order); ok {
			continue
		}
		orderBys = append(orderBys, order)
	}
	return orderBys, nil
}

// nodeOrder handles *tree.Order nodes.
func nodeOrder(ctx *Context, node *tree.Order) (*vitess.Order, error) {
	if node.OrderType != tree.OrderByColumn {
		return nil, errors.Errorf("ORDER BY type is not yet supported")
	}
	direction := vitess.AscScr
	nullsOrder := vitess.NullsLastStr
	switch node.Direction {
	case tree.DefaultDirection, tree.Ascending:
	case tree.Descending:
		direction = vitess.DescScr
		nullsOrder = vitess.NullsFirstStr
	default:
		return nil, errors.Errorf("unknown ORDER BY sorting direction")
	}
	switch node.NullsOrder {
	case tree.DefaultNullsOrder:
	case tree.NullsFirst:
		nullsOrder = vitess.NullsFirstStr
	case tree.NullsLast:
		nullsOrder = vitess.NullsLastStr
	default:
		return nil, errors.Errorf("unknown NULL ordering in ORDER BY")
	}
	expr, err := nodeExpr(ctx, node.Expr)
	if err != nil {
		return nil, err
	}
	return &vitess.Order{
		Expr:       expr,
		Direction:  direction,
		NullsOrder: nullsOrder,
	}, nil
}

// orderLiteral returns the literal that the given order sorts by, if it sorts by a literal.
func orderLiteral(order *vitess.Order) (*expression.Literal, bool) {
	if injectedExpr, ok := order.Expr.(vitess.InjectedExpr); ok {
		if literal, ok := injectedExpr.Expression.(*expression.Literal); ok {
			return literal, true
		}
	}
	return nil, false
}
