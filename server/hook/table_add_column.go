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
	"github.com/dolthub/go-mysql-server/sql/plan"

	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/core/id"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// BeforeTableAddColumn handles validation that's unique to Doltgres.
func BeforeTableAddColumn(ctx *sql.Context, runner sql.StatementRunner, nodeInterface sql.Node) (sql.Node, error) {
	n, ok := nodeInterface.(*plan.AddColumn)
	if !ok {
		return nil, errors.Errorf("ADD COLUMN pre-hook expected `*plan.AddColumn` but received `%T`", nodeInterface)
	}
	// If the column being added doesn't have a default value, then we don't have anything to check (for now)
	if n.Column().Default == nil {
		return n, nil
	}

	// Grab the table being altered
	doltTable := core.SQLNodeToDoltTable(n.Table)
	if doltTable == nil {
		// If this table isn't a Dolt table then we don't have anything to do
		return n, nil
	}
	_, root, err := core.GetRootFromContext(ctx)
	if err != nil {
		return n, nil
	}
	tableName := doltTable.TableName()
	tableAsType := id.NewType(tableName.Schema, tableName.Name)
	allTableNames, err := root.GetAllTableNames(ctx, false)
	if err != nil {
		return nil, err
	}

	for _, otherTableName := range allTableNames {
		if doltdb.IsSystemTable(otherTableName) {
			// System tables don't use any table types
			continue
		}
		otherTable, ok, err := root.GetTable(ctx, otherTableName)
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, errors.Errorf("root returned table name `%s` but it could not be found?", otherTableName.String())
		}
		otherTableSch, err := otherTable.GetSchema(ctx)
		if err != nil {
			return nil, err
		}
		for _, otherCol := range otherTableSch.GetAllCols().GetColumns() {
			colType := otherCol.TypeInfo.ToSqlType()
			dgtype, ok := colType.(*pgtypes.DoltgresType)
			if !ok {
				// If this isn't a Doltgres type, then it can't be a table type so we can ignore it
				continue
			}
			if dgtype.ID != tableAsType {
				// This column isn't our table type, so we can ignore it
				continue
			}
			return nil, errors.Errorf(`cannot alter table "%s" because column "%s.%s" uses its row type`,
				tableName.Name, otherTableName.Name, otherCol.Name)
		}
	}
	return n, nil
}

// AfterTableAddColumn handles updating various table columns, alongside other validation that's unique to Doltgres.
func AfterTableAddColumn(ctx *sql.Context, runner sql.StatementRunner, nodeInterface sql.Node) error {
	n, ok := nodeInterface.(*plan.AddColumn)
	if !ok {
		return errors.Errorf("ADD COLUMN post-hook expected `*plan.AddColumn` but received `%T`", nodeInterface)
	}

	// Grab the table being altered
	doltTable := core.SQLNodeToDoltTable(n.Table)
	if doltTable == nil {
		// If this table isn't a Dolt table then we don't have anything to do
		return nil
	}
	_, root, err := core.GetRootFromContext(ctx)
	if err != nil {
		return err
	}
	tableName := doltTable.TableName()
	tableAsType := id.NewType(tableName.Schema, tableName.Name)
	allTableNames, err := root.GetAllTableNames(ctx, false)
	if err != nil {
		return err
	}
	sch := doltTable.Schema()

	for _, otherTableName := range allTableNames {
		if doltdb.IsSystemTable(otherTableName) {
			// System tables don't use any table types
			continue
		}
		otherTable, ok, err := root.GetTable(ctx, otherTableName)
		if err != nil {
			return err
		}
		if !ok {
			return errors.Errorf("root returned table name `%s` but it could not be found?", otherTableName.String())
		}
		otherTableSch, err := otherTable.GetSchema(ctx)
		if err != nil {
			return err
		}
		for _, otherCol := range otherTableSch.GetAllCols().GetColumns() {
			colType := otherCol.TypeInfo.ToSqlType()
			dgtype, ok := colType.(*pgtypes.DoltgresType)
			if !ok {
				// If this isn't a Doltgres type, then it can't be a table type so we can ignore it
				continue
			}
			if dgtype.ID != tableAsType {
				// This column isn't our table type, so we can ignore it
				continue
			}
			// Build the UPDATE statement that we'll run for this table
			rowValues := make([]string, len(sch)+1)
			for i, col := range sch {
				rowValues[i] = fmt.Sprintf(`("%s")."%s"`, otherCol.Name, col.Name)
			}
			rowValues[len(rowValues)-1] = "NULL"
			// The UPDATE changes the values in the table
			updateStr := fmt.Sprintf(`UPDATE "%s"."%s" SET "%s" = ROW(%s)::"%s"."%s" WHERE length("%s"::text) > 0;`,
				otherTableName.Schema, otherTableName.Name, otherCol.Name, strings.Join(rowValues, ","), tableName.Schema, tableName.Name, otherCol.Name)
			// The ALTER updates the type on the schema since it still has the old one
			alterStr := fmt.Sprintf(`ALTER TABLE "%s"."%s" ALTER COLUMN "%s" TYPE "%s"."%s";`,
				otherTableName.Schema, otherTableName.Name, otherCol.Name, tableName.Schema, tableName.Name)
			// We run the statements as though they were interpreted since we're running new statements inside the original
			_, err = sql.RunInterpreted(ctx, func(subCtx *sql.Context) ([]sql.Row, error) {
				_, rowIter, _, err := runner.QueryWithBindings(subCtx, updateStr, nil, nil, nil)
				if err != nil {
					return nil, err
				}
				_, err = sql.RowIterToRows(subCtx, rowIter)
				if err != nil {
					return nil, err
				}
				_, rowIter, _, err = runner.QueryWithBindings(subCtx, alterStr, nil, nil, nil)
				if err != nil {
					return nil, err
				}
				return sql.RowIterToRows(subCtx, rowIter)
			})
			if err != nil {
				return err
			}
		}
	}
	return nil
}
