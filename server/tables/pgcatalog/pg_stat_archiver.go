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
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/server/tables"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// PgStatArchiverName is a constant to the pg_stat_archiver name.
const PgStatArchiverName = "pg_stat_archiver"

// InitPgStatArchiver handles registration of the pg_stat_archiver handler.
func InitPgStatArchiver() {
	tables.AddHandler(PgCatalogName, PgStatArchiverName, PgStatArchiverHandler{})
}

// PgStatArchiverHandler is the handler for the pg_stat_archiver table.
type PgStatArchiverHandler struct{}

var _ tables.Handler = PgStatArchiverHandler{}

// Name implements the interface tables.Handler.
func (p PgStatArchiverHandler) Name() string {
	return PgStatArchiverName
}

// RowIter implements the interface tables.Handler.
func (p PgStatArchiverHandler) RowIter(ctx *sql.Context, partition sql.Partition) (sql.RowIter, error) {
	// Doltgres does not have a WAL archiver process, so this returns a single row with zero
	// counters and NULL values, matching what a freshly-started Postgres server reports.
	return sql.RowsToRowIter(sql.Row{
		int64(0), // archived_count
		nil,      // last_archived_wal
		nil,      // last_archived_time
		int64(0), // failed_count
		nil,      // last_failed_wal
		nil,      // last_failed_time
		nil,      // stats_reset
	}), nil
}

// PkSchema implements the interface tables.Handler.
func (p PgStatArchiverHandler) PkSchema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgStatArchiverSchema,
		PkOrdinals: nil,
	}
}

// pgStatArchiverSchema is the schema for pg_stat_archiver.
var pgStatArchiverSchema = sql.Schema{
	{Name: "archived_count", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatArchiverName},
	{Name: "last_archived_wal", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgStatArchiverName},
	{Name: "last_archived_time", Type: pgtypes.TimestampTZ, Default: nil, Nullable: true, Source: PgStatArchiverName},
	{Name: "failed_count", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatArchiverName},
	{Name: "last_failed_wal", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgStatArchiverName},
	{Name: "last_failed_time", Type: pgtypes.TimestampTZ, Default: nil, Nullable: true, Source: PgStatArchiverName},
	{Name: "stats_reset", Type: pgtypes.TimestampTZ, Default: nil, Nullable: true, Source: PgStatArchiverName},
}
