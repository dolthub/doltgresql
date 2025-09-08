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

// PgReplicationOriginStatusName is a constant to the pg_replication_origin_status name.
const PgReplicationOriginStatusName = "pg_replication_origin_status"

// InitPgReplicationOriginStatus handles registration of the pg_replication_origin_status handler.
func InitPgReplicationOriginStatus() {
	tables.AddHandler(PgCatalogName, PgReplicationOriginStatusName, PgReplicationOriginStatusHandler{})
}

// PgReplicationOriginStatusHandler is the handler for the pg_replication_origin_status table.
type PgReplicationOriginStatusHandler struct{}

var _ tables.Handler = PgReplicationOriginStatusHandler{}

// Name implements the interface tables.Handler.
func (p PgReplicationOriginStatusHandler) Name() string {
	return PgReplicationOriginStatusName
}

// RowIter implements the interface tables.Handler.
func (p PgReplicationOriginStatusHandler) RowIter(ctx *sql.Context, partition sql.Partition) (sql.RowIter, error) {
	// TODO: Implement pg_replication_origin_status row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgReplicationOriginStatusHandler) PkSchema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgReplicationOriginStatusSchema,
		PkOrdinals: nil,
	}
}

// pgReplicationOriginStatusSchema is the schema for pg_replication_origin_status.
var pgReplicationOriginStatusSchema = sql.Schema{
	{Name: "local_id", Type: pgtypes.Oid, Default: nil, Nullable: true, Source: PgReplicationOriginStatusName},
	{Name: "external_id", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgReplicationOriginStatusName},
	{Name: "remote_lsn", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgReplicationOriginStatusName}, // TODO: pg_lsn type
	{Name: "local_lsn", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgReplicationOriginStatusName},  // TODO: pg_lsn type
}

// pgReplicationOriginStatusRowIter is the sql.RowIter for the pg_replication_origin_status table.
type pgReplicationOriginStatusRowIter struct {
}

var _ sql.RowIter = (*pgReplicationOriginStatusRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgReplicationOriginStatusRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgReplicationOriginStatusRowIter) Close(ctx *sql.Context) error {
	return nil
}
