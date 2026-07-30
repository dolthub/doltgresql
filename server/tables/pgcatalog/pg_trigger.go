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
	"sort"

	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/core/triggers"
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
func (p PgTriggerHandler) RowIter(ctx *sql.Context, partition sql.Partition) (sql.RowIter, error) {
	// Use cached data from this process if it exists
	pgCatalogCache, err := getPgCatalogCache(ctx)
	if err != nil {
		return nil, err
	}

	if pgCatalogCache.triggers == nil {
		err = cachePgTriggers(ctx, pgCatalogCache)
		if err != nil {
			return nil, err
		}
	}

	return &pgTriggerRowIter{
		triggers: pgCatalogCache.triggers,
		idx:      0,
	}, nil
}

// cachePgTriggers caches the pg_trigger data for the current database in the session.
func cachePgTriggers(ctx *sql.Context, pgCatalogCache *pgCatalogCache) error {
	collection, err := core.GetTriggersCollectionFromContext(ctx, ctx.GetCurrentDatabase())
	if err != nil {
		return err
	}
	var trigs []triggers.Trigger
	err = collection.IterateTriggers(ctx, func(t triggers.Trigger) (stop bool, err error) {
		trigs = append(trigs, t)
		return false, nil
	})
	if err != nil {
		return err
	}
	// Sort for deterministic output: by schema, then table, then trigger name
	sort.Slice(trigs, func(i, j int) bool {
		if trigs[i].ID.SchemaName() != trigs[j].ID.SchemaName() {
			return trigs[i].ID.SchemaName() < trigs[j].ID.SchemaName()
		}
		if trigs[i].ID.TableName() != trigs[j].ID.TableName() {
			return trigs[i].ID.TableName() < trigs[j].ID.TableName()
		}
		return trigs[i].ID.TriggerName() < trigs[j].ID.TriggerName()
	})
	pgCatalogCache.triggers = trigs
	return nil
}

// PkSchema implements the interface tables.Handler.
func (p PgTriggerHandler) PkSchema() sql.PrimaryKeySchema {
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
	{Name: "tgenabled", Type: pgtypes.InternalChar, Default: nil, Nullable: false, Source: PgTriggerName},
	{Name: "tgisinternal", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgTriggerName},
	{Name: "tgconstrrelid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgTriggerName},
	{Name: "tgconstrindid", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgTriggerName},
	{Name: "tgconstraint", Type: pgtypes.Oid, Default: nil, Nullable: false, Source: PgTriggerName},
	{Name: "tgdeferrable", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgTriggerName},
	{Name: "tginitdeferred", Type: pgtypes.Bool, Default: nil, Nullable: false, Source: PgTriggerName},
	{Name: "tgnargs", Type: pgtypes.Int16, Default: nil, Nullable: false, Source: PgTriggerName},
	{Name: "tgattr", Type: pgtypes.Int16vector, Default: nil, Nullable: false, Source: PgTriggerName},
	{Name: "tgargs", Type: pgtypes.Bytea, Default: nil, Nullable: false, Source: PgTriggerName},
	{Name: "tgqual", Type: pgtypes.Text, Default: nil, Nullable: true, Source: PgTriggerName}, // TODO: type pg_node_tree, collation C
	{Name: "tgoldtable", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgTriggerName},
	{Name: "tgnewtable", Type: pgtypes.Name, Default: nil, Nullable: true, Source: PgTriggerName},
}

// Trigger type bits for the tgtype column, matching Postgres' TRIGGER_TYPE_* defines.
const (
	triggerTypeRow      int16 = 1 << 0
	triggerTypeBefore   int16 = 1 << 1
	triggerTypeInsert   int16 = 1 << 2
	triggerTypeDelete   int16 = 1 << 3
	triggerTypeUpdate   int16 = 1 << 4
	triggerTypeTruncate int16 = 1 << 5
	triggerTypeInstead  int16 = 1 << 6
)

// triggerType computes the tgtype bitmask for the given trigger, following Postgres semantics.
func triggerType(t triggers.Trigger) int16 {
	var tgtype int16
	if t.ForEachRow {
		tgtype |= triggerTypeRow
	}
	switch t.Timing {
	case triggers.TriggerTiming_Before:
		tgtype |= triggerTypeBefore
	case triggers.TriggerTiming_InsteadOf:
		tgtype |= triggerTypeInstead
	}
	for _, event := range t.Events {
		switch event.Type {
		case triggers.TriggerEventType_Insert:
			tgtype |= triggerTypeInsert
		case triggers.TriggerEventType_Update:
			tgtype |= triggerTypeUpdate
		case triggers.TriggerEventType_Delete:
			tgtype |= triggerTypeDelete
		case triggers.TriggerEventType_Truncate:
			tgtype |= triggerTypeTruncate
		}
	}
	return tgtype
}

// triggerArgs encodes the trigger's arguments in the same form that Postgres uses for tgargs: each
// argument is followed by a NUL terminator.
func triggerArgs(args []string) []byte {
	encoded := make([]byte, 0, len(args)*8)
	for _, arg := range args {
		encoded = append(encoded, arg...)
		encoded = append(encoded, 0)
	}
	return encoded
}

// pgTriggerRowIter is the sql.RowIter for the pg_trigger table.
type pgTriggerRowIter struct {
	triggers []triggers.Trigger
	idx      int
}

var _ sql.RowIter = (*pgTriggerRowIter)(nil)

// Next implements the interface sql.RowIter.
func (iter *pgTriggerRowIter) Next(ctx *sql.Context) (sql.Row, error) {
	if iter.idx >= len(iter.triggers) {
		return nil, io.EOF
	}
	iter.idx++
	t := iter.triggers[iter.idx-1]

	tableOid := id.NewTable(t.ID.SchemaName(), t.ID.TableName()).AsId()
	constrRelOid := id.Null
	if t.ReferencedTableName.IsValid() {
		constrRelOid = t.ReferencedTableName.AsId()
	}
	var oldTable any
	if len(t.OldTransitionName) > 0 {
		oldTable = t.OldTransitionName
	}
	var newTable any
	if len(t.NewTransitionName) > 0 {
		newTable = t.NewTransitionName
	}

	return sql.Row{
		t.ID.AsId(),        // oid
		tableOid,           // tgrelid
		id.Null,            // tgparentid
		t.ID.TriggerName(), // tgname
		t.Function.AsId(),  // tgfoid
		triggerType(t),     // tgtype
		"O",                // tgenabled
		false,              // tgisinternal
		constrRelOid,       // tgconstrrelid
		id.Null,            // tgconstrindid
		id.Null,            // tgconstraint (TODO: constraint triggers do not yet have pg_constraint entries)
		t.Deferrable != triggers.TriggerDeferrable_NotDeferrable,      // tgdeferrable
		t.Deferrable == triggers.TriggerDeferrable_DeferrableDeferred, // tginitdeferred
		int16(len(t.Arguments)),  // tgnargs
		[]any{},                  // tgattr (TODO: column numbers for UPDATE OF column lists)
		triggerArgs(t.Arguments), // tgargs
		nil,                      // tgqual (TODO: node tree for the WHEN condition)
		oldTable,                 // tgoldtable
		newTable,                 // tgnewtable
	}, nil
}

// Close implements the interface sql.RowIter.
func (iter *pgTriggerRowIter) Close(ctx *sql.Context) error {
	return nil
}
