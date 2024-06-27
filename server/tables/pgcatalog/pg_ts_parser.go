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

// PgTsParserName is a constant to the pg_ts_parser name.
const PgTsParserName = "pg_ts_parser"

// InitPgTsParser handles registration of the pg_ts_parser handler.
func InitPgTsParser() {
	tables.AddHandler(PgCatalogName, PgTsParserName, PgTsParserHandler{})
}

// PgTsParserHandler is the handler for the pg_ts_parser table.
type PgTsParserHandler struct{}

var _ tables.Handler = PgTsParserHandler{}

// Name implements the interface tables.Handler.
func (p PgTsParserHandler) Name() string {
	return PgTsParserName
}

// RowIter implements the interface tables.Handler.
func (p PgTsParserHandler) RowIter(ctx *sql.Context) (sql.RowIter, error) {
	// TODO: Implement pg_ts_parser row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgTsParserHandler) Schema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgTsParserSchema,
		PkOrdinals: nil,
	}
}

// pgTsParserSchema is the schema for pg_ts_parser.
var pgTsParserSchema = sql.Schema{
	{Name: "oid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgTsParserName},
	{Name: "prsname", Type: pgtypes.Name, Default: nil, Nullable: false, Source: PgTsParserName},
	{Name: "prsnamespace", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgTsParserName},
	{Name: "prsstart", Type: pgtypes.Text, Default: nil, Nullable: false, Source: PgTsParserName},    // TODO: regproc type
	{Name: "prstoken", Type: pgtypes.Text, Default: nil, Nullable: false, Source: PgTsParserName},    // TODO: regproc type
	{Name: "prsend", Type: pgtypes.Text, Default: nil, Nullable: false, Source: PgTsParserName},      // TODO: regproc type
	{Name: "prsheadline", Type: pgtypes.Text, Default: nil, Nullable: false, Source: PgTsParserName}, // TODO: regproc type
	{Name: "prslextype", Type: pgtypes.Text, Default: nil, Nullable: false, Source: PgTsParserName},  // TODO: regproc type
}

// pgTsParserRowIter is the sql.RowIter for the pg_ts_parser table.
type pgTsParserRowIter struct {
}

var _ sql.RowIter = (*pgTsParserRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgTsParserRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgTsParserRowIter) Close(ctx *sql.Context) error {
	return nil
}
