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
		var err error
		var statement *vitess.DDL
		switch cmd := cmd.(type) {
		case *tree.AlterTableAddConstraint:
			statement, err = nodeAlterTableAddConstraint(cmd, tableName, ifExists)
		case *tree.AlterTableAddColumn:
			statement, err = nodeAlterTableAddColumn(cmd, tableName, ifExists)
		case *tree.AlterTableDropColumn:
			statement, err = nodeAlterTableDropColumn(cmd, tableName, ifExists)
		case *tree.AlterTableRenameColumn:
			statement, err = nodeAlterTableRenameColumn(cmd, tableName, ifExists)
		case *tree.AlterTableSetDefault:
			statement, err = nodeAlterTableSetDefault(cmd, tableName, ifExists)
		case *tree.AlterTableDropNotNull:
			statement, err = nodeAlterTableDropNotNull(cmd, tableName, ifExists)
		case *tree.AlterTableSetNotNull:
			statement, err = nodeAlterTableSetNotNull(cmd, tableName, ifExists)
		case *tree.AlterTableAlterColumnType:
			statement, err = nodeAlterTableAlterColumnType(cmd, tableName, ifExists)
		default:
			return nil, fmt.Errorf("ALTER TABLE with unsupported command type %T", cmd)
		}

		if err != nil {
			return nil, err
		}
		vitessDdlCmds = append(vitessDdlCmds, statement)
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

// nodeAlterTableAddColumn converts a tree.AlterTableAddColumn instance into an equivalent vitess.DDL instance.
func nodeAlterTableAddColumn(node *tree.AlterTableAddColumn, tableName vitess.TableName, ifExists bool) (*vitess.DDL, error) {
	if node.IfNotExists {
		return nil, fmt.Errorf("IF NOT EXISTS on a column in an ADD COLUMN statement is not supported yet")
	}

	vitessColumnDef, checkConstraints, err := nodeColumnTableDef(node.ColumnDef)
	if err != nil {
		return nil, err
	}

	return &vitess.DDL{
		Action:       "alter",
		ColumnAction: "add",
		Table:        tableName,
		IfExists:     ifExists,
		Column:       vitessColumnDef.Name,
		TableSpec: &vitess.TableSpec{
			Columns: []*vitess.ColumnDefinition{
				{
					Name: vitessColumnDef.Name,
					Type: vitessColumnDef.Type,
				},
			},
			Constraints: checkConstraints,
		},
	}, nil
}

// nodeAlterTableDropColumn converts a tree.AlterTableDropColumn instance into an equivalent vitess.DDL instance.
func nodeAlterTableDropColumn(node *tree.AlterTableDropColumn, tableName vitess.TableName, ifExists bool) (*vitess.DDL, error) {
	if node.IfExists {
		return nil, fmt.Errorf("IF EXISTS on a column in a DROP COLUMN statement is not supported yet")
	}

	switch node.DropBehavior {
	case tree.DropDefault:
	case tree.DropRestrict:
		return nil, fmt.Errorf("ALTER TABLE DROP COLUMN does not support RESTRICT option")
	case tree.DropCascade:
		return nil, fmt.Errorf("ALTER TABLE DROP COLUMN does not support CASCADE option")
	default:
		return nil, fmt.Errorf("ALTER TABLE with unsupported drop behavior %v", node.DropBehavior)
	}

	return &vitess.DDL{
		Action:       "alter",
		ColumnAction: "drop",
		Table:        tableName,
		IfExists:     ifExists,
		Column:       vitess.NewColIdent(node.Column.String()),
	}, nil
}

// nodeAlterTableRenameColumn converts a tree.AlterTableRenameColumn instance into an equivalent vitess.DDL instance.
func nodeAlterTableRenameColumn(node *tree.AlterTableRenameColumn, tableName vitess.TableName, ifExists bool) (*vitess.DDL, error) {
	return &vitess.DDL{
		Action:       "alter",
		ColumnAction: "rename",
		Table:        tableName,
		IfExists:     ifExists,
		Column:       vitess.NewColIdent(node.Column.String()),
		ToColumn:     vitess.NewColIdent(node.NewName.String()),
	}, nil
}

// nodeAlterTableSetDefault converts a tree.AlterTableSetDefault instance into an equivalent vitess.DDL instance.
func nodeAlterTableSetDefault(node *tree.AlterTableSetDefault, tableName vitess.TableName, ifExists bool) (*vitess.DDL, error) {
	expr, err := nodeExpr(node.Default)
	if err != nil {
		return nil, err
	}

	// GMS requires the AST to wrap function expressions in parens
	if _, ok := expr.(*vitess.FuncExpr); ok {
		expr = &vitess.ParenExpr{Expr: expr}
	}

	return &vitess.DDL{
		Action:   "alter",
		Table:    tableName,
		IfExists: ifExists,
		DefaultSpec: &vitess.DefaultSpec{
			Action: "set",
			Column: vitess.NewColIdent(node.Column.String()),
			Value:  expr,
		},
	}, nil
}

// nodeAlterTableAlterColumnType converts a tree.AlterTableAlterColumnType instance into an equivalent vitess.DDL instance.
func nodeAlterTableAlterColumnType(node *tree.AlterTableAlterColumnType, tableName vitess.TableName, ifExists bool) (*vitess.DDL, error) {
	if node.Collation != "" {
		return nil, fmt.Errorf("ALTER TABLE with COLLATE is not supported yet")
	}

	if node.Using != nil {
		return nil, fmt.Errorf("ALTER TABLE with USING is not supported yet")
	}

	convertType, resolvedType, err := nodeResolvableTypeReference(node.ToType)
	if err != nil {
		return nil, err
	}

	return &vitess.DDL{
		Action:   "alter",
		Table:    tableName,
		IfExists: ifExists,
		ColumnTypeSpec: &vitess.ColumnTypeSpec{
			Column: vitess.NewColIdent(node.Column.String()),
			Type: vitess.ColumnType{
				Type:         convertType.Type,
				ResolvedType: resolvedType,
				Length:       convertType.Length,
				Scale:        convertType.Scale,
				Charset:      convertType.Charset,
			},
		},
	}, nil
}

// nodeAlterTableDropNotNull converts a tree.AlterTableDropNotNull instance into an equivalent vitess.DDL instance.
func nodeAlterTableDropNotNull(node *tree.AlterTableDropNotNull, tableName vitess.TableName, ifExists bool) (*vitess.DDL, error) {
	return &vitess.DDL{
		Action:   "alter",
		Table:    tableName,
		IfExists: ifExists,
		NotNullSpec: &vitess.NotNullSpec{
			Action: "drop",
			Column: vitess.NewColIdent(node.Column.String()),
		},
	}, nil
}

// nodeAlterTableSetNotNull converts a tree.AlterTableSetNotNull instance into an equivalent vitess.DDL instance.
func nodeAlterTableSetNotNull(node *tree.AlterTableSetNotNull, tableName vitess.TableName, ifExists bool) (*vitess.DDL, error) {
	return &vitess.DDL{
		Action:   "alter",
		Table:    tableName,
		IfExists: ifExists,
		NotNullSpec: &vitess.NotNullSpec{
			Action: "set",
			Column: vitess.NewColIdent(node.Column.String()),
		},
	}, nil
}
