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
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/expression"
	"github.com/dolthub/go-mysql-server/sql/plan"
	"github.com/dolthub/go-mysql-server/sql/transform"
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
	SchemaName string
	TableName  string
	ColumnName string
	IfExists   bool
	usingExpr  sql.Expression
	// table is the resolved target table, assigned during analysis by the resolveTableForDDL rule. It is nil when
	// the table does not exist, which analysis only tolerates when IfExists is set; execution is then a no-op.
	table sql.TableNode
}

var _ sql.ExecSourceRel = (*AlterTableColumnTypeUsing)(nil)
var _ sql.Expressioner = (*AlterTableColumnTypeUsing)(nil)
var _ sql.MultiDatabaser = (*AlterTableColumnTypeUsing)(nil)
var _ vitess.Injectable = (*AlterTableColumnTypeUsing)(nil)

// NewAlterTableColumnTypeUsing returns a new *AlterTableColumnTypeUsing. The USING expression is provided through
// WithResolvedChildren as the single injected child.
func NewAlterTableColumnTypeUsing(schemaName, tableName, columnName string, newType *pgtypes.DoltgresType, ifExists bool) *AlterTableColumnTypeUsing {
	return &AlterTableColumnTypeUsing{
		NewType:    newType,
		SchemaName: schemaName,
		TableName:  tableName,
		ColumnName: columnName,
		IfExists:   ifExists,
	}
}

// Children implements sql.ExecSourceRel.
func (a *AlterTableColumnTypeUsing) Children() []sql.Node { return nil }

// IsReadOnly implements sql.ExecSourceRel.
func (a *AlterTableColumnTypeUsing) IsReadOnly() bool { return false }

// Resolved implements sql.ExecSourceRel. The table field is not part of this check because a missing table with
// IF EXISTS legitimately leaves it nil.
func (a *AlterTableColumnTypeUsing) Resolved() bool {
	return a.DbProvider != nil && a.usingExpr != nil && a.usingExpr.Resolved()
}

