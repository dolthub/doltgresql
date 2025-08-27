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

// PgFileSettingsName is a constant to the pg_file_settings name.
const PgFileSettingsName = "pg_file_settings"

// InitPgFileSettings handles registration of the pg_file_settings handler.
func InitPgFileSettings() {
	tables.AddHandler(PgCatalogName, PgFileSettingsName, PgFileSettingsHandler{})
}

// PgFileSettingsHandler is the handler for the pg_file_settings table.
type PgFileSettingsHandler struct{}

var _ tables.Handler = PgFileSettingsHandler{}

// Name implements the interface tables.Handler.
func (p PgFileSettingsHandler) Name() string {
	return PgFileSettingsName
}

// RowIter implements the interface tables.Handler.
func (p PgFileSettingsHandler) RowIter(ctx *sql.Context, partition sql.Partition) (sql.RowIter, error) {
	// TODO: Implement pg_file_settings row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgFileSettingsHandler) Schema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgFileSettingsSchema,
		PkOrdinals: nil,
	}
}

// pgFileSettingsSchema is the schema for pg_file_settings.
var pgFileSettingsSchema = sql.Schema{
	{Name: "sourcefile", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgFileSettingsName},
	{Name: "sourceline", Type: pgtypes.Int32, Default: nil, Nullable: true, Source: PgFileSettingsName},
	{Name: "seqno", Type: pgtypes.Int32, Default: nil, Nullable: true, Source: PgFileSettingsName},
	{Name: "name", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgFileSettingsName},
	{Name: "setting", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgFileSettingsName},
	{Name: "applied", Type: pgtypes.Bool, Default: nil, Nullable: true, Source: PgFileSettingsName},
	{Name: "error", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgFileSettingsName},
}

// pgFileSettingsRowIter is the sql.RowIter for the pg_file_settings table.
type pgFileSettingsRowIter struct {
}

var _ sql.RowIter = (*pgFileSettingsRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgFileSettingsRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgFileSettingsRowIter) Close(ctx *sql.Context) error {
	return nil
}
