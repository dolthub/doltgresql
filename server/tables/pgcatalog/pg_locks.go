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

// PgLocksName is a constant to the pg_locks name.
const PgLocksName = "pg_locks"

// InitPgLocks handles registration of the pg_locks handler.
func InitPgLocks() {
	tables.AddHandler(PgCatalogName, PgLocksName, PgLocksHandler{})
}

// PgLocksHandler is the handler for the pg_locks table.
type PgLocksHandler struct{}

var _ tables.Handler = PgLocksHandler{}

// Name implements the interface tables.Handler.
func (p PgLocksHandler) Name() string {
	return PgLocksName
}

// RowIter implements the interface tables.Handler.
func (p PgLocksHandler) RowIter(ctx *sql.Context) (sql.RowIter, error) {
	// TODO: Implement pg_locks row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgLocksHandler) Schema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgLocksSchema,
		PkOrdinals: nil,
	}
}

// pgLocksSchema is the schema for pg_locks.
var pgLocksSchema = sql.Schema{
	{Name: "locktype", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgLocksName},
	{Name: "database", Type: pgtypes.Oid, Default: nil, Nullable: true, Source: PgLocksName},
	{Name: "relation", Type: pgtypes.Oid, Default: nil, Nullable: true, Source: PgLocksName},
	{Name: "page", Type: pgtypes.Int32, Default: nil, Nullable: true, Source: PgLocksName},
	{Name: "tuple", Type: pgtypes.Int16, Default: nil, Nullable: true, Source: PgLocksName},
	{Name: "virtualxid", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgLocksName},
	{Name: "transactionid", Type: pgtypes.Xid, Default: nil, Nullable: true, Source: PgLocksName},
	{Name: "classid", Type: pgtypes.Oid, Default: nil, Nullable: true, Source: PgLocksName},
	{Name: "objid", Type: pgtypes.Oid, Default: nil, Nullable: true, Source: PgLocksName},
	{Name: "objsubid", Type: pgtypes.Int16, Default: nil, Nullable: true, Source: PgLocksName},
	{Name: "virtualtransaction", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgLocksName},
	{Name: "pid", Type: pgtypes.Int32, Default: nil, Nullable: true, Source: PgLocksName},
	{Name: "mode", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgLocksName},
	{Name: "granted", Type: pgtypes.Bool, Default: nil, Nullable: true, Source: PgLocksName},
	{Name: "fastpath", Type: pgtypes.Bool, Default: nil, Nullable: true, Source: PgLocksName},
	{Name: "waitstart", Type: pgtypes.TimestampTZ, Default: nil, Nullable: true, Source: PgLocksName},
}

// pgLocksRowIter is the sql.RowIter for the pg_locks table.
type pgLocksRowIter struct {
}

var _ sql.RowIter = (*pgLocksRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgLocksRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgLocksRowIter) Close(ctx *sql.Context) error {
	return nil
}
