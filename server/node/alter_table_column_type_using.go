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
	usingExpr  sql.Expression
	schemaName string
	tableName  string
	columnName string
	ifExists   bool
	// table is the resolved target table, set by the ResolveType analyzer rule via ResolveTable. It is nil when the
	// table does not exist, which resolution only tolerates when ifExists is set.
	table         sql.Table
	tableResolved bool
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
		schemaName: schemaName,
		tableName:  tableName,
		columnName: columnName,
		ifExists:   ifExists,
	}
}

// Children implements sql.ExecSourceRel.
func (a *AlterTableColumnTypeUsing) Children() []sql.Node { return nil }

// IsReadOnly implements sql.ExecSourceRel.
func (a *AlterTableColumnTypeUsing) IsReadOnly() bool { return false }

// Resolved implements sql.ExecSourceRel.
func (a *AlterTableColumnTypeUsing) Resolved() bool {
	return a.DbProvider != nil && a.usingExpr != nil && a.usingExpr.Resolved() && a.tableResolved
}

// TableResolved returns whether the target table has been resolved by the analyzer.
func (a *AlterTableColumnTypeUsing) TableResolved() bool {
	return a.tableResolved
}

// ResolveTable resolves the target table of the statement, returning a copy of the node with the resolved table
// stored on it. The search path is used for unqualified table names. A missing table is an error unless the statement
// specified IF EXISTS, in which case the node executes as a no-op.
func (a *AlterTableColumnTypeUsing) ResolveTable(ctx *sql.Context) (*AlterTableColumnTypeUsing, error) {
	tbl, err := resolveUsingTable(ctx, a.schemaName, a.tableName)
	if err != nil {
		return nil, err
	}
	if tbl == nil && !a.ifExists {
		return nil, sql.ErrTableNotFound.New(a.tableName)
	}
	na := *a
	na.table = tbl
	na.tableResolved = true
	return &na, nil
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

	if !a.tableResolved {
		return nil, errors.Errorf("target table for ALTER TABLE ... ALTER COLUMN ... TYPE ... USING was not resolved during analysis")
	}
	if a.table == nil {
		// The table does not exist; ResolveTable only permits this with IF EXISTS, making this statement a no-op
		if a.ifExists {
			return sql.RowsToRowIter(), nil
		}
		return nil, sql.ErrTableNotFound.New(a.tableName)
	}

	rwt, ok := a.table.(sql.RewritableTable)
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
	if err := a.checkForeignKeyUsage(ctx, a.table, oldCol.Name); err != nil {
		return nil, err
	}

	// Reject the type change if this table's implicit row type is used as a column type anywhere else.
	if doltTable := core.SQLTableToDoltTable(a.table); doltTable != nil {
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

// resolveUsingTable resolves the target table of an ALTER TABLE ... ALTER COLUMN ... TYPE ... USING statement,
// honoring the search path for unqualified table names. Returns nil (without an error) if the table does not exist.
// This is the single resolution mechanism shared by the node's analyzer-rule resolution (ResolveTable) and the
// UsingColumn placeholder's build-time column type resolution, so the two can never disagree about the target table.
func resolveUsingTable(ctx *sql.Context, schemaName string, tableName string) (sql.Table, error) {
	return core.GetSqlTableFromContext(ctx, "", doltdb.TableName{Name: tableName, Schema: schemaName})
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
// the target table, if the table exists.
func (u *UsingColumn) WithResolvedChildren(ctx context.Context, children []any) (any, error) {
	if len(children) != 0 {
		return nil, ErrVitessChildCount.New(0, len(children))
	}
	sqlCtx, ok := ctx.(*sql.Context)
	if !ok {
		return u, nil
	}
	tbl, err := resolveUsingTable(sqlCtx, u.SchemaName, u.TableName)
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
