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
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
	"github.com/dolthub/doltgresql/server/auth"
)

// nodeUpdate handles *tree.Update nodes.
func nodeUpdate(ctx *Context, node *tree.Update) (update *vitess.Update, err error) {
	if node == nil {
		return nil, nil
	}
	ctx.Auth().PushAuthType(auth.AuthType_UPDATE)
	defer ctx.Auth().PopAuthType()

	var returningExprs vitess.SelectExprs
	if returning, ok := node.Returning.(*tree.ReturningExprs); ok {
		returningExprs, err = nodeSelectExprs(ctx, tree.SelectExprs(*returning))
		if err != nil {
			return nil, err
		}
	}

	with, err := nodeWith(ctx, node.With)
	if err != nil {
		return nil, err
	}
	table, err := nodeTableExpr(ctx, node.Table)
	if err != nil {
		return nil, err
	}

	tableExprs := vitess.TableExprs{table}
	if len(node.From) > 0 {
		vitessTableExprs := make(vitess.TableExprs, len(node.From))
		for i, tableExpr := range node.From {
			vitessTableExpr, err := nodeTableExpr(ctx, tableExpr)
			if err != nil {
				return nil, err
			}
			vitessTableExprs[i] = vitessTableExpr
		}

		tableExprs = []vitess.TableExpr{
			&vitess.JoinTableExpr{
				Join:      vitess.JoinStr,
				LeftExpr:  buildJoinTableExpressionTree(ctx, vitessTableExprs),
				RightExpr: table,
			},
		}
	}

	exprs, err := nodeUpdateExprs(ctx, node.Exprs)
	if err != nil {
		return nil, err
	}
	where, err := nodeWhere(ctx, node.Where)
	if err != nil {
		return nil, err
	}
	orderBy, err := nodeOrderBy(ctx, node.OrderBy)
	if err != nil {
		return nil, err
	}
	limit, err := nodeLimit(ctx, node.Limit)
	if err != nil {
		return nil, err
	}
	return &vitess.Update{
		TableExprs: tableExprs,
		With:       with,
		Exprs:      exprs,
		Where:      where,
		OrderBy:    orderBy,
		Limit:      limit,
		Returning:  returningExprs,
	}, nil
}

// buildJoinTableExpressionTree returns an expression tree of JoinTableExprs with |tableExprs| as the
// leaf nodes. If |tableExprs| is empty or nil, then nil is returned.
func buildJoinTableExpressionTree(ctx *Context, tableExprs vitess.TableExprs) vitess.TableExpr {
	switch len(tableExprs) {
	case 0:
		return nil
	case 1:
		return tableExprs[0]
	case 2:
		return &vitess.JoinTableExpr{
			Join:      vitess.JoinStr,
			LeftExpr:  tableExprs[0],
			RightExpr: tableExprs[1],
		}
	default:
		subtree := buildJoinTableExpressionTree(ctx, tableExprs[0:2])
		return buildJoinTableExpressionTree(ctx, append(tableExprs[2:], subtree))
	}
}
