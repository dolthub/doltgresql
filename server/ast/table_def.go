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
)

// assignTableDef handles tree.TableDef nodes for *vitess.DDL targets.
func assignTableDef(node tree.TableDef, target *vitess.DDL) error {
	switch node := node.(type) {
	case *tree.CheckConstraintTableDef:
		if target.TableSpec == nil {
			target.TableSpec = &vitess.TableSpec{}
		}
		expr, err := nodeExpr(node.Expr)
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
		columnDef, err := nodeColumnTableDef(node)
		if err != nil {
			return err
		}
		target.TableSpec.Columns = append(target.TableSpec.Columns, columnDef)
		return nil
	case *tree.FamilyTableDef:
		return fmt.Errorf("FAMILY is not yet supported")
	case *tree.ForeignKeyConstraintTableDef:
		if target.TableSpec == nil {
			target.TableSpec = &vitess.TableSpec{}
		}
		fkDef, err := nodeForeignKeyConstraintTableDef(node)
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
		indexDef, err := nodeIndexTableDef(node)
		if err != nil {
			return err
		}
		target.TableSpec.Indexes = append(target.TableSpec.Indexes, indexDef)
		return nil
	case *tree.LikeTableDef:
		if len(node.Options) > 0 {
			return fmt.Errorf("options for LIKE are not yet supported")
		}
		tableName, err := nodeTableName(&node.Name)
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
		indexDef, err := nodeIndexTableDef(&node.IndexTableDef)
		if err != nil {
			return err
		}
		indexDef.Info.Unique = true
		indexDef.Info.Primary = node.PrimaryKey
		target.TableSpec.Indexes = append(target.TableSpec.Indexes, indexDef)
		return nil
	case nil:
		return nil
	default:
		return fmt.Errorf("unknown table definition encountered")
	}
}

// assignTableDefs handles tree.TableDefs nodes for *vitess.DDL targets.
func assignTableDefs(node tree.TableDefs, target *vitess.DDL) error {
	for i := range node {
		if err := assignTableDef(node[i], target); err != nil {
			return err
		}
	}
	return nil
}
