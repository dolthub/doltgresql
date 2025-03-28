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

	"github.com/cockroachdb/errors"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
)

// nodeAlterTable handles *tree.AlterTable nodes.
func nodeAlterTable(ctx *Context, node *tree.AlterTable) (vitess.Statement, error) {
	if node == nil {
		return nil, nil
	}

	treeTableName := node.Table.ToTableName()
	tableName, err := nodeTableName(ctx, &treeTableName)
	if err != nil {
		return nil, err
	}

	statements, noOps, err := nodeAlterTableCmds(ctx, node.Cmds, tableName, node.IfExists)
	if err != nil {
		return nil, err
	}

	// If there are no valid statements return a no-op statement
	if len(noOps) > 0 && len(statements) == 0 {
		return NewNoOp(noOps), nil
	}

	// Otherwise emit warnings now, then return an AlterTable statement
	// TODO: we don't have a way to send or store the warnings alongside a valid AlterTable statement. We could either
	//  get a *sql.Context here and emit warnings, or we could store the warnings in the Context and make the caller
	//  emit them before it sends |ReadyForQuery|
	// if len(noOps) > 0 {
	//  emit warnings
	// }

	return &vitess.AlterTable{
		Table:      tableName,
		Statements: statements,
	}, nil
}

// nodeAlterTableSetSchema handles *tree.AlterTableSetSchema nodes.
func nodeAlterTableSetSchema(ctx *Context, node *tree.AlterTableSetSchema) (vitess.Statement, error) {
	if node == nil {
		return nil, nil
	}
	return nil, errors.Errorf("ALTER TABLE SET SCHEMA is not yet supported")
}

// nodeAlterTableCmds converts tree.AlterTableCmds into a slice of vitess.DDL instances that can be executed by GMS.
// |tableName| identifies the table being altered, and |ifExists| indicates whether the IF EXISTS clause was specified.
// A second slice of unsupported but safely ignored statements is return as well.
func nodeAlterTableCmds(
	ctx *Context,
	node tree.AlterTableCmds,
	tableName vitess.TableName,
	ifExists bool) ([]*vitess.DDL, []string, error) {

	if len(node) == 0 {
		return nil, nil, errors.Errorf("no commands specified for ALTER TABLE statement")
	}

	vitessDdlCmds := make([]*vitess.DDL, 0, len(node))
	var unsupportedWarnings []string
	for _, cmd := range node {
		var err error
		var statement *vitess.DDL
		switch cmd := cmd.(type) {
		case *tree.AlterTableAddConstraint:
			statement, err = nodeAlterTableAddConstraint(ctx, cmd, tableName, ifExists)
			if err != nil {
				return nil, nil, err
			}
			vitessDdlCmds = append(vitessDdlCmds, statement)
		case *tree.AlterTableAddColumn:
			statement, err = nodeAlterTableAddColumn(ctx, cmd, tableName, ifExists)
			if err != nil {
				return nil, nil, err
			}
			vitessDdlCmds = append(vitessDdlCmds, statement)

			// If inline constraints have been specified, set the ConstraintAction so that they get processed
			if len(statement.TableSpec.Constraints) > 0 {
				statement.ConstraintAction = vitess.AddStr
			}

		case *tree.AlterTableDropColumn:
			statement, err = nodeAlterTableDropColumn(ctx, cmd, tableName, ifExists)
			if err != nil {
				return nil, nil, err
			}
			vitessDdlCmds = append(vitessDdlCmds, statement)
		case *tree.AlterTableDropConstraint:
			statement, err = nodeAlterTableDropConstraint(ctx, cmd, tableName, ifExists)
			if err != nil {
				return nil, nil, err
			}
			vitessDdlCmds = append(vitessDdlCmds, statement)
		case *tree.AlterTableRenameColumn:
			statement, err = nodeAlterTableRenameColumn(ctx, cmd, tableName, ifExists)
			if err != nil {
				return nil, nil, err
			}
			vitessDdlCmds = append(vitessDdlCmds, statement)
		case *tree.AlterTableSetDefault:
			statement, err = nodeAlterTableSetDefault(ctx, cmd, tableName, ifExists)
			if err != nil {
				return nil, nil, err
			}
			vitessDdlCmds = append(vitessDdlCmds, statement)
		case *tree.AlterTableDropNotNull:
			statement, err = nodeAlterTableDropNotNull(ctx, cmd, tableName, ifExists)
			if err != nil {
				return nil, nil, err
			}
			vitessDdlCmds = append(vitessDdlCmds, statement)
		case *tree.AlterTableSetNotNull:
			statement, err = nodeAlterTableSetNotNull(ctx, cmd, tableName, ifExists)
			if err != nil {
				return nil, nil, err
			}
			vitessDdlCmds = append(vitessDdlCmds, statement)
		case *tree.AlterTableAlterColumnType:
			statement, err = nodeAlterTableAlterColumnType(ctx, cmd, tableName, ifExists)
			if err != nil {
				return nil, nil, err
			}
			vitessDdlCmds = append(vitessDdlCmds, statement)
		case *tree.AlterTableOwner:
			unsupportedWarnings = append(unsupportedWarnings, fmt.Sprintf("ALTER TABLE %s OWNER TO %s", tableName.String(), cmd.Owner))
		default:
			return nil, nil, errors.Errorf("ALTER TABLE with unsupported command type %T", cmd)
		}
	}

	return vitessDdlCmds, unsupportedWarnings, nil
}

