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

// PgEventTriggerName is a constant to the pg_event_trigger name.
const PgEventTriggerName = "pg_event_trigger"

// InitPgEventTrigger handles registration of the pg_event_trigger handler.
func InitPgEventTrigger() {
	tables.AddHandler(PgCatalogName, PgEventTriggerName, PgEventTriggerHandler{})
}

// PgEventTriggerHandler is the handler for the pg_event_trigger table.
type PgEventTriggerHandler struct{}

var _ tables.Handler = PgEventTriggerHandler{}

// Name implements the interface tables.Handler.
func (p PgEventTriggerHandler) Name() string {
	return PgEventTriggerName
}

// RowIter implements the interface tables.Handler.
func (p PgEventTriggerHandler) RowIter(ctx *sql.Context) (sql.RowIter, error) {
	// TODO: Implement pg_event_trigger row iter
	return emptyRowIter()
}

// Schema implements the interface tables.Handler.
func (p PgEventTriggerHandler) Schema() sql.PrimaryKeySchema {
	return sql.PrimaryKeySchema{
		Schema:     PgEventTriggerSchema,
		PkOrdinals: nil,
	}
}

// PgEventTriggerSchema is the schema for pg_event_trigger.
var PgEventTriggerSchema = sql.Schema{
	{Name: "oid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgEventTriggerName},
	{Name: "evtname", Type: pgtypes.Name, Default: nil, Nullable: false, Source: PgEventTriggerName},
	{Name: "evtevent", Type: pgtypes.Name, Default: nil, Nullable: false, Source: PgEventTriggerName},
	{Name: "evtowner", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgEventTriggerName},
	{Name: "evtfoid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgEventTriggerName},
	{Name: "evtenabled", Type: pgtypes.InternalChar, Default: nil, Nullable: false, Source: PgEventTriggerName},
	{Name: "evttags", Type: pgtypes.TextArray, Default: nil, Nullable: true, Source: PgEventTriggerName}, // TODO: collation C
}

// pgEventTriggerRowIter is the sql.RowIter for the pg_event_trigger table.
type pgEventTriggerRowIter struct {
}

var _ sql.RowIter = (*pgEventTriggerRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgEventTriggerRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	return nil, io.EOF
}

// Close implements the interface sql.RowIter.
func (iter *pgEventTriggerRowIter) Close(ctx *sql.Context) error {
	return nil
}
