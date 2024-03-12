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
	pgexprs "github.com/dolthub/doltgresql/server/expression"
)

// nodeSelectClause handles tree.SelectClause nodes.
func nodeSelectClause(node *tree.SelectClause) (*vitess.Select, error) {
	if node == nil {
		return nil, nil
	}
	selectExprs, err := nodeSelectExprs(node.Exprs)
	if err != nil {
		return nil, err
	}
	from, err := nodeFrom(node.From)
	if err != nil {
		return nil, err
	}
	// We use TableFuncExprs to represent queries on functions that behave as though they were tables. This is something
	// that we have to situationally support, as inner nodes do not have the proper context to output a TableFuncExpr,
	// since TableFuncExprs pertain only to SELECT statements.
	for i, fromExpr := range from {
		// Nodes are very liberal in wrapping themselves within other nodes, which gives them a technically correct
		// tree, however GMS makes assumptions about the makeup of the trees that it receives. We'll eventually
		// generalize this on the GMS side, but for now we need to transform our tree in case we need to use a TableFuncExpr.
		if aliasedTableExpr, ok := fromExpr.(*vitess.AliasedTableExpr); ok {
			subquery, ok := aliasedTableExpr.Expr.(*vitess.Subquery)
			// If all of these are true, then the AliasedTableExpr is probably a wrapper around a subquery, but we have
			// to confirm that the subquery contains a *Select with a single child in its From expressions.
			if !aliasedTableExpr.Lateral &&
				aliasedTableExpr.Hints == nil &&
				len(aliasedTableExpr.Partitions) == 0 &&
				ok && len(subquery.Columns) == 0 {
				// If this is true, then we can confirm that it's just a wrapper (and not an explicit AliasedTableExpr).
				// This may seem like a lot of fragile checks, but AliasedTableExpr explicitly sets its state to this in
				// this circumstance. We do not want to create a TableFuncExpr except under very specific circumstances.
				if subquerySelect, ok := subquery.Select.(*vitess.Select); ok && len(subquerySelect.From) == 1 {
					if valuesStatement, ok := subquerySelect.From[0].(*vitess.ValuesStatement); ok {
						if len(valuesStatement.Columns) == 0 && len(valuesStatement.Rows) == 1 && len(valuesStatement.Rows[0]) == 1 {
							if funcExpr, ok := valuesStatement.Rows[0][0].(*vitess.FuncExpr); ok {
								// It appears that GMS hardcodes the expectation of vitess literals here, so we have to
								// convert from Doltgres literals to GMS literals. Eventually we need to remove this
								// hardcoded behavior.
								for _, fExpr := range funcExpr.Exprs {
									if aliasedExpr, ok := fExpr.(*vitess.AliasedExpr); ok {
										if injectedExpr, ok := aliasedExpr.Expr.(vitess.InjectedExpr); ok {
											if literal, ok := injectedExpr.Expression.(*pgexprs.Literal); ok {
												aliasedExpr.Expr = literal.ToVitessLiteral()
											}
										}
									}
								}
								from[i] = &vitess.TableFuncExpr{
									Name:  funcExpr.Name.String(),
									Exprs: funcExpr.Exprs,
									Alias: aliasedTableExpr.As,
								}
							}
						}
					}
				}
			}
		}
	}
	if len(node.DistinctOn) > 0 {
		return nil, fmt.Errorf("DISTINCT ON is not yet supported")
	}
	where, err := nodeWhere(node.Where)
	if err != nil {
		return nil, err
	}
	having, err := nodeWhere(node.Having)
	if err != nil {
		return nil, err
	}
	groupBy, err := nodeGroupBy(node.GroupBy)
	if err != nil {
		return nil, err
	}
	window, err := nodeWindow(node.Window)
	if err != nil {
		return nil, err
	}
	return &vitess.Select{
		QueryOpts:   vitess.QueryOpts{Distinct: node.Distinct},
		SelectExprs: selectExprs,
		From:        from,
		Where:       where,
		GroupBy:     groupBy,
		Having:      having,
		Window:      window,
	}, nil
}
