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
	"github.com/dolthub/doltgresql/server/auth"
)

// nodeTableExpr handles tree.TableExpr nodes.
func nodeTableExpr(ctx *Context, node tree.TableExpr) (vitess.TableExpr, error) {
	switch node := node.(type) {
	case *tree.AliasedTableExpr:
		return nodeAliasedTableExpr(ctx, node)
	case *tree.JoinTableExpr:
		left, err := nodeTableExpr(ctx, node.Left)
		if err != nil {
			return nil, err
		}
		right, err := nodeTableExpr(ctx, node.Right)
		if err != nil {
			return nil, err
		}
		var condition vitess.JoinCondition
		switch treeCondition := node.Cond.(type) {
		case tree.NaturalJoinCond:
			// Nothing to do, the default value is equivalent
		case *tree.OnJoinCond:
			onExpr, err := nodeExpr(ctx, treeCondition.Expr)
			if err != nil {
				return nil, err
			}
			condition.On = onExpr
		case *tree.UsingJoinCond:
			condition.Using = make([]vitess.ColIdent, len(treeCondition.Cols))
			for i := range treeCondition.Cols {
				condition.Using[i] = vitess.NewColIdent(string(treeCondition.Cols[i]))
			}
		case nil:
			// cross join (no join condition)
		default:
			return nil, errors.Errorf("unknown JOIN condition: `%T`", treeCondition)
		}
		var joinType string
		switch node.JoinType {
		case tree.AstFull:
			joinType = vitess.FullOuterJoinStr
		case tree.AstLeft:
			if condition.On == nil && len(condition.Using) == 0 {
				joinType = vitess.NaturalLeftJoinStr
			} else {
				joinType = vitess.LeftJoinStr
			}
		case tree.AstRight:
			if condition.On == nil && len(condition.Using) == 0 {
				joinType = vitess.NaturalRightJoinStr
			} else {
				joinType = vitess.RightJoinStr
			}
		case tree.AstCross, tree.AstInner:
			joinType = vitess.JoinStr
		case "":
			if condition.On == nil && len(condition.Using) == 0 {
				joinType = vitess.NaturalJoinStr
			} else {
				joinType = vitess.JoinStr
			}
		default:
			return nil, errors.Errorf("unknown JOIN type: `%s`", node.JoinType)
		}
		return &vitess.JoinTableExpr{
			LeftExpr:  left,
			Join:      joinType,
			RightExpr: right,
			Condition: condition,
		}, nil
	case *tree.ParenTableExpr:
		tableExpr, err := nodeTableExpr(ctx, node.Expr)
		if err != nil {
			return nil, err
		}
		return &vitess.ParenTableExpr{
			Exprs: vitess.TableExprs{tableExpr},
		}, nil
	case *tree.RowsFromExpr:
		exprs, err := nodeExprs(ctx, node.Items)
		if err != nil {
			return nil, err
		}
		//TODO: not sure if this is correct at all. I think we want to return one result per row, but maybe not.
		// This needs to be tested to verify.
		rows := make([]vitess.ValTuple, len(exprs))
		for i := range exprs {
			rows[i] = vitess.ValTuple{exprs[i]}
		}
		return &vitess.ValuesStatement{
			Rows: rows,
		}, nil
	case *tree.StatementSource:
		return nil, errors.Errorf("this statement is not yet supported")
	case *tree.Subquery:
		return nodeSubqueryToTableExpr(ctx, node)
	case *tree.TableName:
		tableName, err := nodeTableName(ctx, node)
		if err != nil {
			return nil, err
		}
		return &vitess.AliasedTableExpr{
			Expr: tableName,
			Auth: vitess.AuthInformation{
				AuthType:    ctx.Auth().PeekAuthType(),
				TargetType:  auth.AuthTargetType_TableIdentifiers,
				TargetNames: []string{tableName.DbQualifier.String(), tableName.SchemaQualifier.String(), tableName.Name.String()},
			},
		}, nil
	case *tree.TableRef:
		return nil, errors.Errorf("table refs are not yet supported")
	case *tree.UnresolvedObjectName:
		tableName, err := nodeUnresolvedObjectName(ctx, node)
		if err != nil {
			return nil, err
		}
		return &vitess.AliasedTableExpr{
			Expr: tableName,
			Auth: vitess.AuthInformation{
				AuthType:    ctx.Auth().PeekAuthType(),
				TargetType:  auth.AuthTargetType_TableIdentifiers,
				TargetNames: []string{tableName.DbQualifier.String(), tableName.SchemaQualifier.String(), tableName.Name.String()},
			},
		}, nil
	default:
		return nil, errors.Errorf("unknown table expression: `%T`", node)
	}
}

// nodeTableExprs handles tree.TableExprs nodes.
func nodeTableExprs(ctx *Context, node tree.TableExprs) (vitess.TableExprs, error) {
	if len(node) == 0 {
		return nil, nil
	}
	exprs := make(vitess.TableExprs, len(node))
	for i := range node {
		var err error
		exprs[i], err = nodeTableExpr(ctx, node[i])
		if err != nil {
			return nil, err
		}
	}
	return exprs, nil
}
