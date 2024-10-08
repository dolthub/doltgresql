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
	// Multiple tables in the FROM column with an "equals" filter for some columns within each table should be treated
	// as a join. The analyzer should catch this, however GMS processes this form of a join differently than a standard
	// join, which is currently incompatible with Doltgres expressions. As a workaround, we rewrite the tree so that we
	// pass along a join node.
	// TODO: handle more than two tables, also make this more robust with handling more node types
	if len(node.From.Tables) == 2 && node.Where != nil {
		tableNames := make(map[tree.TableName]int)
		tableAliases := make(map[tree.TableName]int)
		// First we need to get the table names and aliases, since they'll be referenced by the filters
		for i := range node.From.Tables {
			switch table := node.From.Tables[i].(type) {
			case *tree.AliasedTableExpr:
				if tableName, ok := table.Expr.(*tree.TableName); ok {
					tableNames[*tableName] = i
				} else {
					goto PostJoinRewrite
				}
				tableAliases[tree.MakeUnqualifiedTableName(table.As.Alias)] = i
			case *tree.TableName:
				tableNames[*table] = i
			case *tree.UnresolvedObjectName:
				tableNames[table.ToTableName()] = i
			default:
				goto PostJoinRewrite
			}
		}
		// For now, we'll check if the entire filter should be moved into the join condition. Eventually, this should
		// move only the needed expressions into the join condition.
		var delveExprs func(expr tree.Expr) bool
		delveExprs = func(expr tree.Expr) bool {
			switch expr := expr.(type) {
			case *tree.AndExpr:
				return delveExprs(expr.Left) && delveExprs(expr.Right)
			case *tree.OrExpr:
				return delveExprs(expr.Left) && delveExprs(expr.Right)
			case *tree.ComparisonExpr:
				if expr.Operator != tree.EQ {
					return false
				}
				var refTables [2]int
				for argIndex, arg := range []tree.Expr{expr.Left, expr.Right} {
					switch arg := arg.(type) {
					case *tree.UnresolvedName:
						refTable := arg.GetUnresolvedObjectName().ToTableName()
						if aliasIndex, ok := tableAliases[refTable]; ok {
							refTables[argIndex] = aliasIndex
						} else if tableIndex, ok := tableNames[refTable]; ok {
							refTables[argIndex] = tableIndex
						} else {
							return false
						}
					default:
						return false
					}
				}
				// In this case, the expression does not reference multiple tables, so it's not a join condition
				if refTables[0] == refTables[1] {
					return false
				}
				return true
			default:
				return false
			}
		}
		if !delveExprs(node.Where.Expr) {
			goto PostJoinRewrite
		}
		// The filter condition represents a join, so we need to rewrite our FROM node to be a join node
		node.From.Tables = tree.TableExprs{&tree.JoinTableExpr{
			JoinType: "",
			Left:     node.From.Tables[0],
			Right:    node.From.Tables[1],
			Cond:     &tree.OnJoinCond{Expr: node.Where.Expr},
		}}
		node.Where = nil
	}
PostJoinRewrite:
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
