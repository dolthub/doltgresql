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

// PgStatProgressCopyName is a constant to the pg_stat_progress_copy name.
const PgStatProgressCopyName = "pg_stat_progress_copy"

// InitPgStatProgressCopy handles registration of the pg_stat_progress_copy handler.
func InitPgStatProgressCopy() {
	tables.AddHandler(PgCatalogName, PgStatProgressCopyName, PgStatProgressCopyHandler{})
}

// PgStatProgressCopyHandler is the handler for the pg_stat_progress_copy table.
type PgStatProgressCopyHandler struct{}

var _ tables.Handler = PgStatProgressCopyHandler{}

// Name implements the interface tables.Handler.
func (p PgStatProgressCopyHandler) Name() string {
	return PgStatProgressCopyName
}

// RowIter implements the interface tables.Handler.
func (p PgStatProgressCopyHandler) RowIter(ctx *sql.Context, partition sql.Partition) (sql.RowIter, error) {
	// TODO: Implement pg_stat_progress_copy row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgStatProgressCopyHandler) Schema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgStatProgressCopySchema,
		PkOrdinals: nil,
	}
}

// pgStatProgressCopySchema is the schema for pg_stat_progress_copy.
var pgStatProgressCopySchema = sql.Schema{
	{Name: "pid", Type: pgtypes.Int32, Default: nil, Nullable: true, Source: PgStatProgressCopyName},
	{Name: "datid", Type: pgtypes.Oid, Default: nil, Nullable: true, Source: PgStatProgressCopyName},
	{Name: "datname", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgStatProgressCopyName},
	{Name: "relid", Type: pgtypes.Oid, Default: nil, Nullable: true, Source: PgStatProgressCopyName},
	{Name: "command", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgStatProgressCopyName},
	{Name: "type", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgStatProgressCopyName},
	{Name: "bytes_processed", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatProgressCopyName},
	{Name: "bytes_total", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatProgressCopyName},
	{Name: "tuples_processed", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatProgressCopyName},
	{Name: "tuples_excluded", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatProgressCopyName},
}

// pgStatProgressCopyRowIter is the sql.RowIter for the pg_stat_progress_copy table.
type pgStatProgressCopyRowIter struct {
}

var _ sql.RowIter = (*pgStatProgressCopyRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgStatProgressCopyRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgStatProgressCopyRowIter) Close(ctx *sql.Context) error {
	return nil
}
