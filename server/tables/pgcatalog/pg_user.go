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

// PgUserName is a constant to the pg_user name.
const PgUserName = "pg_user"

// InitPgUser handles registration of the pg_user handler.
func InitPgUser() {
	tables.AddHandler(PgCatalogName, PgUserName, PgUserHandler{})
}

// PgUserHandler is the handler for the pg_user table.
type PgUserHandler struct{}

var _ tables.Handler = PgUserHandler{}

// Name implements the interface tables.Handler.
func (p PgUserHandler) Name() string {
	return PgUserName
}

// RowIter implements the interface tables.Handler.
func (p PgUserHandler) RowIter(ctx *sql.Context) (sql.RowIter, error) {
	// TODO: Implement pg_user row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgUserHandler) Schema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgUserSchema,
		PkOrdinals: nil,
	}
}

// pgUserSchema is the schema for pg_user.
var pgUserSchema = sql.Schema{
	{Name: "usename", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgUserName},
	{Name: "usesysid", Type: pgtypes.Oid, Default: nil, Nullable: true, Source: PgUserName},
	{Name: "usecreatedb", Type: pgtypes.Bool, Default: nil, Nullable: true, Source: PgUserName},
	{Name: "usesuper", Type: pgtypes.Bool, Default: nil, Nullable: true, Source: PgUserName},
	{Name: "userepl", Type: pgtypes.Bool, Default: nil, Nullable: true, Source: PgUserName},
	{Name: "usebypassrls", Type: pgtypes.Bool, Default: nil, Nullable: true, Source: PgUserName},
	{Name: "passwd", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgUserName},
	{Name: "valuntil", Type: pgtypes.TimestampTZ, Default: nil, Nullable: true, Source: PgUserName},
	{Name: "useconfig", Type: pgtypes.TextArray, Default: nil, Nullable: true, Source: PgUserName}, // TODO: collation C
}

// pgUserRowIter is the sql.RowIter for the pg_user table.
type pgUserRowIter struct {
}

var _ sql.RowIter = (*pgUserRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgUserRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgUserRowIter) Close(ctx *sql.Context) error {
	return nil
}
