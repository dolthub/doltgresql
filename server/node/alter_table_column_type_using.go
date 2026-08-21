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
	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/expranalysis"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/plan"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/core"
	pgexprs "github.com/dolthub/doltgresql/server/expression"
	"github.com/dolthub/doltgresql/server/hook"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// AlterTableColumnTypeUsing handles `ALTER TABLE ... ALTER COLUMN ... TYPE ... USING <expr>` statements. Unlike a
// plain column type change (which converts the existing values directly to the new type), the USING form computes each
// row's new value by evaluating the given expression against the old row.
type AlterTableColumnTypeUsing struct {
	DbProvider sql.DatabaseProvider
	NewType    *pgtypes.DoltgresType
	schemaName string
	tableName  string
	columnName string
	usingExpr  string
	ifExists   bool
}

var _ sql.ExecSourceRel = (*AlterTableColumnTypeUsing)(nil)
var _ sql.MultiDatabaser = (*AlterTableColumnTypeUsing)(nil)
var _ vitess.Injectable = (*AlterTableColumnTypeUsing)(nil)

// NewAlterTableColumnTypeUsing returns a new *AlterTableColumnTypeUsing. |usingExpr| is the textual (Postgres syntax)
// form of the USING expression, which is resolved against the table at execution time.
func NewAlterTableColumnTypeUsing(schemaName, tableName, columnName string, newType *pgtypes.DoltgresType, usingExpr string, ifExists bool) *AlterTableColumnTypeUsing {
	return &AlterTableColumnTypeUsing{
		NewType:    newType,
		schemaName: schemaName,
		tableName:  tableName,
		columnName: columnName,
		usingExpr:  usingExpr,
		ifExists:   ifExists,
	}
}

// Children implements sql.ExecSourceRel.
func (a *AlterTableColumnTypeUsing) Children() []sql.Node { return nil }

// IsReadOnly implements sql.ExecSourceRel.
func (a *AlterTableColumnTypeUsing) IsReadOnly() bool { return false }

// Resolved implements sql.ExecSourceRel.
func (a *AlterTableColumnTypeUsing) Resolved() bool {
	return a.DbProvider != nil
}

// Schema implements sql.ExecSourceRel.
func (a *AlterTableColumnTypeUsing) Schema(ctx *sql.Context) sql.Schema { return nil }

// String implements sql.ExecSourceRel.
func (a *AlterTableColumnTypeUsing) String() string {
	return "ALTER TABLE ALTER COLUMN TYPE USING"
}

// WithChildren implements sql.ExecSourceRel.
func (a *AlterTableColumnTypeUsing) WithChildren(ctx *sql.Context, children ...sql.Node) (sql.Node, error) {
	return plan.NillaryWithChildren(a, children...)
}

// WithResolvedChildren implements vitess.Injectable.
func (a *AlterTableColumnTypeUsing) WithResolvedChildren(ctx context.Context, children []any) (any, error) {
	if len(children) != 0 {
		return nil, ErrVitessChildCount.New(0, len(children))
	}
	return a, nil
}

// DatabaseProvider implements sql.MultiDatabaser.
func (a *AlterTableColumnTypeUsing) DatabaseProvider() sql.DatabaseProvider { return a.DbProvider }

// WithDatabaseProvider implements sql.MultiDatabaser.
func (a *AlterTableColumnTypeUsing) WithDatabaseProvider(provider sql.DatabaseProvider) (sql.Node, error) {
	na := *a
	na.DbProvider = provider
	return &na, nil
}

