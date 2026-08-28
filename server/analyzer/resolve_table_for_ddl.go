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
		tbl, schema, err := resolveTableUsingSearchPath(ctx, node.DbName, node.SchemaName, node.TableName, node.IndexName)
		if err != nil {
			return nil, transform.SameTree, err
		}
		if tbl == nil {
			if !node.IfExists {
				return nil, transform.SameTree, errors.Errorf(`relation "%s" does not exist`, node.IndexName)
			}
			return node.WithResolvedTable(nil), transform.NewTree, nil
		}
		if _, ok := tbl.UnderlyingTable().(sql.IndexAlterableTable); !ok {
			return nil, transform.SameTree, errors.Errorf(`table "%s" does not support renaming indexes`, tbl.Name())
		}
		if validator, ok := tables.WrapSqlDatabase(schema).(sql.SchemaObjectNameValidator); ok {
			if _, err := validator.ValidateNewIndexName(ctx, node.NewName, false); err != nil {
				return nil, transform.SameTree, err
			}
		}
		return node.WithResolvedTable(tbl), transform.NewTree, nil
	case *pgnodes.AlterTableColumnTypeUsing:
		tbl, _, err := resolveTableUsingSearchPath(ctx, "", node.SchemaName, node.TableName, "")
		if err != nil {
			return nil, transform.SameTree, err
		}
		if tbl == nil {
			if !node.IfExists {
				return nil, transform.SameTree, sql.ErrTableNotFound.New(node.TableName)
			}
			// The missing table makes this statement a no-op, which the node's execution handles
			return n, transform.SameTree, nil
		}
		return node.WithResolvedTable(tbl), transform.NewTree, nil
	default:
		return n, transform.SameTree, nil
	}
}

// resolveTableUsingSearchPath finds the target table of a DDL statement, searching the schemas on the search path (or
// only the explicitly named schema) and returning the table as a sql.TableNode along with the schema containing it.
// When indexName is non-empty, the table name may be empty: Postgres does not name the table in ALTER INDEX
// statements, so we search for a table owning an index with that name. Returns nils (with no error) if no matching
// table was found.
func resolveTableUsingSearchPath(ctx *sql.Context, dbName, schemaName, tableName, indexName string) (sql.TableNode, sql.DatabaseSchema, error) {
	db, err := core.GetSqlDatabaseFromContext(ctx, dbName)
	if err != nil {
		return nil, nil, err
	}
	if db == nil {
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
	if len(schemaName) > 0 {
		searchPath = []string{schemaName}
	} else {
		searchPath, err = core.SearchPath(ctx)
		if err != nil {
			return nil, nil, err
		}
	}
	for _, searchSchemaName := range searchPath {
		// System schemas hold read-only virtual tables, which can never be the target of a DDL statement
		if searchSchemaName == "pg_catalog" || searchSchemaName == "information_schema" {
			continue
		}
		schema, ok, err := schemaDb.GetSchema(ctx, searchSchemaName)
		if err != nil {
			return nil, nil, err
		}
		if !ok {
			continue
		}
		var tableNames []string
		if len(tableName) > 0 {
			tableNames = []string{tableName}
		} else {
			tableNames, err = schema.GetTableNames(ctx)
			if err != nil {
				return nil, nil, err
			}
		}
		for _, searchTableName := range tableNames {
			tbl, ok, err := schema.GetTableInsensitive(ctx, searchTableName)
			if err != nil {
				return nil, nil, err
			}
			if !ok {
				continue
			}
			if len(indexName) > 0 {
				hasIndex, err := tableHasIndex(ctx, tbl, indexName)
				if err != nil {
					return nil, nil, err
				}
				if !hasIndex {
					continue
				}
			}
			return plan.NewResolvedTable(tbl, schema, nil), schema, nil
		}
	}
	return nil, nil, nil
}

// tableHasIndex returns whether the given table has an index with the given name (compared case-insensitively).
func tableHasIndex(ctx *sql.Context, tbl sql.Table, indexName string) (bool, error) {
	idxTbl, ok := tbl.(sql.IndexAddressableTable)
	if !ok {
		return false, nil
	}
	indexes, err := idxTbl.GetIndexes(ctx)
	if err != nil {
		return false, err
	}
	for _, index := range indexes {
		if strings.EqualFold(index.ID(), indexName) {
			return true, nil
		}
	}
	return false, nil
}
