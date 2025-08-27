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

// PgStatProgressCreateIndexName is a constant to the pg_stat_progress_create_index name.
const PgStatProgressCreateIndexName = "pg_stat_progress_create_index"

// InitPgStatProgressCreateIndex handles registration of the pg_stat_progress_create_index handler.
func InitPgStatProgressCreateIndex() {
	tables.AddHandler(PgCatalogName, PgStatProgressCreateIndexName, PgStatProgressCreateIndexHandler{})
}

// PgStatProgressCreateIndexHandler is the handler for the pg_stat_progress_create_index table.
type PgStatProgressCreateIndexHandler struct{}

var _ tables.Handler = PgStatProgressCreateIndexHandler{}

// Name implements the interface tables.Handler.
func (p PgStatProgressCreateIndexHandler) Name() string {
	return PgStatProgressCreateIndexName
}

// RowIter implements the interface tables.Handler.
func (p PgStatProgressCreateIndexHandler) RowIter(ctx *sql.Context, partition sql.Partition) (sql.RowIter, error) {
	// TODO: Implement pg_stat_progress_create_index row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgStatProgressCreateIndexHandler) Schema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgStatProgressCreateIndexSchema,
		PkOrdinals: nil,
	}
}

// pgStatProgressCreateIndexSchema is the schema for pg_stat_progress_create_index.
var pgStatProgressCreateIndexSchema = sql.Schema{
	{Name: "pid", Type: pgtypes.Int32, Default: nil, Nullable: true, Source: PgStatProgressCreateIndexName},
	{Name: "datid", Type: pgtypes.Oid, Default: nil, Nullable: true, Source: PgStatProgressCreateIndexName},
	{Name: "datname", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgStatProgressCreateIndexName},
	{Name: "relid", Type: pgtypes.Oid, Default: nil, Nullable: true, Source: PgStatProgressCreateIndexName},
	{Name: "index_relid", Type: pgtypes.Oid, Default: nil, Nullable: true, Source: PgStatProgressCreateIndexName},
	{Name: "command", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgStatProgressCreateIndexName},
	{Name: "phase", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgStatProgressCreateIndexName},
	{Name: "lockers_total", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatProgressCreateIndexName},
	{Name: "lockers_done", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatProgressCreateIndexName},
	{Name: "current_locker_pid", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatProgressCreateIndexName},
	{Name: "blocks_total", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatProgressCreateIndexName},
	{Name: "blocks_done", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatProgressCreateIndexName},
	{Name: "tuples_total", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatProgressCreateIndexName},
	{Name: "tuples_done", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatProgressCreateIndexName},
	{Name: "partitions_total", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatProgressCreateIndexName},
	{Name: "partitions_done", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatProgressCreateIndexName},
}

// pgStatProgressCreateIndexRowIter is the sql.RowIter for the pg_stat_progress_create_index table.
type pgStatProgressCreateIndexRowIter struct {
}

var _ sql.RowIter = (*pgStatProgressCreateIndexRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgStatProgressCreateIndexRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgStatProgressCreateIndexRowIter) Close(ctx *sql.Context) error {
	return nil
}
