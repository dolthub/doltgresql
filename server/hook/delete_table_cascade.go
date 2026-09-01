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

package hook

import (
	"fmt"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/core/functions"
	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/core/procedures"
	"github.com/dolthub/doltgresql/postgres/parser/parser"
	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
	"github.com/dolthub/doltgresql/server/settings"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// cascadeDropDependencies drops the objects that depend on the tables being dropped, implementing the CASCADE drop
// behavior. Views whose definitions reference the dropped tables (directly, or through other dependent views) are
// dropped, as are foreign keys on other tables that reference the dropped tables, functions and procedures with a
// parameter of a dropped table's row type, and columns of other tables whose type is a dropped table's row type.
// Sequences owned by the dropped tables' columns (e.g. SERIAL columns) are dropped by the standard DROP TABLE path
// itself. Foreign keys declared by tables that are themselves being dropped (including self-referential keys) are
// also removed by the standard path.
func cascadeDropDependencies(ctx *sql.Context, runner sql.StatementRunner, allDeletedTables []doltdb.TableName) error {
	if err := cascadeDropViews(ctx, runner, allDeletedTables); err != nil {
		return err
	}
	if err := cascadeDropForeignKeys(ctx, allDeletedTables); err != nil {
		return err
	}
	if err := cascadeDropRoutines(ctx, allDeletedTables); err != nil {
		return err
	}
	return cascadeDropDependentColumns(ctx, runner, allDeletedTables)
}

// relationKey identifies a relation (table or view) by its lowercased schema and name.
type relationKey struct {
	schema string
	name   string
}

// newRelationKey returns the relationKey for the given schema and name.
func newRelationKey(schema string, name string) relationKey {
	return relationKey{schema: strings.ToLower(schema), name: strings.ToLower(name)}
}

// cascadeView is a view in the current database, along with the table names its definition references.
type cascadeView struct {
	schema    string
	name      string
	refs      []*tree.TableName
	dependent bool
}

// cascadeDropViews drops the views whose definitions reference the tables being dropped. A view depends on a dropped
// table if its definition references the table directly, or references another view that does. Dependencies are
// determined by parsing each view's definition and resolving the referenced names the same way the engine would when
// the view is queried: explicitly qualified names match as given, and unqualified names resolve against the search
// path. Column-level dependencies are not tracked, so a view is only dropped when a dropped relation appears in its
// definition by name.
func cascadeDropViews(ctx *sql.Context, runner sql.StatementRunner, allDeletedTables []doltdb.TableName) error {
	views, viewExists, err := loadDatabaseViews(ctx)
	if err != nil || len(views) == 0 {
		return err
	}
	_, root, err := core.GetRootFromContext(ctx)
	if err != nil {
		return err
	}
	searchPath, err := settings.GetCurrentSchemas(ctx)
	if err != nil {
		return err
	}

	// Compute the closure of dependent relations: dropping a view may make further views dependent, so iterate until
	// no new views are added.
	closure := make(map[relationKey]struct{}, len(allDeletedTables))
	for _, tblName := range allDeletedTables {
		closure[newRelationKey(tblName.Schema, tblName.Name)] = struct{}{}
	}
	for changed := true; changed; {
		changed = false
		for _, view := range views {
			if view.dependent {
				continue
			}
			dependent, err := viewDependsOnClosure(ctx, root, view, closure, searchPath, viewExists)
			if err != nil {
				return err
			}
			if dependent {
				view.dependent = true
				closure[newRelationKey(view.schema, view.name)] = struct{}{}
				changed = true
			}
		}
	}

	for _, view := range views {
		if !view.dependent {
			continue
		}
		// TODO: issue a notice that the view is being dropped ("drop cascades to view ...")
		dropStmt := fmt.Sprintf(`DROP VIEW %s.%s;`, quoteIdentifier(view.schema), quoteIdentifier(view.name))
		if err = runStatement(ctx, runner, dropStmt); err != nil {
			return err
		}
	}
	return nil
}

// loadDatabaseViews returns all views in the current database with their parsed table references, along with a set of
// the views' relation keys for name resolution.
func loadDatabaseViews(ctx *sql.Context) ([]*cascadeView, map[relationKey]struct{}, error) {
	db, err := core.GetSqlDatabaseFromContext(ctx, "")
	if err != nil {
		return nil, nil, err
	}
	schemaDb, ok := db.(sql.SchemaDatabase)
	if !ok {
		return nil, nil, nil
	}
	schemas, err := schemaDb.AllSchemas(ctx)
	if err != nil {
		return nil, nil, err
	}
	var views []*cascadeView
	viewExists := make(map[relationKey]struct{})
	for _, schema := range schemas {
		viewDb, ok := schema.(sql.ViewDatabase)
		if !ok {
			continue
		}
		defs, err := viewDb.AllViews(ctx)
		if err != nil {
			return nil, nil, err
		}
		for _, def := range defs {
			stmts, err := parser.Parse(def.CreateViewStatement)
			if err != nil || len(stmts) == 0 {
				return nil, nil, errors.Newf("could not parse the definition of view %s.%s: %v",
					schema.SchemaName(), def.Name, err)
			}
			createView, ok := stmts[0].AST.(*tree.CreateView)
			if !ok {
				// Definitions that aren't plain CREATE VIEW statements (e.g. materialized views) don't have their
				// dependencies tracked yet, so they are left in place.
				continue
			}
			collector := newTableRefCollector()
			collector.collectSelect(createView.AsSource)
			views = append(views, &cascadeView{
				schema: schema.SchemaName(),
				name:   def.Name,
				refs:   collector.refs,
			})
			viewExists[newRelationKey(schema.SchemaName(), def.Name)] = struct{}{}
		}
	}
	return views, viewExists, nil
}

// viewDependsOnClosure returns whether any of the given view's referenced names resolve to a relation in the given
// closure. Unqualified names resolve to the first schema on the search path containing a relation with that name,
// mirroring how the engine resolves the name when the view is queried.
func viewDependsOnClosure(ctx *sql.Context, root *core.RootValue, view *cascadeView, closure map[relationKey]struct{},
	searchPath []string, viewExists map[relationKey]struct{}) (bool, error) {
	for _, ref := range view.refs {
		if ref.ExplicitSchema {
			if _, ok := closure[newRelationKey(ref.Schema(), ref.Table())]; ok {
				return true, nil
			}
			continue
		}
		for _, schema := range searchPath {
			key := newRelationKey(schema, ref.Table())
			if _, ok := closure[key]; ok {
				return true, nil
			}
			// If the name resolves to a relation that isn't being dropped, then it shadows any same-named relation
			// later on the search path.
			if _, ok := viewExists[key]; ok {
				break
			}
			hasTable, err := root.HasTable(ctx, doltdb.TableName{Schema: schema, Name: ref.Table()})
			if err != nil {
				return false, err
			}
			if hasTable {
				break
			}
		}
	}
	return false, nil
}

// cascadeDropForeignKeys drops the foreign keys on other tables that reference the tables being dropped. These
// constraints depend on the dropped tables, so CASCADE removes them. This uses the same interface calls that the
// engine's DROP TABLE execution uses for a dropped table's declared keys.
func cascadeDropForeignKeys(ctx *sql.Context, allDeletedTables []doltdb.TableName) error {
	_, root, err := core.GetRootFromContext(ctx)
	if err != nil {
		return err
	}
	fkc, err := root.GetForeignKeyCollection(ctx)
	if err != nil {
		return err
	}
	for _, tblName := range allDeletedTables {
		_, referencedByFk := fkc.KeysForTable(tblName)
		if len(referencedByFk) == 0 {
			continue
		}
		sqlTable, err := core.GetSqlTableFromContext(ctx, "", tblName)
		if err != nil {
			return err
		}
		if sqlTable == nil {
			return errors.Newf(`table "%s" was resolved but could not be found`, tblName.Name)
		}
		fkTable, ok := sqlTable.(sql.ForeignKeyTable)
		if !ok {
			continue
		}
		for _, fk := range referencedByFk {
			if tableNameInSet(fk.TableName, allDeletedTables) {
				continue
			}
			// TODO: issue a notice that the constraint is being dropped ("drop cascades to constraint ... on table ...")
			if err = fkTable.DropForeignKey(ctx, fk.Name, fk.TableName.Name, fk.TableName.Schema); err != nil {
				return err
			}
		}
	}
	return nil
}

// cascadeDropRoutines drops the functions and procedures that have a parameter typed with a dropped table's row type.
func cascadeDropRoutines(ctx *sql.Context, allDeletedTables []doltdb.TableName) error {
	deletedTypes := deletedTableTypes(allDeletedTables)
	funcsColl, err := core.GetFunctionsCollectionFromContext(ctx, "")
	if err != nil {
		return err
	}
	var funcIDs []id.Function
	err = funcsColl.IterateFunctions(ctx, func(f functions.Function) (stop bool, err error) {
		for _, param := range f.AllParams {
			if _, ok := deletedTypes[param.Type]; ok {
				funcIDs = append(funcIDs, f.ID)
				break
			}
		}
		return false, nil
	})
	if err != nil {
		return err
	}
	// TODO: issue a notice for each dropped routine ("drop cascades to function ...")
	if err = funcsColl.DropFunction(ctx, funcIDs...); err != nil {
		return err
	}
	procsColl, err := core.GetProceduresCollectionFromContext(ctx, "")
	if err != nil {
		return err
	}
	var procIDs []id.Procedure
	err = procsColl.IterateProcedures(ctx, func(p procedures.Procedure) (stop bool, err error) {
		for _, param := range p.AllParams {
			if _, ok := deletedTypes[param.Type]; ok {
				procIDs = append(procIDs, p.ID)
				break
			}
		}
		return false, nil
	})
	if err != nil {
		return err
	}
	return procsColl.DropProcedure(ctx, procIDs...)
}

// cascadeDropDependentColumns drops the columns of other tables whose type is a dropped table's row type. This matches
// Postgres, which drops just the dependent column rather than the whole table.
func cascadeDropDependentColumns(ctx *sql.Context, runner sql.StatementRunner, allDeletedTables []doltdb.TableName) error {
	_, root, err := core.GetRootFromContext(ctx)
	if err != nil {
		return err
	}
	deletedTypes := deletedTableTypes(allDeletedTables)
	allTableNames, err := root.GetAllTableNames(ctx, false)
	if err != nil {
		return err
	}
	type dependentColumn struct {
		table  doltdb.TableName
		column string
	}
	// Collect all the dependent columns before dropping any, so that the scan works from a consistent root.
	var dependentColumns []dependentColumn
	for _, otherTableName := range allTableNames {
		if doltdb.IsSystemTable(otherTableName) {
			// System tables don't use any table types
			continue
		}
		if tableNameInSet(otherTableName, allDeletedTables) {
			// If we're also deleting this table, then it doesn't matter what the columns have
			continue
		}
		otherTable, ok, err := root.GetTable(ctx, otherTableName)
		if err != nil {
			return err
		}
		if !ok {
			return errors.Newf("root returned table name `%s` but it could not be found?", otherTableName.String())
		}
		otherTableSch, err := otherTable.GetSchema(ctx)
		if err != nil {
			return err
		}
		for _, col := range otherTableSch.GetAllCols().GetColumns() {
			dgtype, ok := col.TypeInfo.ToSqlType().(*pgtypes.DoltgresType)
			if !ok {
				// If this isn't a Doltgres type, then it can't be a table type so we can ignore it
				continue
			}
			if _, ok = deletedTypes[dgtype.ID]; ok {
				dependentColumns = append(dependentColumns, dependentColumn{table: otherTableName, column: col.Name})
			}
		}
	}
	for _, depCol := range dependentColumns {
		// TODO: issue a notice that the column is being dropped ("drop cascades to column ... of table ...")
		alterStmt := fmt.Sprintf(`ALTER TABLE %s DROP COLUMN %s;`,
			quotedQualifiedName(depCol.table), quoteIdentifier(depCol.column))
		if err = runStatement(ctx, runner, alterStmt); err != nil {
			return err
		}
	}
	return nil
}

// deletedTableTypes returns the set of row types belonging to the given tables.
func deletedTableTypes(allDeletedTables []doltdb.TableName) map[id.Type]struct{} {
	deletedTypes := make(map[id.Type]struct{}, len(allDeletedTables))
	for _, tblName := range allDeletedTables {
		deletedTypes[id.NewType(tblName.Schema, tblName.Name)] = struct{}{}
	}
	return deletedTypes
}

// tableNameInSet returns whether the given table name is in the given set of table names, ignoring case.
func tableNameInSet(name doltdb.TableName, set []doltdb.TableName) bool {
	for _, candidate := range set {
		if candidate.EqualFold(name) {
			return true
		}
	}
	return false
}

// runStatement runs the given statement on the given runner, draining and discarding any returned rows. The statement
// runs as though it were interpreted, since it's a new statement running inside the original one.
func runStatement(ctx *sql.Context, runner sql.StatementRunner, statement string) error {
	_, err := sql.RunInterpreted(ctx, func(subCtx *sql.Context) (struct{}, error) {
		_, rowIter, _, err := runner.QueryWithBindings(subCtx, statement, nil, nil, nil)
		if err != nil {
			return struct{}{}, err
		}
		_, err = sql.RowIterToRows(subCtx, rowIter)
		return struct{}{}, err
	})
	return err
}

// quoteIdentifier returns the given identifier in its quoted form.
func quoteIdentifier(name string) string {
	return `"` + strings.ReplaceAll(name, `"`, `""`) + `"`
}

// quotedQualifiedName returns the given table name in its quoted, schema-qualified form. Names without a schema are
// returned unqualified.
func quotedQualifiedName(name doltdb.TableName) string {
	if len(name.Schema) == 0 {
		return quoteIdentifier(name.Name)
	}
	return quoteIdentifier(name.Schema) + "." + quoteIdentifier(name.Name)
}

// tableRefCollector collects the table names referenced by a view definition. Common table expression names are
// tracked so that references to them are not mistaken for table references.
type tableRefCollector struct {
	refs     []*tree.TableName
	cteNames map[string]struct{}
}

var _ tree.Visitor = (*tableRefCollector)(nil)

// newTableRefCollector returns a new *tableRefCollector.
func newTableRefCollector() *tableRefCollector {
	return &tableRefCollector{cteNames: make(map[string]struct{})}
}

// VisitPre implements the interface tree.Visitor. It recurses into subqueries appearing in expression position.
func (c *tableRefCollector) VisitPre(expr tree.Expr) (recurse bool, newExpr tree.Expr) {
	if subquery, ok := expr.(*tree.Subquery); ok {
		c.collectSelectStatement(subquery.Select)
	}
	return true, expr
}

// VisitPost implements the interface tree.Visitor.
func (c *tableRefCollector) VisitPost(expr tree.Expr) tree.Expr {
	return expr
}

// collectExpr collects the table references from any subqueries within the given expression.
func (c *tableRefCollector) collectExpr(expr tree.Expr) {
	if expr != nil {
		tree.WalkExpr(c, expr)
	}
}

// collectSelect collects the table references from the given SELECT statement, including its CTEs, ORDER BY, and
// LIMIT clauses.
func (c *tableRefCollector) collectSelect(sel *tree.Select) {
	if sel == nil {
		return
	}
	if sel.With != nil {
		for _, cte := range sel.With.CTEList {
			c.collectStatement(cte.Stmt)
			c.cteNames[strings.ToLower(string(cte.Name.Alias))] = struct{}{}
		}
	}
	c.collectSelectStatement(sel.Select)
	for _, order := range sel.OrderBy {
		c.collectExpr(order.Expr)
	}
	if sel.Limit != nil {
		c.collectExpr(sel.Limit.Count)
		c.collectExpr(sel.Limit.Offset)
	}
}

// collectSelectStatement collects the table references from the given select statement variant.
func (c *tableRefCollector) collectSelectStatement(stmt tree.SelectStatement) {
	switch stmt := stmt.(type) {
	case *tree.ParenSelect:
		c.collectSelect(stmt.Select)
	case *tree.SelectClause:
		for _, tableExpr := range stmt.From.Tables {
			c.collectTableExpr(tableExpr)
		}
		for i := range stmt.Exprs {
			c.collectExpr(stmt.Exprs[i].Expr)
		}
		if stmt.Where != nil {
			c.collectExpr(stmt.Where.Expr)
		}
		for _, expr := range stmt.GroupBy {
			c.collectExpr(expr)
		}
		if stmt.Having != nil {
			c.collectExpr(stmt.Having.Expr)
		}
		for _, expr := range stmt.DistinctOn {
			c.collectExpr(expr)
		}
	case *tree.UnionClause:
		c.collectSelect(stmt.Left)
		c.collectSelect(stmt.Right)
	case *tree.ValuesClause:
		for _, row := range stmt.Rows {
			for _, expr := range row {
				c.collectExpr(expr)
			}
		}
	}
}

// collectStatement collects the table references from the given statement, which may be any statement form that can
// appear in a CTE.
func (c *tableRefCollector) collectStatement(stmt tree.Statement) {
	switch stmt := stmt.(type) {
	case *tree.Select:
		c.collectSelect(stmt)
	case tree.SelectStatement:
		c.collectSelectStatement(stmt)
	}
}

// collectTableExpr collects the table references from the given table expression.
func (c *tableRefCollector) collectTableExpr(expr tree.TableExpr) {
	switch expr := expr.(type) {
	case *tree.AliasedTableExpr:
		c.collectTableExpr(expr.Expr)
	case *tree.ParenTableExpr:
		c.collectTableExpr(expr.Expr)
	case *tree.JoinTableExpr:
		c.collectTableExpr(expr.Left)
		c.collectTableExpr(expr.Right)
		if cond, ok := expr.Cond.(*tree.OnJoinCond); ok {
			c.collectExpr(cond.Expr)
		}
	case *tree.TableName:
		if !expr.ExplicitSchema {
			if _, ok := c.cteNames[strings.ToLower(expr.Table())]; ok {
				return
			}
		}
		c.refs = append(c.refs, expr)
	case *tree.Subquery:
		c.collectSelectStatement(expr.Select)
	case *tree.StatementSource:
		c.collectStatement(expr.Statement)
	case *tree.RowsFromExpr:
		for _, item := range expr.Items {
			c.collectExpr(item)
		}
	}
}
