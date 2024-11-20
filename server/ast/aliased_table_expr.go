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
	"github.com/dolthub/doltgresql/server/auth"
)

// nodeAliasedTableExpr handles *tree.AliasedTableExpr nodes.
func nodeAliasedTableExpr(ctx *Context, node *tree.AliasedTableExpr) (*vitess.AliasedTableExpr, error) {
	if node.Ordinality {
		return nil, fmt.Errorf("ordinality is not yet supported")
	}
	if node.IndexFlags != nil {
		return nil, fmt.Errorf("index flags are not yet supported")
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
			return nil, fmt.Errorf("expected *vitess.AliasedTableExpr, found %T", tableExpr)
		}

		var selectStmt vitess.SelectStatement
		switch ate.Expr.(type) {
		case *vitess.Subquery:
			selectStmt = ate.Expr.(*vitess.Subquery).Select
		default:
			return nil, fmt.Errorf("unhandled subquery table expression: `%T`", tableExpr)
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
	case *tree.RowsFromExpr:
		tableExpr, err := nodeTableExpr(ctx, expr)
		if err != nil {
			return nil, err
		}

		// TODO: this should be represented as a table function more directly
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
		aliasExpr = subquery
	default:
		return nil, fmt.Errorf("unhandled table expression: `%T`", expr)
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
