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

// PgPublicationTablesName is a constant to the pg_publication_tables name.
const PgPublicationTablesName = "pg_publication_tables"

// InitPgPublicationTables handles registration of the pg_publication_tables handler.
func InitPgPublicationTables() {
	tables.AddHandler(PgCatalogName, PgPublicationTablesName, PgPublicationTablesHandler{})
}

// PgPublicationTablesHandler is the handler for the pg_publication_tables table.
type PgPublicationTablesHandler struct{}

var _ tables.Handler = PgPublicationTablesHandler{}

// Name implements the interface tables.Handler.
func (p PgPublicationTablesHandler) Name() string {
	return PgPublicationTablesName
}

// RowIter implements the interface tables.Handler.
func (p PgPublicationTablesHandler) RowIter(ctx *sql.Context, partition sql.Partition) (sql.RowIter, error) {
	// TODO: Implement pg_publication_tables row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgPublicationTablesHandler) PkSchema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgPublicationTablesSchema,
		PkOrdinals: nil,
	}
}

// pgPublicationTablesSchema is the schema for pg_publication_tables.
var pgPublicationTablesSchema = sql.Schema{
	{Name: "pubname", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgPublicationTablesName},
	{Name: "schemaname", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgPublicationTablesName},
	{Name: "tablename", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgPublicationTablesName},
	{Name: "attnames", Type: pgtypes.NameArray, Default: nil, Nullable: true, Source: PgPublicationTablesName},
	{Name: "rowfilter", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgPublicationTablesName},
}

// pgPublicationTablesRowIter is the sql.RowIter for the pg_publication_tables table.
type pgPublicationTablesRowIter struct {
}

var _ sql.RowIter = (*pgPublicationTablesRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgPublicationTablesRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgPublicationTablesRowIter) Close(ctx *sql.Context) error {
	return nil
}
