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

// PgStatUserIndexesName is a constant to the pg_stat_user_indexes name.
const PgStatUserIndexesName = "pg_stat_user_indexes"

// InitPgStatUserIndexes handles registration of the pg_stat_user_indexes handler.
func InitPgStatUserIndexes() {
	tables.AddHandler(PgCatalogName, PgStatUserIndexesName, PgStatUserIndexesHandler{})
}

// PgStatUserIndexesHandler is the handler for the pg_stat_user_indexes table.
type PgStatUserIndexesHandler struct{}

var _ tables.Handler = PgStatUserIndexesHandler{}

// Name implements the interface tables.Handler.
func (p PgStatUserIndexesHandler) Name() string {
	return PgStatUserIndexesName
}

// RowIter implements the interface tables.Handler.
func (p PgStatUserIndexesHandler) RowIter(ctx *sql.Context, partition sql.Partition) (sql.RowIter, error) {
	// TODO: Implement pg_stat_user_indexes row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgStatUserIndexesHandler) Schema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgStatUserIndexesSchema,
		PkOrdinals: nil,
	}
}

// pgStatUserIndexesSchema is the schema for pg_stat_user_indexes.
var pgStatUserIndexesSchema = sql.Schema{
	{Name: "relid", Type: pgtypes.Oid, Default: nil, Nullable: true, Source: PgStatUserIndexesName},
	{Name: "indexrelid", Type: pgtypes.Oid, Default: nil, Nullable: true, Source: PgStatUserIndexesName},
	{Name: "schemaname", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgStatUserIndexesName},
	{Name: "relname", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgStatUserIndexesName},
	{Name: "indexrelname", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgStatUserIndexesName},
	{Name: "idx_scan", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatUserIndexesName},
	{Name: "last_idx_scan", Type: pgtypes.TimestampTZ, Default: nil, Nullable: true, Source: PgStatUserIndexesName},
	{Name: "idx_tup_read", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatUserIndexesName},
	{Name: "idx_tup_fetch", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatUserIndexesName},
}

// pgStatUserIndexesRowIter is the sql.RowIter for the pg_stat_user_indexes table.
type pgStatUserIndexesRowIter struct {
}

var _ sql.RowIter = (*pgStatUserIndexesRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgStatUserIndexesRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgStatUserIndexesRowIter) Close(ctx *sql.Context) error {
	return nil
}
