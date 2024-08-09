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
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// AssignInsertCasts adds the appropriate assign casts for insertions.
func AssignInsertCasts(ctx *sql.Context, a *analyzer.Analyzer, node sql.Node, scope *plan.Scope, selector analyzer.RuleSelector, qFlags *sql.QueryFlags) (sql.Node, transform.TreeIdentity, error) {
	insertInto, ok := node.(*plan.InsertInto)
	if !ok {
		return node, transform.SameTree, nil
	}
	// First we'll make a map for each column, so that it's easier to match a name to a type. We also ensure that the
	// types use Doltgres types, as casts rely on them. At this point, we shouldn't have any GMS types floating around
	// anymore, so no need to include a lot of additional code to handle them.
	destinationNameToType := make(map[string]pgtypes.DoltgresType)
	for _, col := range insertInto.Destination.Schema() {
		colType, ok := col.Type.(pgtypes.DoltgresType)
		if !ok {
			return nil, transform.NewTree, fmt.Errorf("INSERT: non-Doltgres type found in destination: %s", col.Type.String())
		}
		destinationNameToType[col.Name] = colType
	}
	// Create the destination type slice that will match each inserted column
	destinationTypes := make([]pgtypes.DoltgresType, len(insertInto.ColumnNames))
	for i, colName := range insertInto.ColumnNames {
		destinationTypes[i], ok = destinationNameToType[colName]
		if !ok {
			return nil, transform.NewTree, fmt.Errorf("INSERT: cannot find destination column with name `%s`", colName)
		}
	}
	// Replace expressions with casts as needed
	if values, ok := insertInto.Source.(*plan.Values); ok {
		// Values do not return the correct Schema since each row may contain different types, so we must handle it differently
		newValues := make([][]sql.Expression, len(values.ExpressionTuples))
		for rowIndex, rowExprs := range values.ExpressionTuples {
			newValues[rowIndex] = make([]sql.Expression, len(rowExprs))
			for columnIndex, colExpr := range rowExprs {
				fromColType, ok := colExpr.Type().(pgtypes.DoltgresType)
				if !ok {
					return nil, transform.NewTree, fmt.Errorf("INSERT: non-Doltgres type found in values source: %s", fromColType.String())
				}
				toColType := destinationTypes[columnIndex]
				// We only assign the existing expression if the types perfectly match (same parameters), otherwise we'll cast
				if fromColType.Equals(toColType) {
					newValues[rowIndex][columnIndex] = colExpr
				} else {
					newValues[rowIndex][columnIndex] = pgexprs.NewAssignmentCast(colExpr, fromColType, toColType)
				}
			}
		}
		return insertInto.WithSource(plan.NewValues(newValues)), transform.NewTree, nil
	} else {
		sourceSchema := insertInto.Source.Schema()
		projections := make([]sql.Expression, len(sourceSchema))
		for i, col := range sourceSchema {
			fromColType, ok := col.Type.(pgtypes.DoltgresType)
			if !ok {
				return nil, transform.NewTree, fmt.Errorf("INSERT: non-Doltgres type found in source: %s", fromColType.String())
			}
			toColType := destinationTypes[i]
			getField := expression.NewGetField(i, fromColType, col.Name, true)
			// We only assign the GetField if the types perfectly match (same parameters), otherwise we'll cast
			if fromColType.Equals(toColType) {
				projections[i] = getField
			} else {
				projections[i] = pgexprs.NewAssignmentCast(getField, fromColType, toColType)
			}
		}
		return insertInto.WithSource(plan.NewProject(projections, insertInto.Source)), transform.NewTree, nil
	}
}
