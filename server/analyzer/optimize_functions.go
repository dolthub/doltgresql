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

	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/analyzer"
	"github.com/dolthub/go-mysql-server/sql/expression"
	"github.com/dolthub/go-mysql-server/sql/plan"
	"github.com/dolthub/go-mysql-server/sql/planbuilder"
	"github.com/dolthub/go-mysql-server/sql/transform"

	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtransform "github.com/dolthub/doltgresql/server/transform"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// OptimizeFunctions replaces all functions that fit specific criteria with their optimized variants. Also handles
// SRFs (set-returning functions) by setting the `IncludesNestedIters` flag on the Project node if any SRF is found
// inside projection expressions.
func OptimizeFunctions(ctx *sql.Context, a *analyzer.Analyzer, node sql.Node, scope *plan.Scope, selector analyzer.RuleSelector, qFlags *sql.QueryFlags) (sql.Node, transform.TreeIdentity, error) {
	_, isInsertNode := node.(*plan.InsertInto)
	return pgtransform.NodeWithOpaque(ctx, node, func(ctx *sql.Context, n sql.Node) (sql.Node, transform.TreeIdentity, error) {
		projectNode, ok := n.(*plan.Project)
		if !ok {
			return n, transform.SameTree, nil
		}

		hasMultipleExpressionTuples := false
		hasSRF := false
		// Check if there is set returning function in the source node (e.g. SELECT * FROM unnest())
		n, sameNode, err := transform.NodeExprsWithNode(ctx, projectNode.Child, func(ctx *sql.Context, in sql.Node, expr sql.Expression) (sql.Expression, transform.TreeIdentity, error) {
			if compiledFunction, ok := expr.(*framework.CompiledFunction); ok {
				// TODO: need better way to detect sequence usage
				switch compiledFunction.FunctionName() {
				case "nextval", "setval", "currval":
					err := authCheckSequenceFromExpr(ctx, a.Catalog.AuthHandler, compiledFunction.Arguments[0])
					if err != nil {
						return nil, transform.SameTree, err
					}
				}
				hasSRF = hasSRF || compiledFunction.IsSRF()
				if quickFunction := compiledFunction.GetQuickFunction(ctx); quickFunction != nil {
					return quickFunction, transform.NewTree, nil
				}

				// fill in default exprs if applicable
				if err := compiledFunction.ResolveDefaultValues(ctx, func(defExpr string) (sql.Expression, error) {
					return getDefaultExpr(ctx, a.Catalog, defExpr)
				}); err != nil {
					return nil, transform.SameTree, err
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
		refreshedProjections, sameTypes, err := refreshProjectedAggregateTypes(ctx, projectNode.Child, projectNode.Projections)
		if err != nil {
			return nil, transform.SameTree, err
		}

		// insert node cannot have more than 1 row value if it has set returning function
		if isInsertNode && hasMultipleExpressionTuples && hasSRF {
			return nil, false, errors.Errorf("set-returning functions are not allowed in VALUES")
		}

		// Check if there is set returning function in the projection expressions (e.g. SELECT unnest() [FROM table/srf])
		hasSRFInProjection := false
		exprs, sameExprs, err := transform.Exprs(ctx, refreshedProjections, func(ctx *sql.Context, expr sql.Expression) (sql.Expression, transform.TreeIdentity, error) {
			if compiledFunction, ok := expr.(*framework.CompiledFunction); ok {
				hasSRFInProjection = hasSRFInProjection || compiledFunction.IsSRF()
				if quickFunction := compiledFunction.GetQuickFunction(ctx); quickFunction != nil {
					return quickFunction, transform.NewTree, nil
				}
				// TODO: need better way to detect sequence usage
				switch compiledFunction.FunctionName() {
				case "nextval", "setval", "currval":
					err = authCheckSequenceFromExpr(ctx, a.Catalog.AuthHandler, compiledFunction.Arguments[0])
					if err != nil {
						return nil, transform.SameTree, err
					}
				}

				// fill in default exprs if applicablea
				if err = compiledFunction.ResolveDefaultValues(ctx, func(defExpr string) (sql.Expression, error) {
					return getDefaultExpr(ctx, a.Catalog, defExpr)
				}); err != nil {
					return nil, transform.SameTree, err
				}
			}
			return expr, transform.SameTree, nil
		})
		if err != nil {
			return nil, transform.SameTree, err
		}
		if !sameTypes || !sameExprs {
			n, err = projectNode.WithExpressions(ctx, exprs...)
			if err != nil {
				return nil, transform.SameTree, err
			}
			projectNode = n.(*plan.Project)
		}

		// nested iter is used for set returning functions in the projections only
		if hasSRFInProjection {
			// Under some conditions, there will be no quick-function replacement, but changing the Project node to include
			// nested iterators is still a change we need to tell the transform functions about.
			sameExprs = transform.NewTree
			projectNode = projectNode.WithIncludesNestedIters(true)
		}

		return projectNode, sameNode && sameTypes && sameExprs, err
	})
}

// refreshProjectedAggregateTypes updates fields whose aggregate type resolved after plan construction.
func refreshProjectedAggregateTypes(ctx *sql.Context, child sql.Node, projections []sql.Expression) ([]sql.Expression, transform.TreeIdentity, error) {
	return transform.Exprs(ctx, projections, func(ctx *sql.Context, expr sql.Expression) (sql.Expression, transform.TreeIdentity, error) {
		field, ok := expr.(*expression.GetField)
		if !ok || field.Type(ctx) != pgtypes.Unknown {
			return expr, transform.SameTree, nil
		}
		sourceType := resolvedAggregateType(ctx, child, field.Id())
		if sourceType == nil {
			return expr, transform.SameTree, nil
		}
		refreshed := expression.NewGetFieldWithTable(
			field.Index(), int(field.TableId()), sourceType,
			field.Database(), field.Table(), field.Name(), field.IsNullable(ctx),
		).WithId(field.Id())
		return refreshed, transform.NewTree, nil
	})
}

// resolvedAggregateType returns the concrete type of the aggregate identified by id beneath node.
func resolvedAggregateType(ctx *sql.Context, node sql.Node, id sql.ColumnId) sql.Type {
	var typ sql.Type
	transform.InspectExpressions(ctx, node, func(ctx *sql.Context, expr sql.Expression) bool {
		agg, ok := expr.(framework.AggregateFunction)
		if !ok || agg.Id() != id || agg.Type(ctx) == pgtypes.Unknown {
			return true
		}
		typ = agg.Type(ctx)
		return false
	})
	return typ
}

// getDefaultExpr takes the default value definition, parses, builds and returns sql.ColumnDefaultValue.
func getDefaultExpr(ctx *sql.Context, c sql.Catalog, defExpr string) (sql.Expression, error) {
	builder := planbuilder.New(ctx, c, nil)
	proj, _, _, _, err := builder.Parse(fmt.Sprintf("select %s", defExpr), nil, false)
	if err != nil {
		return nil, err
	}
	parsedExpr := proj.(*plan.Project).Projections[0]
	if a, ok := parsedExpr.(*expression.Alias); ok {
		parsedExpr = a.Child
	}
	return parsedExpr, nil
}
