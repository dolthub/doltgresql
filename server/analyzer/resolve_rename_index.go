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

package analyzer

import (
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/analyzer"
	"github.com/dolthub/go-mysql-server/sql/plan"
	"github.com/dolthub/go-mysql-server/sql/transform"

	"github.com/dolthub/doltgresql/core"
	pgnodes "github.com/dolthub/doltgresql/server/node"
	"github.com/dolthub/doltgresql/server/tables"
)

// resolveRenameIndex resolves the table owning the index named in an ALTER INDEX ... RENAME TO statement and fills
// it in on the RenameIndex node. It also validates that the new index name is not already in use on that table.
func resolveRenameIndex(ctx *sql.Context, a *analyzer.Analyzer, n sql.Node, scope *plan.Scope, selector analyzer.RuleSelector, qFlags *sql.QueryFlags) (sql.Node, transform.TreeIdentity, error) {
	ri, ok := n.(*pgnodes.RenameIndex)
	if !ok || ri.Resolved() {
		return n, transform.SameTree, nil
	}

	tbl, schema, err := findRenameIndexTable(ctx, ri)
	if err != nil {
		return nil, transform.SameTree, err
	}
	if tbl == nil {
		if !ri.IfExists {
			return nil, transform.SameTree, errors.Errorf(`relation "%s" does not exist`, ri.IndexName)
		}
		return ri.WithResolvedTable(nil), transform.NewTree, nil
	}

	alterableTbl, ok := tbl.(sql.IndexAlterableTable)
	if !ok {
		return nil, transform.SameTree, errors.Errorf(`table "%s" does not support renaming indexes`, tbl.Name())
	}
	if validator, ok := tables.WrapSqlDatabase(schema).(sql.SchemaObjectNameValidator); ok {
		if _, err := validator.ValidateNewIndexName(ctx, ri.NewName, false); err != nil {
			return nil, transform.SameTree, err
		}
	}

	return ri.WithResolvedTable(alterableTbl), transform.NewTree, nil
}

// findRenameIndexTable finds the table that owns the index being renamed, along with the schema containing it.
// Postgres does not name the table in ALTER INDEX statements, so we search the schemas on the search path (or the
// explicitly named schema) for a table with a matching index. Returns nils (with no error) if no matching index
// was found.
func findRenameIndexTable(ctx *sql.Context, ri *pgnodes.RenameIndex) (sql.Table, sql.DatabaseSchema, error) {
	db, err := core.GetSqlDatabaseFromContext(ctx, ri.DbName)
	if err != nil {
		return nil, nil, err
	}
	if db == nil {
		dbName := ri.DbName
		if len(dbName) == 0 {
			dbName = ctx.GetCurrentDatabase()
		}
		return nil, nil, sql.ErrDatabaseNotFound.New(dbName)
	}
	schemaDb, ok := db.(sql.SchemaDatabase)
	if !ok {
		return nil, nil, errors.Errorf(`database "%s" does not support schemas`, db.Name())
	}
	var searchPath []string
	if len(ri.SchemaName) > 0 {
		searchPath = []string{ri.SchemaName}
	} else {
		searchPath, err = core.SearchPath(ctx)
		if err != nil {
			return nil, nil, err
		}
	}
	for _, schemaName := range searchPath {
		// System schemas hold read-only virtual tables, so we never look for user indexes there
		if schemaName == "pg_catalog" || schemaName == "information_schema" {
			continue
		}
		schema, ok, err := schemaDb.GetSchema(ctx, schemaName)
		if err != nil {
			return nil, nil, err
		}
		if !ok {
			continue
		}
		var tableNames []string
		if len(ri.TableName) > 0 {
			tableNames = []string{ri.TableName}
		} else {
			tableNames, err = schema.GetTableNames(ctx)
			if err != nil {
				return nil, nil, err
			}
		}
		for _, tableName := range tableNames {
			tbl, ok, err := schema.GetTableInsensitive(ctx, tableName)
			if err != nil {
				return nil, nil, err
			}
			if !ok {
				continue
			}
			idxTbl, ok := tbl.(sql.IndexAddressableTable)
			if !ok {
				continue
			}
			indexes, err := idxTbl.GetIndexes(ctx)
			if err != nil {
				return nil, nil, err
			}
			for _, index := range indexes {
				if strings.EqualFold(index.ID(), ri.IndexName) {
					return tbl, schema, nil
				}
			}
		}
	}
	return nil, nil, nil
}
