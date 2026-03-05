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
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/analyzer"
	"github.com/dolthub/go-mysql-server/sql/expression"
	"github.com/dolthub/go-mysql-server/sql/plan"
	"github.com/dolthub/go-mysql-server/sql/transform"

	pgexprs "github.com/dolthub/doltgresql/server/expression"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtransform "github.com/dolthub/doltgresql/server/transform"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// ResolveValuesTypes determines the common type for each column in a VALUES clause
// by examining all rows, following PostgreSQL's type resolution rules.
// This ensures VALUES(1),(2.01),(3) correctly infers numeric type, not integer.
func ResolveValuesTypes(ctx *sql.Context, a *analyzer.Analyzer, node sql.Node, scope *plan.Scope, selector analyzer.RuleSelector, qFlags *sql.QueryFlags) (sql.Node, transform.TreeIdentity, error) {
	// Walk the tree and wrap mixed-type VALUES columns with ImplicitCast.
	// We record which VDTs changed so we can fix up GetField types afterward.
	transformedVDTs := make(map[sql.TableId]sql.Schema)
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

	// Now, fix GetField types that reference a transformed VDT. For example,
	// after wrapping VALUES(1),(2.5) with ImplicitCast to numeric, any
	// GetField reading column "n" from that VDT still says int4 and needs
	// to be updated to numeric.
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

			// We match by column name because GetField indices are global
			// across all tables in a JOIN (e.g., a.n=0, b.id=1, b.label=2).
			// We can't convert a global index to a per-table position without
			// knowing the table's starting offset, which we don't have here.
			schemaIdx := -1
			for i, col := range newSch {
				if col.Name == gf.Name() {
					schemaIdx = i
					break
				}
			}
			if schemaIdx < 0 {
				return expr, transform.SameTree, nil
			}

			newType := newSch[schemaIdx].Type
			if gf.Type() == newType {
				return expr, transform.SameTree, nil
			}

			return expression.NewGetFieldWithTable(
				gf.Index(), int(gf.TableId()), newType,
				gf.Database(), gf.Table(), gf.Name(), gf.IsNullable(),
			), transform.NewTree, nil
		})
		if err != nil {
			return nil, transform.SameTree, err
		}

		// The pass above only fixed GetFields that read directly from a VDT
		// (matched by tableId). But changing a VDT column's type can have a
		// ripple effect: if that column feeds into an aggregate like MIN or
		// MAX, the aggregate's return type changes too. Parent nodes that
		// read the aggregate result still have the old type. For example:
		//
		//   SELECT MIN(n) FROM (VALUES(1),(2.5)) v(n)
		//
		//   Project [GetField("min(v.n)", tableId=GroupBy, type=int4)]
		//     └── GroupBy [MIN(GetField("n", tableId=VDT, type=numeric))]
		//           └── VDT [n: int4 → numeric]
		//
		// The pass above fixed "n" inside MIN because its tableId=VDT.
		// MIN now returns numeric, so GroupBy produces numeric. But the
		// Project's GetField still says int4 because its tableId=GroupBy,
		// which wasn't in transformedVDTs. At runtime this causes a panic
		// because the actual value is decimal.Decimal but the type says int32.
		//
		// This pass catches those: for each GetField, check if its type
		// disagrees with what the child node actually produces.
		node, _, err = pgtransform.NodeExprsWithNodeWithOpaque(node, func(n sql.Node, expr sql.Expression) (sql.Expression, transform.TreeIdentity, error) {
			gf, ok := expr.(*expression.GetField)
			if !ok {
				return expr, transform.SameTree, nil
			}
			// Skip VDT GetFields — the first pass already handled these
			if _, isVDT := transformedVDTs[gf.TableId()]; isVDT {
				return expr, transform.SameTree, nil
			}
			// Collect the schema that this node's children produce
			var childSchema sql.Schema
			for _, child := range n.Children() {
				childSchema = append(childSchema, child.Schema()...)
			}
			// TODO: GMS is case-insensitive for identifiers, so aggregate
			// GetField names and child schema names may differ in casing.
			// We use strings.ToLower to handle this, but Postgres requires
			// case-sensitivity for quoted identifiers, which this breaks.
			gfName := strings.ToLower(gf.Name())
			for _, col := range childSchema {
				if strings.ToLower(col.Name) == gfName && gf.Type() != col.Type {
					return expression.NewGetFieldWithTable(
						gf.Index(), int(gf.TableId()), col.Type,
						gf.Database(), gf.Table(), gf.Name(), gf.IsNullable(),
					), transform.NewTree, nil
				}
			}
			return expr, transform.SameTree, nil
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
			return nil, transform.NewTree, errors.Errorf("VALUES: row %d has %d columns, expected %d", i+1, len(values.ExpressionTuples[i]), numCols)
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
				return nil, transform.NewTree, errors.Errorf("VALUES: non-Doltgres type found in row %d, column %d: %s", rowIdx, colIdx, exprType.String())
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
	flatExprs := make([]sql.Expression, 0, len(newTuples)*len(newTuples[0]))
	for _, row := range newTuples {
		flatExprs = append(flatExprs, row...)
	}
	newNode, err := expressionerNode.WithExpressions(flatExprs...)
	if err != nil {
		return nil, transform.NewTree, err
	}
	return newNode, transform.NewTree, nil
}
