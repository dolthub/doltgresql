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

// PgStatSslName is a constant to the pg_stat_ssl name.
const PgStatSslName = "pg_stat_ssl"

// InitPgStatSsl handles registration of the pg_stat_ssl handler.
func InitPgStatSsl() {
	tables.AddHandler(PgCatalogName, PgStatSslName, PgStatSslHandler{})
}

// PgStatSslHandler is the handler for the pg_stat_ssl table.
type PgStatSslHandler struct{}

var _ tables.Handler = PgStatSslHandler{}

// Name implements the interface tables.Handler.
func (p PgStatSslHandler) Name() string {
	return PgStatSslName
}

// RowIter implements the interface tables.Handler.
func (p PgStatSslHandler) RowIter(ctx *sql.Context, partition sql.Partition) (sql.RowIter, error) {
	// TODO: Implement pg_stat_ssl row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgStatSslHandler) PkSchema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgStatSslSchema,
		PkOrdinals: nil,
	}
}

// pgStatSslSchema is the schema for pg_stat_ssl.
var pgStatSslSchema = sql.Schema{
	{Name: "pid", Type: pgtypes.Int32, Default: nil, Nullable: true, Source: PgStatSslName},
	{Name: "ssl", Type: pgtypes.Bool, Default: nil, Nullable: true, Source: PgStatSslName},
	{Name: "version", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgStatSslName},
	{Name: "cipher", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgStatSslName},
	{Name: "bits", Type: pgtypes.Int32, Default: nil, Nullable: true, Source: PgStatSslName},
	{Name: "client_dn", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgStatSslName},
	{Name: "client_serial", Type: pgtypes.Numeric, Default: nil, Nullable: true, Source: PgStatSslName},
	{Name: "issuer_dn", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgStatSslName},
}

// pgStatSslRowIter is the sql.RowIter for the pg_stat_ssl table.
type pgStatSslRowIter struct {
}

var _ sql.RowIter = (*pgStatSslRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgStatSslRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgStatSslRowIter) Close(ctx *sql.Context) error {
	return nil
}