// WithResolvedTable returns a copy of this node with the target table assigned.
func (a *AlterTableColumnTypeUsing) WithResolvedTable(table sql.TableNode) *AlterTableColumnTypeUsing {
	na := *a
	na.table = table
	return &na
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

// Expressions implements sql.Expressioner.
func (a *AlterTableColumnTypeUsing) Expressions() []sql.Expression {
	return []sql.Expression{a.usingExpr}
}

// WithExpressions implements sql.Expressioner.
func (a *AlterTableColumnTypeUsing) WithExpressions(ctx *sql.Context, exprs ...sql.Expression) (sql.Node, error) {
	if len(exprs) != 1 {
		return nil, sql.ErrInvalidChildrenNumber.New(a, len(exprs), 1)
	}
	na := *a
	na.usingExpr = exprs[0]
	return &na, nil
}

// WithResolvedChildren implements vitess.Injectable.
func (a *AlterTableColumnTypeUsing) WithResolvedChildren(ctx context.Context, children []any) (any, error) {
	if len(children) != 1 {
		return nil, ErrVitessChildCount.New(1, len(children))
	}
	usingExpr, ok := children[0].(sql.Expression)
	if !ok {
		return nil, errors.Errorf("expected vitess child to be an expression but has type `%T`", children[0])
	}
	na := *a
	na.usingExpr = usingExpr
	return &na, nil
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

	if a.table == nil {
		// The table does not exist; analysis only permits this with IF EXISTS, making this statement a no-op
		if a.IfExists {
			return sql.RowsToRowIter(), nil
		}
		return nil, sql.ErrTableNotFound.New(a.TableName)
	}

	tbl := a.table.UnderlyingTable()
	rwt, ok := tbl.(sql.RewritableTable)
	if !ok {
		return nil, errors.Errorf("ALTER COLUMN ... TYPE ... USING is not supported for table %s", a.TableName)
	}

	sch := rwt.Schema(ctx)
	colIdx := sch.IndexOfColName(a.ColumnName)
	if colIdx < 0 {
		return nil, sql.ErrTableColumnNotFound.New(a.TableName, a.ColumnName)
	}
	oldCol := sch[colIdx]
	newCol := *oldCol
	newCol.Type = a.NewType

	// Reject type changes for columns that participate in a foreign key, mirroring the behavior of the
	// non-USING column type change.
	if err := a.checkForeignKeyUsage(ctx, tbl, oldCol.Name); err != nil {
		return nil, err
	}

	// Reject the type change if this table's implicit row type is used as a column type anywhere else.
	if doltTable := core.SQLTableToDoltTable(tbl); doltTable != nil {
		if err := hook.ValidateColumnTypeChangeForTable(ctx, doltTable.TableName()); err != nil {
			return nil, err
		}
	}

	// Bind any column references in the USING expression to the table's schema.
	usingExpr, err := a.bindColumnReferences(ctx, sch)
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

	if err := a.rewriteRows(ctx, rwt, oldPkSchema, newPkSchema, oldCol, &newCol, conversionExpr, usingTypeOk, colIdx); err != nil {
		return nil, err
	}
	return sql.RowsToRowIter(), nil
}

// rewriteRows rewrites the given table, computing the new value of the altered column for each row by evaluating
// |conversionExpr| against the old row. If any row fails to convert, the rewrite is aborted and the original table is
// left untouched. It is critical that the rewrite inserter is never finalized (which swaps in the rewritten table)
// after a failure: DiscardChanges must always record the failure before Close is called, which prevents Close from
// flushing the partially rewritten table.
func (a *AlterTableColumnTypeUsing) rewriteRows(
	ctx *sql.Context,
	rwt sql.RewritableTable,
	oldPkSchema, newPkSchema sql.PrimaryKeySchema,
	oldCol, newCol *sql.Column,
	conversionExpr sql.Expression,
	usingTypeOk bool,
	colIdx int,
) (err error) {
	inserter, err := rwt.RewriteInserter(ctx, oldPkSchema, newPkSchema, oldCol, newCol, nil)
	if err != nil {
		return err
	}

	// Any exit before the rewrite completes successfully (including a panic) must abort the rewrite, so that the
	// inserter never finalizes the partially rewritten table.
	rewriteComplete := false
	defer func() {
		if !rewriteComplete {
			err = abortRewrite(ctx, inserter, err)
		}
	}()

	partitions, err := rwt.Partitions(ctx)
	if err != nil {
		return err
	}

	rowIter := sql.NewTableRowIter(ctx, rwt, partitions)
	defer func() {
		if closeErr := rowIter.Close(ctx); closeErr != nil && err == nil && !rewriteComplete {
			err = closeErr
		}
	}()

	for {
		row, err := rowIter.Next(ctx)
		if err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		newVal, err := conversionExpr.Eval(ctx, row)
		if err != nil {
			return err
		}
		if !usingTypeOk && newVal != nil {
			// The expression didn't produce a Doltgres type, so fall back to a direct conversion to the new type
			newVal, _, err = a.NewType.Convert(ctx, newVal)
			if err != nil {
				return err
			}
		}
		if newVal == nil && !newCol.Nullable {
			return errors.Errorf(`column "%s" of relation "%s" contains null values`, newCol.Name, rwt.Name())
		}

		newRow := make(sql.Row, len(row))
		copy(newRow, row)
		newRow[colIdx] = newVal
		if err = inserter.Insert(ctx, newRow); err != nil {
			return err
		}
	}

	// The rewrite is only finalized (swapping in the rewritten table) by this Close call on the success path.
	// rewriteComplete is set first so that a failure during Close doesn't attempt to abort the half-closed inserter.
	rewriteComplete = true
	return inserter.Close(ctx)
}

// bindColumnReferences replaces any UsingColumn placeholders in the USING expression with GetField expressions bound
// to the given table schema, so that the expression can be evaluated against the table's rows.
func (a *AlterTableColumnTypeUsing) bindColumnReferences(ctx *sql.Context, sch sql.Schema) (sql.Expression, error) {
	boundExpr, _, err := transform.Expr(ctx, a.usingExpr, func(ctx *sql.Context, e sql.Expression) (sql.Expression, transform.TreeIdentity, error) {
		if usingCol, ok := e.(*UsingColumn); ok {
			idx := sch.IndexOfColName(usingCol.ColumnName)
			if idx < 0 {
				return nil, transform.NewTree, errors.Errorf(`column "%s" does not exist`, usingCol.ColumnName)
			}
			col := sch[idx]
			return expression.NewGetField(idx, col.Type, col.Name, col.Nullable), transform.NewTree, nil
		}
		return e, transform.SameTree, nil
	})
	if err != nil {
		return nil, err
	}
	return boundExpr, nil
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

// abortRewrite aborts a failed table rewrite, leaving the original table untouched: DiscardChanges records the failure
// on the inserter, which guarantees that the subsequent Close does not finalize (swap in) the partially rewritten
// table. Returns the original error, with any cleanup failures wrapped alongside it rather than swallowed.
func abortRewrite(ctx *sql.Context, inserter sql.RowInserter, cause error) error {
	if cause == nil {
		// This should be impossible, but DiscardChanges must always be given a non-nil error: it's what prevents
		// Close from finalizing the rewritten table.
		cause = errors.New("column type rewrite aborted")
	}
	discardErr := inserter.DiscardChanges(ctx, cause)
	closeErr := inserter.Close(ctx)
	// Close returns the error recorded by DiscardChanges when it aborts the rewrite, which isn't a distinct
	// cleanup failure, so we don't report it as one.
	if errors.Is(closeErr, cause) {
		closeErr = nil
	}
	if discardErr != nil {
		cause = errors.Wrapf(cause, "discarding the failed column type rewrite errored: %v; original error", discardErr)
	}
	if closeErr != nil {
		cause = errors.Wrapf(cause, "closing the failed column type rewrite errored: %v; original error", closeErr)
	}
	return cause
}

// UsingColumn is a placeholder expression for a column reference within the USING expression of an
// ALTER TABLE ... ALTER COLUMN ... TYPE ... USING statement. It resolves its type against the target table when the
// statement is built, so that operator and function overloads over the column are chosen correctly, and is replaced
// with a GetField expression bound to the table's schema when the statement executes.
type UsingColumn struct {
	SchemaName string
	TableName  string
	ColumnName string
	typ        sql.Type
}

var _ sql.Expression = (*UsingColumn)(nil)
var _ vitess.Injectable = (*UsingColumn)(nil)

// Resolved implements the interface sql.Expression.
func (u *UsingColumn) Resolved() bool {
	return true
}

// String implements the interface sql.Expression.
func (u *UsingColumn) String() string {
	return u.ColumnName
}

// Type implements the interface sql.Expression.
func (u *UsingColumn) Type(ctx *sql.Context) sql.Type {
	if u.typ == nil {
		return pgtypes.Unknown
	}
	return u.typ
}

// IsNullable implements the interface sql.Expression.
func (u *UsingColumn) IsNullable(ctx *sql.Context) bool {
	return true
}

// Eval implements the interface sql.Expression.
func (u *UsingColumn) Eval(ctx *sql.Context, row sql.Row) (interface{}, error) {
	return nil, errors.Errorf(`column reference "%s" was not bound to the table being altered`, u.ColumnName)
}

// Children implements the interface sql.Expression.
func (u *UsingColumn) Children() []sql.Expression {
	return nil
}

// WithChildren implements the interface sql.Expression.
func (u *UsingColumn) WithChildren(ctx *sql.Context, children ...sql.Expression) (sql.Expression, error) {
	if len(children) != 0 {
		return nil, sql.ErrInvalidChildrenNumber.New(u, len(children), 0)
	}
	return u, nil
}

// WithResolvedChildren implements the interface vitess.Injectable. It resolves the referenced column's type against
// the target table, if the table exists, honoring the search path for unqualified table names just like the
// resolveTableForDDL analyzer rule does when it resolves the statement's target table.
func (u *UsingColumn) WithResolvedChildren(ctx context.Context, children []any) (any, error) {
	if len(children) != 0 {
		return nil, ErrVitessChildCount.New(0, len(children))
	}
	sqlCtx, ok := ctx.(*sql.Context)
	if !ok {
		return u, nil
	}
	tbl, err := core.GetSqlTableFromContext(sqlCtx, "", doltdb.TableName{Name: u.TableName, Schema: u.SchemaName})
	if err != nil {
		return nil, err
	}
	if tbl == nil {
		// The table may legitimately not exist (e.g. ALTER TABLE IF EXISTS), which is handled during execution
		return u, nil
	}
	sch := tbl.Schema(sqlCtx)
	idx := sch.IndexOfColName(u.ColumnName)
	if idx < 0 {
		return nil, errors.Errorf(`column "%s" does not exist`, u.ColumnName)
	}
	nu := *u
	nu.typ = sch[idx].Type
	return &nu, nil
}
