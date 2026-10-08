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

// PgStatDatabaseConflictsName is a constant to the pg_stat_database_conflicts name.
const PgStatDatabaseConflictsName = "pg_stat_database_conflicts"

// InitPgStatDatabaseConflicts handles registration of the pg_stat_database_conflicts handler.
func InitPgStatDatabaseConflicts() {
	tables.AddHandler(PgCatalogName, PgStatDatabaseConflictsName, PgStatDatabaseConflictsHandler{})
}

// PgStatDatabaseConflictsHandler is the handler for the pg_stat_database_conflicts table.
type PgStatDatabaseConflictsHandler struct{}

var _ tables.Handler = PgStatDatabaseConflictsHandler{}

// Name implements the interface tables.Handler.
func (p PgStatDatabaseConflictsHandler) Name() string {
	return PgStatDatabaseConflictsName
}

// RowIter implements the interface tables.Handler.
func (p PgStatDatabaseConflictsHandler) RowIter(ctx *sql.Context, partition sql.Partition) (sql.RowIter, error) {
	// Doltgres does not track recovery conflict statistics, so all counters are zero,
	// matching what a freshly-started Postgres server reports.
	// TODO: fill in real values when recovery conflict statistics are tracked
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

	rows := make([]sql.Row, 0, len(dbs))
	for _, db := range dbs {
		rows = append(rows, sql.Row{
			id.NewDatabase(db.Name()).AsId(), // datid
			db.Name(),                        // datname
			int64(0),                         // confl_tablespace
			int64(0),                         // confl_lock
			int64(0),                         // confl_snapshot
			int64(0),                         // confl_bufferpin
			int64(0),                         // confl_deadlock
			int64(0),                         // confl_active_logicalslot
		})
	}
	return sql.RowsToRowIter(rows...), nil
}

// PkSchema implements the interface tables.Handler.
func (p PgStatDatabaseConflictsHandler) PkSchema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgStatDatabaseConflictsSchema,
		PkOrdinals: nil,
	}
}

// pgStatDatabaseConflictsSchema is the schema for pg_stat_database_conflicts.
var pgStatDatabaseConflictsSchema = sql.Schema{
	{Name: "datid", Type: pgtypes.Oid, Default: nil, Nullable: true, Source: PgStatDatabaseConflictsName},
	{Name: "datname", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgStatDatabaseConflictsName},
	{Name: "confl_tablespace", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatDatabaseConflictsName},
	{Name: "confl_lock", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatDatabaseConflictsName},
	{Name: "confl_snapshot", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatDatabaseConflictsName},
	{Name: "confl_bufferpin", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatDatabaseConflictsName},
	{Name: "confl_deadlock", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatDatabaseConflictsName},
	{Name: "confl_active_logicalslot", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatDatabaseConflictsName},
}
