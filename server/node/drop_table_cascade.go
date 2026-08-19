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
	"fmt"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/resolve"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/plan"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/core"
)

// DropTableCascade handles the DROP TABLE ... CASCADE statement. It first drops any objects that the dependency
// tracking knows depend on the dropped tables (currently foreign key constraints on other tables that reference the
// dropped tables), then delegates the actual table drop to the standard DROP TABLE path. Sequences owned by the
// dropped tables' columns (e.g. SERIAL columns) are dropped by the standard path itself.
type DropTableCascade struct {
	// database is the database qualifier given in the statement, if any. Only the current database is supported.
	database string
	// tables are the (possibly schema-qualified) names of the tables to drop.
	tables   []doltdb.TableName
	ifExists bool
}

var _ sql.ExecSourceRel = (*DropTableCascade)(nil)
var _ vitess.Injectable = (*DropTableCascade)(nil)

// NewDropTableCascade returns a new *DropTableCascade.
func NewDropTableCascade(ifExists bool, database string, tables []doltdb.TableName) *DropTableCascade {
	return &DropTableCascade{
		database: database,
		tables:   tables,
		ifExists: ifExists,
	}
}

// Children implements the interface sql.ExecSourceRel.
func (c *DropTableCascade) Children() []sql.Node {
	return nil
}

// IsReadOnly implements the interface sql.ExecSourceRel.
func (c *DropTableCascade) IsReadOnly() bool {
	return false
}

// Resolved implements the interface sql.ExecSourceRel.
func (c *DropTableCascade) Resolved() bool {
	return true
}

// RowIter implements the interface sql.ExecSourceRel.
func (c *DropTableCascade) RowIter(ctx *sql.Context, r sql.Row) (sql.RowIter, error) {
	if len(c.database) > 0 && c.database != ctx.GetCurrentDatabase() {
		return nil, errors.Errorf("DROP TABLE CASCADE is currently only supported for the current database")
	}
	runner, err := core.GetRunnerFromContext(ctx)
	if err != nil {
		return nil, err
	}
	if runner == nil {
		return nil, errors.Errorf("DROP TABLE CASCADE requires a statement runner, but one was not found in the context")
	}
	_, root, err := core.GetRootFromContext(ctx)
	if err != nil {
		return nil, err
	}

	// Resolve each of the given table names. Names without an explicit schema are resolved against the search path,
	// matching the behavior of the regular DROP TABLE path.
	var dropTables []doltdb.TableName
	for _, tblName := range c.tables {
		resolvedName, found, err := c.resolveTable(ctx, root, tblName)
		if err != nil {
			return nil, err
		}
		if !found {
			if c.ifExists {
				// TODO: issue a notice that the table is being skipped
				continue
			}
			return nil, errors.Errorf(`table "%s" does not exist`, tblName.Name)
		}
		dropTables = append(dropTables, resolvedName)
	}
	if len(dropTables) == 0 {
		return sql.RowsToRowIter(), nil
	}
	inDropSet := func(name doltdb.TableName) bool {
		for _, dropped := range dropTables {
			if dropped.EqualFold(name) {
				return true
			}
		}
		return false
	}

	// Drop the foreign keys on other tables that reference the tables being dropped. These constraints depend on the
	// dropped tables, so CASCADE removes them. Foreign keys declared by tables that are themselves being dropped
	// (including self-referential keys) are removed along with their tables by the standard DROP TABLE path. This uses
	// the same interface calls that the engine's DROP TABLE execution uses for a dropped table's declared keys.
	fkc, err := root.GetForeignKeyCollection(ctx)
	if err != nil {
		return nil, err
	}
	for _, tblName := range dropTables {
		_, referencedByFk := fkc.KeysForTable(tblName)
		if len(referencedByFk) == 0 {
			continue
		}
		sqlTable, err := core.GetSqlTableFromContext(ctx, "", tblName)
		if err != nil {
			return nil, err
		}
		if sqlTable == nil {
			return nil, errors.Errorf(`table "%s" was resolved but could not be found`, tblName.Name)
		}
		fkTable, ok := sqlTable.(sql.ForeignKeyTable)
		if !ok {
			continue
		}
		for _, fk := range referencedByFk {
			if inDropSet(fk.TableName) {
				continue
			}
			// TODO: issue a notice that the constraint is being dropped ("drop cascades to constraint ... on table ...")
			if err = fkTable.DropForeignKey(ctx, fk.Name, fk.TableName.Name, fk.TableName.Schema); err != nil {
				return nil, err
			}
		}
	}

	// Drop the tables themselves by running the equivalent DROP TABLE statement. This reuses the standard path,
	// including its own dependency bookkeeping (such as dropping sequences owned by the tables' columns).
	_, err = sql.RunInterpreted(ctx, func(subCtx *sql.Context) (struct{}, error) {
		quotedNames := make([]string, len(dropTables))
		for i, tblName := range dropTables {
			quotedNames[i] = quotedQualifiedName(tblName)
		}
		dropTable := fmt.Sprintf(`DROP TABLE %s;`, strings.Join(quotedNames, ", "))
		return struct{}{}, runStatement(subCtx, runner, dropTable)
	})
	if err != nil {
		return nil, err
	}
	return sql.RowsToRowIter(), nil
}

