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
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/plan"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/core"
)

// RenameIndex handles the ALTER INDEX ... RENAME TO ... statement.
type RenameIndex struct {
	ifExists   bool
	dbName     string // optional catalog qualifier for the index name
	schemaName string // optional schema qualifier for the index name
	tableName  string // optional table name (from the CockroachDB `table@index` syntax)
	indexName  string
	newName    string
}

var _ sql.ExecSourceRel = (*RenameIndex)(nil)
var _ vitess.Injectable = (*RenameIndex)(nil)

// NewRenameIndex returns a new *RenameIndex.
func NewRenameIndex(ifExists bool, dbName, schemaName, tableName, indexName, newName string) *RenameIndex {
	return &RenameIndex{
		ifExists:   ifExists,
		dbName:     dbName,
		schemaName: schemaName,
		tableName:  tableName,
		indexName:  indexName,
		newName:    newName,
	}
}

// Children implements the interface sql.ExecSourceRel.
func (r *RenameIndex) Children() []sql.Node {
	return nil
}

// IsReadOnly implements the interface sql.ExecSourceRel.
func (r *RenameIndex) IsReadOnly() bool {
	return false
}

// Resolved implements the interface sql.ExecSourceRel.
func (r *RenameIndex) Resolved() bool {
	return true
}

// RowIter implements the interface sql.ExecSourceRel.
func (r *RenameIndex) RowIter(ctx *sql.Context, _ sql.Row) (sql.RowIter, error) {
	tbl, err := r.findIndexTable(ctx)
	if err != nil {
		return nil, err
	}
	if tbl == nil {
		if r.ifExists {
			// TODO: send notice "relation ... does not exist, skipping"
			return sql.RowsToRowIter(), nil
		}
		return nil, errors.Errorf(`relation "%s" does not exist`, r.indexName)
	}
	alterableTbl, ok := tbl.(sql.IndexAlterableTable)
	if !ok {
		return nil, errors.Errorf(`table "%s" does not support renaming indexes`, tbl.Name())
	}
	if idxTbl, ok := tbl.(sql.IndexAddressableTable); ok {
		indexes, err := idxTbl.GetIndexes(ctx)
		if err != nil {
			return nil, err
		}
		for _, index := range indexes {
			if strings.EqualFold(index.ID(), r.newName) {
				return nil, errors.Errorf(`relation "%s" already exists`, r.newName)
			}
		}
	}
	if err = alterableTbl.RenameIndex(ctx, r.indexName, r.newName); err != nil {
		return nil, err
	}
	return sql.RowsToRowIter(), nil
}

// findIndexTable finds the table that owns the index being renamed. Postgres does not name the table in ALTER INDEX
// statements, so we search the schemas on the search path (or the explicitly named schema) for a table with a matching
// index. Returns nil (with no error) if no matching index was found.
func (r *RenameIndex) findIndexTable(ctx *sql.Context) (sql.Table, error) {
	db, err := core.GetSqlDatabaseFromContext(ctx, r.dbName)
	if err != nil {
		return nil, err
	}
	if db == nil {
		dbName := r.dbName
		if len(dbName) == 0 {
			dbName = ctx.GetCurrentDatabase()
		}
		return nil, sql.ErrDatabaseNotFound.New(dbName)
	}
	schemaDb, ok := db.(sql.SchemaDatabase)
	if !ok {
		return nil, errors.Errorf(`database "%s" does not support schemas`, db.Name())
	}
	var searchPath []string
	if len(r.schemaName) > 0 {
		searchPath = []string{r.schemaName}
	} else {
		searchPath, err = core.SearchPath(ctx)
		if err != nil {
			return nil, err
		}
	}
	for _, schemaName := range searchPath {
		// System schemas hold read-only virtual tables, so we never look for user indexes there
		if schemaName == "pg_catalog" || schemaName == "information_schema" {
			continue
		}
		schema, ok, err := schemaDb.GetSchema(ctx, schemaName)
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}
		var tableNames []string
		if len(r.tableName) > 0 {
			tableNames = []string{r.tableName}
		} else {
			tableNames, err = schema.GetTableNames(ctx)
			if err != nil {
				return nil, err
			}
		}
		for _, tableName := range tableNames {
			tbl, ok, err := schema.GetTableInsensitive(ctx, tableName)
			if err != nil {
				return nil, err
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
				return nil, err
			}
			for _, index := range indexes {
				if strings.EqualFold(index.ID(), r.indexName) {
					return tbl, nil
				}
			}
		}
	}
	return nil, nil
}

// Schema implements the interface sql.ExecSourceRel.
func (r *RenameIndex) Schema(ctx *sql.Context) sql.Schema {
	return nil
}

// String implements the interface sql.ExecSourceRel.
func (r *RenameIndex) String() string {
	return "ALTER INDEX RENAME"
}

// WithChildren implements the interface sql.ExecSourceRel.
func (r *RenameIndex) WithChildren(ctx *sql.Context, children ...sql.Node) (sql.Node, error) {
	return plan.NillaryWithChildren(r, children...)
}

// WithResolvedChildren implements the interface vitess.Injectable.
func (r *RenameIndex) WithResolvedChildren(ctx context.Context, children []any) (any, error) {
	if len(children) != 0 {
		return nil, ErrVitessChildCount.New(0, len(children))
	}
	return r, nil
}
