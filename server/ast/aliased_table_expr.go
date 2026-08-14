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
	"strings"

	"github.com/cockroachdb/errors"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
	"github.com/dolthub/doltgresql/server/auth"
)

// nodeAliasedTableExpr handles *tree.AliasedTableExpr nodes.
func nodeAliasedTableExpr(ctx *Context, node *tree.AliasedTableExpr) (*vitess.AliasedTableExpr, error) {
	if node.Ordinality {
		if _, ok := node.Expr.(*tree.RowsFromExpr); !ok {
			return nil, errors.Errorf("WITH ORDINALITY is only supported for functions")
		}
	}
	if node.IndexFlags != nil {
		return nil, errors.Errorf("index flags are not yet supported")
	}
	var aliasExpr vitess.SimpleTableExpr
	var authInfo vitess.AuthInformation

	switch expr := node.Expr.(type) {
	case *tree.TableName:
		tableName, err := nodeTableName(ctx, expr)
		if err != nil {
			return nil, err
		}
		aliasExpr = tableName
		authInfo = vitess.AuthInformation{
			AuthType:    ctx.Auth().PeekAuthType(),
			TargetType:  auth.AuthTargetType_TableIdentifiers,
			TargetNames: []string{tableName.DbQualifier.String(), tableName.SchemaQualifier.String(), tableName.Name.String()},
		}
	case *tree.Subquery:
		tableExpr, err := nodeTableExpr(ctx, expr)
		if err != nil {
			return nil, err
		}

		ate, ok := tableExpr.(*vitess.AliasedTableExpr)
		if !ok {
			return nil, errors.Errorf("expected *vitess.AliasedTableExpr, found %T", tableExpr)
		}

		var selectStmt vitess.SelectStatement
		switch ate.Expr.(type) {
		case *vitess.Subquery:
			selectStmt = ate.Expr.(*vitess.Subquery).Select
		default:
			return nil, errors.Errorf("unhandled subquery table expression: `%T`", tableExpr)
		}

		// If the subquery is a VALUES statement, it should be represented more directly
		innerSelect := selectStmt
		if parentSelect, ok := innerSelect.(*vitess.ParenSelect); ok {
			innerSelect = parentSelect.Select
		}
		if inSelect, ok := innerSelect.(*vitess.Select); ok {
			if isTrivialSelectStar(inSelect) {
				if aliasedTblExpr, ok := inSelect.From[0].(*vitess.AliasedTableExpr); ok {
					if valuesStmt, ok := aliasedTblExpr.Expr.(*vitess.ValuesStatement); ok {
						if len(node.As.Cols) > 0 {
							columns := make([]vitess.ColIdent, len(node.As.Cols))
							for i := range node.As.Cols {
								columns[i] = vitess.NewColIdent(string(node.As.Cols[i]))
							}
							valuesStmt.Columns = columns
						}
						aliasExpr = valuesStmt
						break
					}
				}
			}
		}

		subquery := &vitess.Subquery{
			Select: selectStmt,
		}

		if len(node.As.Cols) > 0 {
			columns := make([]vitess.ColIdent, len(node.As.Cols))
			for i := range node.As.Cols {
				columns[i] = vitess.NewColIdent(string(node.As.Cols[i]))
			}
			subquery.Columns = columns
		}
		aliasExpr = subquery
	case *tree.RowsFromExpr:
		var selectStmt vitess.SelectStatement
		if node.Ordinality {
			// WITH ORDINALITY appends a bigint column numbering the function's result rows, named
			// "ordinality" unless renamed by a column alias list. The numbering projection has to
			// live one level above the function's expansion, so we expand the function in the
			// select list of a wrapped subquery.
			items, err := nodeExprs(ctx, expr.Items)
			if err != nil {
				return nil, err
			}
			innerExprs := make(vitess.SelectExprs, len(items))
			for i := range items {
				innerExprs[i] = &vitess.AliasedExpr{Expr: items[i]}
			}
			selectStmt = &vitess.Select{
				SelectExprs: vitess.SelectExprs{
					&vitess.StarExpr{},
					&vitess.AliasedExpr{
						Expr: &vitess.FuncExpr{
							Name: vitess.NewColIdent("row_number"),
							Over: &vitess.Over{},
						},
						As: vitess.NewColIdent("ordinality"),
					},
				},
				From: vitess.TableExprs{
					&vitess.AliasedTableExpr{
						Expr: &vitess.Subquery{Select: &vitess.Select{SelectExprs: innerExprs}},
						As:   vitess.NewTableIdent("with_ordinality"),
					},
				},
			}
		} else {
			tableExpr, err := nodeTableExpr(ctx, expr)
			if err != nil {
				return nil, err
			}

			// TODO: this should be represented as a table function more directly
			selectStmt = &vitess.Select{
				From: vitess.TableExprs{tableExpr},
			}
		}
		subquery := &vitess.Subquery{
			Select: selectStmt,
		}

		if len(node.As.Cols) > 0 {
			columns := make([]vitess.ColIdent, len(node.As.Cols))
			for i := range node.As.Cols {
				columns[i] = vitess.NewColIdent(string(node.As.Cols[i]))
			}
			subquery.Columns = columns
		}
		aliasExpr = subquery
	default:
		return nil, errors.Errorf("unhandled table expression: `%T`", expr)
	}
	alias := string(node.As.Alias)
	if alias == "" && node.Ordinality {
		// A derived table needs an alias; the implicit alias of a function called in FROM is the
		// function's name, matching Postgres
		alias = "with_ordinality"
		if rf, ok := node.Expr.(*tree.RowsFromExpr); ok && len(rf.Items) == 1 {
			if fe, ok := rf.Items[0].(*tree.FuncExpr); ok {
				nameParts := strings.Split(fe.Func.String(), ".")
				alias = strings.ToLower(nameParts[len(nameParts)-1])
			}
		}
	}

	var asOf *vitess.AsOf
	if node.AsOf != nil {
		asOfExpr, err := nodeExpr(ctx, node.AsOf.Expr)
		if err != nil {
			return nil, err
		}
		// TODO: other forms of AS OF (not just point in time)
		asOf = &vitess.AsOf{
			Time: asOfExpr,
		}
	}

	return &vitess.AliasedTableExpr{
		Expr:    aliasExpr,
		As:      vitess.NewTableIdent(alias),
		AsOf:    asOf,
		Lateral: node.Lateral,
		Auth:    authInfo,
	}, nil
}

// isTrivialSelectStar returns true when the Select is just "SELECT * FROM <single table>"
// with no other clauses that would alter semantics (no WHERE, ORDER BY, LIMIT, GROUP BY,
// HAVING, DISTINCT, or WITH).
func isTrivialSelectStar(s *vitess.Select) bool {
	if len(s.From) != 1 ||
		s.QueryOpts.Distinct ||
		s.With != nil ||
		s.Limit != nil ||
		len(s.OrderBy) != 0 ||
		s.Where != nil ||
		len(s.GroupBy) != 0 ||
		s.Having != nil ||
		len(s.SelectExprs) != 1 {
		return false
	}
	starExpr, ok := s.SelectExprs[0].(*vitess.StarExpr)
	if !ok {
		return false
	}
	return starExpr.TableName.IsEmpty()
}
