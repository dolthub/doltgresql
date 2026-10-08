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
	"github.com/cockroachdb/apd/v3"
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/server/tables"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// PgStatWalName is a constant to the pg_stat_wal name.
const PgStatWalName = "pg_stat_wal"

// InitPgStatWal handles registration of the pg_stat_wal handler.
func InitPgStatWal() {
	tables.AddHandler(PgCatalogName, PgStatWalName, PgStatWalHandler{})
}

// PgStatWalHandler is the handler for the pg_stat_wal table.
type PgStatWalHandler struct{}

var _ tables.Handler = PgStatWalHandler{}

// Name implements the interface tables.Handler.
func (p PgStatWalHandler) Name() string {
	return PgStatWalName
}

// RowIter implements the interface tables.Handler.
func (p PgStatWalHandler) RowIter(ctx *sql.Context, partition sql.Partition) (sql.RowIter, error) {
	// Doltgres does not track WAL activity statistics, so this returns a single row with zero
	// counters and a NULL stats_reset, matching what a freshly-started Postgres server reports.
	return sql.RowsToRowIter(sql.Row{
		int64(0),      // wal_records
		int64(0),      // wal_fpi
		apd.New(0, 0), // wal_bytes
		int64(0),      // wal_buffers_full
		int64(0),      // wal_write
		int64(0),      // wal_sync
		float64(0),    // wal_write_time
		float64(0),    // wal_sync_time
		nil,           // stats_reset
	}), nil
}

// PkSchema implements the interface tables.Handler.
func (p PgStatWalHandler) PkSchema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgStatWalSchema,
		PkOrdinals: nil,
	}
}

// pgStatWalSchema is the schema for pg_stat_wal.
var pgStatWalSchema = sql.Schema{
	{Name: "wal_records", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatWalName},
	{Name: "wal_fpi", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatWalName},
	{Name: "wal_bytes", Type: pgtypes.Numeric, Default: nil, Nullable: true, Source: PgStatWalName},
	{Name: "wal_buffers_full", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatWalName},
	{Name: "wal_write", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatWalName},
	{Name: "wal_sync", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatWalName},
	{Name: "wal_write_time", Type: pgtypes.Float64, Default: nil, Nullable: true, Source: PgStatWalName},
	{Name: "wal_sync_time", Type: pgtypes.Float64, Default: nil, Nullable: true, Source: PgStatWalName},
	{Name: "stats_reset", Type: pgtypes.TimestampTZ, Default: nil, Nullable: true, Source: PgStatWalName},
}
