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

// PgStatioUserIndexesName is a constant to the pg_statio_user_indexes name.
const PgStatioUserIndexesName = "pg_statio_user_indexes"

// InitPgStatioUserIndexes handles registration of the pg_statio_user_indexes handler.
func InitPgStatioUserIndexes() {
	tables.AddHandler(PgCatalogName, PgStatioUserIndexesName, PgStatioUserIndexesHandler{})
}

// PgStatioUserIndexesHandler is the handler for the pg_statio_user_indexes table.
type PgStatioUserIndexesHandler struct{}

var _ tables.Handler = PgStatioUserIndexesHandler{}

// Name implements the interface tables.Handler.
func (p PgStatioUserIndexesHandler) Name() string {
	return PgStatioUserIndexesName
}

// RowIter implements the interface tables.Handler.
func (p PgStatioUserIndexesHandler) RowIter(ctx *sql.Context) (sql.RowIter, error) {
	// TODO: Implement pg_statio_user_indexes row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgStatioUserIndexesHandler) Schema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgStatioUserIndexesSchema,
		PkOrdinals: nil,
	}
}

// pgStatioUserIndexesSchema is the schema for pg_statio_user_indexes.
var pgStatioUserIndexesSchema = sql.Schema{
	{Name: "relid", Type: pgtypes.Oid, Default: nil, Nullable: true, Source: PgStatioUserIndexesName},
	{Name: "indexrelid", Type: pgtypes.Oid, Default: nil, Nullable: true, Source: PgStatioUserIndexesName},
	{Name: "schemaname", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgStatioUserIndexesName},
	{Name: "relname", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgStatioUserIndexesName},
	{Name: "indexrelname", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgStatioUserIndexesName},
	{Name: "idx_blks_read", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatioUserIndexesName},
	{Name: "idx_blks_hit", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatioUserIndexesName},
}

// pgStatioUserIndexesRowIter is the sql.RowIter for the pg_statio_user_indexes table.
type pgStatioUserIndexesRowIter struct {
}

var _ sql.RowIter = (*pgStatioUserIndexesRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgStatioUserIndexesRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgStatioUserIndexesRowIter) Close(ctx *sql.Context) error {
	return nil
}
