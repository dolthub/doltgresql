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
	"net"
	"sort"
	"strconv"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/core/id"
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
	// One row per client connection, from the engine's process list. Sessions between queries show as 'idle',
	// matching Postgres.
	if ctx.ProcessList == nil {
		return emptyRowIter()
	}
	processes := ctx.ProcessList.Processes()
	sort.Slice(processes, func(i, j int) bool {
		return processes[i].Connection < processes[j].Connection
	})
	return &pgStatActivityRowIter{
		processes: processes,
		idx:       0,
	}, nil
}

// PkSchema implements the interface tables.Handler.
func (p PgStatActivityHandler) PkSchema() sql.PrimaryKeySchema {
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
	processes []sql.Process
	idx       int
}

var _ sql.RowIter = (*pgStatActivityRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgStatActivityRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	if iter.idx >= len(iter.processes) {
		return nil, io.EOF
	}
	iter.idx++
	process := iter.processes[iter.idx-1]

	var datid any
	var datname any
	if process.Database != "" {
		datid = id.NewDatabase(process.Database).AsId()
		datname = process.Database
	}
	var usesysid any
	var usename any
	if process.User != "" {
		usesysid = roleOid(process.User)
		usename = process.User
	}
	// The host is usually in address:port form, but may be a bare address
	var clientAddr any
	var clientPort any
	if host, port, err := net.SplitHostPort(process.Host); err == nil {
		clientAddr = host
		if portNum, err := strconv.Atoi(port); err == nil {
			clientPort = int32(portNum)
		}
	} else if process.Host != "" {
		clientAddr = process.Host
	}
	state := "idle"
	query := process.Query
	var queryStart any
	if process.Command == sql.ProcessCommandQuery {
		state = "active"
		queryStart = process.StartedAt
	}

	return sql.Row{
		datid,                     // datid
		datname,                   // datname
		int32(process.Connection), // pid
		nil,                       // leader_pid
		usesysid,                  // usesysid
		usename,                   // usename
		"",                        // application_name (TODO: not tracked)
		clientAddr,                // client_addr
		nil,                       // client_hostname
		clientPort,                // client_port
		nil,                       // backend_start (TODO: connection start time is not tracked)
		nil,                       // xact_start (TODO: transaction start time is not tracked)
		queryStart,                // query_start
		nil,                       // state_change (TODO: not tracked)
		nil,                       // wait_event_type
		nil,                       // wait_event
		state,                     // state
		nil,                       // backend_xid
		nil,                       // backend_xmin
		nil,                       // query_id
		query,                     // query
		"client backend",          // backend_type
	}, nil
}

// Close implements the interface sql.RowIter.
func (iter *pgStatActivityRowIter) Close(ctx *sql.Context) error {
	return nil
}
