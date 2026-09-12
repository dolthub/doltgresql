// Copyright 2026 Dolthub, Inc.
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

package information_schema

import (
	"fmt"
	"sort"
	"strings"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/information_schema"
	"github.com/lib/pq"

	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/core/triggers"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// newTriggersTable creates a new information_schema.TRIGGERS table.
func newTriggersTable() *information_schema.InformationSchemaTable {
	return &information_schema.InformationSchemaTable{
		TableName:   information_schema.TriggersTableName,
		TableSchema: triggersSchema,
		Reader:      triggersRowIter,
	}
}

// triggersSchema is the schema for the information_schema.TRIGGERS table.
var triggersSchema = sql.Schema{
	{Name: "trigger_catalog", Type: sql_identifier, Default: nil, Nullable: true, Source: information_schema.TriggersTableName},
	{Name: "trigger_schema", Type: sql_identifier, Default: nil, Nullable: true, Source: information_schema.TriggersTableName},
	{Name: "trigger_name", Type: sql_identifier, Default: nil, Nullable: true, Source: information_schema.TriggersTableName},
	{Name: "event_manipulation", Type: character_data, Default: nil, Nullable: true, Source: information_schema.TriggersTableName},
	{Name: "event_object_catalog", Type: sql_identifier, Default: nil, Nullable: true, Source: information_schema.TriggersTableName},
	{Name: "event_object_schema", Type: sql_identifier, Default: nil, Nullable: true, Source: information_schema.TriggersTableName},
	{Name: "event_object_table", Type: sql_identifier, Default: nil, Nullable: true, Source: information_schema.TriggersTableName},
	{Name: "action_order", Type: cardinal_number, Default: nil, Nullable: true, Source: information_schema.TriggersTableName},
	{Name: "action_condition", Type: character_data, Default: nil, Nullable: true, Source: information_schema.TriggersTableName},
	{Name: "action_statement", Type: character_data, Default: nil, Nullable: true, Source: information_schema.TriggersTableName},
	{Name: "action_orientation", Type: character_data, Default: nil, Nullable: true, Source: information_schema.TriggersTableName},
	{Name: "action_timing", Type: character_data, Default: nil, Nullable: true, Source: information_schema.TriggersTableName},
	{Name: "action_reference_old_table", Type: sql_identifier, Default: nil, Nullable: true, Source: information_schema.TriggersTableName},
	{Name: "action_reference_new_table", Type: sql_identifier, Default: nil, Nullable: true, Source: information_schema.TriggersTableName},
	{Name: "action_reference_old_row", Type: sql_identifier, Default: nil, Nullable: true, Source: information_schema.TriggersTableName},
	{Name: "action_reference_new_row", Type: sql_identifier, Default: nil, Nullable: true, Source: information_schema.TriggersTableName},
	{Name: "created", Type: pgtypes.TimestampTZ, Default: nil, Nullable: true, Source: information_schema.TriggersTableName},
}

// triggersRowIter implements the sql.RowIter for the information_schema.TRIGGERS table.
func triggersRowIter(ctx *sql.Context, catalog sql.Catalog) (sql.RowIter, error) {
	collection, err := core.GetTriggersCollectionFromContext(ctx, "")
	if err != nil {
		return nil, err
	}
	var trigs []triggers.Trigger
	err = collection.IterateTriggers(ctx, func(t triggers.Trigger) (stop bool, err error) {
		trigs = append(trigs, t)
		return false, nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(trigs, func(i, j int) bool {
		if trigs[i].ID.SchemaName() != trigs[j].ID.SchemaName() {
			return trigs[i].ID.SchemaName() < trigs[j].ID.SchemaName()
		}
		if trigs[i].ID.TableName() != trigs[j].ID.TableName() {
			return trigs[i].ID.TableName() < trigs[j].ID.TableName()
		}
		return trigs[i].ID.TriggerName() < trigs[j].ID.TriggerName()
	})

	catName := ctx.GetCurrentDatabase()
	actionOrders := make(map[string]int32)
	var rows []sql.Row
	for _, t := range trigs {
		orientation := "STATEMENT"
		if t.ForEachRow {
			orientation = "ROW"
		}
		timing := triggerTiming(t.Timing)
		args := make([]string, len(t.Arguments))
		for i, arg := range t.Arguments {
			args[i] = pq.QuoteLiteral(arg)
		}
		statement := fmt.Sprintf("EXECUTE FUNCTION %s(%s)", t.Function.FunctionName(), strings.Join(args, ", "))
		var oldTable any
		if len(t.OldTransitionName) > 0 {
			oldTable = t.OldTransitionName
		}
		var newTable any
		if len(t.NewTransitionName) > 0 {
			newTable = t.NewTransitionName
		}
		for _, event := range t.Events {
			manipulation := triggerEventManipulation(event.Type)
			if len(manipulation) == 0 {
				continue
			}
			orderKey := strings.Join([]string{t.ID.SchemaName(), t.ID.TableName(), manipulation, orientation, timing}, "\x00")
			actionOrders[orderKey]++
			//TODO: action_condition needs the WHEN condition, which is only kept in its compiled form
			rows = append(rows, sql.Row{
				catName,                // trigger_catalog
				t.ID.SchemaName(),      // trigger_schema
				t.ID.TriggerName(),     // trigger_name
				manipulation,           // event_manipulation
				catName,                // event_object_catalog
				t.ID.SchemaName(),      // event_object_schema
				t.ID.TableName(),       // event_object_table
				actionOrders[orderKey], // action_order
				nil,                    // action_condition
				statement,              // action_statement
				orientation,            // action_orientation
				timing,                 // action_timing
				oldTable,               // action_reference_old_table
				newTable,               // action_reference_new_table
				nil,                    // action_reference_old_row
				nil,                    // action_reference_new_row
				nil,                    // created
			})
		}
	}
	return sql.RowsToRowIter(rows...), nil
}

// triggerTiming returns the action_timing value for the given trigger timing.
func triggerTiming(timing triggers.TriggerTiming) string {
	switch timing {
	case triggers.TriggerTiming_Before:
		return "BEFORE"
	case triggers.TriggerTiming_After:
		return "AFTER"
	default:
		return "INSTEAD OF"
	}
}

// triggerEventManipulation returns the event_manipulation value for the given event type, which is empty for TRUNCATE.
func triggerEventManipulation(eventType triggers.TriggerEventType) string {
	switch eventType {
	case triggers.TriggerEventType_Insert:
		return "INSERT"
	case triggers.TriggerEventType_Update:
		return "UPDATE"
	case triggers.TriggerEventType_Delete:
		return "DELETE"
	default:
		return ""
	}
}
