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

// PgReplicationOriginName is a constant to the pg_replication_origin name.
const PgReplicationOriginName = "pg_replication_origin"

// InitPgReplicationOrigin handles registration of the pg_replication_origin handler.
func InitPgReplicationOrigin() {
	tables.AddHandler(PgCatalogName, PgReplicationOriginName, PgReplicationOriginHandler{})
}

// PgReplicationOriginHandler is the handler for the pg_replication_origin table.
type PgReplicationOriginHandler struct{}

var _ tables.Handler = PgReplicationOriginHandler{}

// Name implements the interface tables.Handler.
func (p PgReplicationOriginHandler) Name() string {
	return PgReplicationOriginName
}

// RowIter implements the interface tables.Handler.
func (p PgReplicationOriginHandler) RowIter(ctx *sql.Context, partition sql.Partition) (sql.RowIter, error) {
	// TODO: Implement pg_replication_origin row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgReplicationOriginHandler) PkSchema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgReplicationOriginSchema,
		PkOrdinals: nil,
	}
}

// pgReplicationOriginSchema is the schema for pg_replication_origin.
var pgReplicationOriginSchema = sql.Schema{
	{Name: "roident", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgReplicationOriginName},
	{Name: "roname", Type: pgtypes.Text, Default: nil, Nullable: false, Source: PgReplicationOriginName}, // TODO: collation C
}

// pgReplicationOriginRowIter is the sql.RowIter for the pg_replication_origin table.
type pgReplicationOriginRowIter struct {
}

var _ sql.RowIter = (*pgReplicationOriginRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgReplicationOriginRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgReplicationOriginRowIter) Close(ctx *sql.Context) error {
	return nil
}
