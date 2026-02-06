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
func nodeAliasedTableExpr(ctx *Context, node *tree.AliasedTableExpr) (vitess.TableExpr, error) {
	if node.IndexFlags != nil {
		return nil, errors.Errorf("index flags are not yet supported")
	}

	// Handle RowsFromExpr specially - it can have WITH ORDINALITY and column aliases
	if rowsFrom, ok := node.Expr.(*tree.RowsFromExpr); ok {
		// Handle multi-argument UNNEST specially: UNNEST(arr1, arr2, ...)
		// is syntactic sugar for ROWS FROM(unnest(arr1), unnest(arr2), ...)
		// We need to detect this case and expand it to use RowsFromExpr.
		if len(rowsFrom.Items) == 1 {
			if funcExpr, ok := rowsFrom.Items[0].(*tree.FuncExpr); ok {
				funcName := funcExpr.Func.String()
				if strings.EqualFold(funcName, "unnest") && len(funcExpr.Exprs) > 1 {
					// Expand multi-arg UNNEST into separate unnest calls
					selectExprs := make(vitess.SelectExprs, len(funcExpr.Exprs))
					for i, arg := range funcExpr.Exprs {
						argExpr, err := nodeExpr(ctx, arg)
						if err != nil {
							return nil, err
						}
						selectExprs[i] = &vitess.AliasedExpr{
							Expr: &vitess.FuncExpr{
								Name:  vitess.NewColIdent("unnest"),
								Exprs: vitess.SelectExprs{&vitess.AliasedExpr{Expr: argExpr}},
							},
						}
					}

					var columns vitess.Columns
					if len(node.As.Cols) > 0 {
						columns = make(vitess.Columns, len(node.As.Cols))
						for i := range node.As.Cols {
							columns[i] = vitess.NewColIdent(string(node.As.Cols[i]))
						}
					}

					return &vitess.RowsFromExpr{
						Exprs:          selectExprs,
						WithOrdinality: node.Ordinality,
						Alias:          vitess.NewTableIdent(string(node.As.Alias)),
						Columns:        columns,
					}, nil
				}
			}
		}

		// Use RowsFromExpr for:
		// 1. Multiple functions: ROWS FROM(func1(), func2()) AS alias
		// 2. WITH ORDINALITY: ROWS FROM(func()) WITH ORDINALITY
		if len(rowsFrom.Items) > 1 || node.Ordinality {
			selectExprs := make(vitess.SelectExprs, len(rowsFrom.Items))
			for i, item := range rowsFrom.Items {
				expr, err := nodeExpr(ctx, item)
				if err != nil {
					return nil, err
				}
				selectExprs[i] = &vitess.AliasedExpr{Expr: expr}
			}

			var columns vitess.Columns
			if len(node.As.Cols) > 0 {
				columns = make(vitess.Columns, len(node.As.Cols))
				for i := range node.As.Cols {
					columns[i] = vitess.NewColIdent(string(node.As.Cols[i]))
				}
			}

			return &vitess.RowsFromExpr{
				Exprs:          selectExprs,
				WithOrdinality: node.Ordinality,
				Alias:          vitess.NewTableIdent(string(node.As.Alias)),
				Columns:        columns,
			}, nil
		}

		// For single function without ordinality, fall through to use the existing
		// table function infrastructure via nodeTableExpr
		tableExpr, err := nodeTableExpr(ctx, rowsFrom)
		if err != nil {
			return nil, err
		}

		// Wrap in a subquery as the original code did
		subquery := &vitess.Subquery{
			Select: &vitess.Select{
				From: vitess.TableExprs{tableExpr},
			},
		}

		if len(node.As.Cols) > 0 {
			columns := make([]vitess.ColIdent, len(node.As.Cols))
			for i := range node.As.Cols {
				columns[i] = vitess.NewColIdent(string(node.As.Cols[i]))
			}
			subquery.Columns = columns
		}

		return &vitess.AliasedTableExpr{
			Expr:    subquery,
			As:      vitess.NewTableIdent(string(node.As.Alias)),
			Lateral: node.Lateral,
		}, nil
	}

	// For non-RowsFromExpr expressions, ordinality is not yet supported
	if node.Ordinality {
		return nil, errors.Errorf("ordinality is only supported for ROWS FROM expressions")
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
			if len(inSelect.From) == 1 {
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
	default:
		return nil, errors.Errorf("unhandled table expression: `%T`", expr)
	}
	alias := string(node.As.Alias)

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
