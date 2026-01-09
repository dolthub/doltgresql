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
	"github.com/dolthub/go-mysql-server/sql/expression"
	"github.com/dolthub/go-mysql-server/sql/plan"
	"github.com/dolthub/go-mysql-server/sql/transform"

	pgexprs "github.com/dolthub/doltgresql/server/expression"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// ResolveValuesTypes determines the common type for each column in a VALUES clause
// by examining all rows, following PostgreSQL's type resolution rules.
// This ensures VALUES(1),(2.01),(3) correctly infers numeric type, not integer.
func ResolveValuesTypes(ctx *sql.Context, a *analyzer.Analyzer, node sql.Node, scope *plan.Scope, selector analyzer.RuleSelector, qFlags *sql.QueryFlags) (sql.Node, transform.TreeIdentity, error) {
	// Track which VDTs we transform so we can update parent nodes
	transformedVDTs := make(map[*plan.ValueDerivedTable]sql.Schema)

	// First pass: transform VDTs and record their new schemas
	node, same1, err := transform.NodeWithOpaque(node, func(n sql.Node) (sql.Node, transform.TreeIdentity, error) {
		newNode, same, err := transformValuesNode(n)
		if err != nil {
			return nil, same, err
		}
		if !same {
			if vdt, ok := newNode.(*plan.ValueDerivedTable); ok {
				transformedVDTs[vdt] = vdt.Schema()
			}
		}
		return newNode, same, err
	})
	if err != nil {
		return nil, transform.SameTree, err
	}

	// Second pass: update GetField types in parent nodes that reference transformed VDTs
	if len(transformedVDTs) > 0 {
		node, _, err = transform.NodeWithOpaque(node, func(n sql.Node) (sql.Node, transform.TreeIdentity, error) {
			return updateGetFieldTypes(n, transformedVDTs)
		})
		if err != nil {
			return nil, transform.SameTree, err
		}
	}

	return node, same1, nil
}

// getSourceSchema traverses through wrapper nodes (GroupBy, Filter, etc.) to find
// the actual source schema from a VDT or other data source. This is needed because
// nodes like GroupBy produce a different output schema than their input schema.
func getSourceSchema(n sql.Node) sql.Schema {
	switch node := n.(type) {
	case *plan.GroupBy:
		// GroupBy's Schema() returns aggregate output, but we need the source schema
		return getSourceSchema(node.Child)
	case *plan.Filter:
		return getSourceSchema(node.Child)
	case *plan.Sort:
		return getSourceSchema(node.Child)
	case *plan.Limit:
		return getSourceSchema(node.Child)
	case *plan.Offset:
		return getSourceSchema(node.Child)
	case *plan.Distinct:
		return getSourceSchema(node.Child)
	case *plan.SubqueryAlias:
		// SubqueryAlias wraps a VDT - get the child's schema
		return node.Child.Schema()
	case *plan.ValueDerivedTable:
		return node.Schema()
	default:
		// For other nodes, return their schema directly
		return n.Schema()
	}
}

// updateGetFieldTypes updates GetField expressions that reference transformed VDT columns
func updateGetFieldTypes(n sql.Node, transformedVDTs map[*plan.ValueDerivedTable]sql.Schema) (sql.Node, transform.TreeIdentity, error) {
	// Only handle nodes that have expressions (like Project)
	exprNode, ok := n.(sql.Expressioner)
	if !ok {
		return n, transform.SameTree, nil
	}

	// Get the source schema by traversing through wrapper nodes like GroupBy
	// This ensures we get the VDT's schema, not the aggregate output schema
	var childSchema sql.Schema
	switch node := n.(type) {
	case *plan.Project:
		childSchema = getSourceSchema(node.Child)
	case *plan.SubqueryAlias:
		childSchema = node.Child.Schema()
	default:
		return n, transform.SameTree, nil
	}

	if childSchema == nil {
		return n, transform.SameTree, nil
	}

	// Transform expressions to update GetField types (recursively for nested expressions)
	exprs := exprNode.Expressions()
	newExprs := make([]sql.Expression, len(exprs))
	changed := false

	for i, expr := range exprs {
		newExpr, exprChanged, err := updateGetFieldExprRecursive(expr, childSchema)
		if err != nil {
			return nil, transform.SameTree, err
		}
		newExprs[i] = newExpr
		if exprChanged {
			changed = true
		}
	}

	if !changed {
		return n, transform.SameTree, nil
	}

	newNode, err := exprNode.WithExpressions(newExprs...)
	if err != nil {
		return nil, transform.SameTree, err
	}
	return newNode.(sql.Node), transform.NewTree, nil
}

