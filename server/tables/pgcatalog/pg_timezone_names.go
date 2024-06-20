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

// PgTimezoneNamesName is a constant to the pg_timezone_names name.
const PgTimezoneNamesName = "pg_timezone_names"

// InitPgTimezoneNames handles registration of the pg_timezone_names handler.
func InitPgTimezoneNames() {
	tables.AddHandler(PgCatalogName, PgTimezoneNamesName, PgTimezoneNamesHandler{})
}

// PgTimezoneNamesHandler is the handler for the pg_timezone_names table.
type PgTimezoneNamesHandler struct{}

var _ tables.Handler = PgTimezoneNamesHandler{}

// Name implements the interface tables.Handler.
func (p PgTimezoneNamesHandler) Name() string {
	return PgTimezoneNamesName
}

// RowIter implements the interface tables.Handler.
func (p PgTimezoneNamesHandler) RowIter(ctx *sql.Context) (sql.RowIter, error) {
	// TODO: Implement pg_timezone_names row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgTimezoneNamesHandler) Schema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgTimezoneNamesSchema,
		PkOrdinals: nil,
	}
}

// pgTimezoneNamesSchema is the schema for pg_timezone_names.
var pgTimezoneNamesSchema = sql.Schema{
	{Name: "name", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgTimezoneNamesName},
	{Name: "abbrev", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgTimezoneNamesName},
	{Name: "utc_offset", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgTimezoneNamesName}, // TODO: interval type
	{Name: "is_dst", Type: pgtypes.Bool, Default: nil, Nullable: true, Source: PgTimezoneNamesName},
}

// pgTimezoneNamesRowIter is the sql.RowIter for the pg_timezone_names table.
type pgTimezoneNamesRowIter struct {
}

var _ sql.RowIter = (*pgTimezoneNamesRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgTimezoneNamesRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgTimezoneNamesRowIter) Close(ctx *sql.Context) error {
	return nil
}
