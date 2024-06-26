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
func (p PgStatDatabaseConflictsHandler) RowIter(ctx *sql.Context) (sql.RowIter, error) {
	// TODO: Implement pg_stat_database_conflicts row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgStatDatabaseConflictsHandler) Schema() sql.PrimaryKeySchema {
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

// pgStatDatabaseConflictsRowIter is the sql.RowIter for the pg_stat_database_conflicts table.
type pgStatDatabaseConflictsRowIter struct {
}

var _ sql.RowIter = (*pgStatDatabaseConflictsRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgStatDatabaseConflictsRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgStatDatabaseConflictsRowIter) Close(ctx *sql.Context) error {
	return nil
}
