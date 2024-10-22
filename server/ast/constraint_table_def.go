// Copyright 2024 Dolthub, Inc.
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

// nodeCheckConstraintTableDef converts a tree.CheckConstraintTableDef instance
// into a vitess.DDL instance that can be executed by GMS. |tableName| identifies
// the table being altered, and |ifExists| indicates whether the IF EXISTS clause
// was specified.
func nodeCheckConstraintTableDef(
	node *tree.CheckConstraintTableDef,
	tableName vitess.TableName,
	ifExists bool) (*vitess.DDL, error) {

	if node.NoInherit {
		return nil, fmt.Errorf("NO INHERIT is not yet supported for check constraints")
	}

	expr, err := nodeExpr(node.Expr)
	if err != nil {
		return nil, err
	}

	return &vitess.DDL{
		Action:           "alter",
		Table:            tableName,
		IfExists:         ifExists,
		ConstraintAction: "add",
		TableSpec: &vitess.TableSpec{
			Constraints: []*vitess.ConstraintDefinition{
				{
					Name: node.Name.String(),
					Details: &vitess.CheckConstraintDefinition{
						Expr:     expr,
						Enforced: true,
					},
				},
			},
		},
	}, nil
}

// nodeAlterTableDropConstraint converts a tree.AlterTableDropConstraint instance
// into a vitess.DDL instance that can be executed by GMS. |tableName| identifies
// the table being altered, and |ifExists| indicates whether the IF EXISTS clause
// was specified.
func nodeAlterTableDropConstraint(
	node *tree.AlterTableDropConstraint,
	tableName vitess.TableName,
	ifExists bool) (*vitess.DDL, error) {

	if node.DropBehavior == tree.DropCascade {
		return nil, fmt.Errorf("CASCADE is not yet supported for drop constraint")
	}

	if node.IfExists {
		return nil, fmt.Errorf("IF EXISTS is not yet supported for drop constraint")
	}

	return &vitess.DDL{
		Action:           "alter",
		Table:            tableName,
		IfExists:         ifExists,
		ConstraintAction: "drop",
		TableSpec: &vitess.TableSpec{
			Constraints: []*vitess.ConstraintDefinition{
				{Name: node.Constraint.String()},
			},
		},
	}, nil
}

// nodeUniqueConstraintTableDef converts a tree.UniqueConstraintTableDef instance
// into a vitess.DDL instance that can be executed by GMS. |tableName| identifies
// the table being altered, and |ifExists| indicates whether the IF EXISTS clause
// was specified.
func nodeUniqueConstraintTableDef(
	node *tree.UniqueConstraintTableDef,
	tableName vitess.TableName,
	ifExists bool) (*vitess.DDL, error) {

	if len(node.IndexParams.StorageParams) > 0 {
		return nil, fmt.Errorf("STORAGE parameters not yet supported for indexes")
	}

	if node.IndexParams.Tablespace != "" {
		return nil, fmt.Errorf("TABLESPACE is not yet supported")
	}

	if node.NullsNotDistinct {
		return nil, fmt.Errorf("NULLS NOT DISTINCT is not yet supported")
	}

	columns, err := nodeIndexElemList(node.Columns)
	if err != nil {
		return nil, err
	}

	indexType := "unique"
	if node.PrimaryKey {
		indexType = "primary"
	}

	return &vitess.DDL{
		Action:   "alter",
		Table:    tableName,
		IfExists: ifExists,
		IndexSpec: &vitess.IndexSpec{
			Action:  "create",
			Type:    indexType,
			Columns: columns,
		},
	}, nil
}
