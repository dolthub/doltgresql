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

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/analyzer"
	"github.com/dolthub/go-mysql-server/sql/expression"
	"github.com/dolthub/go-mysql-server/sql/plan"
	"github.com/dolthub/go-mysql-server/sql/transform"

	pgexprs "github.com/dolthub/doltgresql/server/expression"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// ComparisonCasts handles implicit conversions for comparisons. We cannot yet implement the comparison operators as GMS
// relies on specific expressions some cases. Eventually, we'll need to decouple the dependency, so that we may use the
// appropriate functions. This is technically incorrect, but should let us make progress until then.
func ComparisonCasts(ctx *sql.Context, a *analyzer.Analyzer, node sql.Node, scope *plan.Scope, selector analyzer.RuleSelector) (sql.Node, transform.TreeIdentity, error) {
	if disjointedNode, ok := node.(plan.DisjointedChildrenNode); ok {
		return handleDisjointedNodes(ctx, a, disjointedNode, scope, selector, ComparisonCasts)
	}
	return transform.NodeExprsWithOpaque(node, func(expr sql.Expression) (sql.Expression, transform.TreeIdentity, error) {
		switch comparison := expr.(type) {
		case *expression.Equals:
			left, right, ok, err := comparisonCasts(comparison.LeftChild, comparison.RightChild)
			if err != nil {
				return nil, transform.NewTree, err
			}
			if !ok {
				return comparison, transform.SameTree, nil
			}
			return expression.NewEquals(left, right), transform.NewTree, nil
		case *expression.NullSafeEquals:
			left, right, ok, err := comparisonCasts(comparison.LeftChild, comparison.RightChild)
			if err != nil {
				return nil, transform.NewTree, err
			}
			if !ok {
				return comparison, transform.SameTree, nil
			}
			return expression.NewNullSafeEquals(left, right), transform.NewTree, nil
		case *expression.GreaterThan:
			left, right, ok, err := comparisonCasts(comparison.LeftChild, comparison.RightChild)
			if err != nil {
				return nil, transform.NewTree, err
			}
			if !ok {
				return comparison, transform.SameTree, nil
			}
			return expression.NewGreaterThan(left, right), transform.NewTree, nil
		case *expression.GreaterThanOrEqual:
			left, right, ok, err := comparisonCasts(comparison.LeftChild, comparison.RightChild)
			if err != nil {
				return nil, transform.NewTree, err
			}
			if !ok {
				return comparison, transform.SameTree, nil
			}
			return expression.NewGreaterThanOrEqual(left, right), transform.NewTree, nil
		case *expression.LessThan:
			left, right, ok, err := comparisonCasts(comparison.LeftChild, comparison.RightChild)
			if err != nil {
				return nil, transform.NewTree, err
			}
			if !ok {
				return comparison, transform.SameTree, nil
			}
			return expression.NewLessThan(left, right), transform.NewTree, nil
		case *expression.LessThanOrEqual:
			left, right, ok, err := comparisonCasts(comparison.LeftChild, comparison.RightChild)
			if err != nil {
				return nil, transform.NewTree, err
			}
			if !ok {
				return comparison, transform.SameTree, nil
			}
			return expression.NewLessThanOrEqual(left, right), transform.NewTree, nil
		default:
			// The expression is not a comparison, so we'll simply return it
			return expr, transform.SameTree, nil
		}
	})
}

// comparisonCasts handles casting either side of a comparison.
func comparisonCasts(left sql.Expression, right sql.Expression) (sql.Expression, sql.Expression, bool, error) {
	leftType, ok := left.Type().(pgtypes.DoltgresType)
	if !ok {
		left = pgexprs.NewGMSCast(left)
		leftType = left.Type().(pgtypes.DoltgresType)
	}
	rightType, ok := right.Type().(pgtypes.DoltgresType)
	if !ok {
		right = pgexprs.NewGMSCast(right)
		rightType = right.Type().(pgtypes.DoltgresType)
	}
	if leftType.Equals(rightType) {
		return left, right, false, nil
	}
	rightToLeftCast := framework.GetImplicitCast(rightType.BaseID(), leftType.BaseID())
	if rightToLeftCast != nil {
		return left, pgexprs.NewImplicitCast(right, rightType, leftType), true, nil
	}
	leftToRightCast := framework.GetImplicitCast(leftType.BaseID(), rightType.BaseID())
	if leftToRightCast != nil {
		return pgexprs.NewImplicitCast(left, leftType, rightType), right, true, nil
	}
	return nil, nil, false, fmt.Errorf("COMPARISON: types are incompatible: %s and %s", leftType.String(), rightType.String())
}
