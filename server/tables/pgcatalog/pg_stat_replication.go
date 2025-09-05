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

// PgStatReplicationName is a constant to the pg_stat_replication name.
const PgStatReplicationName = "pg_stat_replication"

// InitPgStatReplication handles registration of the pg_stat_replication handler.
func InitPgStatReplication() {
	tables.AddHandler(PgCatalogName, PgStatReplicationName, PgStatReplicationHandler{})
}

// PgStatReplicationHandler is the handler for the pg_stat_replication table.
type PgStatReplicationHandler struct{}

var _ tables.Handler = PgStatReplicationHandler{}

// Name implements the interface tables.Handler.
func (p PgStatReplicationHandler) Name() string {
	return PgStatReplicationName
}

// RowIter implements the interface tables.Handler.
func (p PgStatReplicationHandler) RowIter(ctx *sql.Context, partition sql.Partition) (sql.RowIter, error) {
	// TODO: Implement pg_stat_replication row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgStatReplicationHandler) PkSchema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgStatReplicationSchema,
		PkOrdinals: nil,
	}
}

// pgStatReplicationSchema is the schema for pg_stat_replication.
var pgStatReplicationSchema = sql.Schema{
	{Name: "pid", Type: pgtypes.Int32, Default: nil, Nullable: true, Source: PgStatReplicationName},
	{Name: "usesysid", Type: pgtypes.Oid, Default: nil, Nullable: true, Source: PgStatReplicationName},
	{Name: "usename", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgStatReplicationName},
	{Name: "application_name", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgStatReplicationName},
	{Name: "client_addr", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgStatReplicationName}, // TODO: inet type
	{Name: "client_hostname", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgStatReplicationName},
	{Name: "client_port", Type: pgtypes.Int32, Default: nil, Nullable: true, Source: PgStatReplicationName},
	{Name: "backend_start", Type: pgtypes.TimestampTZ, Default: nil, Nullable: true, Source: PgStatReplicationName},
	{Name: "backend_xmin", Type: pgtypes.Xid, Default: nil, Nullable: true, Source: PgStatReplicationName},
	{Name: "state", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgStatReplicationName},
	{Name: "sent_lsn", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgStatReplicationName},   // TODO: pg_lsn type
	{Name: "write_lsn", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgStatReplicationName},  // TODO: pg_lsn type
	{Name: "flush_lsn", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgStatReplicationName},  // TODO: pg_lsn type
	{Name: "replay_lsn", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgStatReplicationName}, // TODO: pg_lsn type
	{Name: "write_lag", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgStatReplicationName},  // TODO: interval type
	{Name: "flush_lag", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgStatReplicationName},  // TODO: interval type
	{Name: "replay_lag", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgStatReplicationName}, // TODO: interval type
	{Name: "sync_priority", Type: pgtypes.Int32, Default: nil, Nullable: true, Source: PgStatReplicationName},
	{Name: "sync_state", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgStatReplicationName},
	{Name: "reply_time", Type: pgtypes.TimestampTZ, Default: nil, Nullable: true, Source: PgStatReplicationName},
}

// pgStatReplicationRowIter is the sql.RowIter for the pg_stat_replication table.
type pgStatReplicationRowIter struct {
}

var _ sql.RowIter = (*pgStatReplicationRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgStatReplicationRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgStatReplicationRowIter) Close(ctx *sql.Context) error {
	return nil
}
