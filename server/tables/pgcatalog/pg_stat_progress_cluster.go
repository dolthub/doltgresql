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

// PgStatProgressClusterName is a constant to the pg_stat_progress_cluster name.
const PgStatProgressClusterName = "pg_stat_progress_cluster"

// InitPgStatProgressCluster handles registration of the pg_stat_progress_cluster handler.
func InitPgStatProgressCluster() {
	tables.AddHandler(PgCatalogName, PgStatProgressClusterName, PgStatProgressClusterHandler{})
}

// PgStatProgressClusterHandler is the handler for the pg_stat_progress_cluster table.
type PgStatProgressClusterHandler struct{}

var _ tables.Handler = PgStatProgressClusterHandler{}

// Name implements the interface tables.Handler.
func (p PgStatProgressClusterHandler) Name() string {
	return PgStatProgressClusterName
}

// RowIter implements the interface tables.Handler.
func (p PgStatProgressClusterHandler) RowIter(ctx *sql.Context) (sql.RowIter, error) {
	// TODO: Implement pg_stat_progress_cluster row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgStatProgressClusterHandler) Schema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgStatProgressClusterSchema,
		PkOrdinals: nil,
	}
}

// pgStatProgressClusterSchema is the schema for pg_stat_progress_cluster.
var pgStatProgressClusterSchema = sql.Schema{
	{Name: "pid", Type: pgtypes.Int32, Default: nil, Nullable: true, Source: PgStatProgressClusterName},
	{Name: "datid", Type: pgtypes.Oid, Default: nil, Nullable: true, Source: PgStatProgressClusterName},
	{Name: "datname", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgStatProgressClusterName},
	{Name: "relid", Type: pgtypes.Oid, Default: nil, Nullable: true, Source: PgStatProgressClusterName},
	{Name: "command", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgStatProgressClusterName},
	{Name: "phase", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgStatProgressClusterName},
	{Name: "cluster_index_relid", Type: pgtypes.Oid, Default: nil, Nullable: true, Source: PgStatProgressClusterName},
	{Name: "heap_tuples_scanned", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatProgressClusterName},
	{Name: "heap_tuples_written", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatProgressClusterName},
	{Name: "heap_blks_total", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatProgressClusterName},
	{Name: "heap_blks_scanned", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatProgressClusterName},
	{Name: "index_rebuild_count", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatProgressClusterName},
}

// pgStatProgressClusterRowIter is the sql.RowIter for the pg_stat_progress_cluster table.
type pgStatProgressClusterRowIter struct {
}

var _ sql.RowIter = (*pgStatProgressClusterRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgStatProgressClusterRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgStatProgressClusterRowIter) Close(ctx *sql.Context) error {
	return nil
}
