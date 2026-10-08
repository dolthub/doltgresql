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
)

// preserveSetReturningFilters prevents GMS's subquery filter pushdown from substituting an SRF into a scalar
// predicate. Such predicates must run after the projection expands its rows. Other subqueries retain pushdown.
func preserveSetReturningFilters(apply analyzer.RuleFunc) analyzer.RuleFunc {
	return func(ctx *sql.Context, a *analyzer.Analyzer, node sql.Node, scope *plan.Scope, selector analyzer.RuleSelector, qFlags *sql.QueryFlags) (sql.Node, transform.TreeIdentity, error) {
		mappings := make(map[sql.TableId]map[sql.ColumnId]sql.Expression)
		protected, _, err := transform.Node(ctx, node, func(ctx *sql.Context, n sql.Node) (sql.Node, transform.TreeIdentity, error) {
			alias, ok := n.(*plan.SubqueryAlias)
			if !ok || alias.ScopeMapping == nil || !hasSetReturningProjection(ctx, alias.Child) {
				return n, transform.SameTree, nil
			}
			// A nil scope mapping makes GMS leave filters above this subquery alias. Restore the
			// mapping after the rule so subsequent analyzer rules can still resolve its columns.
			mappings[alias.Id()] = alias.ScopeMapping
			return alias.WithScopeMapping(nil), transform.NewTree, nil
		})
		if err != nil {
			return nil, transform.SameTree, err
		}
		result, same, err := apply(ctx, a, protected, scope, selector, qFlags)
		if err != nil || len(mappings) == 0 {
			return result, same, err
		}
		result, _, err = transform.Node(ctx, result, func(ctx *sql.Context, n sql.Node) (sql.Node, transform.TreeIdentity, error) {
			if alias, ok := n.(*plan.SubqueryAlias); ok {
				if mapping, ok := mappings[alias.Id()]; ok {
					return alias.WithScopeMapping(mapping), transform.NewTree, nil
				}
			}
			return n, transform.SameTree, nil
		})
		return result, same, err
	}
}

func hasSetReturningProjection(ctx *sql.Context, node sql.Node) bool {
	return transform.InspectUp(ctx, node, func(ctx *sql.Context, n sql.Node) bool {
		project, ok := n.(*plan.Project)
		if !ok {
			return false
		}
		for _, expr := range project.Projections {
			if transform.InspectExpr(ctx, expr, func(ctx *sql.Context, e sql.Expression) bool {
				srf, ok := e.(sql.RowIterExpression)
				return ok && srf.ReturnsRowIter()
			}) {
				return true
			}
		}
		return false
	})
}
