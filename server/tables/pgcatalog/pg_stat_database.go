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
	"sort"

	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/dsess"
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/server/tables"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// PgStatDatabaseName is a constant to the pg_stat_database name.
const PgStatDatabaseName = "pg_stat_database"

// InitPgStatDatabase handles registration of the pg_stat_database handler.
func InitPgStatDatabase() {
	tables.AddHandler(PgCatalogName, PgStatDatabaseName, PgStatDatabaseHandler{})
}

// PgStatDatabaseHandler is the handler for the pg_stat_database table.
type PgStatDatabaseHandler struct{}

var _ tables.Handler = PgStatDatabaseHandler{}

// Name implements the interface tables.Handler.
func (p PgStatDatabaseHandler) Name() string {
	return PgStatDatabaseName
}

// RowIter implements the interface tables.Handler.
func (p PgStatDatabaseHandler) RowIter(ctx *sql.Context, partition sql.Partition) (sql.RowIter, error) {
	// Doltgres does not track database statistics yet, so all counters are zero and all
	// timestamps are NULL, matching what a freshly-started Postgres server reports.
	// TODO: fill in real values when database statistics are tracked
	doltSession := dsess.DSessFromSess(ctx.Session)
	databases := doltSession.Provider().AllDatabases(ctx)
	dbs := make([]sql.Database, 0, len(databases))
	for _, db := range databases {
		name := db.Name()
		if name == "information_schema" || name == "pg_catalog" || name == "performance_schema" {
			continue
		}
		dbs = append(dbs, db)
	}
	sort.Slice(dbs, func(i, j int) bool {
		return dbs[i].Name() < dbs[j].Name()
	})

	// The first row reports stats for shared objects, with a zero OID and a NULL database name.
	rows := make([]sql.Row, 0, len(dbs)+1)
	rows = append(rows, pgStatDatabaseRow(id.Null, nil))
	for _, db := range dbs {
		rows = append(rows, pgStatDatabaseRow(id.NewDatabase(db.Name()).AsId(), db.Name()))
	}
	return sql.RowsToRowIter(rows...), nil
}

// pgStatDatabaseRow returns a pg_stat_database row for the given database OID and name, with all
// counters zero and all timestamps NULL.
func pgStatDatabaseRow(datid id.Id, datname any) sql.Row {
	return sql.Row{
		datid,      // datid
		datname,    // datname
		int32(0),   // numbackends
		int64(0),   // xact_commit
		int64(0),   // xact_rollback
		int64(0),   // blks_read
		int64(0),   // blks_hit
		int64(0),   // tup_returned
		int64(0),   // tup_fetched
		int64(0),   // tup_inserted
		int64(0),   // tup_updated
		int64(0),   // tup_deleted
		int64(0),   // conflicts
		int64(0),   // temp_files
		int64(0),   // temp_bytes
		int64(0),   // deadlocks
		int64(0),   // checksum_failures
		nil,        // checksum_last_failure
		float64(0), // blk_read_time
		float64(0), // blk_write_time
		float64(0), // session_time
		float64(0), // active_time
		float64(0), // idle_in_transaction_time
		int64(0),   // sessions
		int64(0),   // sessions_abandoned
		int64(0),   // sessions_fatal
		int64(0),   // sessions_killed
		nil,        // stats_reset
	}
}

// PkSchema implements the interface tables.Handler.
func (p PgStatDatabaseHandler) PkSchema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgStatDatabaseSchema,
		PkOrdinals: nil,
	}
}

// pgStatDatabaseSchema is the schema for pg_stat_database.
var pgStatDatabaseSchema = sql.Schema{
	{Name: "datid", Type: pgtypes.Oid, Default: nil, Nullable: true, Source: PgStatDatabaseName},
	{Name: "datname", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgStatDatabaseName},
	{Name: "numbackends", Type: pgtypes.Int32, Default: nil, Nullable: true, Source: PgStatDatabaseName},
	{Name: "xact_commit", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatDatabaseName},
	{Name: "xact_rollback", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatDatabaseName},
	{Name: "blks_read", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatDatabaseName},
	{Name: "blks_hit", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatDatabaseName},
	{Name: "tup_returned", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatDatabaseName},
	{Name: "tup_fetched", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatDatabaseName},
	{Name: "tup_inserted", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatDatabaseName},
	{Name: "tup_updated", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatDatabaseName},
	{Name: "tup_deleted", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatDatabaseName},
	{Name: "conflicts", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatDatabaseName},
	{Name: "temp_files", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatDatabaseName},
	{Name: "temp_bytes", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatDatabaseName},
	{Name: "deadlocks", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatDatabaseName},
	{Name: "checksum_failures", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatDatabaseName},
	{Name: "checksum_last_failure", Type: pgtypes.TimestampTZ, Default: nil, Nullable: true, Source: PgStatDatabaseName},
	{Name: "blk_read_time", Type: pgtypes.Float64, Default: nil, Nullable: true, Source: PgStatDatabaseName},
	{Name: "blk_write_time", Type: pgtypes.Float64, Default: nil, Nullable: true, Source: PgStatDatabaseName},
	{Name: "session_time", Type: pgtypes.Float64, Default: nil, Nullable: true, Source: PgStatDatabaseName},
	{Name: "active_time", Type: pgtypes.Float64, Default: nil, Nullable: true, Source: PgStatDatabaseName},
	{Name: "idle_in_transaction_time", Type: pgtypes.Float64, Default: nil, Nullable: true, Source: PgStatDatabaseName},
	{Name: "sessions", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatDatabaseName},
	{Name: "sessions_abandoned", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatDatabaseName},
	{Name: "sessions_fatal", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatDatabaseName},
	{Name: "sessions_killed", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatDatabaseName},
	{Name: "stats_reset", Type: pgtypes.TimestampTZ, Default: nil, Nullable: true, Source: PgStatDatabaseName},
}
