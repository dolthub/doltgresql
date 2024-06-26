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

// PgStatReplicationSlotsName is a constant to the pg_stat_replication_slots name.
const PgStatReplicationSlotsName = "pg_stat_replication_slots"

// InitPgStatReplicationSlots handles registration of the pg_stat_replication_slots handler.
func InitPgStatReplicationSlots() {
	tables.AddHandler(PgCatalogName, PgStatReplicationSlotsName, PgStatReplicationSlotsHandler{})
}

// PgStatReplicationSlotsHandler is the handler for the pg_stat_replication_slots table.
type PgStatReplicationSlotsHandler struct{}

var _ tables.Handler = PgStatReplicationSlotsHandler{}

// Name implements the interface tables.Handler.
func (p PgStatReplicationSlotsHandler) Name() string {
	return PgStatReplicationSlotsName
}

// RowIter implements the interface tables.Handler.
func (p PgStatReplicationSlotsHandler) RowIter(ctx *sql.Context) (sql.RowIter, error) {
	// TODO: Implement pg_stat_replication_slots row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgStatReplicationSlotsHandler) Schema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgStatReplicationSlotsSchema,
		PkOrdinals: nil,
	}
}

// pgStatReplicationSlotsSchema is the schema for pg_stat_replication_slots.
var pgStatReplicationSlotsSchema = sql.Schema{
	{Name: "slot_name", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgStatReplicationSlotsName},
	{Name: "spill_txns", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatReplicationSlotsName},
	{Name: "spill_count", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatReplicationSlotsName},
	{Name: "spill_bytes", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatReplicationSlotsName},
	{Name: "stream_txns", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatReplicationSlotsName},
	{Name: "stream_count", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatReplicationSlotsName},
	{Name: "stream_bytes", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatReplicationSlotsName},
	{Name: "total_txns", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatReplicationSlotsName},
	{Name: "total_bytes", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatReplicationSlotsName},
	{Name: "stats_reset", Type: pgtypes.TimestampTZ, Default: nil, Nullable: true, Source: PgStatReplicationSlotsName},
}

// pgStatReplicationSlotsRowIter is the sql.RowIter for the pg_stat_replication_slots table.
type pgStatReplicationSlotsRowIter struct {
}

var _ sql.RowIter = (*pgStatReplicationSlotsRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgStatReplicationSlotsRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgStatReplicationSlotsRowIter) Close(ctx *sql.Context) error {
	return nil
}
