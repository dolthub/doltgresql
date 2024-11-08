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
	"sort"

	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
	"github.com/dolthub/doltgresql/utils"
)

// assignTableDef handles tree.TableDef nodes for *vitess.DDL targets. Some table defs, such as indexes, affect other
// defs, such as columns, and they're therefore dependent on columns being handled first. It is up to the caller to
// ensure that all defs have been ordered properly before calling. assignTableDefs handles the sort for you, so this
// notice is only relevant when individually calling assignTableDef.
func assignTableDef(ctx *Context, node tree.TableDef, target *vitess.DDL) error {
	switch node := node.(type) {
	case *tree.CheckConstraintTableDef:
		if target.TableSpec == nil {
			target.TableSpec = &vitess.TableSpec{}
		}
		expr, err := nodeExpr(ctx, node.Expr)
		if err != nil {
			return err
		}
		target.TableSpec.Constraints = append(target.TableSpec.Constraints, &vitess.ConstraintDefinition{
			Name: string(node.Name),
			Details: &vitess.CheckConstraintDefinition{
				Expr:     expr,
				Enforced: true,
			},
		})
		return nil
	case *tree.ColumnTableDef:
		if target.TableSpec == nil {
			target.TableSpec = &vitess.TableSpec{}
		}
		columnDef, err := nodeColumnTableDef(ctx, node)
		if err != nil {
			return err
		}
		target.TableSpec.AddColumn(columnDef)
		return nil
	case *tree.ForeignKeyConstraintTableDef:
		if target.TableSpec == nil {
			target.TableSpec = &vitess.TableSpec{}
		}
		fkDef, err := nodeForeignKeyConstraintTableDef(ctx, node)
		if err != nil {
			return err
		}
		target.TableSpec.Constraints = append(target.TableSpec.Constraints, &vitess.ConstraintDefinition{
			Name:    string(node.Name),
			Details: fkDef,
		})
		return nil
	case *tree.IndexTableDef:
		if target.TableSpec == nil {
			target.TableSpec = &vitess.TableSpec{}
		}
		indexDef, err := nodeIndexTableDef(ctx, node)
		if err != nil {
			return err
		}
		target.TableSpec.Indexes = append(target.TableSpec.Indexes, indexDef)
		return nil
	case *tree.LikeTableDef:
		if len(node.Options) > 0 {
			return fmt.Errorf("options for LIKE are not yet supported")
		}
		tableName, err := nodeTableName(ctx, &node.Name)
		if err != nil {
			return err
		}
		target.OptLike = &vitess.OptLike{
			LikeTable: tableName,
		}
		return nil
	case *tree.UniqueConstraintTableDef:
		if target.TableSpec == nil {
			target.TableSpec = &vitess.TableSpec{}
		}
		indexDef, err := nodeIndexTableDef(ctx, &node.IndexTableDef)
		if err != nil {
			return err
		}
		indexDef.Info.Unique = true
		indexDef.Info.Primary = node.PrimaryKey
		// If we're setting a primary key, then we need to make sure that all of the columns are also set to NOT NULL
		if indexDef.Info.Primary {
			tableColumns := utils.SliceToMapValues(target.TableSpec.Columns, func(col *vitess.ColumnDefinition) string {
				return col.Name.String()
			})
			for _, indexedColumn := range indexDef.Columns {
				if column, ok := tableColumns[indexedColumn.Column.String()]; ok {
					column.Type.Null = false
					column.Type.NotNull = true
				}
			}
		}
		target.TableSpec.Indexes = append(target.TableSpec.Indexes, indexDef)
		return nil
	case nil:
		return nil
	default:
		return fmt.Errorf("unknown table definition encountered")
	}
}

// assignTableDefs handles tree.TableDefs nodes for *vitess.DDL targets. This also sorts table defs by whether they're
// dependent on other table defs evaluating first. Some table defs, such as indexes, affect other defs, such as columns,
// and they're therefore dependent on columns being handled first.
func assignTableDefs(ctx *Context, node tree.TableDefs, target *vitess.DDL) error {
	sortedNode := make(tree.TableDefs, len(node))
	copy(sortedNode, node)
	sort.Slice(sortedNode, func(i, j int) bool {
		var cmps [2]int
		for cmpsIdx := range []tree.TableDef{sortedNode[i], sortedNode[j]} {
			switch sortedNode[i].(type) {
			case *tree.IndexTableDef:
				cmps[cmpsIdx] = 1
			case *tree.UniqueConstraintTableDef:
				cmps[cmpsIdx] = 2
			default:
				cmps[cmpsIdx] = 0
			}
		}
		return cmps[0] < cmps[1]
	})
	for i := range sortedNode {
		if err := assignTableDef(ctx, sortedNode[i], target); err != nil {
			return err
		}
	}
	return nil
}
