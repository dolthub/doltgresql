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

// PgStatRecoveryPrefetchName is a constant to the pg_stat_recovery_prefetch name.
const PgStatRecoveryPrefetchName = "pg_stat_recovery_prefetch"

// InitPgStatRecoveryPrefetch handles registration of the pg_stat_recovery_prefetch handler.
func InitPgStatRecoveryPrefetch() {
	tables.AddHandler(PgCatalogName, PgStatRecoveryPrefetchName, PgStatRecoveryPrefetchHandler{})
}

// PgStatRecoveryPrefetchHandler is the handler for the pg_stat_recovery_prefetch table.
type PgStatRecoveryPrefetchHandler struct{}

var _ tables.Handler = PgStatRecoveryPrefetchHandler{}

// Name implements the interface tables.Handler.
func (p PgStatRecoveryPrefetchHandler) Name() string {
	return PgStatRecoveryPrefetchName
}

// RowIter implements the interface tables.Handler.
func (p PgStatRecoveryPrefetchHandler) RowIter(ctx *sql.Context, partition sql.Partition) (sql.RowIter, error) {
	// TODO: Implement pg_stat_recovery_prefetch row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgStatRecoveryPrefetchHandler) PkSchema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgStatRecoveryPrefetchSchema,
		PkOrdinals: nil,
	}
}

// pgStatRecoveryPrefetchSchema is the schema for pg_stat_recovery_prefetch.
var pgStatRecoveryPrefetchSchema = sql.Schema{
	{Name: "stats_reset", Type: pgtypes.TimestampTZ, Default: nil, Nullable: true, Source: PgStatRecoveryPrefetchName},
	{Name: "prefetch", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatRecoveryPrefetchName},
	{Name: "hit", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatRecoveryPrefetchName},
	{Name: "skip_init", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatRecoveryPrefetchName},
	{Name: "skip_new", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatRecoveryPrefetchName},
	{Name: "skip_fpw", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatRecoveryPrefetchName},
	{Name: "skip_rep", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatRecoveryPrefetchName},
	{Name: "wal_distance", Type: pgtypes.Int32, Default: nil, Nullable: true, Source: PgStatRecoveryPrefetchName},
	{Name: "block_distance", Type: pgtypes.Int32, Default: nil, Nullable: true, Source: PgStatRecoveryPrefetchName},
	{Name: "io_depth", Type: pgtypes.Int32, Default: nil, Nullable: true, Source: PgStatRecoveryPrefetchName},
}

// pgStatRecoveryPrefetchRowIter is the sql.RowIter for the pg_stat_recovery_prefetch table.
type pgStatRecoveryPrefetchRowIter struct {
}

var _ sql.RowIter = (*pgStatRecoveryPrefetchRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgStatRecoveryPrefetchRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgStatRecoveryPrefetchRowIter) Close(ctx *sql.Context) error {
	return nil
}
