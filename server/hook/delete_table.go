// Copyright 2025 Dolthub, Inc.
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
	"github.com/cockroachdb/errors"
	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/dolthub/dolt/go/libraries/doltcore/sqle"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/plan"

	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/core/functions"
	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/core/procedures"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// BeforeTableDeletion performs all validation necessary to ensure that table deletion does not leave the database in an
// invalid state.
func BeforeTableDeletion(ctx *sql.Context, runner sql.StatementRunner, nodeInterface sql.Node) (sql.Node, error) {
	n, ok := nodeInterface.(*plan.DropTable)
	if !ok {
		return nil, errors.Newf("DROP TABLE pre-hook expected `*plan.DropTable` but received `%T`", nodeInterface)
	}
	var resolvedTables []*sqle.DoltTable
	var allTableNames []doltdb.TableName
	for _, tbl := range n.Tables {
		doltTable := core.SQLNodeToDoltTable(tbl)
		if doltTable == nil {
			// If this table isn't a Dolt table then we ignore it
			continue
		}
		resolvedTables = append(resolvedTables, doltTable)
		allTableNames = append(allTableNames, doltTable.TableName())
	}
	// TODO: handle DROP TABLE CASCADE
	for _, doltTable := range resolvedTables {
		// Check if the table is in a column
		if err := beforeTableDeletionCheckTableColumns(ctx, doltTable, allTableNames); err != nil {
			return nil, err
		}
		// Check if the table is in a function/procedure parameter
		if err := beforeTableDeletionCheckFuncsProcs(ctx, doltTable, allTableNames); err != nil {
			return nil, err
		}
	}
	return n, nil
}

// beforeTableDeletionCheckTableColumns is called from BeforeTableDeletion and handles checking for columns that use a
// table's type.
func beforeTableDeletionCheckTableColumns(ctx *sql.Context, doltTable *sqle.DoltTable, allDeletedTables []doltdb.TableName) error {
	tableName := doltTable.TableName()
	_, root, err := core.GetRootFromContext(ctx)
	if err != nil {
		return err
	}
	tableAsType := id.NewType(tableName.Schema, tableName.Name)
	allTableNames, err := root.GetAllTableNames(ctx, false)
	if err != nil {
		return err
	}
OuterLoop:
	for _, otherTableName := range allTableNames {
		if doltdb.IsSystemTable(otherTableName) {
			// System tables don't use any table types
			continue OuterLoop
		}
		for _, deletedTable := range allDeletedTables {
			if deletedTable.EqualFold(otherTableName) {
				// If we're also deleting this table, then it doesn't matter what the columns have
				continue OuterLoop
			}
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
			colType := col.TypeInfo.ToSqlType()
			dgtype, ok := colType.(*pgtypes.DoltgresType)
			if !ok {
				// If this isn't a Doltgres type, then it can't be a table type so we can ignore it
				continue
			}
			if dgtype.ID == tableAsType {
				// TODO: portion after newline should be in DETAILS but we don't yet support that in our error messages
				return errors.Newf("cannot drop table %s because other objects depend on it\ncolumn %s of table %s depends on type %s",
					tableName.Name, col.Name, otherTableName.Name, tableName.Name)
			}
		}
	}
	return nil
}

// beforeTableDeletionCheckFuncsProcs is called from BeforeTableDeletion and handles checking for function and procedure
// parameters that use a table's type.
func beforeTableDeletionCheckFuncsProcs(ctx *sql.Context, doltTable *sqle.DoltTable, allDeletedTables []doltdb.TableName) error {
	tableName := doltTable.TableName()
	tableAsType := id.NewType(tableName.Schema, tableName.Name)
	funcsColl, err := core.GetFunctionsCollectionFromContext(ctx)
	if err != nil {
		return err
	}
	err = funcsColl.IterateFunctions(ctx, func(f functions.Function) (stop bool, err error) {
		for _, paramType := range f.ParameterTypes {
			if paramType == tableAsType {
				// TODO: portion after newline should be in DETAILS but we don't yet support that in our error messages
				return true, errors.Newf("cannot drop table %s because other objects depend on it\nfunction %s depends on type %s",
					tableName.Name, f.Name().Name, tableName.Name)
			}
		}
		return false, nil
	})
	if err != nil {
		return err
	}
	procsColl, err := core.GetProceduresCollectionFromContext(ctx)
	if err != nil {
		return err
	}
	err = procsColl.IterateProcedures(ctx, func(p procedures.Procedure) (stop bool, err error) {
		for _, paramType := range p.ParameterTypes {
			if paramType == tableAsType {
				// TODO: portion after newline should be in DETAILS but we don't yet support that in our error messages
				return true, errors.Newf("cannot drop table %s because other objects depend on it\nfunction %s depends on type %s",
					tableName.Name, p.Name().Name, tableName.Name)
			}
		}
		return false, nil
	})
	if err != nil {
		return err
	}
	return nil
}
