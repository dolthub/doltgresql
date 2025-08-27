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

// PgSettingsName is a constant to the pg_settings name.
const PgSettingsName = "pg_settings"

// InitPgSettings handles registration of the pg_settings handler.
func InitPgSettings() {
	tables.AddHandler(PgCatalogName, PgSettingsName, PgSettingsHandler{})
}

// PgSettingsHandler is the handler for the pg_settings table.
type PgSettingsHandler struct{}

var _ tables.Handler = PgSettingsHandler{}

// Name implements the interface tables.Handler.
func (p PgSettingsHandler) Name() string {
	return PgSettingsName
}

// RowIter implements the interface tables.Handler.
func (p PgSettingsHandler) RowIter(ctx *sql.Context, partition sql.Partition) (sql.RowIter, error) {
	// TODO: Implement pg_settings row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgSettingsHandler) Schema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgSettingsSchema,
		PkOrdinals: nil,
	}
}

// pgSettingsSchema is the schema for pg_settings.
var pgSettingsSchema = sql.Schema{
	{Name: "name", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgSettingsName},
	{Name: "setting", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgSettingsName},
	{Name: "unit", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgSettingsName},
	{Name: "category", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgSettingsName},
	{Name: "short_desc", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgSettingsName},
	{Name: "extra_desc", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgSettingsName},
	{Name: "context", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgSettingsName},
	{Name: "vartype", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgSettingsName},
	{Name: "source", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgSettingsName},
	{Name: "min_val", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgSettingsName},
	{Name: "max_val", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgSettingsName},
	{Name: "enumvals", Type: pgtypes.TextArray, Default: nil, Nullable: true, Source: PgSettingsName},
	{Name: "boot_val", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgSettingsName},
	{Name: "reset_val", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgSettingsName},
	{Name: "sourcefile", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgSettingsName},
	{Name: "sourceline", Type: pgtypes.Int32, Default: nil, Nullable: true, Source: PgSettingsName},
	{Name: "pending_restart", Type: pgtypes.Bool, Default: nil, Nullable: true, Source: PgSettingsName},
}

// pgSettingsRowIter is the sql.RowIter for the pg_settings table.
type pgSettingsRowIter struct {
}

var _ sql.RowIter = (*pgSettingsRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgSettingsRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgSettingsRowIter) Close(ctx *sql.Context) error {
	return nil
}