// updateGetFieldExprRecursive recursively updates GetField expressions in the expression tree
func updateGetFieldExprRecursive(expr sql.Expression, childSchema sql.Schema) (sql.Expression, bool, error) {
	// First try to update if this is a GetField
	if _, ok := expr.(*expression.GetField); ok {
		return updateGetFieldExpr(expr, childSchema)
	}

	// Recursively process children
	children := expr.Children()
	if len(children) == 0 {
		return expr, false, nil
	}

	newChildren := make([]sql.Expression, len(children))
	changed := false
	for i, child := range children {
		newChild, childChanged, err := updateGetFieldExprRecursive(child, childSchema)
		if err != nil {
			return nil, false, err
		}
		newChildren[i] = newChild
		if childChanged {
			changed = true
		}
	}

	if !changed {
		return expr, false, nil
	}

	newExpr, err := expr.WithChildren(newChildren...)
	if err != nil {
		return nil, false, err
	}
	return newExpr, true, nil
}

// updateGetFieldExpr updates a GetField expression to use the correct type from the child schema
func updateGetFieldExpr(expr sql.Expression, childSchema sql.Schema) (sql.Expression, bool, error) {
	gf, ok := expr.(*expression.GetField)
	if !ok {
		return expr, false, nil
	}

	idx := gf.Index()
	// GetField indices are 1-based in GMS planbuilder, so subtract 1 for schema access
	schemaIdx := idx - 1
	if schemaIdx < 0 || schemaIdx >= len(childSchema) {
		return expr, false, nil
	}

	newType := childSchema[schemaIdx].Type
	if gf.Type() == newType {
		return expr, false, nil
	}

	// Create a new GetField with the updated type
	newGf := expression.NewGetFieldWithTable(
		idx,
		int(gf.TableId()),
		newType,
		gf.Database(),
		gf.Table(),
		gf.Name(),
		gf.IsNullable(),
	)
	return newGf, true, nil
}

// transformValuesNode transforms a VALUES or ValueDerivedTable node to use common types
func transformValuesNode(n sql.Node) (sql.Node, transform.TreeIdentity, error) {
	// Handle both ValueDerivedTable and Values nodes
	var values *plan.Values
	var vdt *plan.ValueDerivedTable
	var isVDT bool

	switch v := n.(type) {
	case *plan.ValueDerivedTable:
		vdt = v
		values = v.Values
		isVDT = true
	case *plan.Values:
		values = v
		isVDT = false
	default:
		return n, transform.SameTree, nil
	}

	// Skip if no rows or single row (nothing to unify)
	if len(values.ExpressionTuples) <= 1 {
		return n, transform.SameTree, nil
	}

	numCols := len(values.ExpressionTuples[0])
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
				// Non-DoltgresType encountered - should have been sanitized
				// Return unchanged and let TypeSanitizer handle it
				return n, transform.SameTree, nil
			}
		}
	}

	// Find common type for each column
	commonTypes := make([]*pgtypes.DoltgresType, numCols)
	for colIdx := 0; colIdx < numCols; colIdx++ {
		commonType, err := framework.FindCommonType(columnTypes[colIdx])
		if err != nil {
			return nil, transform.NewTree, err
		}
		commonTypes[colIdx] = commonType
	}

	// Check if any changes are needed
	needsChange := false
	for colIdx := 0; colIdx < numCols; colIdx++ {
		for rowIdx := 0; rowIdx < len(values.ExpressionTuples); rowIdx++ {
			if !columnTypes[colIdx][rowIdx].Equals(commonTypes[colIdx]) {
				needsChange = true
				break
			}
		}
		if needsChange {
			break
		}
	}

	if !needsChange {
		return n, transform.SameTree, nil
	}

	// Create new expression tuples with implicit casts where needed
	newTuples := make([][]sql.Expression, len(values.ExpressionTuples))
	for rowIdx, row := range values.ExpressionTuples {
		newTuples[rowIdx] = make([]sql.Expression, numCols)
		for colIdx, expr := range row {
			fromType := columnTypes[colIdx][rowIdx]
			toType := commonTypes[colIdx]
			if fromType.Equals(toType) {
				newTuples[rowIdx][colIdx] = expr
			} else if fromType.ID == pgtypes.Unknown.ID {
				// Unknown type can be coerced to any type without explicit cast
				// Use UnknownCoercion to report the target type while passing through values
				newTuples[rowIdx][colIdx] = pgexprs.NewUnknownCoercion(expr, toType)
			} else {
				newTuples[rowIdx][colIdx] = pgexprs.NewImplicitCast(expr, fromType, toType)
			}
		}
	}

	// Flatten the new tuples into a single expression slice for WithExpressions
	var flatExprs []sql.Expression
	for _, row := range newTuples {
		flatExprs = append(flatExprs, row...)
	}

	if isVDT {
		// Use WithExpressions to preserve all VDT fields (name, columns, id, cols)
		// while updating the expressions and recalculating the schema
		newNode, err := vdt.WithExpressions(flatExprs...)
		if err != nil {
			return nil, transform.NewTree, err
		}
		return newNode, transform.NewTree, nil
	}

	// For standalone Values node, use WithExpressions as well
	newNode, err := values.WithExpressions(flatExprs...)
	if err != nil {
		return nil, transform.NewTree, err
	}
	return newNode, transform.NewTree, nil
}
