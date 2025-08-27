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

// PgStatSubscriptionStatsName is a constant to the pg_stat_subscription_stats name.
const PgStatSubscriptionStatsName = "pg_stat_subscription_stats"

// InitPgStatSubscriptionStats handles registration of the pg_stat_subscription_stats handler.
func InitPgStatSubscriptionStats() {
	tables.AddHandler(PgCatalogName, PgStatSubscriptionStatsName, PgStatSubscriptionStatsHandler{})
}

// PgStatSubscriptionStatsHandler is the handler for the pg_stat_subscription_stats table.
type PgStatSubscriptionStatsHandler struct{}

var _ tables.Handler = PgStatSubscriptionStatsHandler{}

// Name implements the interface tables.Handler.
func (p PgStatSubscriptionStatsHandler) Name() string {
	return PgStatSubscriptionStatsName
}

// RowIter implements the interface tables.Handler.
func (p PgStatSubscriptionStatsHandler) RowIter(ctx *sql.Context, partition sql.Partition) (sql.RowIter, error) {
	// TODO: Implement pg_stat_subscription_stats row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgStatSubscriptionStatsHandler) Schema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgStatSubscriptionStatsSchema,
		PkOrdinals: nil,
	}
}

// pgStatSubscriptionStatsSchema is the schema for pg_stat_subscription_stats.
var pgStatSubscriptionStatsSchema = sql.Schema{
	{Name: "subid", Type: pgtypes.Oid, Default: nil, Nullable: true, Source: PgStatSubscriptionStatsName},
	{Name: "subname", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgStatSubscriptionStatsName},
	{Name: "apply_error_count", Type: pgtypes.Int32, Default: nil, Nullable: true, Source: PgStatSubscriptionStatsName},
	{Name: "sync_error_count", Type: pgtypes.Int32, Default: nil, Nullable: true, Source: PgStatSubscriptionStatsName},
	{Name: "stats_reset", Type: pgtypes.TimestampTZ, Default: nil, Nullable: true, Source: PgStatSubscriptionStatsName},
}

// pgStatSubscriptionStatsRowIter is the sql.RowIter for the pg_stat_subscription_stats table.
type pgStatSubscriptionStatsRowIter struct {
}

var _ sql.RowIter = (*pgStatSubscriptionStatsRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgStatSubscriptionStatsRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgStatSubscriptionStatsRowIter) Close(ctx *sql.Context) error {
	return nil
}
