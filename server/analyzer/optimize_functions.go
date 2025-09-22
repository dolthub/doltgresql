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
	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/analyzer"
	"github.com/dolthub/go-mysql-server/sql/plan"
	"github.com/dolthub/go-mysql-server/sql/transform"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtransform "github.com/dolthub/doltgresql/server/transform"
)

// OptimizeFunctions replaces all functions that fit specific criteria with their optimized variants. Also handles
// SRFs (set-returning functions) by setting the `IncludesNestedIters` flag on the Project node if any SRF is found
// inside projection expressions.
func OptimizeFunctions(ctx *sql.Context, a *analyzer.Analyzer, node sql.Node, scope *plan.Scope, selector analyzer.RuleSelector, qFlags *sql.QueryFlags) (sql.Node, transform.TreeIdentity, error) {
	// This is supposed to be one of the last rules to run. Subqueries break that assumption, so we skip this rule in such cases.
	if scope != nil && scope.CurrentNodeIsFromSubqueryExpression {
		return node, transform.SameTree, nil
	}

	_, isInsertNode := node.(*plan.InsertInto)
	return pgtransform.NodeWithOpaque(node, func(n sql.Node) (sql.Node, transform.TreeIdentity, error) {
		projectNode, ok := n.(*plan.Project)
		if !ok {
			return n, transform.SameTree, nil
		}

		hasMultipleExpressionTuples := false
		hasSRF := false
		// Check if there is set returning function in the source node (e.g. SELECT * FROM unnest())
		n, sameNode, err := transform.NodeExprsWithNode(projectNode.Child, func(in sql.Node, expr sql.Expression) (sql.Expression, transform.TreeIdentity, error) {
			if compiledFunction, ok := expr.(*framework.CompiledFunction); ok {
				hasSRF = hasSRF || compiledFunction.IsSRF()
				if quickFunction := compiledFunction.GetQuickFunction(); quickFunction != nil {
					return quickFunction, transform.NewTree, nil
				}
			}
			if v, ok := in.(*plan.Values); ok {
				hasMultipleExpressionTuples = len(v.ExpressionTuples) > 1
			}
			return expr, transform.SameTree, nil
		})
		if err != nil {
			return nil, transform.SameTree, err
		}
		if !sameNode {
			projectNode.Child = n
		}

		// insert node cannot have more than 1 row value if it has set returning function
		if isInsertNode && hasMultipleExpressionTuples && hasSRF {
			return nil, false, errors.Errorf("set-returning functions are not allowed in VALUES")
		}

		// Check if there is set returning function in the projection expressions (e.g. SELECT unnest() [FROM table/srf])
		hasSRFInProjection := false
		exprs, sameExprs, err := transform.Exprs(projectNode.Projections, func(expr sql.Expression) (sql.Expression, transform.TreeIdentity, error) {
			if compiledFunction, ok := expr.(*framework.CompiledFunction); ok {
				hasSRFInProjection = hasSRFInProjection || compiledFunction.IsSRF()
				if quickFunction := compiledFunction.GetQuickFunction(); quickFunction != nil {
					return quickFunction, transform.NewTree, nil
				}
			}
			return expr, transform.SameTree, nil
		})
		if err != nil {
			return nil, transform.SameTree, err
		}
		if !sameExprs {
			projectNode.Projections = exprs
		}

		// nested iter is used for set returning functions in the projections only
		if hasSRFInProjection {
			// Under some conditions, there will be no quick-function replacement, but changing the Project node to include
			// nested iterators is still a change we need to tell the transform functions about.
			sameExprs = transform.NewTree
			projectNode = projectNode.WithIncludesNestedIters(true)
		}

		return projectNode, sameNode && sameExprs, err
	})
}
