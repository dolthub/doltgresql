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

// PgStatAllIndexesName is a constant to the pg_stat_all_indexes name.
const PgStatAllIndexesName = "pg_stat_all_indexes"

// InitPgStatAllIndexes handles registration of the pg_stat_all_indexes handler.
func InitPgStatAllIndexes() {
	tables.AddHandler(PgCatalogName, PgStatAllIndexesName, PgStatAllIndexesHandler{})
}

// PgStatAllIndexesHandler is the handler for the pg_stat_all_indexes table.
type PgStatAllIndexesHandler struct{}

var _ tables.Handler = PgStatAllIndexesHandler{}

// Name implements the interface tables.Handler.
func (p PgStatAllIndexesHandler) Name() string {
	return PgStatAllIndexesName
}

// RowIter implements the interface tables.Handler.
func (p PgStatAllIndexesHandler) RowIter(ctx *sql.Context) (sql.RowIter, error) {
	// TODO: Implement pg_stat_all_indexes row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgStatAllIndexesHandler) Schema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgStatAllIndexesSchema,
		PkOrdinals: nil,
	}
}

// pgStatAllIndexesSchema is the schema for pg_stat_all_indexes.
var pgStatAllIndexesSchema = sql.Schema{
	{Name: "relid", Type: pgtypes.Oid, Default: nil, Nullable: true, Source: PgStatAllIndexesName},
	{Name: "indexrelid", Type: pgtypes.Oid, Default: nil, Nullable: true, Source: PgStatAllIndexesName},
	{Name: "schemaname", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgStatAllIndexesName},
	{Name: "relname", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgStatAllIndexesName},
	{Name: "indexrelname", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgStatAllIndexesName},
	{Name: "idx_scan", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatAllIndexesName},
	{Name: "last_idx_scan", Type: pgtypes.TimestampTZ, Default: nil, Nullable: true, Source: PgStatAllIndexesName},
	{Name: "idx_tup_read", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatAllIndexesName},
	{Name: "idx_tup_fetch", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatAllIndexesName},
}

// pgStatAllIndexesRowIter is the sql.RowIter for the pg_stat_all_indexes table.
type pgStatAllIndexesRowIter struct {
}

var _ sql.RowIter = (*pgStatAllIndexesRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgStatAllIndexesRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgStatAllIndexesRowIter) Close(ctx *sql.Context) error {
	return nil
}
