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
	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/analyzer"
	"github.com/dolthub/go-mysql-server/sql/expression"
	"github.com/dolthub/go-mysql-server/sql/plan"
	"github.com/dolthub/go-mysql-server/sql/transform"

	pgtransform "github.com/dolthub/doltgresql/server/transform"

	pgexprs "github.com/dolthub/doltgresql/server/expression"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// ResolveValuesTypes determines the common type for each column in a VALUES clause
// by examining all rows, following PostgreSQL's type resolution rules.
// This ensures VALUES(1),(2.01),(3) correctly infers numeric type, not integer.
func ResolveValuesTypes(ctx *sql.Context, a *analyzer.Analyzer, node sql.Node, scope *plan.Scope, selector analyzer.RuleSelector, qFlags *sql.QueryFlags) (sql.Node, transform.TreeIdentity, error) {
	// Track which VDTs we transform so we can update GetField nodes
	transformedVDTs := make(map[sql.TableId]sql.Schema)
	// First we transform VDTs and record their new schemas
	node, same, err := transform.NodeWithOpaque(node, func(n sql.Node) (sql.Node, transform.TreeIdentity, error) {
		newNode, same, err := transformValuesNode(n)
		if err != nil {
			return nil, same, err
		}
		if !same {
			if vdt, ok := newNode.(*plan.ValueDerivedTable); ok {
				transformedVDTs[vdt.Id()] = vdt.Schema()
			}
		}
		return newNode, same, err
	})
	if err != nil {
		return nil, transform.SameTree, err
	}

	// Next we update all GetField expressions that refer to a transformed VDT
	if len(transformedVDTs) > 0 {
		node, _, err = pgtransform.NodeExprsWithOpaque(node, func(expr sql.Expression) (sql.Expression, transform.TreeIdentity, error) {
			gf, ok := expr.(*expression.GetField)
			if !ok {
				return expr, transform.SameTree, nil
			}
			newSch, ok := transformedVDTs[gf.TableId()]
			if !ok {
				return expr, transform.SameTree, nil
			}

			// GetField indices are 1-based in GMS planbuilder, so subtract 1 for schema access
			schemaIdx := gf.Index() - 1
			if schemaIdx < 0 || schemaIdx >= len(newSch) {
				return nil, transform.NewTree, errors.Errorf("VALUES: GetField `%s` on table `%s` uses invalid index `%d`",
					gf.Name(), gf.Table(), gf.Index())
			}

			newType := newSch[schemaIdx].Type
			if gf.Type() == newType {
				return expr, transform.SameTree, nil
			}

			// Create a new expression with the updated type
			newGf := expression.NewGetFieldWithTable(
				gf.Index(),
				int(gf.TableId()),
				newType,
				gf.Database(),
				gf.Table(),
				gf.Name(),
				gf.IsNullable(),
			)
			return newGf, transform.NewTree, nil
		})
		if err != nil {
			return nil, transform.SameTree, err
		}
	}

	return node, same, nil
}

// transformValuesNode transforms a plan.Values or plan.ValueDerivedTable node to use common types
func transformValuesNode(n sql.Node) (sql.Node, transform.TreeIdentity, error) {
	var values *plan.Values
	var expressionerNode sql.Expressioner
	switch v := n.(type) {
	case *plan.ValueDerivedTable:
		values = v.Values
		expressionerNode = v
	case *plan.Values:
		values = v
		expressionerNode = v
	default:
		return n, transform.SameTree, nil
	}

	// Skip if no rows or single row (nothing to unify)
	if len(values.ExpressionTuples) <= 1 {
		return n, transform.SameTree, nil
	}
	numCols := len(values.ExpressionTuples[0])
	for i := 1; i < len(values.ExpressionTuples); i++ {
		if len(values.ExpressionTuples[i]) != numCols {
			return nil, transform.NewTree, errors.New("VALUES: VALUES lists must all be the same length")
		}
	}
	if numCols == 0 {
		return n, transform.SameTree, nil
	}

	// Collect types for each column across all rows
	columnTypes := make([][]*pgtypes.DoltgresType, numCols)
	for colIdx := 0; colIdx < numCols; colIdx++ {
		columnTypes[colIdx] = make([]*pgtypes.DoltgresType, len(values.ExpressionTuples))
		for rowIdx, row := range values.ExpressionTuples {
			exprType := row[colIdx].Type()
			if exprType == nil {
				columnTypes[colIdx][rowIdx] = pgtypes.Unknown
			} else if pgType, ok := exprType.(*pgtypes.DoltgresType); ok {
				columnTypes[colIdx][rowIdx] = pgType
			} else {
				return nil, transform.NewTree, errors.New("VALUES: VALUES cannot use GMS types")
			}
		}
	}

	// Find common type for each column
	var newTuples [][]sql.Expression
	for colIdx := 0; colIdx < numCols; colIdx++ {
		commonType, requiresCasts, err := framework.FindCommonType(columnTypes[colIdx])
		if err != nil {
			return nil, transform.NewTree, err
		}
		// If we require any casts, then we'll add casting to all expressions in the list
		if requiresCasts {
			if len(newTuples) == 0 {
				// Deep copy to avoid mutating the original expression tuples.
				newTuples = make([][]sql.Expression, len(values.ExpressionTuples))
				for i, row := range values.ExpressionTuples {
					newTuples[i] = make([]sql.Expression, len(row))
					copy(newTuples[i], row)
				}
			}
			for rowIdx := 0; rowIdx < len(newTuples); rowIdx++ {
				newTuples[rowIdx][colIdx] = pgexprs.NewImplicitCast(
					newTuples[rowIdx][colIdx], columnTypes[colIdx][rowIdx], commonType)
			}
		}
	}
	// If we didn't require any casts, then we can simply return our old node
	if len(newTuples) == 0 {
		return n, transform.SameTree, nil
	}

	// Flatten the new tuples into a single expression slice for WithExpressions
	var flatExprs []sql.Expression
	for _, row := range newTuples {
		flatExprs = append(flatExprs, row...)
	}
	newNode, err := expressionerNode.WithExpressions(flatExprs...)
	if err != nil {
		return nil, transform.NewTree, err
	}
	return newNode, transform.NewTree, nil
}
