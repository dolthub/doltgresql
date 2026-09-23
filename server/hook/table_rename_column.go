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
	"bytes"
	"fmt"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/plan"

	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/postgres/parser/lex"
	"github.com/dolthub/doltgresql/postgres/parser/parser"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// AfterTableRenameColumn handles updating various table columns, alongside other validation that's unique to Doltgres.
func AfterTableRenameColumn(ctx *sql.Context, runner sql.StatementRunner, nodeInterface sql.Node) error {
	n, ok := nodeInterface.(*plan.RenameColumn)
	if !ok {
		return errors.Errorf("RENAME COLUMN post-hook expected `*plan.RenameColumn` but received `%T`", nodeInterface)
	}
	if n.ColumnName == n.NewColumnName {
		return nil
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
			// The ALTER updates the type on the schema since it still has the old one
			alterStr := fmt.Sprintf(`ALTER TABLE "%s"."%s" ALTER COLUMN "%s" TYPE "%s"."%s";`,
				otherTableName.Schema, otherTableName.Name, otherCol.Name, tableName.Schema, tableName.Name)
			// We run the statement as though it were interpreted since we're running new statements inside the original
			_, err = sql.RunInterpreted(ctx, func(subCtx *sql.Context) ([]sql.Row, error) {
				_, rowIter, _, err := runner.QueryWithBindings(subCtx, alterStr, nil, nil, nil)
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
	trigColl, err := core.GetTriggersCollectionFromContext(ctx, "")
	if err != nil {
		return err
	}
	for _, trig := range trigColl.GetTriggersForTable(ctx, id.NewTable(tableName.Schema, tableName.Name)) {
		renamed := false
		for _, event := range trig.Events {
			for i, colName := range event.ColumnNames {
				if strings.EqualFold(colName, n.ColumnName) {
					event.ColumnNames[i] = n.NewColumnName
					renamed = true
				}
			}
		}
		if !renamed {
			continue
		}
		trig.Definition = renameTriggerDefinitionColumn(trig.Definition, n.ColumnName, n.NewColumnName)
		if err = trigColl.DropTrigger(ctx, trig.ID); err != nil {
			return err
		}
		if err = trigColl.AddTrigger(ctx, trig); err != nil {
			return err
		}
	}
	return nil
}

// renameTriggerDefinitionColumn returns the given CREATE TRIGGER definition with `oldName` replaced by `newName` in
// every UPDATE OF column list.
func renameTriggerDefinitionColumn(definition string, oldName string, newName string) string {
	tokens, _ := parser.Tokens(definition)
	var buf bytes.Buffer
	lastEnd := 0
	for i := 0; i+1 < len(tokens); i++ {
		if tokens[i].TokenID != lex.UPDATE || tokens[i+1].TokenID != lex.OF {
			continue
		}
		for i += 2; i < len(tokens); i += 2 {
			if strings.EqualFold(tokens[i].Str, oldName) {
				buf.WriteString(definition[lastEnd:tokens[i].Start])
				lex.EncodeRestrictedSQLIdent(&buf, newName, lex.EncNoFlags)
				lastEnd = tokens[i].End
			}
			if i+1 >= len(tokens) || tokens[i+1].TokenID != ',' {
				break
			}
		}
	}
	buf.WriteString(definition[lastEnd:])
	return buf.String()
}
