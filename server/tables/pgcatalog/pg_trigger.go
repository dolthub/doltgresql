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

// PgTriggerName is a constant to the pg_trigger name.
const PgTriggerName = "pg_trigger"

// InitPgTrigger handles registration of the pg_trigger handler.
func InitPgTrigger() {
	tables.AddHandler(PgCatalogName, PgTriggerName, PgTriggerHandler{})
}

// PgTriggerHandler is the handler for the pg_trigger table.
type PgTriggerHandler struct{}

var _ tables.Handler = PgTriggerHandler{}

// Name implements the interface tables.Handler.
func (p PgTriggerHandler) Name() string {
	return PgTriggerName
}

// RowIter implements the interface tables.Handler.
func (p PgTriggerHandler) RowIter(ctx *sql.Context) (sql.RowIter, error) {
	// TODO: Implement pg_trigger row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgTriggerHandler) Schema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     pgTriggerSchema,
		PkOrdinals: nil,
	}
}

// pgTriggerSchema is the schema for pg_trigger.
var pgTriggerSchema = sql.Schema{
	{Name: "oid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgTriggerName},
	{Name: "tgrelid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgTriggerName},
	{Name: "tgparentid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgTriggerName},
	{Name: "tgname", Type: pgtypes.Name, Default: nil, Nullable: false, Source: PgTriggerName},
	{Name: "tgfoid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgTriggerName},
	{Name: "tgtype", Type: pgtypes.Int16, Default: nil, Nullable: false, Source: PgTriggerName},
	{Name: "tgenabled", Type: pgtypes.BpChar, Default: nil, Nullable: false, Source: PgTriggerName},
	{Name: "tgisinternal", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgTriggerName},
	{Name: "tgconstrrelid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgTriggerName},
	{Name: "tgconstrindid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgTriggerName},
	{Name: "tgconstraint", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgTriggerName},
	{Name: "tgdeferrable", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgTriggerName},
	{Name: "tginitdeferred", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgTriggerName},
	{Name: "tgnargs", Type: pgtypes.Int16, Default: nil, Nullable: false, Source: PgTriggerName},
	{Name: "tgattr", Type: pgtypes.Int16, Default: nil, Nullable: false, Source: PgTriggerName}, // TODO: type int2vector
	{Name: "tgargs", Type: pgtypes.Bytea, Default: nil, Nullable: false, Source: PgTriggerName},
	{Name: "tgqual", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgTriggerName}, // TODO: type pg_node_tree, collation C
	{Name: "tgoldtable", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgTriggerName},
	{Name: "tgnewtable", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgTriggerName},
}

// pgTriggerRowIter is the sql.RowIter for the pg_trigger table.
type pgTriggerRowIter struct {
	idx int
}

var _ sql.RowIter = (*pgTriggerRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgTriggerRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgTriggerRowIter) Close(ctx *sql.Context) error {
	return nil
}
