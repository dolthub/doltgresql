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

package ast

import (
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/core/triggers"
	pgnodes "github.com/dolthub/doltgresql/server/node"

	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
)

// nodeCreateTrigger handles *tree.CreateTrigger nodes.
func nodeCreateTrigger(ctx *Context, node *tree.CreateTrigger) (vitess.Statement, error) {
	if node.Constraint {
		return NotYetSupportedError("CREATE CONSTRAINT TRIGGER is not yet supported")
	}
	if !node.RefTable.IsEmpty() {
		return NotYetSupportedError("FROM is not yet supported for CREATE TRIGGER")
	}
	if node.Deferrable != tree.TriggerNotDeferrable {
		return NotYetSupportedError("DEFERRABLE is not yet supported for CREATE TRIGGER")
	}
	if len(node.Relations) > 0 {
		return NotYetSupportedError("REFERENCING is not yet supported for CREATE TRIGGER")
	}
	if !node.ForEachRow {
		return NotYetSupportedError("FOR EACH STATEMENT is not yet supported for CREATE TRIGGER")
	}
	funcName := node.FuncName.ToTableName()
	var timing triggers.TriggerTiming
	switch node.Time {
	case tree.TriggerTimeBefore:
		timing = triggers.TriggerTiming_Before
	case tree.TriggerTimeAfter:
		timing = triggers.TriggerTiming_After
	case tree.TriggerTimeInsteadOf:
		return NotYetSupportedError("INSTEAD OF is not yet supported for CREATE TRIGGER")
	}
	var events []triggers.TriggerEvent
	for _, event := range node.Events {
		switch event.Type {
		case tree.TriggerEventInsert:
			events = append(events, triggers.TriggerEvent{
				Type: triggers.TriggerEventType_Insert,
			})
		case tree.TriggerEventUpdate:
			if len(event.Cols) > 0 {
				return NotYetSupportedError("UPDATE specific columns are not yet supported for CREATE TRIGGER")
			}
			events = append(events, triggers.TriggerEvent{
				Type:        triggers.TriggerEventType_Update,
				ColumnNames: event.Cols.ToStrings(),
			})
		case tree.TriggerEventDelete:
			events = append(events, triggers.TriggerEvent{
				Type: triggers.TriggerEventType_Delete,
			})
		case tree.TriggerEventTruncate:
			return NotYetSupportedError("TRUNCATE is not yet supported for CREATE TRIGGER")
		default:
			return NotYetSupportedError("UNKNOWN EVENT TYPE is not yet supported for CREATE TRIGGER")
		}
	}
	return vitess.InjectedStatement{
		Statement: pgnodes.NewCreateTrigger(
			id.NewTrigger(node.OnTable.Schema(), node.OnTable.Table(), node.Name.String()),
			id.NewFunction(funcName.Schema(), funcName.Table()),
			node.Replace,
			timing,
			events,
			node.ForEachRow,
			nil, // TODO: node.When (expr)
			node.Args.ToStrings(),
			ctx.originalQuery,
		),
		Children: nil,
	}, nil
}
