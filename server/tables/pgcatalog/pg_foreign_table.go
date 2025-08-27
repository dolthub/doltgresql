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

// PgForeignTableName is a constant to the pg_foreign_table name.
const PgForeignTableName = "pg_foreign_table"

// InitPgForeignTable handles registration of the pg_foreign_table handler.
func InitPgForeignTable() {
	tables.AddHandler(PgCatalogName, PgForeignTableName, PgForeignTableHandler{})
}

// PgForeignTableHandler is the handler for the pg_foreign_table table.
type PgForeignTableHandler struct{}

var _ tables.Handler = PgForeignTableHandler{}

// Name implements the interface tables.Handler.
func (p PgForeignTableHandler) Name() string {
	return PgForeignTableName
}

// RowIter implements the interface tables.Handler.
func (p PgForeignTableHandler) RowIter(ctx *sql.Context, partition sql.Partition) (sql.RowIter, error) {
	// TODO: Implement pg_foreign_table row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgForeignTableHandler) Schema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgForeignTableSchema,
		PkOrdinals: nil,
	}
}

// pgForeignTableSchema is the schema for pg_foreign_table.
var pgForeignTableSchema = sql.Schema{
	{Name: "ftrelid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgForeignTableName},
	{Name: "ftserver", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgForeignTableName},
	{Name: "ftoptions", Type: pgtypes.TextArray, Default: nil, Nullable: true, Source: PgForeignTableName}, // TODO: collation C
}

// pgForeignTableRowIter is the sql.RowIter for the pg_foreign_table table.
type pgForeignTableRowIter struct {
}

var _ sql.RowIter = (*pgForeignTableRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgForeignTableRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgForeignTableRowIter) Close(ctx *sql.Context) error {
	return nil
}
