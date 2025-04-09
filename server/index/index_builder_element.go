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

package index

import (
	"sort"

	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/index"
	"github.com/dolthub/go-mysql-server/sql"

	pgexprs "github.com/dolthub/doltgresql/server/expression"
	"github.com/dolthub/doltgresql/server/functions/framework"
)

// indexBuilderElement is an element within the IndexBuilder, comprised of an index and any related information.
type indexBuilderElement struct {
	index       sql.Index
	columns     []indexBuilderColumn
	columnMap   map[string]int
	needsFilter bool
}

// ColumnCount returns the number of columns that have expressions declared on them. If the column count is less than
// the number of columns in the index, then the index's prefix will be used.
func (element *indexBuilderElement) ColumnCount() int {
	count := 0
	for _, column := range element.columns {
		if len(column.exprs) == 0 {
			break
		}
		count++
	}
	return count
}

// ToRange returns the element as a range.
func (element *indexBuilderElement) ToRange(ctx *sql.Context) index.DoltgresRange {
	// The start and stop expressions will determine where the iterator should begin and end, while the filter
	// expressions will determine whether a specific row should be returned by the iterator. The iterator finds its
	// starting point as the first position where all expressions are true (pushing the start as far back as possible).
	// The stopping point is the first position where at least one expression is true (pushing the stop as far forward
	// as possible).
	var startExprs []sql.Expression
	var stopExprs []sql.Expression
	var filterExprs []sql.Expression
	// Postgres indexes only consider columns further in the index when the previous column was an "equals". This is due
	// to the fact that any other comparison type will always have to traverse all entries within its range anyway, so
	// they can be excluded.
	// https://www.postgresql.org/docs/15/indexes-multicolumn.html
	lastIndexEqual := true
	// This index represents the expressions that should be processed in the iterator to determine whether a specific
	// row should be returned, but should not determine where the iterator begins or ends. The starting index will
	// usually correspond to the first index where `lastIndexEqual == false`.
	filterExprsStartingIndex := len(element.columns)
	for columnIndex, column := range element.columns {
		// Ensure that this column has at least one expression. We're guaranteed to have a valid element, as this is
		// only called when the element will, at a minimum, use the partial index. Also, if the last column index was
		// not an "equals" (=), then we'll also use the partial index.
		if len(column.exprs) == 0 || !lastIndexEqual {
			// Set the filter index as well
			filterExprsStartingIndex = columnIndex
			break
		}
		// If there are multiple expressions on this column, then we'll consider that to be equivalent to having a
		// non-equals strategy.
		if len(column.exprs) > 1 {
			lastIndexEqual = false
		}
		for _, columnExpr := range column.exprs {
			switch columnExpr.strategy {
			case OperatorStrategyNumber_Less:
				lastIndexEqual = false
				stopExprs = append(stopExprs, getQuickFunction(framework.GetBinaryFunction(framework.Operator_BinaryGreaterOrEqual).
					Compile("index_less_stop", columnExpr.column, columnExpr.literal)))
			case OperatorStrategyNumber_LessEquals:
				lastIndexEqual = false
				stopExprs = append(stopExprs, getQuickFunction(framework.GetBinaryFunction(framework.Operator_BinaryGreaterThan).
					Compile("index_less_equals_stop", columnExpr.column, columnExpr.literal)))
			case OperatorStrategyNumber_Equals:
				startExprs = append(startExprs, getQuickFunction(framework.GetBinaryFunction(framework.Operator_BinaryGreaterOrEqual).
					Compile("index_equals_start", columnExpr.column, columnExpr.literal)))
				stopExprs = append(stopExprs, getQuickFunction(framework.GetBinaryFunction(framework.Operator_BinaryGreaterThan).
					Compile("index_equals_stop", columnExpr.column, columnExpr.literal)))
			case OperatorStrategyNumber_GreaterEquals:
				lastIndexEqual = false
				startExprs = append(startExprs, getQuickFunction(framework.GetBinaryFunction(framework.Operator_BinaryGreaterOrEqual).
					Compile("index_greater_equals_start", columnExpr.column, columnExpr.literal)))
			case OperatorStrategyNumber_Greater:
				lastIndexEqual = false
				startExprs = append(startExprs, getQuickFunction(framework.GetBinaryFunction(framework.Operator_BinaryGreaterThan).
					Compile("index_greater_start", columnExpr.column, columnExpr.literal)))
			}
		}
	}
	// Now we grab all the filter expressions
	for columnIndex := filterExprsStartingIndex; columnIndex < len(element.columns); columnIndex++ {
		column := element.columns[columnIndex]
		for _, expr := range column.exprs {
			filterExprs = append(filterExprs, expr.original)
		}
	}
	// If either the start or stop expressions are empty, then we'll insert literal boolean expressions.
	// For start expressions, giving the expression literal "true" will cause the iterator to match the very beginning,
	// since the binary search algorithm finds the first value where "true" is returned. With the "true" literal, we'll
	// essentially force the iterator to start from the beginning. When there are multiple expressions, the expression
	// literal "true" becomes redundant, as search requires that all expressions are "true" for that tuple to be
	// iterated over.
	// Stop expressions use the same binary search algorithm that start expressions do, however there is an important
	// nuance that must be considered. The expression literal "false" will push the iterator to the end, which is what
	// we want in some scenarios. However, since search needs only a single "false" to push the iterator forward, we can
	// end up forcing the iterator to push to the end if a single expression literal "false" is present. This causes the
	// iterator to iterate over too many elements, which can return incorrect results if the range is considered to be a
	// precise match (which removes the high-level filter as an optimization, or some portion of the high-level filter).
	//
	// For a concrete example, let's look at the filter (BETWEEN 4 AND 7). This is equivalent to (x >= 4 AND x <= 7).
	// We can rewrite these as ranges using set notation, which would be [4, ∞) ∩ (∞, 7].
	// We can individually represent [4, ∞) as expressions: start(>=4) | stop(false)
	// We can also represent (∞, 7] as expressions: start(true) | stop(>7)
	// Combining the expressions naively would give: start(>=4, true) | stop(false, >7)
	// Although our start expressions will return the correct results, our stop expressions will always be pushed to the
	// end as the >7 expression is essentially ignored. This is why we only add our literal boolean expressions at the
	// end.
	// This makes sense from the set notation perspective as well: [4, ∞) ∩ (∞, 7] === [4, 7]
	if len(startExprs) == 0 {
		startExprs = []sql.Expression{pgexprs.NewRawLiteralBool(true)}
	}
	if len(stopExprs) == 0 {
		stopExprs = []sql.Expression{pgexprs.NewRawLiteralBool(false)}
	}
	return index.DoltgresRange{
		StartExpressions:  startExprs,
		StopExpressions:   stopExprs,
		FilterExpressions: filterExprs,
		PreciseMatch:      !element.needsFilter,
	}
}

// SortStrategiesByRestrictiveness sorts the strategies for each column such that the most restrictive come first.
func (element *indexBuilderElement) SortStrategiesByRestrictiveness() {
	for _, column := range element.columns {
		if len(column.exprs) == 0 {
			return
		}
		sort.Slice(column.exprs, func(i, j int) bool {
			return column.exprs[i].strategy.IsMoreRestrictive(column.exprs[j].strategy)
		})
	}
}

// getQuickFunction returns the framework.QuickFunction form of this function, if it exists. If one does not exist, then
// this returns the original framework.CompiledFunction.
func getQuickFunction(c *framework.CompiledFunction) sql.FunctionExpression {
	qf := c.GetQuickFunction()
	if qf != nil {
		return qf
	}
	return c
}
