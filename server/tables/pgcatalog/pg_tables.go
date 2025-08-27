// Copyright 2024 Dolthub, Inc.
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

package pgcatalog

import (
	"io"

	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/resolve"
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/server/functions"
	"github.com/dolthub/doltgresql/server/tables"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// PgTablesName is a constant to the pg_tables name.
const PgTablesName = "pg_tables"

// InitPgTables handles registration of the pg_tables handler.
func InitPgTables() {
	tables.AddHandler(PgCatalogName, PgTablesName, PgTablesHandler{})
}

// PgTablesHandler is the handler for the pg_tables table.
type PgTablesHandler struct{}

var _ tables.Handler = PgTablesHandler{}

// Name implements the interface tables.Handler.
func (p PgTablesHandler) Name() string {
	return PgTablesName
}

// RowIter implements the interface tables.Handler.
func (p PgTablesHandler) RowIter(ctx *sql.Context, partition sql.Partition) (sql.RowIter, error) {
	// Use cached data from this process if it exists
	pgCatalogCache, err := getPgCatalogCache(ctx)
	if err != nil {
		return nil, err
	}

	if pgCatalogCache.tables == nil {
		var tables []sql.Table
		var tableSchemas []string
		// TODO: This should include a few information_schema tables
		err := functions.IterateCurrentDatabase(ctx, functions.Callbacks{
			Table: func(ctx *sql.Context, schema functions.ItemSchema, table functions.ItemTable) (cont bool, err error) {
				tables = append(tables, table.Item)
				tableSchemas = append(tableSchemas, schema.Item.SchemaName())
				return true, nil
			},
		})
		if err != nil {
			return nil, err
		}

		if includeSystemTables {
			_, root, err := core.GetRootFromContext(ctx)
			if err != nil {
				return nil, err
			}

			systemTables, err := resolve.GetGeneratedSystemTables(ctx, root)
			if err != nil {
				return nil, err
			}
			pgCatalogCache.systemTables = systemTables
		}

		pgCatalogCache.tables = tables
	}

	return &pgTablesRowIter{
		userTables:       pgCatalogCache.tables,
		systemTableNames: pgCatalogCache.systemTables,
	}, nil
}

// Schema implements the interface tables.Handler.
func (p PgTablesHandler) Schema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgTablesSchema,
		PkOrdinals: nil,
	}
}

// pgTablesSchema is the schema for pg_tables.
var pgTablesSchema = sql.Schema{
	{Name: "schemaname", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgTablesName},
	{Name: "tablename", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgTablesName},
	{Name: "tableowner", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgTablesName},
	{Name: "tablespace", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgTablesName},
	{Name: "hasindexes", Type: pgtypes.Bool, Default: nil, Nullable: true, Source: PgTablesName},
	{Name: "hasrules", Type: pgtypes.Bool, Default: nil, Nullable: true, Source: PgTablesName},
	{Name: "hastriggers", Type: pgtypes.Bool, Default: nil, Nullable: true, Source: PgTablesName},
	{Name: "rowsecurity", Type: pgtypes.Bool, Default: nil, Nullable: true, Source: PgTablesName},
}

// pgTablesRowIter is the sql.RowIter for the pg_tables table.
type pgTablesRowIter struct {
	// userTable are the set of user-defined tables
	userTables []sql.Table
	// systemTableNames is the names of all system tables
	systemTableNames []doltdb.TableName
	// idx is the current index in the iteration through both slices
	idx int
}

var _ sql.RowIter = (*pgTablesRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgTablesRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	if iter.idx >= len(iter.userTables)+len(iter.systemTableNames) {
		return nil, io.EOF
	}
	defer func() {
		iter.idx++
	}()

	var tableName string
	var hasIndexes bool
	var schema string

	if iter.idx < len(iter.userTables) {
		table := iter.userTables[iter.idx]

		switch table := table.(type) {
		case sql.DatabaseSchemaTable:
			schema = table.DatabaseSchema().SchemaName()
		default:
			schema = "information_schema"
		}

		tableName = table.Name()

		if it, ok := table.(sql.IndexAddressable); ok {
			idxs, err := it.GetIndexes(ctx)
			if err != nil {
				return nil, err
			}

			if len(idxs) > 0 {
				hasIndexes = true
			}
		}
	} else {
		tblName := iter.systemTableNames[iter.idx-len(iter.userTables)]
		tableName = tblName.Name
		schema = tblName.Schema
	}

	return sql.Row{
		schema,     // schemaname
		tableName,  // tablename
		"postgres", // tableowner
		nil,        // tablespace
		hasIndexes, // hasindexes
		false,      // hasrules  // TODO
		false,      // hastriggers // TODO
		false,      // rowsecurity
	}, nil
}

// Close implements the interface sql.RowIter.
func (iter *pgTablesRowIter) Close(ctx *sql.Context) error {
	return nil
}
