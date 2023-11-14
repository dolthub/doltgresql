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
	"fmt"

	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
)

// nodeOrderBy handles *tree.OrderBy nodes.
func nodeOrderBy(node tree.OrderBy) (vitess.OrderBy, error) {
	if len(node) == 0 {
		return nil, nil
	}
	orderBys := make([]*vitess.Order, len(node))
	for i := range node {
		if node[i].OrderType != tree.OrderByColumn {
			return nil, fmt.Errorf("ORDER BY type is not yet supported")
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
			return nil, fmt.Errorf("unknown ORDER BY sorting direction")
		}
		switch node[i].NullsOrder {
		case tree.DefaultNullsOrder:
			//TODO: the default NULL order is reversed compared to MySQL, so the default is technically always wrong.
			// To prevent choking on every ORDER BY, we allow this to proceed (even with incorrect results) for now.
			// If the NULL order is explicitly declared, then we want to error rather than return incorrect results.
		case tree.NullsFirst:
			if direction != vitess.AscScr {
				return nil, fmt.Errorf("this NULL ordering is not yet supported for this ORDER BY direction")
			}
		case tree.NullsLast:
			if direction != vitess.DescScr {
				return nil, fmt.Errorf("this NULL ordering is not yet supported for this ORDER BY direction")
			}
		default:
			return nil, fmt.Errorf("unknown NULL ordering in ORDER BY")
		}
		expr, err := nodeExpr(node[i].Expr)
		if err != nil {
			return nil, err
		}
		orderBys[i] = &vitess.Order{
			Expr:      expr,
			Direction: direction,
		}
	}
	return orderBys, nil
}
