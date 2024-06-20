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

// PgStatProgressVacuumName is a constant to the pg_stat_progress_vacuum name.
const PgStatProgressVacuumName = "pg_stat_progress_vacuum"

// InitPgStatProgressVacuum handles registration of the pg_stat_progress_vacuum handler.
func InitPgStatProgressVacuum() {
	tables.AddHandler(PgCatalogName, PgStatProgressVacuumName, PgStatProgressVacuumHandler{})
}

// PgStatProgressVacuumHandler is the handler for the pg_stat_progress_vacuum table.
type PgStatProgressVacuumHandler struct{}

var _ tables.Handler = PgStatProgressVacuumHandler{}

// Name implements the interface tables.Handler.
func (p PgStatProgressVacuumHandler) Name() string {
	return PgStatProgressVacuumName
}

// RowIter implements the interface tables.Handler.
func (p PgStatProgressVacuumHandler) RowIter(ctx *sql.Context) (sql.RowIter, error) {
	// TODO: Implement pg_stat_progress_vacuum row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgStatProgressVacuumHandler) Schema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgStatProgressVacuumSchema,
		PkOrdinals: nil,
	}
}

// pgStatProgressVacuumSchema is the schema for pg_stat_progress_vacuum.
var pgStatProgressVacuumSchema = sql.Schema{
	{Name: "pid", Type: pgtypes.Int32, Default: nil, Nullable: true, Source: PgStatProgressVacuumName},
	{Name: "datid", Type: pgtypes.Oid, Default: nil, Nullable: true, Source: PgStatProgressVacuumName},
	{Name: "datname", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgStatProgressVacuumName},
	{Name: "relid", Type: pgtypes.Oid, Default: nil, Nullable: true, Source: PgStatProgressVacuumName},
	{Name: "phase", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgStatProgressVacuumName},
	{Name: "heap_blks_total", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatProgressVacuumName},
	{Name: "heap_blks_scanned", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatProgressVacuumName},
	{Name: "heap_blks_vacuumed", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatProgressVacuumName},
	{Name: "index_vacuum_count", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatProgressVacuumName},
	{Name: "max_dead_tuples", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatProgressVacuumName},
	{Name: "num_dead_tuples", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatProgressVacuumName},
}

// pgStatProgressVacuumRowIter is the sql.RowIter for the pg_stat_progress_vacuum table.
type pgStatProgressVacuumRowIter struct {
}

var _ sql.RowIter = (*pgStatProgressVacuumRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgStatProgressVacuumRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgStatProgressVacuumRowIter) Close(ctx *sql.Context) error {
	return nil
}
