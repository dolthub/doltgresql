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
	"slices"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/analyzer"
	"github.com/dolthub/go-mysql-server/sql/plan"
	"github.com/dolthub/go-mysql-server/sql/transform"

	"github.com/dolthub/doltgresql/server/functions/framework"
)

// RemoveIdentityUpdates removes the derived assignments that recompute GENERATED ALWAYS AS IDENTITY columns, which are
// stored as generated columns calling nextval. Postgres only assigns a new identity value on insert, or when an UPDATE
// explicitly sets the column to DEFAULT.
func RemoveIdentityUpdates(ctx *sql.Context, a *analyzer.Analyzer, node sql.Node, scope *plan.Scope, selector analyzer.RuleSelector, qFlags *sql.QueryFlags) (sql.Node, transform.TreeIdentity, error) {
	switch node := node.(type) {
	case *plan.Update:
		return transform.Node(ctx, node, removeIdentityUpdatesFromSource)
	case *plan.InsertInto:
		if node.OnDup == nil {
			return node, transform.SameTree, nil
		}
		newOnDup, same, err := transform.Node(ctx, node.OnDup, removeIdentityUpdatesFromSource)
		if err != nil || same {
			return node, same, err
		}
		newInsertInto := *node
		newInsertInto.OnDup = newOnDup
		return &newInsertInto, transform.NewTree, nil
	default:
		return node, transform.SameTree, nil
	}
}

// removeIdentityUpdatesFromSource removes the identity assignments from the update expressions of UPDATE and
// ON CONFLICT DO UPDATE sources.
func removeIdentityUpdatesFromSource(ctx *sql.Context, node sql.Node) (sql.Node, transform.TreeIdentity, error) {
	switch node := node.(type) {
	case *plan.UpdateSource:
		updateExprs, same := removeIdentityUpdateExprs(ctx, node.UpdateExprs)
		if same {
			return node, transform.SameTree, nil
		}
		newNode := *node
		newNode.UpdateExprs = updateExprs
		return &newNode, transform.NewTree, nil
	case *plan.OnDuplicateKeyUpdateSource:
		updateExprs, same := removeIdentityUpdateExprs(ctx, node.UpdateExprs)
		if same {
			return node, transform.SameTree, nil
		}
		newNode := *node
		newNode.UpdateExprs = updateExprs
		return &newNode, transform.NewTree, nil
	default:
		return node, transform.SameTree, nil
	}
}

// removeIdentityUpdateExprs returns the given update expressions without the derived expressions that call nextval.
// Explicit expressions are always kept.
func removeIdentityUpdateExprs(ctx *sql.Context, updateExprs *sql.UpdateExprs) (*sql.UpdateExprs, transform.TreeIdentity) {
	if !updateExprs.HasDerivedUpdates() {
		return updateExprs, transform.SameTree
	}
	explicitExprs := updateExprs.ExplicitUpdateExprs()
	exprs := slices.Clip(explicitExprs)
	for _, expr := range updateExprs.DerivedUpdateExprs() {
		callsNextVal := transform.InspectExpr(ctx, expr, func(ctx *sql.Context, expr sql.Expression) bool {
			compiledFunction, ok := expr.(*framework.CompiledFunction)
			return ok && compiledFunction.FunctionName() == "nextval"
		})
		if !callsNextVal {
			exprs = append(exprs, expr)
		}
	}
	if len(exprs) == updateExprs.Length() {
		return updateExprs, transform.SameTree
	}
	return sql.NewUpdateExprs(exprs, len(explicitExprs)), transform.NewTree
}
