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

	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/dsess"
	sqle "github.com/dolthub/go-mysql-server"
	"github.com/dolthub/go-mysql-server/sql"

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
func (p PgTablesHandler) RowIter(ctx *sql.Context) (sql.RowIter, error) {
	doltSession := dsess.DSessFromSess(ctx.Session)
	c := sqle.NewDefault(doltSession.Provider()).Analyzer.Catalog

	var tables []sql.Table
	var schemas []string

	err := currentDatabaseSchemaIter(ctx, c, func(sch sql.DatabaseSchema) (bool, error) {
		// Get tables and table indexes
		err := sql.DBTableIter(ctx, sch, func(t sql.Table) (cont bool, err error) {
			tables = append(tables, t)
			schemas = append(schemas, sch.SchemaName())
			return true, nil
		})
		if err != nil {
			return false, err
		}

		return true, nil
	})
	if err != nil {
		return nil, err
	}

	return &pgTablesRowIter{
		tables:  tables,
		schemas: schemas,
		idx:     0,
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
	tables  []sql.Table
	schemas []string
	idx     int
}

var _ sql.RowIter = (*pgTablesRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgTablesRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	if iter.idx >= len(iter.tables) {
		return nil, io.EOF
	}
	iter.idx++
	table := iter.tables[iter.idx-1]
	schema := iter.schemas[iter.idx-1]

	hasIndexes := false
	if it, ok := table.(sql.IndexAddressable); ok {
		idxs, err := it.GetIndexes(ctx)
		if err != nil {
			return nil, err
		}

		if len(idxs) > 0 {
			hasIndexes = true
		}
	}

	return sql.Row{
		schema,       // schemaname
		table.Name(), // tablename
		"",           // tableowner
		"",           // tablespace
		hasIndexes,   // hasindexes
		false,        // hasrules
		false,        // hastriggers
		false,        // rowsecurity
	}, nil
}

// Close implements the interface sql.RowIter.
func (iter *pgTablesRowIter) Close(ctx *sql.Context) error {
	return nil
}
