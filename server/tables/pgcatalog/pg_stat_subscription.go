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

// PgStatSubscriptionName is a constant to the pg_stat_subscription name.
const PgStatSubscriptionName = "pg_stat_subscription"

// InitPgStatSubscription handles registration of the pg_stat_subscription handler.
func InitPgStatSubscription() {
	tables.AddHandler(PgCatalogName, PgStatSubscriptionName, PgStatSubscriptionHandler{})
}

// PgStatSubscriptionHandler is the handler for the pg_stat_subscription table.
type PgStatSubscriptionHandler struct{}

var _ tables.Handler = PgStatSubscriptionHandler{}

// Name implements the interface tables.Handler.
func (p PgStatSubscriptionHandler) Name() string {
	return PgStatSubscriptionName
}

// RowIter implements the interface tables.Handler.
func (p PgStatSubscriptionHandler) RowIter(ctx *sql.Context, partition sql.Partition) (sql.RowIter, error) {
	// TODO: Implement pg_stat_subscription row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgStatSubscriptionHandler) Schema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgStatSubscriptionSchema,
		PkOrdinals: nil,
	}
}

// pgStatSubscriptionSchema is the schema for pg_stat_subscription.
var pgStatSubscriptionSchema = sql.Schema{
	{Name: "subid", Type: pgtypes.Oid, Default: nil, Nullable: true, Source: PgStatSubscriptionName},
	{Name: "subname", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgStatSubscriptionName},
	{Name: "pid", Type: pgtypes.Int32, Default: nil, Nullable: true, Source: PgStatSubscriptionName},
	{Name: "leader_pid", Type: pgtypes.Int32, Default: nil, Nullable: true, Source: PgStatSubscriptionName},
	{Name: "relid", Type: pgtypes.Oid, Default: nil, Nullable: true, Source: PgStatSubscriptionName},
	{Name: "received_lsn", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgStatSubscriptionName}, // TODO: pg_lsn type
	{Name: "last_msg_send_time", Type: pgtypes.TimestampTZ, Default: nil, Nullable: true, Source: PgStatSubscriptionName},
	{Name: "last_msg_receipt_time", Type: pgtypes.TimestampTZ, Default: nil, Nullable: true, Source: PgStatSubscriptionName},
	{Name: "latest_end_lsn", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgStatSubscriptionName}, // TODO: pg_lsn type
	{Name: "latest_end_time", Type: pgtypes.TimestampTZ, Default: nil, Nullable: true, Source: PgStatSubscriptionName},
}

// pgStatSubscriptionRowIter is the sql.RowIter for the pg_stat_subscription table.
type pgStatSubscriptionRowIter struct {
}

var _ sql.RowIter = (*pgStatSubscriptionRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgStatSubscriptionRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgStatSubscriptionRowIter) Close(ctx *sql.Context) error {
	return nil
}