// nodeAlterTableAddConstraint converts a tree.AlterTableAddConstraint instance
// into a vitess.DDL instance that can be executed by GMS. |tableName| identifies
// the table being altered, and |ifExists| indicates whether the IF EXISTS clause
// was specified.
func nodeAlterTableAddConstraint(
	ctx *Context,
	node *tree.AlterTableAddConstraint,
	tableName vitess.TableName,
	ifExists bool) (*vitess.DDL, error) {

	if node.ValidationBehavior == tree.ValidationSkip {
		return nil, errors.Errorf("NOT VALID is not supported yet")
	}

	switch constraintDef := node.ConstraintDef.(type) {
	case *tree.CheckConstraintTableDef:
		return nodeCheckConstraintTableDef(ctx, constraintDef, tableName, ifExists)
	case *tree.UniqueConstraintTableDef:
		return nodeUniqueConstraintTableDef(ctx, constraintDef, tableName, ifExists)
	case *tree.ForeignKeyConstraintTableDef:
		foreignKeyDefinition, err := nodeForeignKeyConstraintTableDef(ctx, constraintDef)
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
					{
						Name:    bareIdentifier(constraintDef.Name),
						Details: foreignKeyDefinition,
					},
				},
			},
		}, nil

	default:
		return nil, errors.Errorf("ALTER TABLE with unsupported constraint "+
			"definition type %T", node)
	}
}

// bareIdentifier returns the string representation of a name without any quoting
// (quoted is the default Name.String() behavior)
func bareIdentifier(id tree.Name) string {
	ctx := tree.NewFmtCtx(tree.FmtBareIdentifiers)
	id.Format(ctx)
	return ctx.CloseAndGetString()
}

// nodeAlterTableAddColumn converts a tree.AlterTableAddColumn instance into an equivalent vitess.DDL instance.
func nodeAlterTableAddColumn(ctx *Context, node *tree.AlterTableAddColumn, tableName vitess.TableName, ifExists bool) (*vitess.DDL, error) {
	if node.IfNotExists {
		return nil, errors.Errorf("IF NOT EXISTS on a column in an ADD COLUMN statement is not supported yet")
	}

	vitessColumnDef, err := nodeColumnTableDef(ctx, node.ColumnDef)
	if err != nil {
		return nil, err
	}

	tableSpec := &vitess.TableSpec{}
	tableSpec.AddColumn(vitessColumnDef)

	if node.ColumnDef.References.Table != nil {
		constraintDef, err := nodeForeignKeyDefinitionFromColumnTableDef(ctx, node.ColumnDef.Name, node.ColumnDef)
		if err != nil {
			return nil, err
		}
		tableSpec.AddConstraint(&vitess.ConstraintDefinition{Details: constraintDef})
	}

	return &vitess.DDL{
		Action:       "alter",
		ColumnAction: "add",
		Table:        tableName,
		IfExists:     ifExists,
		Column:       vitessColumnDef.Name,
		TableSpec:    tableSpec,
	}, nil
}

