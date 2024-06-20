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

// PgStatioSysIndexesName is a constant to the pg_statio_sys_indexes name.
const PgStatioSysIndexesName = "pg_statio_sys_indexes"

// InitPgStatioSysIndexes handles registration of the pg_statio_sys_indexes handler.
func InitPgStatioSysIndexes() {
	tables.AddHandler(PgCatalogName, PgStatioSysIndexesName, PgStatioSysIndexesHandler{})
}

// PgStatioSysIndexesHandler is the handler for the pg_statio_sys_indexes table.
type PgStatioSysIndexesHandler struct{}

var _ tables.Handler = PgStatioSysIndexesHandler{}

// Name implements the interface tables.Handler.
func (p PgStatioSysIndexesHandler) Name() string {
	return PgStatioSysIndexesName
}

// RowIter implements the interface tables.Handler.
func (p PgStatioSysIndexesHandler) RowIter(ctx *sql.Context) (sql.RowIter, error) {
	// TODO: Implement pg_statio_sys_indexes row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgStatioSysIndexesHandler) Schema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgStatioSysIndexesSchema,
		PkOrdinals: nil,
	}
}

// pgStatioSysIndexesSchema is the schema for pg_statio_sys_indexes.
var pgStatioSysIndexesSchema = sql.Schema{
	{Name: "relid", Type: pgtypes.Oid, Default: nil, Nullable: true, Source: PgStatioSysIndexesName},
	{Name: "indexrelid", Type: pgtypes.Oid, Default: nil, Nullable: true, Source: PgStatioSysIndexesName},
	{Name: "schemaname", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgStatioSysIndexesName},
	{Name: "relname", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgStatioSysIndexesName},
	{Name: "indexrelname", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgStatioSysIndexesName},
	{Name: "idx_blks_read", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatioSysIndexesName},
	{Name: "idx_blks_hit", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatioSysIndexesName},
}

// pgStatioSysIndexesRowIter is the sql.RowIter for the pg_statio_sys_indexes table.
type pgStatioSysIndexesRowIter struct {
}

var _ sql.RowIter = (*pgStatioSysIndexesRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgStatioSysIndexesRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgStatioSysIndexesRowIter) Close(ctx *sql.Context) error {
	return nil
}
