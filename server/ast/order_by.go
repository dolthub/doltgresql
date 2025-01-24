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
		if node[i].OrderType != tree.OrderByColumn {
			return nil, errors.Errorf("ORDER BY type is not yet supported")
		}
		var direction string
		switch node[i].Direction {
		case tree.DefaultDirection:
			direction = vitess.AscScr
		case tree.Ascending:
			direction = vitess.AscScr
		case tree.Descending:
			direction = vitess.DescScr
		default:
			return nil, errors.Errorf("unknown ORDER BY sorting direction")
		}
		switch node[i].NullsOrder {
		case tree.DefaultNullsOrder:
			//TODO: the default NULL order is reversed compared to MySQL, so the default is technically always wrong.
			// To prevent choking on every ORDER BY, we allow this to proceed (even with incorrect results) for now.
			// If the NULL order is explicitly declared, then we want to error rather than return incorrect results.
		case tree.NullsFirst:
			if direction != vitess.AscScr {
				return nil, errors.Errorf("this NULL ordering is not yet supported for this ORDER BY direction")
			}
		case tree.NullsLast:
			if direction != vitess.DescScr {
				return nil, errors.Errorf("this NULL ordering is not yet supported for this ORDER BY direction")
			}
		default:
			return nil, errors.Errorf("unknown NULL ordering in ORDER BY")
		}
		expr, err := nodeExpr(ctx, node[i].Expr)
		if err != nil {
			return nil, err
		}
		// GMS order by is hardcoded to expect vitess.SQLVal for expressions such as `ORDER BY 1`.
		// In addition, there is the requirement that columns in the order by also need to be referenced somewhere in
		// the query, which is not a requirement for Postgres. Whenever we add that functionality, we also need to
		// remove the dependency on vitess.SQLVal. For now, we'll just convert our literals to a vitess.SQLVal.
		if injectedExpr, ok := expr.(vitess.InjectedExpr); ok {
			if literal, ok := injectedExpr.Expression.(*pgexprs.Literal); ok {
				expr = literal.ToVitessLiteral()
			}
		}
		orderBys[i] = &vitess.Order{
			Expr:      expr,
			Direction: direction,
		}
	}
	return orderBys, nil
}
