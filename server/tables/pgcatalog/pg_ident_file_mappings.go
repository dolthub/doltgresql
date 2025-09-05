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

// PgIdentFileMappingsName is a constant to the pg_ident_file_mappings name.
const PgIdentFileMappingsName = "pg_ident_file_mappings"

// InitPgIdentFileMappings handles registration of the pg_ident_file_mappings handler.
func InitPgIdentFileMappings() {
	tables.AddHandler(PgCatalogName, PgIdentFileMappingsName, PgIdentFileMappingsHandler{})
}

// PgIdentFileMappingsHandler is the handler for the pg_ident_file_mappings table.
type PgIdentFileMappingsHandler struct{}

var _ tables.Handler = PgIdentFileMappingsHandler{}

// Name implements the interface tables.Handler.
func (p PgIdentFileMappingsHandler) Name() string {
	return PgIdentFileMappingsName
}

// RowIter implements the interface tables.Handler.
func (p PgIdentFileMappingsHandler) RowIter(ctx *sql.Context, partition sql.Partition) (sql.RowIter, error) {
	// TODO: Implement pg_ident_file_mappings row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgIdentFileMappingsHandler) PkSchema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgIdentFileMappingsSchema,
		PkOrdinals: nil,
	}
}

// pgIdentFileMappingsSchema is the schema for pg_ident_file_mappings.
var pgIdentFileMappingsSchema = sql.Schema{
	{Name: "line_number", Type: pgtypes.Int32, Default: nil, Nullable: true, Source: PgIdentFileMappingsName},
	{Name: "map_name", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgIdentFileMappingsName},
	{Name: "sys_name", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgIdentFileMappingsName},
	{Name: "pg_username", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgIdentFileMappingsName},
	{Name: "error", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgIdentFileMappingsName},
}

// pgIdentFileMappingsRowIter is the sql.RowIter for the pg_ident_file_mappings table.
type pgIdentFileMappingsRowIter struct {
}

var _ sql.RowIter = (*pgIdentFileMappingsRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgIdentFileMappingsRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgIdentFileMappingsRowIter) Close(ctx *sql.Context) error {
	return nil
}
