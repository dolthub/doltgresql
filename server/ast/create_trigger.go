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
	"fmt"
	"regexp"

	"github.com/cockroachdb/errors"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/core/triggers"
	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
	pgnodes "github.com/dolthub/doltgresql/server/node"
	"github.com/dolthub/doltgresql/server/plpgsql"
)

// createTriggerWhenCapture is a regex that should only capture the contents of the WHEN expression. Although a bit
// complex, this is done to ensure that the capture group contains only the WHEN expression and nothing else.
var createTriggerWhenCapture = regexp.MustCompile(`(?is)create\s+(?:or\s+replace\s+)?(?:constraint\s+)?trigger\s+.*\s+for\s+(?:each\s+)?(?:row|statement)\s+when\s+\((.*)\)\s+execute\s+(?:function|procedure).*`)

// nodeCreateTrigger handles *tree.CreateTrigger nodes.
func nodeCreateTrigger(ctx *Context, node *tree.CreateTrigger) (_ vitess.Statement, err error) {
	if node.Constraint {
		return NotYetSupportedError("CREATE CONSTRAINT TRIGGER is not yet supported")
	}
	if !node.RefTable.IsEmpty() {
		return NotYetSupportedError("FROM for CREATE TRIGGER is not yet supported")
	}
	if node.Deferrable != tree.TriggerNotDeferrable {
		return NotYetSupportedError("DEFERRABLE for CREATE TRIGGER is not yet supported")
	}
	if len(node.Relations) > 0 {
		return NotYetSupportedError("REFERENCING for CREATE TRIGGER is not yet supported")
	}
	if !node.ForEachRow {
		return NotYetSupportedError("FOR EACH STATEMENT for CREATE TRIGGER is not yet supported")
	}
	funcName := node.FuncName.ToTableName()
	var timing triggers.TriggerTiming
	switch node.Time {
	case tree.TriggerTimeBefore:
		timing = triggers.TriggerTiming_Before
	case tree.TriggerTimeAfter:
		timing = triggers.TriggerTiming_After
	case tree.TriggerTimeInsteadOf:
		return NotYetSupportedError("INSTEAD OF for CREATE TRIGGER is not yet supported")
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
				return NotYetSupportedError("UPDATE specific columns for CREATE TRIGGER are not yet supported")
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
			return NotYetSupportedError("TRUNCATE for CREATE TRIGGER is not yet supported")
		default:
			return NotYetSupportedError("UNKNOWN EVENT TYPE for CREATE TRIGGER is not yet supported")
		}
	}
	// WHEN expressions seem to behave identically to interpreted functions, so we'll parse them as interpreted functions.
	// To do this, we need the raw string, and we wrap it as though it were a trigger function (which has special logic
	// for handling NEW and OLD rows). Using a regex for this rather than modifying the parser may seem suboptimal, but
	// we want to retain the parser validation of using an expression, however we cannot rely on the expression's
	// String() function to return the **exact** same string, so we capture it with a regex.
	var whenOps []plpgsql.InterpreterOperation
	if node.When != nil {
		matches := createTriggerWhenCapture.FindStringSubmatch(ctx.originalQuery)
		if len(matches) != 2 {
			return nil, errors.New("unable to parse WHEN expression from CREATE TRIGGER")
		}
		whenOps, err = plpgsql.Parse(fmt.Sprintf(`CREATE FUNCTION when_wrapper() RETURNS TRIGGER AS $$
BEGIN
	RETURN %s;
END;
$$ LANGUAGE plpgsql;`, matches[1]))
		if err != nil {
			return nil, err
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
			whenOps,
			node.Args.ToStrings(),
			ctx.originalQuery,
		),
		Children: nil,
	}, nil
}
