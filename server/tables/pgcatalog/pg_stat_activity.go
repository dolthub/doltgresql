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

// PgStatActivityName is a constant to the pg_stat_activity name.
const PgStatActivityName = "pg_stat_activity"

// InitPgStatActivity handles registration of the pg_stat_activity handler.
func InitPgStatActivity() {
	tables.AddHandler(PgCatalogName, PgStatActivityName, PgStatActivityHandler{})
}

// PgStatActivityHandler is the handler for the pg_stat_activity table.
type PgStatActivityHandler struct{}

var _ tables.Handler = PgStatActivityHandler{}

// Name implements the interface tables.Handler.
func (p PgStatActivityHandler) Name() string {
	return PgStatActivityName
}

// RowIter implements the interface tables.Handler.
func (p PgStatActivityHandler) RowIter(ctx *sql.Context, partition sql.Partition) (sql.RowIter, error) {
	// TODO: Implement pg_stat_activity row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgStatActivityHandler) Schema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgStatActivitySchema,
		PkOrdinals: nil,
	}
}

// pgStatActivitySchema is the schema for pg_stat_activity.
var pgStatActivitySchema = sql.Schema{
	{Name: "datid", Type: pgtypes.Oid, Default: nil, Nullable: true, Source: PgStatActivityName},
	{Name: "datname", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgStatActivityName},
	{Name: "pid", Type: pgtypes.Int32, Default: nil, Nullable: true, Source: PgStatActivityName},
	{Name: "leader_pid", Type: pgtypes.Int32, Default: nil, Nullable: true, Source: PgStatActivityName},
	{Name: "usesysid", Type: pgtypes.Oid, Default: nil, Nullable: true, Source: PgStatActivityName},
	{Name: "usename", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgStatActivityName},
	{Name: "application_name", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgStatActivityName},
	{Name: "client_addr", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgStatActivityName}, // TODO: inet type
	{Name: "client_hostname", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgStatActivityName},
	{Name: "client_port", Type: pgtypes.Int32, Default: nil, Nullable: true, Source: PgStatActivityName},
	{Name: "backend_start", Type: pgtypes.TimestampTZ, Default: nil, Nullable: true, Source: PgStatActivityName},
	{Name: "xact_start", Type: pgtypes.TimestampTZ, Default: nil, Nullable: true, Source: PgStatActivityName},
	{Name: "query_start", Type: pgtypes.TimestampTZ, Default: nil, Nullable: true, Source: PgStatActivityName},
	{Name: "state_change", Type: pgtypes.TimestampTZ, Default: nil, Nullable: true, Source: PgStatActivityName},
	{Name: "wait_event_type", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgStatActivityName},
	{Name: "wait_event", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgStatActivityName},
	{Name: "state", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgStatActivityName},
	{Name: "backend_xid", Type: pgtypes.Xid, Default: nil, Nullable: true, Source: PgStatActivityName},
	{Name: "backend_xmin", Type: pgtypes.Xid, Default: nil, Nullable: true, Source: PgStatActivityName},
	{Name: "query_id", Type: pgtypes.Int64, Default: nil, Nullable: true, Source: PgStatActivityName},
	{Name: "query", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgStatActivityName},
	{Name: "backend_type", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgStatActivityName},
}

// pgStatActivityRowIter is the sql.RowIter for the pg_stat_activity table.
type pgStatActivityRowIter struct {
}

var _ sql.RowIter = (*pgStatActivityRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgStatActivityRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgStatActivityRowIter) Close(ctx *sql.Context) error {
	return nil
}