// resolveTable resolves the given table name to a schema-qualified name, returning whether the table was found. Names
// without an explicit schema are resolved against the search path. Temporary tables (which do not live in the root)
// are returned as-is.
func (c *DropTableCascade) resolveTable(ctx *sql.Context, root *core.RootValue, tblName doltdb.TableName) (doltdb.TableName, bool, error) {
	if len(tblName.Schema) == 0 {
		resolvedName, found, err := resolve.TableName(ctx, root, tblName.Name)
		if err != nil {
			return doltdb.TableName{}, false, err
		}
		if found {
			return resolvedName, true, nil
		}
	} else {
		found, err := root.HasTable(ctx, tblName)
		if err != nil {
			return doltdb.TableName{}, false, err
		}
		if found {
			return tblName, true, nil
		}
	}
	// The table may be a temporary table, which does not live in the root. Temporary tables cannot be referenced by
	// foreign keys on permanent tables, so there are no dependent constraints to find for them.
	relationType, err := core.GetRelationType(ctx, tblName.Schema, tblName.Name)
	if err != nil {
		return doltdb.TableName{}, false, err
	}
	if relationType == core.RelationType_Table {
		return tblName, true, nil
	}
	return doltdb.TableName{}, false, nil
}

// Schema implements the interface sql.ExecSourceRel.
func (c *DropTableCascade) Schema(ctx *sql.Context) sql.Schema {
	return nil
}

// String implements the interface sql.ExecSourceRel.
func (c *DropTableCascade) String() string {
	return "DROP TABLE CASCADE"
}

// WithChildren implements the interface sql.ExecSourceRel.
func (c *DropTableCascade) WithChildren(ctx *sql.Context, children ...sql.Node) (sql.Node, error) {
	return plan.NillaryWithChildren(c, children...)
}

// WithResolvedChildren implements the interface vitess.Injectable.
func (c *DropTableCascade) WithResolvedChildren(ctx context.Context, children []any) (any, error) {
	if len(children) != 0 {
		return nil, ErrVitessChildCount.New(0, len(children))
	}
	return c, nil
}

// runStatement runs the given statement on the given runner, draining and discarding any returned rows.
func runStatement(ctx *sql.Context, runner sql.StatementRunner, statement string) error {
	_, rowIter, _, err := runner.QueryWithBindings(ctx, statement, nil, nil, nil)
	if err != nil {
		return err
	}
	_, err = sql.RowIterToRows(ctx, rowIter)
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
