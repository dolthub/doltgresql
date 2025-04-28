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

package analyzer

import (
	"fmt"
	"sort"

	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/core/triggers"
	pgexprs "github.com/dolthub/doltgresql/server/expression"
	pgnodes "github.com/dolthub/doltgresql/server/node"
	pgtransform "github.com/dolthub/doltgresql/server/transform"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/analyzer"
	"github.com/dolthub/go-mysql-server/sql/plan"
	"github.com/dolthub/go-mysql-server/sql/transform"
)

// AssignTriggers assigns triggers wherever they're needed.
func AssignTriggers(ctx *sql.Context, a *analyzer.Analyzer, node sql.Node, scope *plan.Scope, selector analyzer.RuleSelector, qFlags *sql.QueryFlags) (sql.Node, transform.TreeIdentity, error) {
	return pgtransform.NodeWithOpaque(node, func(node sql.Node) (sql.Node, transform.TreeIdentity, error) {
		switch node := node.(type) {
		case *plan.DeleteFrom, *plan.InsertInto, *plan.Truncate, *plan.Update:
			sch, beforeTrigs, afterTrigs, err := getTriggerInformation(ctx, node)
			if err != nil {
				return nil, transform.NewTree, err
			}
			if len(beforeTrigs) == 0 && len(afterTrigs) == 0 {
				return node, transform.SameTree, nil
			}
			newNode := node
			if len(beforeTrigs) > 0 {
				handling := getTriggerRowHandling(node)
				newNode, err = nodeWithTriggers(ctx, newNode, &pgnodes.TriggerExecution{
					Triggers: beforeTrigs,
					Split:    handling,
					Return:   handling,
					Sch:      sch,
					Source:   getTriggerSource(node),
					Runner:   pgexprs.StatementRunner{Runner: a.Runner},
				})
				if err != nil {
					return nil, transform.NewTree, err
				}
			}
			if len(afterTrigs) > 0 {
				newNode = &pgnodes.TriggerExecution{
					Triggers: afterTrigs,
					Split:    getTriggerRowHandling(node),
					Return:   pgnodes.TriggerExecutionRowHandling_None,
					Sch:      sch,
					Source:   newNode,
					Runner:   pgexprs.StatementRunner{Runner: a.Runner},
				}
			}
			return newNode, transform.NewTree, nil
		default:
			return node, transform.SameTree, nil
		}
	})
}

// getTriggerInformation loads information that is common for the different trigger types.
func getTriggerInformation(ctx *sql.Context, node sql.Node) (sch sql.Schema, beforeTrigs []triggers.Trigger, afterTrigs []triggers.Trigger, err error) {
	var tbl sql.Table
	switch node := node.(type) {
	case *plan.DeleteFrom:
		tbl, err = plan.GetDeletable(node.Child)
		if err != nil {
			return nil, nil, nil, err
		}
	case *plan.InsertInto:
		tbl, err = plan.GetInsertable(node.Destination)
		if err != nil {
			return nil, nil, nil, err
		}
	case *plan.Truncate:
		tbl, err = plan.GetTruncatable(node.Child)
		if err != nil {
			return nil, nil, nil, err
		}
	case *plan.Update:
		tbl, err = plan.GetUpdatable(node.Child)
		if err != nil {
			return nil, nil, nil, err
		}
	default:
		return nil, nil, nil, nil
	}
	tblID, ok, _ := id.GetFromTable(ctx, tbl)
	if !ok {
		return nil, nil, nil, nil
	}
	trigCollection, err := core.GetTriggersCollectionFromContext(ctx)
	if err != nil {
		return nil, nil, nil, err
	}
	allTrigs := trigCollection.GetTriggersForTable(ctx, tblID)
	// Return early if there are no triggers for the table
	if len(allTrigs) == 0 {
		return tbl.Schema(), nil, nil, nil
	}
	// Trigger order is determined by the name
	sort.Slice(allTrigs, func(i, j int) bool {
		return allTrigs[i].ID.TriggerName() < allTrigs[j].ID.TriggerName()
	})
	beforeTrigs = make([]triggers.Trigger, 0, len(allTrigs))
	afterTrigs = make([]triggers.Trigger, 0, len(allTrigs))
	for _, trig := range allTrigs {
		matchesEventType := false
		for _, event := range trig.Events {
			switch node.(type) {
			case *plan.DeleteFrom:
				if event.Type == triggers.TriggerEventType_Delete {
					matchesEventType = true
				}
			case *plan.InsertInto:
				if event.Type == triggers.TriggerEventType_Insert {
					matchesEventType = true
				}
			case *plan.Truncate:
				if event.Type == triggers.TriggerEventType_Truncate {
					matchesEventType = true
				}
			case *plan.Update:
				if event.Type == triggers.TriggerEventType_Update {
					matchesEventType = true
				}
			}
		}
		if !matchesEventType {
			continue
		}
		switch trig.Timing {
		case triggers.TriggerTiming_Before:
			beforeTrigs = append(beforeTrigs, trig)
		case triggers.TriggerTiming_After:
			afterTrigs = append(afterTrigs, trig)
		default:
			return nil, nil, nil, fmt.Errorf("trigger timing has not yet been implemented")
		}
	}
	return tbl.Schema(), beforeTrigs, afterTrigs, nil
}

// getTriggerSource returns the trigger's source node.
func getTriggerSource(node sql.Node) sql.Node {
	switch node := node.(type) {
	case *plan.DeleteFrom:
		return node.Child
	case *plan.InsertInto:
		return node.Source
	case *plan.Truncate:
		return node.Child
	case *plan.Update:
		return node.Child
	default:
		return node
	}
}

// getTriggerRowHandling returns the trigger's row handling type (based on how GMS passes rows in the intermediate
// steps).
func getTriggerRowHandling(node sql.Node) pgnodes.TriggerExecutionRowHandling {
	switch node.(type) {
	case *plan.DeleteFrom:
		return pgnodes.TriggerExecutionRowHandling_Old
	case *plan.InsertInto:
		return pgnodes.TriggerExecutionRowHandling_New
	case *plan.Truncate:
		return pgnodes.TriggerExecutionRowHandling_None
	case *plan.Update:
		return pgnodes.TriggerExecutionRowHandling_OldNew
	default:
		return pgnodes.TriggerExecutionRowHandling_None
	}
}

// nodeWithTriggers calls the appropriate WithX function depending on the node type.
func nodeWithTriggers(ctx *sql.Context, node sql.Node, executionNode *pgnodes.TriggerExecution) (sql.Node, error) {
	switch node := node.(type) {
	case *plan.DeleteFrom:
		return node.WithChildren(executionNode)
	case *plan.InsertInto:
		return node.WithSource(executionNode), nil
	case *plan.Truncate:
		return node.WithChildren(executionNode)
	case *plan.Update:
		return node.WithChildren(executionNode)
	default:
		return nil, fmt.Errorf("unknown node for triggers")
	}
}
