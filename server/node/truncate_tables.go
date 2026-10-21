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

	"github.com/cockroachdb/errors"
	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/plan"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/core/triggers"
)

// TruncateTables validates a complete set of tables before truncating any of them.
// Foreign keys between tables in the set do not prevent truncation.
type TruncateTables struct {
	Tables []vitess.TableName
}

var _ sql.ExecSourceRel = (*TruncateTables)(nil)
var _ vitess.Injectable = (*TruncateTables)(nil)

func (t *TruncateTables) Children() []sql.Node { return nil }

func (t *TruncateTables) IsReadOnly() bool { return false }

func (t *TruncateTables) Resolved() bool { return true }

func (t *TruncateTables) Schema(ctx *sql.Context) sql.Schema { return nil }

func (t *TruncateTables) String() string { return "TRUNCATE TABLE" }

func (t *TruncateTables) WithChildren(ctx *sql.Context, children ...sql.Node) (sql.Node, error) {
	return plan.NillaryWithChildren(t, children...)
}

func (t *TruncateTables) WithResolvedChildren(ctx context.Context, children []any) (any, error) {
	if len(children) != 0 {
		return nil, ErrVitessChildCount.New(0, len(children))
	}
	return t, nil
}

func (t *TruncateTables) RowIter(ctx *sql.Context, row sql.Row) (_ sql.RowIter, err error) {
	tables := make([]sql.TruncateableTable, 0, len(t.Tables))
	selected := make(map[doltdb.TableName]struct{}, len(t.Tables))
	triggerCollection, err := core.GetTriggersCollectionFromContext(ctx, "")
	if err != nil {
		return nil, err
	}
	for _, target := range t.Tables {
		// Preserve database qualifiers, including the information_schema mapping,
		// so a qualified target can never resolve to a same-named local table.
		if database := target.DbQualifier.String(); database != "" && database != ctx.GetCurrentDatabase() {
			return nil, errors.New("cross-database multi-table TRUNCATE is not yet supported")
		}
		name := doltdb.TableName{Schema: target.SchemaQualifier.String(), Name: target.Name.String()}
		table, err := core.GetSqlTableFromContext(ctx, ctx.GetCurrentDatabase(), name)
		if err != nil {
			return nil, err
		}
		if table == nil {
			return nil, sql.ErrTableNotFound.New(name.Name)
		}
		doltTable := core.SQLTableToDoltTable(table)
		if doltTable == nil {
			return nil, errors.Errorf("TRUNCATE is not supported for table %s", name.Name)
		}
		name = doltTable.TableName()
		if _, exists := selected[name]; exists {
			continue
		}
		truncatable, ok := table.(sql.TruncateableTable)
		if !ok {
			return nil, errors.Errorf("TRUNCATE is not supported for table %s", name.Name)
		}
		// Until grouped trigger execution is implemented, reject rather than silently
		// skip statement triggers or fire AFTER triggers before other tables are empty.
		for _, trigger := range triggerCollection.GetTriggersForTable(ctx, id.NewTable(name.Schema, name.Name)) {
			for _, event := range trigger.Events {
				if event.Type == triggers.TriggerEventType_Truncate {
					return nil, errors.New("multi-table TRUNCATE with TRUNCATE triggers is not yet supported")
				}
			}
		}
		selected[name] = struct{}{}
		tables = append(tables, truncatable)
	}
	session, root, err := core.GetRootFromContext(ctx)
	if err != nil {
		return nil, err
	}
	fks, err := root.GetForeignKeyCollection(ctx)
	if err != nil {
		return nil, err
	}
	for name := range selected {
		_, referenced := fks.KeysForTable(name)
		for _, fk := range referenced {
			if _, included := selected[fk.TableName]; !included {
				return nil, sql.ErrTruncateReferencedFromForeignKey.New(name.Name, fk.Name, fk.TableName.Name)
			}
		}
	}
	// Preserve the original working set if any table fails during execution.
	roots, ok := session.GetRoots(ctx, ctx.GetCurrentDatabase())
	if !ok {
		return nil, sql.ErrDatabaseNotFound.New(ctx.GetCurrentDatabase())
	}
	complete := false
	defer func() {
		if !complete {
			if restoreErr := session.SetRoots(ctx, ctx.GetCurrentDatabase(), roots); restoreErr != nil {
				err = errors.CombineErrors(err, restoreErr)
			}
		}
	}()
	for _, table := range tables {
		if _, err = table.Truncate(ctx); err != nil {
			return nil, err
		}
	}
	complete = true
	return sql.RowsToRowIter(), nil
}
