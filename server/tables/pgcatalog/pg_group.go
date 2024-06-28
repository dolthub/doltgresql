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

// PgGroupName is a constant to the pg_group name.
const PgGroupName = "pg_group"

// InitPgGroup handles registration of the pg_group handler.
func InitPgGroup() {
	tables.AddHandler(PgCatalogName, PgGroupName, PgGroupHandler{})
}

// PgGroupHandler is the handler for the pg_group table.
type PgGroupHandler struct{}

var _ tables.Handler = PgGroupHandler{}

// Name implements the interface tables.Handler.
func (p PgGroupHandler) Name() string {
	return PgGroupName
}

// RowIter implements the interface tables.Handler.
func (p PgGroupHandler) RowIter(ctx *sql.Context) (sql.RowIter, error) {
	// TODO: Implement pg_group row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgGroupHandler) Schema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgGroupSchema,
		PkOrdinals: nil,
	}
}

// pgGroupSchema is the schema for pg_group.
var pgGroupSchema = sql.Schema{
	{Name: "groname", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgGroupName},
	{Name: "grosysid", Type: pgtypes.Oid, Default: nil, Nullable: true, Source: PgGroupName},
	{Name: "grolist", Type: pgtypes.OidArray, Default: nil, Nullable: true, Source: PgGroupName},
}

// pgGroupRowIter is the sql.RowIter for the pg_group table.
type pgGroupRowIter struct {
}

var _ sql.RowIter = (*pgGroupRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgGroupRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgGroupRowIter) Close(ctx *sql.Context) error {
	return nil
}
