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

package analyzer

import (
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/analyzer"
	"github.com/dolthub/go-mysql-server/sql/plan"
	"github.com/dolthub/go-mysql-server/sql/transform"

	"github.com/dolthub/doltgresql/core/triggers"
	pgnodes "github.com/dolthub/doltgresql/server/node"
	pgtransform "github.com/dolthub/doltgresql/server/transform"
)

// HoistInsertTriggers lifts an INSERT's BEFORE triggers above the projection that pads the rows an INSERT
// supplies out to the table's full schema. AssignTriggers has to run long before that projection exists, so
// it can only attach the triggers to the INSERT's source, which leaves them underneath it. A trigger reading
// or writing NEW there would see only the columns the INSERT named, in the order it named them.
func HoistInsertTriggers(ctx *sql.Context, a *analyzer.Analyzer, node sql.Node, scope *plan.Scope, selector analyzer.RuleSelector, qFlags *sql.QueryFlags) (sql.Node, transform.TreeIdentity, error) {
	return pgtransform.NodeWithOpaque(ctx, node, func(ctx *sql.Context, node sql.Node) (sql.Node, transform.TreeIdentity, error) {
		insert, ok := node.(*plan.InsertInto)
		if !ok {
			return node, transform.SameTree, nil
		}
		project, ok := insert.Source.(*plan.Project)
		if !ok {
			return node, transform.SameTree, nil
		}
		trigExec, ok := project.Child.(*pgnodes.TriggerExecution)
		if !ok || trigExec.Timing != triggers.TriggerTiming_Before {
			return node, transform.SameTree, nil
		}
		// The projection reads the rows the INSERT supplied, so it keeps the source it already had, and
		// the triggers take its padded output as their own source.
		newProject, err := project.WithChildren(ctx, trigExec.Source)
		if err != nil {
			return nil, transform.NewTree, err
		}
		newTrigExec, err := trigExec.WithChildren(ctx, newProject)
		if err != nil {
			return nil, transform.NewTree, err
		}
		return insert.WithSource(newTrigExec), transform.NewTree, nil
	})
}
