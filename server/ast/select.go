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
	"strings"

	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
	"github.com/dolthub/doltgresql/server/auth"
)

// nodeSelect handles *tree.Select nodes.
func nodeSelect(ctx *Context, node *tree.Select) (vitess.SelectStatement, error) {
	if node == nil {
		return nil, nil
	}
	if node.Select == nil {
		node.Select = &tree.ValuesClause{
			Rows: []tree.Exprs{},
		}
	}
	selectStmt, err := nodeSelectStatement(ctx, node.Select)
	if err != nil {
		return nil, err
	}
	orderBy, err := nodeOrderBy(ctx, node.OrderBy)
	if err != nil {
		return nil, err
	}
	with, err := nodeWith(ctx, node.With)
	if err != nil {
		return nil, err
	}
	limit, err := nodeLimit(ctx, node.Limit)
	if err != nil {
		return nil, err
	}
	_, err = nodeLockingClause(ctx, node.Locking)
	if err != nil {
		return nil, err
	}

	switch selectStmt := selectStmt.(type) {
	case *vitess.ParenSelect:
		// TODO: figure out if this is even correct, not sure what statement would produce this AST
		// perhaps we should use the inner select statement, but maybe it has its own order by, limit, etc.
		return &vitess.Select{
			SelectExprs: vitess.SelectExprs{
				&vitess.StarExpr{
					TableName: vitess.TableName{
						Name: vitess.NewTableIdent("*"),
					},
				},
			},
			From: vitess.TableExprs{
				&vitess.AliasedTableExpr{
					Expr: &vitess.Subquery{
						Select: selectStmt,
					},
				},
			},
			OrderBy: orderBy,
			With:    with,
			Limit:   limit,
		}, nil
	case *vitess.Select:
		selectStmt.OrderBy = orderBy
		selectStmt.With = with
		selectStmt.Limit = limit
		return selectStmt, nil
	case *vitess.SetOp:
		selectStmt.OrderBy = orderBy
		selectStmt.With = with
		selectStmt.Limit = limit
		return selectStmt, nil
	default:
		return nil, fmt.Errorf("SELECT has encountered an unknown clause: `%T`", selectStmt)
	}
}

// nodeSelectStatement handles tree.SelectStatement nodes.
func nodeSelectStatement(ctx *Context, node tree.SelectStatement) (vitess.SelectStatement, error) {
	if node == nil {
		return nil, nil
	}
	ctx.Auth().PushAuthType(auth.AuthType_SELECT)
	defer ctx.Auth().PopAuthType()

	switch node := node.(type) {
	case *tree.ParenSelect:
		return nodeParenSelect(ctx, node)
	case *tree.SelectClause:
		return nodeSelectClause(ctx, node)
	case *tree.UnionClause:
		return nodeUnionClause(ctx, node)
	case *tree.ValuesClause:
		return nodeValuesClause(ctx, node)
	default:
		return nil, fmt.Errorf("unknown type of SELECT statement: `%T`", node)
	}
}

// nodeSelectExpr handles tree.SelectExpr nodes.
func nodeSelectExpr(ctx *Context, node tree.SelectExpr) (vitess.SelectExpr, error) {
	switch expr := node.Expr.(type) {
	case *tree.AllColumnsSelector:
		if expr.TableName.NumParts > 1 {
			return nil, fmt.Errorf("referencing items outside the schema or database is not yet supported")
		}
		return &vitess.StarExpr{
			TableName: vitess.TableName{
				Name: vitess.NewTableIdent(expr.TableName.Parts[0]),
			},
		}, nil
	case tree.UnqualifiedStar:
		return &vitess.StarExpr{}, nil
	case *tree.UnresolvedName:
		if expr.NumParts > 2 {
			return nil, fmt.Errorf("referencing items outside the schema or database is not yet supported")
		}
		if expr.Star {
			var tableName vitess.TableName
			if expr.NumParts == 2 {
				tableName.Name = vitess.NewTableIdent(expr.Parts[1])
			}
			return &vitess.StarExpr{
				TableName: tableName,
			}, nil
		} else {
			var tableName vitess.TableName
			if expr.NumParts == 2 {
				tableName.Name = vitess.NewTableIdent(expr.Parts[1])
			}
			// We don't set the InputExpression for ColName expressions. This matches the behavior in vitess's
			// post-processing found in ast.go. Input expressions are load bearing for some parts of plan building
			// so we need to match the behavior exactly.
			return &vitess.AliasedExpr{
				Expr: &vitess.ColName{
					Name:      vitess.NewColIdent(expr.Parts[0]),
					Qualifier: tableName,
				},
				As: vitess.NewColIdent(string(node.As)),
			}, nil
		}
	default:
		vitessExpr, err := nodeExpr(ctx, expr)
		if err != nil {
			return nil, err
		}
		// cast part is not part of column name, e.g. `id::INT2` should create column name as `id`.
		if ce, ok := expr.(*tree.CastExpr); ok && node.As == "" {
			node.As = tree.UnrestrictedName(tree.AsString(ce.Expr))
		}
		// To be consistent with vitess handling, InputExpression always gets its outer qoutes trimmed
		inputExpression := tree.AsString(&node)
		if strings.HasPrefix(inputExpression, "'") && strings.HasSuffix(inputExpression, "'") {
			inputExpression = inputExpression[1 : len(inputExpression)-1]
		}

		return &vitess.AliasedExpr{
			Expr:            vitessExpr,
			As:              vitess.NewColIdent(string(node.As)),
			InputExpression: inputExpression,
		}, nil
	}
}

// nodeSelectExprs handles tree.SelectExprs nodes.
func nodeSelectExprs(ctx *Context, node tree.SelectExprs) (vitess.SelectExprs, error) {
	if len(node) == 0 {
		return nil, nil
	}
	selectExprs := make(vitess.SelectExprs, len(node))
	for i := range node {
		var err error
		selectExprs[i], err = nodeSelectExpr(ctx, node[i])
		if err != nil {
			return nil, err
		}
	}
	return selectExprs, nil
}

// nodeExprToSelectExpr handles tree.Expr nodes and returns the result as a vitess.SelectExpr.
func nodeExprToSelectExpr(ctx *Context, node tree.Expr) (vitess.SelectExpr, error) {
	if node == nil {
		return nil, nil
	}
	return nodeSelectExpr(ctx, tree.SelectExpr{
		Expr: node,
	})
}

// nodeExprsToSelectExprs handles tree.Exprs nodes and returns the results as vitess.SelectExprs.
func nodeExprsToSelectExprs(ctx *Context, node tree.Exprs) (vitess.SelectExprs, error) {
	if len(node) == 0 {
		return nil, nil
	}
	selectExprs := make(vitess.SelectExprs, len(node))
	for i := range node {
		var err error
		selectExprs[i], err = nodeSelectExpr(ctx, tree.SelectExpr{
			Expr: node[i],
		})
		if err != nil {
			return nil, err
		}
	}
	return selectExprs, nil
}
