// Copyright 2026 Dolthub, Inc.
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

package node

import (
	"context"
	"io"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/expranalysis"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/plan"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/core"
)

// ValidateConstraint handles the ALTER TABLE ... VALIDATE CONSTRAINT statement.
type ValidateConstraint struct {
	DbProvider     sql.DatabaseProvider
	schemaName     string
	tableName      string
	constraintName string
}

var _ sql.ExecSourceRel = (*ValidateConstraint)(nil)
var _ sql.MultiDatabaser = (*ValidateConstraint)(nil)
var _ vitess.Injectable = (*ValidateConstraint)(nil)

// NewValidateConstraint returns a new *ValidateConstraint.
func NewValidateConstraint(schemaName, tableName, constraintName string) *ValidateConstraint {
	return &ValidateConstraint{
		schemaName:     schemaName,
		tableName:      tableName,
		constraintName: constraintName,
	}
}

// Children implements sql.ExecSourceRel.
func (v *ValidateConstraint) Children() []sql.Node { return nil }

// IsReadOnly implements sql.ExecSourceRel.
func (v *ValidateConstraint) IsReadOnly() bool { return false }

// Resolved implements sql.ExecSourceRel.
func (v *ValidateConstraint) Resolved() bool {
	return v.DbProvider != nil
}

// Schema implements sql.ExecSourceRel.
func (v *ValidateConstraint) Schema(ctx *sql.Context) sql.Schema { return nil }

// String implements sql.ExecSourceRel.
func (v *ValidateConstraint) String() string { return "VALIDATE CONSTRAINT" }

// WithChildren implements sql.ExecSourceRel.
func (v *ValidateConstraint) WithChildren(ctx *sql.Context, children ...sql.Node) (sql.Node, error) {
	return plan.NillaryWithChildren(v, children...)
}

// WithResolvedChildren implements vitess.Injectable.
func (v *ValidateConstraint) WithResolvedChildren(ctx context.Context, children []any) (any, error) {
	if len(children) != 0 {
		return nil, ErrVitessChildCount.New(0, len(children))
	}
	return v, nil
}

// DatabaseProvider implements sql.MultiDatabaser.
func (v *ValidateConstraint) DatabaseProvider() sql.DatabaseProvider { return v.DbProvider }

// WithDatabaseProvider implements sql.MultiDatabaser.
func (v *ValidateConstraint) WithDatabaseProvider(provider sql.DatabaseProvider) (sql.Node, error) {
	nv := *v
	nv.DbProvider = provider
	return &nv, nil
}

// RowIter implements sql.ExecSourceRel.
func (v *ValidateConstraint) RowIter(ctx *sql.Context, r sql.Row) (sql.RowIter, error) {
	db, err := v.DbProvider.Database(ctx, ctx.GetCurrentDatabase())
	if err != nil {
		return nil, err
	}

	schemaName := v.schemaName
	if schemaName == "" {
		schemaName, err = core.GetCurrentSchema(ctx)
		if err != nil {
			return nil, err
		}
	}
	schemaDb, ok := db.(sql.SchemaDatabase)
	if !ok {
		return nil, errors.Errorf("database does not support schemas")
	}
	dbSchema, ok, err := schemaDb.GetSchema(ctx, schemaName)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, errors.Errorf("schema %s does not exist", schemaName)
	}
	tblNode, _, err := dbSchema.GetTableInsensitive(ctx, v.tableName)
	if err != nil {
		return nil, err
	}
	if tblNode == nil {
		return nil, sql.ErrTableNotFound.New(v.tableName)
	}

	// validate foreign key constraint
	if fkTbl, ok := tblNode.(sql.ForeignKeyTable); ok {
		fks, err := fkTbl.GetDeclaredForeignKeys(ctx)
		if err != nil {
			return nil, err
		}
		for _, fk := range fks {
			if strings.EqualFold(fk.Name, v.constraintName) {
				return v.validateForeignKey(ctx, db, fkTbl, fk)
			}
		}
	}

	// validate check constraint
	if checkTbl, ok := tblNode.(sql.CheckTable); ok {
		checks, err := checkTbl.GetChecks(ctx)
		if err != nil {
			return nil, err
		}
		for _, check := range checks {
			if strings.EqualFold(check.Name, v.constraintName) {
				return v.validateCheckConstraint(ctx, db, tblNode, check)
			}
		}
	}

	return nil, errors.Errorf(`constraint "%s" of relation "%s" does not exist`, v.constraintName, v.tableName)
}

func (v *ValidateConstraint) validateCheckConstraint(ctx *sql.Context, db sql.Database, tblNode sql.Table, check sql.CheckDefinition) (sql.RowIter, error) {
	if !check.IsNotValid {
		return sql.RowsToRowIter(), nil
	}

	checkExpr, err := expranalysis.ResolveExpression(ctx, v.tableName, check.CheckExpression)
	if err != nil {
		return nil, errors.Errorf("could not parse check expression for constraint %q: %v", check.Name, err)
	}

	partitions, err := tblNode.Partitions(ctx)
	if err != nil {
		return nil, err
	}
	defer partitions.Close(ctx)

	for {
		partition, err := partitions.Next(ctx)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		rows, err := tblNode.PartitionRows(ctx, partition)
		if err != nil {
			return nil, err
		}
		defer rows.Close(ctx)
		for {
			row, err := rows.Next(ctx)
			if err == io.EOF {
				break
			}
			if err != nil {
				return nil, err
			}
			res, err := sql.EvaluateCondition(ctx, checkExpr, row)
			if err != nil {
				return nil, err
			}
			if sql.IsFalse(res) {
				return nil, sql.ErrCheckConstraintViolated.New(check.Name)
			}
		}
	}

	// TODO: clear the IsNotValid flag so pg_constraint reports convalidated = true
	return sql.RowsToRowIter(), nil
}

func (v *ValidateConstraint) validateForeignKey(ctx *sql.Context, db sql.Database, fkTbl sql.ForeignKeyTable, fkDef sql.ForeignKeyConstraint) (sql.RowIter, error) {
	if !fkDef.IsNotValid {
		return sql.RowsToRowIter(), nil
	}

	refTblNode, _, err := db.GetTableInsensitive(ctx, fkDef.ParentTable)
	if err != nil {
		return nil, err
	}
	if refTblNode == nil {
		return nil, sql.ErrTableNotFound.New(fkDef.ParentTable)
	}
	refFkTbl, ok := refTblNode.(sql.ForeignKeyTable)
	if !ok {
		return nil, errors.Errorf("table %s does not support foreign key constraints", fkDef.ParentTable)
	}

	// this is to allow calling plan.ResolveForeignKey function to validate
	fkDef.IsResolved = false
	if err = plan.ResolveForeignKey(ctx, fkTbl, refFkTbl, fkDef, false, true, true); err != nil {
		// TODO: fix - currently error message includes "cannot add or update a child row"
		return nil, err
	}

	// undo - it's not used but safe to set it
	fkDef.IsResolved = true

	// clear the IsNotValid flag so pg_constraint reports convalidated = true
	fkDef.IsNotValid = false
	if err = fkTbl.UpdateForeignKey(ctx, fkDef.Name, fkDef); err != nil {
		return nil, err
	}

	return sql.RowsToRowIter(), nil
}
