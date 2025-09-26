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
	"github.com/cockroachdb/errors"

	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
)

// nodeCheckConstraintTableDef converts a tree.CheckConstraintTableDef instance
// into a vitess.DDL instance that can be executed by GMS. |tableName| identifies
// the table being altered, and |ifExists| indicates whether the IF EXISTS clause
// was specified.
func nodeCheckConstraintTableDef(
	ctx *Context,
	node *tree.CheckConstraintTableDef,
	tableName vitess.TableName,
	ifExists bool) (*vitess.DDL, error) {

	if node.NoInherit {
		return nil, errors.Errorf("NO INHERIT for check constraints is not yet supported")
	}

	expr, err := nodeExpr(ctx, node.Expr)
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
	ctx *Context,
	node *tree.AlterTableDropConstraint,
	tableName vitess.TableName,
	ifExists bool) (*vitess.DDL, error) {

	if node.DropBehavior == tree.DropCascade {
		return nil, errors.Errorf("CASCADE for drop constraint is not yet supported")
	}

	return &vitess.DDL{
		Action:             "alter",
		Table:              tableName,
		IfExists:           ifExists,
		ConstraintAction:   "drop",
		ConstraintIfExists: node.IfExists,
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
	ctx *Context,
	node *tree.UniqueConstraintTableDef,
	tableName vitess.TableName,
	ifExists bool) (*vitess.DDL, error) {

	if len(node.IndexParams.StorageParams) > 0 {
		return nil, errors.Errorf("STORAGE parameters for indexes are not yet supported")
	}

	if node.IndexParams.Tablespace != "" {
		return nil, errors.Errorf("TABLESPACE is not yet supported")
	}

	if node.NullsNotDistinct {
		return nil, errors.Errorf("NULLS NOT DISTINCT is not yet supported")
	}

	columns, err := nodeIndexElemList(ctx, node.Columns)
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
			ToName:  vitess.NewColIdent(bareIdentifier(node.Name)),
			Action:  "create",
			Type:    indexType,
			Columns: columns,
		},
	}, nil
}
