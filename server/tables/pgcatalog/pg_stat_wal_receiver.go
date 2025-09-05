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

// PgStatWalReceiverName is a constant to the pg_stat_wal_receiver name.
const PgStatWalReceiverName = "pg_stat_wal_receiver"

// InitPgStatWalReceiver handles registration of the pg_stat_wal_receiver handler.
func InitPgStatWalReceiver() {
	tables.AddHandler(PgCatalogName, PgStatWalReceiverName, PgStatWalReceiverHandler{})
}

// PgStatWalReceiverHandler is the handler for the pg_stat_wal_receiver table.
type PgStatWalReceiverHandler struct{}

var _ tables.Handler = PgStatWalReceiverHandler{}

// Name implements the interface tables.Handler.
func (p PgStatWalReceiverHandler) Name() string {
	return PgStatWalReceiverName
}

// RowIter implements the interface tables.Handler.
func (p PgStatWalReceiverHandler) RowIter(ctx *sql.Context, partition sql.Partition) (sql.RowIter, error) {
	// TODO: Implement pg_stat_wal_receiver row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgStatWalReceiverHandler) PkSchema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgStatWalReceiverSchema,
		PkOrdinals: nil,
	}
}

// pgStatWalReceiverSchema is the schema for pg_stat_wal_receiver.
var pgStatWalReceiverSchema = sql.Schema{
	{Name: "pid", Type: pgtypes.Int32, Default: nil, Nullable: true, Source: PgStatWalReceiverName},
	{Name: "status", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgStatWalReceiverName},
	{Name: "receive_start_lsn", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgStatWalReceiverName}, // TODO: pg_lsn type
	{Name: "receive_start_tli", Type: pgtypes.Int32, Default: nil, Nullable: true, Source: PgStatWalReceiverName},
	{Name: "written_lsn", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgStatWalReceiverName}, // TODO: pg_lsn type
	{Name: "flushed_lsn", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgStatWalReceiverName}, // TODO: pg_lsn type
	{Name: "received_tli", Type: pgtypes.Int32, Default: nil, Nullable: true, Source: PgStatWalReceiverName},
	{Name: "last_msg_send_time", Type: pgtypes.TimestampTZ, Default: nil, Nullable: true, Source: PgStatWalReceiverName},
	{Name: "last_msg_receipt_time", Type: pgtypes.TimestampTZ, Default: nil, Nullable: true, Source: PgStatWalReceiverName},
	{Name: "latest_end_lsn", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgStatWalReceiverName}, // TODO: pg_lsn type
	{Name: "latest_end_time", Type: pgtypes.TimestampTZ, Default: nil, Nullable: true, Source: PgStatWalReceiverName},
	{Name: "slot_name", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgStatWalReceiverName},
	{Name: "sender_host", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgStatWalReceiverName},
	{Name: "sender_port", Type: pgtypes.Int32, Default: nil, Nullable: true, Source: PgStatWalReceiverName},
	{Name: "conninfo", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgStatWalReceiverName},
}

// pgStatWalReceiverRowIter is the sql.RowIter for the pg_stat_wal_receiver table.
type pgStatWalReceiverRowIter struct {
}

var _ sql.RowIter = (*pgStatWalReceiverRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgStatWalReceiverRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgStatWalReceiverRowIter) Close(ctx *sql.Context) error {
	return nil
}
