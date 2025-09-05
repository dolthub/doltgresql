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

// PgPreparedXactsName is a constant to the pg_prepared_xacts name.
const PgPreparedXactsName = "pg_prepared_xacts"

// InitPgPreparedXacts handles registration of the pg_prepared_xacts handler.
func InitPgPreparedXacts() {
	tables.AddHandler(PgCatalogName, PgPreparedXactsName, PgPreparedXactsHandler{})
}

// PgPreparedXactsHandler is the handler for the pg_prepared_xacts table.
type PgPreparedXactsHandler struct{}

var _ tables.Handler = PgPreparedXactsHandler{}

// Name implements the interface tables.Handler.
func (p PgPreparedXactsHandler) Name() string {
	return PgPreparedXactsName
}

// RowIter implements the interface tables.Handler.
func (p PgPreparedXactsHandler) RowIter(ctx *sql.Context, partition sql.Partition) (sql.RowIter, error) {
	// TODO: Implement pg_prepared_xacts row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgPreparedXactsHandler) PkSchema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgPreparedXactsSchema,
		PkOrdinals: nil,
	}
}

// pgPreparedXactsSchema is the schema for pg_prepared_xacts.
var pgPreparedXactsSchema = sql.Schema{
	{Name: "transaction", Type: pgtypes.Xid, Default: nil, Nullable: true, Source: PgPreparedXactsName},
	{Name: "gid", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgPreparedXactsName},
	{Name: "prepared", Type: pgtypes.TimestampTZ, Default: nil, Nullable: true, Source: PgPreparedXactsName},
	{Name: "owner", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgPreparedXactsName},
	{Name: "database", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgPreparedXactsName},
}

// pgPreparedXactsRowIter is the sql.RowIter for the pg_prepared_xacts table.
type pgPreparedXactsRowIter struct {
}

var _ sql.RowIter = (*pgPreparedXactsRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgPreparedXactsRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgPreparedXactsRowIter) Close(ctx *sql.Context) error {
	return nil
}
