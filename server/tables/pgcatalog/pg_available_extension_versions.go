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

// PgAvailableExtensionVersionsName is a constant to the pg_available_extension_versions name.
const PgAvailableExtensionVersionsName = "pg_available_extension_versions"

// InitPgAvailableExtensionVersions handles registration of the pg_available_extension_versions handler.
func InitPgAvailableExtensionVersions() {
	tables.AddHandler(PgCatalogName, PgAvailableExtensionVersionsName, PgAvailableExtensionVersionsHandler{})
}

// PgAvailableExtensionVersionsHandler is the handler for the pg_available_extension_versions table.
type PgAvailableExtensionVersionsHandler struct{}

var _ tables.Handler = PgAvailableExtensionVersionsHandler{}

// Name implements the interface tables.Handler.
func (p PgAvailableExtensionVersionsHandler) Name() string {
	return PgAvailableExtensionVersionsName
}

// RowIter implements the interface tables.Handler.
func (p PgAvailableExtensionVersionsHandler) RowIter(ctx *sql.Context) (sql.RowIter, error) {
	// TODO: Implement pg_available_extension_versions row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgAvailableExtensionVersionsHandler) Schema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgAvailableExtensionVersionsSchema,
		PkOrdinals: nil,
	}
}

// pgAvailableExtensionVersionsSchema is the schema for pg_available_extension_versions.
var pgAvailableExtensionVersionsSchema = sql.Schema{
	{Name: "name", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgAvailableExtensionVersionsName},
	{Name: "version", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgAvailableExtensionVersionsName},
	{Name: "installed", Type: pgtypes.Bool, Default: nil, Nullable: true, Source: PgAvailableExtensionVersionsName},
	{Name: "superuser", Type: pgtypes.Bool, Default: nil, Nullable: true, Source: PgAvailableExtensionVersionsName},
	{Name: "trusted", Type: pgtypes.Bool, Default: nil, Nullable: true, Source: PgAvailableExtensionVersionsName},
	{Name: "relocatable", Type: pgtypes.Bool, Default: nil, Nullable: true, Source: PgAvailableExtensionVersionsName},
	{Name: "schema", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgAvailableExtensionVersionsName},
	{Name: "requires", Type: pgtypes.NameArray, Default: nil, Nullable: true, Source: PgAvailableExtensionVersionsName},
	{Name: "comment", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgAvailableExtensionVersionsName},
}

// pgAvailableExtensionVersionsRowIter is the sql.RowIter for the pg_available_extension_versions table.
type pgAvailableExtensionVersionsRowIter struct {
}

var _ sql.RowIter = (*pgAvailableExtensionVersionsRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgAvailableExtensionVersionsRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgAvailableExtensionVersionsRowIter) Close(ctx *sql.Context) error {
	return nil
}
