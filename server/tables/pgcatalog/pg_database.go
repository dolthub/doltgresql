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

// PgDatabaseName is a constant to the pg_database name.
const PgDatabaseName = "pg_database"

// InitPgDatabase handles registration of the pg_database handler.
func InitPgDatabase() {
	tables.AddHandler(PgCatalogName, PgDatabaseName, PgDatabaseHandler{})
}

// PgDatabaseHandler is the handler for the pg_database table.
type PgDatabaseHandler struct{}

var _ tables.Handler = PgDatabaseHandler{}

// Name implements the interface tables.Handler.
func (p PgDatabaseHandler) Name() string {
	return PgDatabaseName
}

// emptyRowIter implements the sql.RowIter for empty table.
func emptyRowIter() (sql.RowIter, error) {
	return sql.RowsToRowIter(), nil
}

// RowIter implements the interface tables.Handler.
func (p PgDatabaseHandler) RowIter(ctx *sql.Context) (sql.RowIter, error) {
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgDatabaseHandler) Schema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgDatabaseSchema,
		PkOrdinals: nil,
	}
}

// pgDatabaseSchema is the schema for pg_database.
var pgDatabaseSchema = sql.Schema{
	{Name: "oid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgDatabaseName},
	{Name: "datname", Type: pgtypes.Name, Default: nil, Nullable: false, Source: PgDatabaseName},
	{Name: "datdba", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgDatabaseName},
	{Name: "encoding", Type: pgtypes.Int32, Default: nil, Nullable: false, Source: PgDatabaseName},
	{Name: "datlocprovider", Type: pgtypes.BpChar, Default: nil, Nullable: false, Source: PgDatabaseName},
	{Name: "datistemplate", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgDatabaseName},
	{Name: "datallowconn", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgDatabaseName},
	{Name: "datconnlimit", Type: pgtypes.Int32, Default: nil, Nullable: false, Source: PgDatabaseName},
	{Name: "datfrozenxid", Type: pgtypes.Xid, Default: nil, Nullable: false, Source: PgDatabaseName},
	{Name: "datminmxid", Type: pgtypes.Xid, Default: nil, Nullable: false, Source: PgDatabaseName},
	{Name: "dattablespace", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgDatabaseName},
	{Name: "datcollate", Type: pgtypes.Text, Default: nil, Nullable: false, Source: PgDatabaseName},  // TODO: collation C
	{Name: "datctype", Type: pgtypes.Text, Default: nil, Nullable: false, Source: PgDatabaseName},    // TODO: collation C
	{Name: "daticulocale", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgDatabaseName}, // TODO: collation C
	{Name: "daticurules", Type: pgtypes.Text, Default: nil, Nullable: false, Source: PgDatabaseName},
	{Name: "datcollversion", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgDatabaseName}, // TODO: collation C
	{Name: "datacl", Type: pgtypes.TextArray, Default: nil, Nullable: true, Source: PgDatabaseName},    // TODO: type aclitem[]
}

// pgDatabaseRowIter is the sql.RowIter for the pg_database table.
type pgDatabaseRowIter struct {
	idx int
}

var _ sql.RowIter = (*pgDatabaseRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgDatabaseRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgDatabaseRowIter) Close(ctx *sql.Context) error {
	return nil
}
