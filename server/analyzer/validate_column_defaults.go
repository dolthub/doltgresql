// Copyright 2020-2021 Dolthub, Inc.
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

package analyzer

import (
	pgnode "github.com/dolthub/doltgresql/server/node"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/analyzer"
	"github.com/dolthub/go-mysql-server/sql/expression"
	"github.com/dolthub/go-mysql-server/sql/plan"
	"github.com/dolthub/go-mysql-server/sql/transform"
)

// validateColumnDefaults ensures that newly created column defaults from a DDL statement are legal for the type of
// column, various other business logic checks to match MySQL's logic.
func ValidateColumnDefaults(ctx *sql.Context, _ *analyzer.Analyzer, n sql.Node, _ *plan.Scope, _ analyzer.RuleSelector, qFlags *sql.QueryFlags) (sql.Node, transform.TreeIdentity, error) {
	span, ctx := ctx.Span("validateColumnDefaults")
	defer span.End()

	return transform.Node(n, func(n sql.Node) (sql.Node, transform.TreeIdentity, error) {
		switch node := n.(type) {
		case *plan.AlterDefaultSet:
			table := getResolvedTable(node)
			sch := table.Schema()
			index := sch.IndexOfColName(node.ColumnName)
			if index == -1 {
				return nil, transform.SameTree, sql.ErrColumnNotFound.New(node.ColumnName)
			}
			col := sch[index]
			err := validateColumnDefault(ctx, col, node.Default)
			if err != nil {
				return node, transform.SameTree, err
			}

			return node, transform.SameTree, nil

		case sql.SchemaTarget:
			switch node.(type) {
			case *plan.AlterPK, *plan.AddColumn, *plan.ModifyColumn, *plan.AlterDefaultDrop, *plan.CreateTable, *plan.DropColumn, *pgnode.CreateTable:
				// DDL nodes must validate any new column defaults, continue to logic below
			default:
				// other node types are not altering the schema and therefore don't need validation of column defaults
				return n, transform.SameTree, nil
			}

			// There may be multiple DDL nodes in the plan (ALTER TABLE statements can have many clauses), and for each of them
			// we need to count the column indexes in the very hacky way outlined above.
			i := 0
			return transform.NodeExprs(n, func(e sql.Expression) (sql.Expression, transform.TreeIdentity, error) {
				eWrapper, ok := e.(*expression.Wrapper)
				if !ok {
					return e, transform.SameTree, nil
				}

				defer func() {
					i++
				}()

				eVal := eWrapper.Unwrap()
				if eVal == nil {
					return e, transform.SameTree, nil
				}
				colDefault, ok := eVal.(*sql.ColumnDefaultValue)
				if !ok {
					return e, transform.SameTree, nil
				}

				col, err := lookupColumnForTargetSchema(ctx, node, i)
				if err != nil {
					return nil, transform.SameTree, err
				}

				err = validateColumnDefault(ctx, col, colDefault)
				if err != nil {
					return nil, transform.SameTree, err
				}

				return e, transform.SameTree, nil
			})
		default:
			return node, transform.SameTree, nil
		}
	})
}

// lookupColumnForTargetSchema looks at the target schema for the specified SchemaTarget node and returns
// the column based on the specified index. For most node types, this is simply indexing into the target
// schema but a few types require special handling.
func lookupColumnForTargetSchema(_ *sql.Context, node sql.SchemaTarget, colIndex int) (*sql.Column, error) {
	schema := node.TargetSchema()

	switch n := node.(type) {
	case *plan.ModifyColumn:
		if colIndex < len(schema) {
			return schema[colIndex], nil
		} else {
			return n.NewColumn(), nil
		}
	case *plan.AddColumn:
		if colIndex < len(schema) {
			return schema[colIndex], nil
		} else {
			return n.Column(), nil
		}
	case *plan.AlterDefaultSet:
		index := schema.IndexOfColName(n.ColumnName)
		if index == -1 {
			return nil, sql.ErrTableColumnNotFound.New(n.Table, n.ColumnName)
		}
		return schema[index], nil
	default:
		if colIndex < len(schema) {
			return schema[colIndex], nil
		} else {
			// TODO: sql.ErrColumnNotFound would be a better error here, but we need to add all the different node types to
			//  the switch to get it
			return nil, expression.ErrIndexOutOfBounds.New(colIndex, len(schema))
		}
	}
}

// validateColumnDefault validates that the column default expression is valid for the column type and returns an error
// if not
func validateColumnDefault(ctx *sql.Context, col *sql.Column, colDefault *sql.ColumnDefaultValue) error {
	if colDefault == nil {
		return nil
	}

	var err error
	sql.Inspect(colDefault.Expr, func(e sql.Expression) bool {
		switch e.(type) {
		case sql.FunctionExpression, *expression.UnresolvedFunction:
			// TODO: functions must be deterministic to be used in column defaults
			return true
		case *plan.Subquery:
			err = sql.ErrColumnDefaultSubquery.New(col.Name)
			return false
		case *expression.GetField:
			if !colDefault.IsParenthesized() {
				err = sql.ErrInvalidColumnDefaultValue.New(col.Name)
				return false
			}
			return true
		default:
			return true
		}
	})

	if err != nil {
		return err
	}

	// validate type of default expression
	if err = colDefault.CheckType(ctx); err != nil {
		return err
	}

	return nil
}

// Finds first ResolvedTable node that is a descendant of the node given
// This function will not look inside SubqueryAliases
func getResolvedTable(node sql.Node) *plan.ResolvedTable {
	var table *plan.ResolvedTable
	transform.Inspect(node, func(n sql.Node) bool {
		// Inspect is called on all children of a node even if an earlier child's call returns false.
		// We only want the first TableNode match.
		if table != nil {
			return false
		}
		switch nn := n.(type) {
		case *plan.SubqueryAlias:
			// We should not be matching with ResolvedTables inside SubqueryAliases
			return false
		case *plan.ResolvedTable:
			if !plan.IsDualTable(nn) {
				table = nn
				return false
			}
		case *plan.IndexedTableAccess:
			if rt, ok := nn.TableNode.(*plan.ResolvedTable); ok {
				table = rt
				return false
			}
		}
		return true
	})
	return table
}