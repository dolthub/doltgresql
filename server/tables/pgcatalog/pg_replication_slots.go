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

// PgReplicationSlotsName is a constant to the pg_replication_slots name.
const PgReplicationSlotsName = "pg_replication_slots"

// InitPgReplicationSlots handles registration of the pg_replication_slots handler.
func InitPgReplicationSlots() {
	tables.AddHandler(PgCatalogName, PgReplicationSlotsName, PgReplicationSlotsHandler{})
}

// PgReplicationSlotsHandler is the handler for the pg_replication_slots table.
type PgReplicationSlotsHandler struct{}

var _ tables.Handler = PgReplicationSlotsHandler{}

// Name implements the interface tables.Handler.
func (p PgReplicationSlotsHandler) Name() string {
	return PgReplicationSlotsName
}

// RowIter implements the interface tables.Handler.
func (p PgReplicationSlotsHandler) RowIter(ctx *sql.Context) (sql.RowIter, error) {
	// TODO: Implement pg_replication_slots row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgReplicationSlotsHandler) Schema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgReplicationSlotsSchema,
		PkOrdinals: nil,
	}
}

// pgReplicationSlotsSchema is the schema for pg_replication_slots.
var pgReplicationSlotsSchema = sql.Schema{
	{Name: "slot_name", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgReplicationSlotsName},
	{Name: "plugin", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgReplicationSlotsName},
	{Name: "slot_type", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgReplicationSlotsName},
	{Name: "datoid", Type: pgtypes.Oid, Default: nil, Nullable: true, Source: PgReplicationSlotsName},
	{Name: "database", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgReplicationSlotsName},
	{Name: "temporary", Type: pgtypes.Bool, Default: nil, Nullable: true, Source: PgReplicationSlotsName},
	{Name: "active", Type: pgtypes.Bool, Default: nil, Nullable: true, Source: PgReplicationSlotsName},
	{Name: "active_pid", Type: pgtypes.Int32, Default: nil, Nullable: true, Source: PgReplicationSlotsName},
	{Name: "xmin", Type: pgtypes.Xid, Default: nil, Nullable: true, Source: PgReplicationSlotsName},
	{Name: "catalog_xmin", Type: pgtypes.Xid, Default: nil, Nullable: true, Source: PgReplicationSlotsName},
	{Name: "restart_lsn", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgReplicationSlotsName},         // TODO: pg_lsn type
	{Name: "confirmed_flush_lsn", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgReplicationSlotsName}, // TODO: pg_lsn type
	{Name: "wal_status", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgReplicationSlotsName},
	{Name: "safe_wal_size", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgReplicationSlotsName},
	{Name: "two_phase", Type: pgtypes.Bool, Default: nil, Nullable: true, Source: PgReplicationSlotsName},
}

// pgReplicationSlotsRowIter is the sql.RowIter for the pg_replication_slots table.
type pgReplicationSlotsRowIter struct {
}

var _ sql.RowIter = (*pgReplicationSlotsRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgReplicationSlotsRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgReplicationSlotsRowIter) Close(ctx *sql.Context) error {
	return nil
}