// RowIter implements sql.ExecSourceRel.
func (a *AlterTableColumnTypeUsing) RowIter(ctx *sql.Context, r sql.Row) (sql.RowIter, error) {
	if !a.NewType.IsResolvedType() {
		return nil, pgtypes.ErrTypeDoesNotExist.New(a.NewType.Name())
	}

	db, err := a.DbProvider.Database(ctx, ctx.GetCurrentDatabase())
	if err != nil {
		return nil, err
	}

	schemaName := a.schemaName
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
	tblNode, found, err := dbSchema.GetTableInsensitive(ctx, a.tableName)
	if err != nil {
		return nil, err
	}
	if !found {
		if a.ifExists {
			return sql.RowsToRowIter(), nil
		}
		return nil, sql.ErrTableNotFound.New(a.tableName)
	}

	rwt, ok := tblNode.(sql.RewritableTable)
	if !ok {
		return nil, errors.Errorf("ALTER COLUMN ... TYPE ... USING is not supported for table %s", a.tableName)
	}

	sch := rwt.Schema(ctx)
	colIdx := sch.IndexOfColName(a.columnName)
	if colIdx < 0 {
		return nil, sql.ErrTableColumnNotFound.New(a.tableName, a.columnName)
	}
	oldCol := sch[colIdx]
	newCol := *oldCol
	newCol.Type = a.NewType

	// Reject type changes for columns that participate in a foreign key, mirroring the behavior of the
	// non-USING column type change.
	if err = a.checkForeignKeyUsage(ctx, tblNode, oldCol.Name); err != nil {
		return nil, err
	}

	// Reject the type change if this table's implicit row type is used as a column type anywhere else.
	if err = hook.ValidateColumnTypeChangeForTable(ctx, doltdb.TableName{Name: rwt.Name(), Schema: schemaName}); err != nil {
		return nil, err
	}

	// Resolve the USING expression against the table's schema so that it can be evaluated against each row.
	usingExpr, err := expranalysis.ResolveExpression(ctx, quoteIdentifier(schemaName)+"."+quoteIdentifier(rwt.Name()), a.usingExpr)
	if err != nil {
		return nil, err
	}

	// Values produced by the USING expression are converted to the new column type with an assignment cast,
	// matching Postgres behavior.
	conversionExpr := usingExpr
	usingDoltgresType, usingTypeOk := usingExpr.Type(ctx).(*pgtypes.DoltgresType)
	if usingTypeOk && !usingDoltgresType.Equals(a.NewType) {
		conversionExpr = pgexprs.NewAssignmentCast(usingExpr, usingDoltgresType, a.NewType)
	}

	newSch := make(sql.Schema, len(sch))
	copy(newSch, sch)
	newSch[colIdx] = &newCol

	oldPkSchema := sql.SchemaToPrimaryKeySchema(ctx, rwt, sch)
	newPkSchema := sql.SchemaToPrimaryKeySchema(ctx, rwt, newSch)

	inserter, err := rwt.RewriteInserter(ctx, oldPkSchema, newPkSchema, oldCol, &newCol, nil)
	if err != nil {
		return nil, err
	}

	partitions, err := rwt.Partitions(ctx)
	if err != nil {
		return nil, err
	}

	rowIter := sql.NewTableRowIter(ctx, rwt, partitions)
	for {
		row, err := rowIter.Next(ctx)
		if err == io.EOF {
			break
		} else if err != nil {
			return nil, abortRewrite(ctx, inserter, err)
		}

		newVal, err := conversionExpr.Eval(ctx, row)
		if err != nil {
			return nil, abortRewrite(ctx, inserter, err)
		}
		if !usingTypeOk && newVal != nil {
			// The expression didn't produce a Doltgres type, so fall back to a direct conversion to the new type
			newVal, _, err = a.NewType.Convert(ctx, newVal)
			if err != nil {
				return nil, abortRewrite(ctx, inserter, err)
			}
		}
		if newVal == nil && !newCol.Nullable {
			err = errors.Errorf(`column "%s" of relation "%s" contains null values`, newCol.Name, rwt.Name())
			return nil, abortRewrite(ctx, inserter, err)
		}

		newRow := make(sql.Row, len(row))
		copy(newRow, row)
		newRow[colIdx] = newVal
		if err = inserter.Insert(ctx, newRow); err != nil {
			return nil, abortRewrite(ctx, inserter, err)
		}
	}

	if err = inserter.Close(ctx); err != nil {
		return nil, err
	}
	return sql.RowsToRowIter(), nil
}

// checkForeignKeyUsage returns an error if the named column is used by any foreign key on the given table (either as
// a child or a parent column), since changing its type would invalidate the foreign key.
func (a *AlterTableColumnTypeUsing) checkForeignKeyUsage(ctx *sql.Context, tblNode sql.Table, colName string) error {
	fkTable, ok := tblNode.(sql.ForeignKeyTable)
	if !ok {
		return nil
	}
	lowerColName := strings.ToLower(colName)
	declaredFks, err := fkTable.GetDeclaredForeignKeys(ctx)
	if err != nil {
		return err
	}
	for _, fk := range declaredFks {
		for _, fkCol := range fk.Columns {
			if strings.ToLower(fkCol) == lowerColName {
				return sql.ErrForeignKeyTypeChange.New(colName)
			}
		}
	}
	referencedFks, err := fkTable.GetReferencedForeignKeys(ctx)
	if err != nil {
		return err
	}
	for _, fk := range referencedFks {
		for _, fkCol := range fk.ParentColumns {
			if strings.ToLower(fkCol) == lowerColName {
				return sql.ErrForeignKeyTypeChange.New(colName)
			}
		}
	}
	return nil
}

// abortRewrite discards any changes made by the given inserter and closes it, returning the original error.
func abortRewrite(ctx *sql.Context, inserter sql.RowInserter, err error) error {
	_ = inserter.DiscardChanges(ctx, err)
	_ = inserter.Close(ctx)
	return err
}

// quoteIdentifier quotes the given identifier so that it can be safely embedded in a query string.
func quoteIdentifier(s string) string {
	return `"` + strings.ReplaceAll(s, `"`, `""`) + `"`
}
