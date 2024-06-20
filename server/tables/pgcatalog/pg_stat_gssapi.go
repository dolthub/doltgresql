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

// PgStatGssapiName is a constant to the pg_stat_gssapi name.
const PgStatGssapiName = "pg_stat_gssapi"

// InitPgStatGssapi handles registration of the pg_stat_gssapi handler.
func InitPgStatGssapi() {
	tables.AddHandler(PgCatalogName, PgStatGssapiName, PgStatGssapiHandler{})
}

// PgStatGssapiHandler is the handler for the pg_stat_gssapi table.
type PgStatGssapiHandler struct{}

var _ tables.Handler = PgStatGssapiHandler{}

// Name implements the interface tables.Handler.
func (p PgStatGssapiHandler) Name() string {
	return PgStatGssapiName
}

// RowIter implements the interface tables.Handler.
func (p PgStatGssapiHandler) RowIter(ctx *sql.Context) (sql.RowIter, error) {
	// TODO: Implement pg_stat_gssapi row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgStatGssapiHandler) Schema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgStatGssapiSchema,
		PkOrdinals: nil,
	}
}

// pgStatGssapiSchema is the schema for pg_stat_gssapi.
var pgStatGssapiSchema = sql.Schema{
	{Name: "pid", Type: pgtypes.Int32, Default: nil, Nullable: true, Source: PgStatGssapiName},
	{Name: "gss_authenticated", Type: pgtypes.Bool, Default: nil, Nullable: true, Source: PgStatGssapiName},
	{Name: "principal", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgStatGssapiName},
	{Name: "encrypted", Type: pgtypes.Bool, Default: nil, Nullable: true, Source: PgStatGssapiName},
	{Name: "credentials_delegated", Type: pgtypes.Bool, Default: nil, Nullable: true, Source: PgStatGssapiName},
}

// pgStatGssapiRowIter is the sql.RowIter for the pg_stat_gssapi table.
type pgStatGssapiRowIter struct {
}

var _ sql.RowIter = (*pgStatGssapiRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgStatGssapiRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgStatGssapiRowIter) Close(ctx *sql.Context) error {
	return nil
}
