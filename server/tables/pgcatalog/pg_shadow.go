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

// PgShadowName is a constant to the pg_shadow name.
const PgShadowName = "pg_shadow"

// InitPgShadow handles registration of the pg_shadow handler.
func InitPgShadow() {
	tables.AddHandler(PgCatalogName, PgShadowName, PgShadowHandler{})
}

// PgShadowHandler is the handler for the pg_shadow table.
type PgShadowHandler struct{}

var _ tables.Handler = PgShadowHandler{}

// Name implements the interface tables.Handler.
func (p PgShadowHandler) Name() string {
	return PgShadowName
}

// RowIter implements the interface tables.Handler.
func (p PgShadowHandler) RowIter(ctx *sql.Context, partition sql.Partition) (sql.RowIter, error) {
	// TODO: Implement pg_shadow row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgShadowHandler) PkSchema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgShadowSchema,
		PkOrdinals: nil,
	}
}

// pgShadowSchema is the schema for pg_shadow.
var pgShadowSchema = sql.Schema{
	{Name: "usename", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgShadowName},
	{Name: "usesysid", Type: pgtypes.Oid, Default: nil, Nullable: true, Source: PgShadowName},
	{Name: "usecreatedb", Type: pgtypes.Bool, Default: nil, Nullable: true, Source: PgShadowName},
	{Name: "usesuper", Type: pgtypes.Bool, Default: nil, Nullable: true, Source: PgShadowName},
	{Name: "userepl", Type: pgtypes.Bool, Default: nil, Nullable: true, Source: PgShadowName},
	{Name: "usebypassrls", Type: pgtypes.Bool, Default: nil, Nullable: true, Source: PgShadowName},
	{Name: "passwd", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgShadowName}, // TODO: collation C
	{Name: "valuntil", Type: pgtypes.TimestampTZ, Default: nil, Nullable: true, Source: PgShadowName},
	{Name: "useconfig", Type: pgtypes.TimeArray, Default: nil, Nullable: true, Source: PgShadowName}, // TODO: collation C
}

// pgShadowRowIter is the sql.RowIter for the pg_shadow table.
type pgShadowRowIter struct {
}

var _ sql.RowIter = (*pgShadowRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgShadowRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgShadowRowIter) Close(ctx *sql.Context) error {
	return nil
}
