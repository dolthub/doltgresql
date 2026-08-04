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

	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
	pgexprs "github.com/dolthub/doltgresql/server/expression"
	"github.com/dolthub/doltgresql/server/functions/framework"

	vitess "github.com/dolthub/vitess/go/vt/sqlparser"
)

// nodeWith handles *tree.CTE nodes.
func nodeCTE(ctx *Context, node *tree.CTE) (*vitess.CommonTableExpr, error) {
	if node == nil {
		return nil, nil
	}

	alias := vitess.NewTableIdent(string(node.Name.Alias))
	cols := make([]vitess.ColIdent, len(node.Name.Cols))
	colMap := make(map[string]int)
	for i, col := range node.Name.Cols {
		cols[i] = vitess.NewColIdent(string(col))
		colMap[string(col)] = i
	}

	subSelect, ok := node.Stmt.(*tree.Select)
	if !ok {
		return nil, errors.Errorf("unsupported CTE statement type: %T", node.Stmt)
	}

	selectStmt, err := nodeSelect(ctx, subSelect)
	if err != nil {
		return nil, err
	}

	if len(node.Cycle.Fields) > 0 {
		// Ensure that the cycle columns are unique
		if _, ok = colMap[string(node.Cycle.Set)]; ok {
			return nil, errors.Errorf(`column reference "%s" is ambiguous`, string(node.Cycle.Set))
		}
		if _, ok = colMap[string(node.Cycle.Using)]; ok {
			return nil, errors.Errorf(`column reference "%s" is ambiguous`, string(node.Cycle.Using))
		}
		if node.Cycle.Set == node.Cycle.Using {
			return nil, errors.New(`cycle mark column name and cycle path column name are the same`)
		}
		cols = append(cols, vitess.NewColIdent(string(node.Cycle.Set)), vitess.NewColIdent(string(node.Cycle.Using)))
		setOp, ok := selectStmt.(*vitess.SetOp)
		// Verify the structure of the RECURSIVE statement
		if !ok || (setOp.Type != vitess.UnionStr && setOp.Type != vitess.UnionAllStr) {
			return nil, errors.New("WITH query is not recursive")
		}
		leftSelect, ok := setOp.Left.(*vitess.Select)
		if !ok {
			return nil, errors.New("with a SEARCH or CYCLE clause, the left side of the UNION must be a SELECT")
		}
		rightSelect, ok := setOp.Right.(*vitess.Select)
		if !ok {
			return nil, errors.New("with a SEARCH or CYCLE clause, the right side of the UNION must be a SELECT")
		}
		// Build the expressions that will represent the cycle check
		leftArrayExpr, err := pgexprs.NewArray(nil)
		if err != nil {
			return nil, err
		}
		leftArrayChildExprs := make(vitess.Exprs, len(node.Cycle.Fields))
		rightAnyComparators := make(vitess.Exprs, len(node.Cycle.Fields))
		rightConcatenateExprs := make(vitess.Exprs, len(node.Cycle.Fields))
		seenFields := make(map[string]struct{})
		for i, field := range node.Cycle.Fields {
			fieldName := string(field)
			// Check for repeated fields and error if a duplicate is found
			if _, ok = seenFields[fieldName]; ok {
				return nil, errors.Errorf(`cycle column "%s" specified more than once`, fieldName)
			} else {
				seenFields[fieldName] = struct{}{}
			}
			// Then verify that the field is actually in the CTE name list
			colIdx, ok := colMap[fieldName]
			if !ok {
				return nil, errors.Errorf(`cycle column "%s" not in WITH query column list`, fieldName)
			}
			leftFieldExpr, ok := leftSelect.SelectExprs[colIdx].(*vitess.AliasedExpr)
			if !ok {
				return nil, errors.Errorf("expected CYCLE field target to be `AliasedExpr` but received `%T`", leftSelect.SelectExprs[colIdx])
			}
			leftArrayChildExprs[i] = leftFieldExpr.Expr
			rightFieldExpr, ok := rightSelect.SelectExprs[colIdx].(*vitess.AliasedExpr)
			if !ok {
				return nil, errors.Errorf("expected CYCLE field target to be `AliasedExpr` but received `%T`", rightSelect.SelectExprs[colIdx])
			}
			rightAnyComparators[i] = rightFieldExpr.Expr
			rightConcatenateExprs[i] = rightFieldExpr.Expr
		}
		// Insert the cycle check expressions
		leftSelect.SelectExprs = append(leftSelect.SelectExprs,
			&vitess.AliasedExpr{
				Expr: vitess.InjectedExpr{Expression: pgexprs.NewRawLiteralBool(false)},
			},
			&vitess.AliasedExpr{
				Expr: vitess.InjectedExpr{
					Expression: leftArrayExpr,
					Children: vitess.Exprs{
						vitess.InjectedExpr{
							Expression: pgexprs.NewRecordExpr(),
							Children:   leftArrayChildExprs,
						},
					},
				},
			})
		rightSelect.SelectExprs = append(rightSelect.SelectExprs,
			&vitess.AliasedExpr{
				Expr: vitess.InjectedExpr{
					Expression: pgexprs.NewAnyExpr("="),
					Children: vitess.Exprs{
						vitess.InjectedExpr{
							Expression: pgexprs.NewRecordExpr(),
							Children:   rightAnyComparators,
						},
						vitess.NewColName(string(node.Cycle.Using)),
					},
				},
			},
			&vitess.AliasedExpr{
				Expr: vitess.InjectedExpr{
					Expression: pgexprs.NewBinaryOperator(framework.Operator_BinaryConcatenate),
					Children: vitess.Exprs{
						vitess.NewColName(string(node.Cycle.Using)),
						vitess.InjectedExpr{
							Expression: pgexprs.NewRecordExpr(),
							Children:   rightConcatenateExprs,
						},
					},
				},
			},
		)
		// Insert the cycle-ending condition
		if rightSelect.Where == nil {
			rightSelect.Where = &vitess.Where{
				Expr: &vitess.NotExpr{Expr: vitess.NewColName(string(node.Cycle.Set))},
				Type: vitess.WhereStr,
			}
		} else {
			rightSelect.Where.Expr = &vitess.AndExpr{
				Left:  rightSelect.Where.Expr,
				Right: &vitess.NotExpr{Expr: vitess.NewColName(string(node.Cycle.Set))},
			}
		}
	}

	subQuery := &vitess.Subquery{
		Select: selectStmt,
	}

	return &vitess.CommonTableExpr{
		AliasedTableExpr: &vitess.AliasedTableExpr{
			Expr: subQuery,
			As:   alias,
			Auth: vitess.AuthInformation{AuthType: vitess.AuthType_IGNORE},
		},
		Columns: cols,
	}, nil
}

// nodeWith handles *tree.With nodes.
func nodeWith(ctx *Context, node *tree.With) (*vitess.With, error) {
	if node == nil {
		return nil, nil
	}

	ctes := make([]*vitess.CommonTableExpr, len(node.CTEList))
	for i, cte := range node.CTEList {
		var err error
		ctes[i], err = nodeCTE(ctx, cte)
		if err != nil {
			return nil, err
		}
	}

	return &vitess.With{
		Recursive: node.Recursive,
		Ctes:      ctes,
	}, nil
}
