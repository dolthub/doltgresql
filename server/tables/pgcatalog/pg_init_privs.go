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

// PgInitPrivsName is a constant to the pg_init_privs name.
const PgInitPrivsName = "pg_init_privs"

// InitPgInitPrivs handles registration of the pg_init_privs handler.
func InitPgInitPrivs() {
	tables.AddHandler(PgCatalogName, PgInitPrivsName, PgInitPrivsHandler{})
}

// PgInitPrivsHandler is the handler for the pg_init_privs table.
type PgInitPrivsHandler struct{}

var _ tables.Handler = PgInitPrivsHandler{}

// Name implements the interface tables.Handler.
func (p PgInitPrivsHandler) Name() string {
	return PgInitPrivsName
}

// RowIter implements the interface tables.Handler.
func (p PgInitPrivsHandler) RowIter(ctx *sql.Context, partition sql.Partition) (sql.RowIter, error) {
	// TODO: Implement pg_init_privs row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgInitPrivsHandler) PkSchema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgInitPrivsSchema,
		PkOrdinals: nil,
	}
}

// pgInitPrivsSchema is the schema for pg_init_privs.
var pgInitPrivsSchema = sql.Schema{
	{Name: "objoid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgInitPrivsName},
	{Name: "classoid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgInitPrivsName},
	{Name: "objsubid", Type: pgtypes.Int32, Default: nil, Nullable: false, Source: PgInitPrivsName},
	{Name: "privtype", Type: pgtypes.InternalChar, Default: nil, Nullable: false, Source: PgInitPrivsName},
	{Name: "initprivs", Type: pgtypes.TextArray, Default: nil, Nullable: false, Source: PgInitPrivsName}, // TODO: aclitem[] type
}

// pgInitPrivsRowIter is the sql.RowIter for the pg_init_privs table.
type pgInitPrivsRowIter struct {
}

var _ sql.RowIter = (*pgInitPrivsRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgInitPrivsRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgInitPrivsRowIter) Close(ctx *sql.Context) error {
	return nil
}