// nodeAlterTableDropColumn converts a tree.AlterTableDropColumn instance into an equivalent vitess.DDL instance.
func nodeAlterTableDropColumn(ctx *Context, node *tree.AlterTableDropColumn, tableName vitess.TableName, ifExists bool) (*vitess.DDL, error) {
	if node.IfExists {
		return nil, errors.Errorf("IF EXISTS on a column in a DROP COLUMN statement is not supported yet")
	}

	switch node.DropBehavior {
	case tree.DropDefault:
	case tree.DropRestrict:
		return nil, errors.Errorf("ALTER TABLE DROP COLUMN does not support RESTRICT option")
	case tree.DropCascade:
		return nil, errors.Errorf("ALTER TABLE DROP COLUMN does not support CASCADE option")
	default:
		return nil, errors.Errorf("ALTER TABLE with unsupported drop behavior %v", node.DropBehavior)
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
func nodeAlterTableRenameColumn(ctx *Context, node *tree.AlterTableRenameColumn, tableName vitess.TableName, ifExists bool) (*vitess.DDL, error) {
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
func nodeAlterTableSetDefault(ctx *Context, node *tree.AlterTableSetDefault, tableName vitess.TableName, ifExists bool) (*vitess.DDL, error) {
	expr, err := nodeExpr(ctx, node.Default)
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
func nodeAlterTableAlterColumnType(ctx *Context, node *tree.AlterTableAlterColumnType, tableName vitess.TableName, ifExists bool) (*vitess.DDL, error) {
	if node.Collation != "" {
		return nil, errors.Errorf("ALTER TABLE with COLLATE is not supported yet")
	}

	if node.Using != nil {
		return nil, errors.Errorf("ALTER TABLE with USING is not supported yet")
	}

	convertType, resolvedType, err := nodeResolvableTypeReference(ctx, node.ToType)
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
func nodeAlterTableDropNotNull(ctx *Context, node *tree.AlterTableDropNotNull, tableName vitess.TableName, ifExists bool) (*vitess.DDL, error) {
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
func nodeAlterTableSetNotNull(ctx *Context, node *tree.AlterTableSetNotNull, tableName vitess.TableName, ifExists bool) (*vitess.DDL, error) {
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

// nodeAlterTableSetNotNull converts a tree.AlterTablePartition instance into an equivalent vitess.DDL instance.
func nodeAlterTablePartition(ctx *Context, node *tree.AlterTablePartition) (*vitess.AlterTable, error) {
	if node == nil {
		return nil, nil
	}

	// TODO: This is an incomplete translation because our GMS implementation doesn't support the MySQL
	//   equivalent of these statements either. Regardless, these are all no-ops.
	treeTableName := node.Name.ToTableName()
	tableName, err := nodeTableName(ctx, &treeTableName)
	if err != nil {
		return nil, err
	}

	switch node.Spec.Type {
	case tree.PartitionBoundIn:
	case tree.PartitionBoundFromTo:
	case tree.PartitionBoundWith:
	default:
		return nil, errors.Errorf("ALTER TABLE with unsupported partition type %v", node.Spec.Type)
	}

	partitionDef := &vitess.PartitionDefinition{
		Name: vitess.NewColIdent(node.Partition.String()),
	}

	actionStr := ""
	if node.IsDetach {
		actionStr = vitess.DropStr
	} else {
		actionStr = vitess.AddStr
	}

	partitionSpec := &vitess.PartitionSpec{
		Action:      actionStr,
		Definitions: []*vitess.PartitionDefinition{partitionDef},
	}

	return &vitess.AlterTable{
		Table:          tableName,
		PartitionSpecs: []*vitess.PartitionSpec{partitionSpec},
	}, nil
}
