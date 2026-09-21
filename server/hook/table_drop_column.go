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

// BeforeTableDropColumn drops the foreign keys that use the column being dropped. Foreign keys declared on the column
// are dropped with it, while foreign keys referencing it from other tables require CASCADE.
func BeforeTableDropColumn(ctx *sql.Context, runner sql.StatementRunner, nodeInterface sql.Node) (sql.Node, error) {
	n, ok := nodeInterface.(*plan.DropColumn)
	if !ok {
		return nil, errors.Errorf("DROP COLUMN pre-hook expected `*plan.DropColumn` but received `%T`", nodeInterface)
	}
	doltTable := core.SQLNodeToDoltTable(n.Table)
	if doltTable == nil {
		return n, nil
	}
	tableName := doltTable.TableName()
	sqlTable, err := core.GetSqlTableFromContext(ctx, "", tableName)
	if err != nil {
		return nil, err
	}
	fkTable, ok := sqlTable.(sql.ForeignKeyTable)
	if !ok {
		return n, nil
	}
	referenced, err := fkTable.GetReferencedForeignKeys(ctx)
	if err != nil {
		return nil, err
	}
	for _, fk := range referenced {
		if !foreignKeyUsesColumn(fk.ParentColumns, n.Column) {
			continue
		}
		if !n.Cascade {
			// TODO: portion after newline should be in DETAILS but we don't yet support that in our error messages
			return nil, errors.Errorf("cannot drop column %s of table %s because other objects depend on it\nconstraint %s on table %s depends on column %s of table %s",
				n.Column, tableName.Name, fk.Name, fk.Table, n.Column, tableName.Name)
		}
		if err = fkTable.DropForeignKey(ctx, fk.Name, fk.Table, fk.SchemaName); err != nil {
			return nil, err
		}
	}
	declared, err := fkTable.GetDeclaredForeignKeys(ctx)
	if err != nil {
		return nil, err
	}
	for _, fk := range declared {
		if !foreignKeyUsesColumn(fk.Columns, n.Column) {
			continue
		}
		if err = fkTable.DropForeignKey(ctx, fk.Name, fk.Table, fk.SchemaName); err != nil {
			return nil, err
		}
	}
	return n, nil
}

// foreignKeyUsesColumn returns whether the given foreign key columns include the named column.
func foreignKeyUsesColumn(fkColumns []string, column string) bool {
	for _, fkColumn := range fkColumns {
		if strings.EqualFold(fkColumn, column) {
			return true
		}
	}
	return false
}

// AfterTableDropColumn handles updating various table columns, alongside other validation that's unique to Doltgres.
func AfterTableDropColumn(ctx *sql.Context, runner sql.StatementRunner, nodeInterface sql.Node) error {
	n, ok := nodeInterface.(*plan.DropColumn)
	if !ok {
		return errors.Errorf("DROP COLUMN post-hook expected `*plan.DropColumn` but received `%T`", nodeInterface)
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
	sch := n.TargetSchema()

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
			trimIdx := -1
			for i, col := range sch {
				if col.Name == n.Column {
					trimIdx = i
					break
				}
			}
			if trimIdx == -1 {
				return errors.New("DROP COLUMN post-hook could not find the index of the column to remove")
			}
			// The UPDATE changes the values in the table
			updateStr := fmt.Sprintf(`UPDATE "%s"."%s" SET "%s" = dolt_recordtrim("%s", %d)::"%s"."%s";`,
				otherTableName.Schema, otherTableName.Name, otherCol.Name, otherCol.Name, trimIdx, tableName.Schema, tableName.Name)
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
