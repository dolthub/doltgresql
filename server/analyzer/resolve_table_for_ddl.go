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

// resolveTableForDDL resolves the target table for DDL node types that need a table assigned to them during analysis
// in order to perform their runtime operation. The resolved table is assigned to the node as a sql.TableNode, and is
// nil when the statement tolerates a missing table (IF EXISTS), in which case execution is a no-op.
func resolveTableForDDL(ctx *sql.Context, a *analyzer.Analyzer, n sql.Node, scope *plan.Scope, selector analyzer.RuleSelector, qFlags *sql.QueryFlags) (sql.Node, transform.TreeIdentity, error) {
	switch node := n.(type) {
	case *pgnodes.RenameIndex:
		if node.Resolved() {
			return n, transform.SameTree, nil
		}
		return resolveRenameIndexTable(ctx, node)
	case *pgnodes.AlterTableColumnTypeUsing:
		if node.TableResolved() {
			return n, transform.SameTree, nil
		}
		return resolveAlterColumnTypeUsingTable(ctx, node)
	default:
		return n, transform.SameTree, nil
	}
}

// resolveRenameIndexTable resolves the table owning the index named in an ALTER INDEX ... RENAME TO statement and
// assigns it to the RenameIndex node. It also validates that the new index name is not already in use on that table.
func resolveRenameIndexTable(ctx *sql.Context, ri *pgnodes.RenameIndex) (sql.Node, transform.TreeIdentity, error) {
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

	if _, ok := tbl.(sql.IndexAlterableTable); !ok {
		return nil, transform.SameTree, errors.Errorf(`table "%s" does not support renaming indexes`, tbl.Name())
	}
	if validator, ok := tables.WrapSqlDatabase(schema).(sql.SchemaObjectNameValidator); ok {
		if _, err := validator.ValidateNewIndexName(ctx, ri.NewName, false); err != nil {
			return nil, transform.SameTree, err
		}
	}

	return ri.WithResolvedTable(plan.NewResolvedTable(tbl, schema, nil)), transform.NewTree, nil
}

// resolveAlterColumnTypeUsingTable resolves the target table of an ALTER TABLE ... ALTER COLUMN ... TYPE ... USING
// statement and assigns it to the AlterTableColumnTypeUsing node. The search path is used for unqualified table
// names. A missing table is an error unless the statement specified IF EXISTS.
func resolveAlterColumnTypeUsingTable(ctx *sql.Context, atu *pgnodes.AlterTableColumnTypeUsing) (sql.Node, transform.TreeIdentity, error) {
	tbl, err := pgnodes.ResolveUsingTable(ctx, atu.SchemaName, atu.TableName)
	if err != nil {
		return nil, transform.SameTree, err
	}
	if tbl == nil {
		if !atu.IfExists {
			return nil, transform.SameTree, sql.ErrTableNotFound.New(atu.TableName)
		}
		return atu.WithResolvedTable(nil), transform.NewTree, nil
	}
	db, err := core.GetSqlDatabaseFromContext(ctx, "")
	if err != nil {
		return nil, transform.SameTree, err
	}
	return atu.WithResolvedTable(plan.NewResolvedTable(tbl, db, nil)), transform.NewTree, nil
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
