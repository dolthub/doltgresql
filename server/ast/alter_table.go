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

// nodeAlterTable handles *tree.AlterTable nodes.
func nodeAlterTable(node *tree.AlterTable) (*vitess.AlterTable, error) {
	if node == nil {
		return nil, nil
	}

	treeTableName := node.Table.ToTableName()
	tableName, err := nodeTableName(&treeTableName)
	if err != nil {
		return nil, err
	}

	statements, err := nodeAlterTableCmds(node.Cmds, tableName, node.IfExists)
	if err != nil {
		return nil, err
	}

	return &vitess.AlterTable{
		Table:      tableName,
		Statements: statements,
	}, nil
}

// nodeAlterTableSetSchema handles *tree.AlterTableSetSchema nodes.
func nodeAlterTableSetSchema(node *tree.AlterTableSetSchema) (vitess.Statement, error) {
	if node == nil {
		return nil, nil
	}
	return nil, fmt.Errorf("ALTER TABLE SET SCHEMA is not yet supported")
}

// nodeAlterTableCmds converts tree.AlterTableCmds into a slice of vitess.DDL
// instances that can be executed by GMS. |tableName| identifies the table
// being altered, and |ifExists| indicates whether the IF EXISTS clause was
// specified.
func nodeAlterTableCmds(
	node tree.AlterTableCmds,
	tableName vitess.TableName,
	ifExists bool) ([]*vitess.DDL, error) {

	if len(node) == 0 {
		return nil, fmt.Errorf("no commands specified for ALTER TABLE statement")
	} else if len(node) > 1 {
		return nil, fmt.Errorf("ALTER TABLE with multiple commands is not yet supported")
	}

	vitessDdlCmds := make([]*vitess.DDL, 0, len(node))
	for _, cmd := range node {
		switch cmd := cmd.(type) {
		case *tree.AlterTableAddConstraint:
			statement, err := nodeAlterTableAddConstraint(cmd, tableName, ifExists)
			if err != nil {
				return nil, err
			}
			vitessDdlCmds = append(vitessDdlCmds, statement)

		default:
			return nil, fmt.Errorf("ALTER TABLE with unsupported command type %T", cmd)
		}
	}

	return vitessDdlCmds, nil
}

// nodeAlterTableAddConstraint converts a tree.AlterTableAddConstraint instance
// into a vitess.DDL instance that can be executed by GMS. |tableName| identifies
// the table being altered, and |ifExists| indicates whether the IF EXISTS clause
// was specified.
func nodeAlterTableAddConstraint(
	node *tree.AlterTableAddConstraint,
	tableName vitess.TableName,
	ifExists bool) (*vitess.DDL, error) {

	if node.ValidationBehavior == tree.ValidationSkip {
		return nil, fmt.Errorf("NOT VALID is not supported yet")
	}

	switch constraintDef := node.ConstraintDef.(type) {
	case *tree.UniqueConstraintTableDef:
		return nodeUniqueConstraintTableDef(constraintDef, tableName, ifExists)
	case *tree.ForeignKeyConstraintTableDef:
		foreignKeyDefinition, err := nodeForeignKeyConstraintTableDef(constraintDef)
		if err != nil {
			return nil, err
		}
		return &vitess.DDL{
			Action:           "alter",
			ConstraintAction: "add",
			Table:            tableName,
			IfExists:         ifExists,
			TableSpec: &vitess.TableSpec{
				Constraints: []*vitess.ConstraintDefinition{
					{Details: foreignKeyDefinition},
				},
			},
		}, nil

	default:
		return nil, fmt.Errorf("ALTER TABLE with unsupported constraint "+
			"definition type %T", node)
	}
}
