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

// PgStatioAllIndexesName is a constant to the pg_statio_all_indexes name.
const PgStatioAllIndexesName = "pg_statio_all_indexes"

// InitPgStatioAllIndexes handles registration of the pg_statio_all_indexes handler.
func InitPgStatioAllIndexes() {
	tables.AddHandler(PgCatalogName, PgStatioAllIndexesName, PgStatioAllIndexesHandler{})
}

// PgStatioAllIndexesHandler is the handler for the pg_statio_all_indexes table.
type PgStatioAllIndexesHandler struct{}

var _ tables.Handler = PgStatioAllIndexesHandler{}

// Name implements the interface tables.Handler.
func (p PgStatioAllIndexesHandler) Name() string {
	return PgStatioAllIndexesName
}

// RowIter implements the interface tables.Handler.
func (p PgStatioAllIndexesHandler) RowIter(ctx *sql.Context) (sql.RowIter, error) {
	// TODO: Implement pg_statio_all_indexes row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgStatioAllIndexesHandler) Schema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgStatioAllIndexesSchema,
		PkOrdinals: nil,
	}
}

// pgStatioAllIndexesSchema is the schema for pg_statio_all_indexes.
var pgStatioAllIndexesSchema = sql.Schema{
	{Name: "relid", Type: pgtypes.Oid, Default: nil, Nullable: true, Source: PgStatioAllIndexesName},
	{Name: "indexrelid", Type: pgtypes.Oid, Default: nil, Nullable: true, Source: PgStatioAllIndexesName},
	{Name: "schemaname", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgStatioAllIndexesName},
	{Name: "relname", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgStatioAllIndexesName},
	{Name: "indexrelname", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgStatioAllIndexesName},
	{Name: "idx_blks_read", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatioAllIndexesName},
	{Name: "idx_blks_hit", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatioAllIndexesName},
}

// pgStatioAllIndexesRowIter is the sql.RowIter for the pg_statio_all_indexes table.
type pgStatioAllIndexesRowIter struct {
}

var _ sql.RowIter = (*pgStatioAllIndexesRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgStatioAllIndexesRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgStatioAllIndexesRowIter) Close(ctx *sql.Context) error {
	return nil
}
