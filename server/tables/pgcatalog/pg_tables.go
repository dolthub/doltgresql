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
	// TODO: Implement pg_tables row iter
	return emptyRowIter()
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
}

var _ sql.RowIter = (*pgTablesRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgTablesRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgTablesRowIter) Close(ctx *sql.Context) error {
	return nil
}
