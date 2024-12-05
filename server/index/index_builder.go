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
	"fmt"
	"strings"

	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/index"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/expression"

	pgexprs "github.com/dolthub/doltgresql/server/expression"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// OperatorStrategyNumber corresponds to the strategy number used within an operator class. These are specifically for
// B-Tree indexes, as other index types use different numbers to represent their strategy types.
// https://www.postgresql.org/docs/current/sql-createopclass.html
type OperatorStrategyNumber uint8

const (
	OperatorStrategyNumber_Less OperatorStrategyNumber = iota + 1
	OperatorStrategyNumber_LessEquals
	OperatorStrategyNumber_Equals
	OperatorStrategyNumber_GreaterEquals
	OperatorStrategyNumber_Greater
)

// IndexBuilder builds an index following the same rules as a B-Tree index in Postgres.
type IndexBuilder struct {
	elements []*indexBuilderElement
}

// NewIndexBuilder creates a new IndexBuilder for the given indexes.
func NewIndexBuilder(ctx *sql.Context, indexes []sql.Index) (*IndexBuilder, error) {
	elements := make([]*indexBuilderElement, len(indexes))
	for elementsIndex, index := range indexes {
		columnExpressionTypes := index.ColumnExpressionTypes()
		columns := make([]indexBuilderColumn, len(columnExpressionTypes))
		columnMap := make(map[string]int)
		for columnIndex, columnType := range index.ColumnExpressionTypes() {
			// Dolt uses the form "table_name.column_name" for its index expressions, so we must remove the table name
			// to get the column name.
			columnName := strings.Replace(columnType.Expression, index.Table()+".", "", 1)
			columnMap[columnName] = columnIndex
			var ok bool
			columns[columnIndex].typ, ok = columnType.Type.(*pgtypes.DoltgresType)
			if !ok {
				return nil, fmt.Errorf("encountered a GMS type in the index `%s` on table `%s`", index.ID(), index.Table())
			}
		}
		elements[elementsIndex] = &indexBuilderElement{
			index:       index,
			columns:     columns,
			columnMap:   columnMap,
			needsFilter: false,
		}
	}
	return &IndexBuilder{
		elements: elements,
	}, nil
}

// AddExpression adds the given expression to the builder.
func (ib *IndexBuilder) AddExpression(ctx *sql.Context, expr sql.Expression) {
	indexExpr := ib.convertExpression(ctx, expr)
	// An invalid strategy is returned when the expression will not conform to a B-Tree index
	if !indexExpr.isValid {
		// Since we've encountered an expression that cannot be handled by the index iterator, we must require the use
		// of a filter
		for _, element := range ib.elements {
			element.needsFilter = true
		}
		return
	}
	for _, element := range ib.elements {
		columnIndex, ok := element.columnMap[indexExpr.column.Name()]
		if !ok {
			// This involves a column that is not in the index, so we must require the use of a filter
			element.needsFilter = true
			continue
		}
		column := &element.columns[columnIndex]
		column.exprs = append(column.exprs, indexExpr.withIndex(columnIndex))
	}
}

// GetLookup returns a fully-formed index lookup. If an index lookup could not be created, then it will contain a nil
// index.
func (ib *IndexBuilder) GetLookup(ctx *sql.Context) sql.IndexLookup {
	var targetElement *indexBuilderElement
	for _, element := range ib.elements {
		columnCount := element.ColumnCount()
		// We'll skip if no columns have expressions
		if columnCount == 0 {
			continue
		}
		// If we've not yet selected an element, then we'll use this one.
		if targetElement == nil {
			element.SortStrategiesByRestrictiveness()
			targetElement = element
			continue
		}
		targetColumnCount := targetElement.ColumnCount()
		// If this new element has more matching columns, then it should be used.
		if columnCount > targetColumnCount {
			element.SortStrategiesByRestrictiveness()
			targetElement = element
			continue
		} else if columnCount < targetColumnCount {
			continue
		}
		// The counts are equal, so we'll just choose one by using the index's name for determinism.
		if element.index.ID() < targetElement.index.ID() {
			element.SortStrategiesByRestrictiveness()
			targetElement = element
			continue
		}
	}
	// If we were unable to match an element, then the expressions given will not make use of an index.
	if targetElement == nil {
		return sql.IndexLookup{}
	}
	// Grab the range and return the lookup
	rang := targetElement.ToRange(ctx)
	return sql.IndexLookup{
		Index:           targetElement.index,
		Ranges:          index.DoltgresRangeCollection{rang},
		IsPointLookup:   false,
		IsEmptyRange:    false,
		IsSpatialLookup: false,
		IsReverse:       false,
	}
}

// convertExpression converts an expression into the form needed by the index builder.
func (ib *IndexBuilder) convertExpression(ctx *sql.Context, expr sql.Expression) indexBuilderExpr {
	// TODO: OR should create separate ranges
	switch expr := expr.(type) {
	case *pgexprs.BinaryOperator:
		return ib.convertBinaryExpression(ctx, expr)
	case *pgexprs.InTuple:
		return ib.convertInExpression(ctx, expr)
	default:
		// The expression must not be valid for use with an index, so we'll return an invalid builder expression
		return indexBuilderExpr{isValid: false}
	}
}

// convertBinaryExpression is called by convertExpression to handle binary operators.
func (ib *IndexBuilder) convertBinaryExpression(ctx *sql.Context, expr *pgexprs.BinaryOperator) indexBuilderExpr {
	operator := expr.Operator()
	var valueExpr sql.Expression
	getField, ok := expr.Left().(*expression.GetField)
	if ok {
		valueExpr = expr.Right()
	} else {
		// Postgres requires that an operator have the `COMMUTATOR` clause in order to allow rearranging the values.
		// Without it, the operator is assumed to not be commutative. Since Postgres requires that the column is on
		// the left, this results in the expression not being valid for index usage.
		// https://www.postgresql.org/docs/15/xoper-optimization.html
		// TODO: look up whether the specific operator has the `COMMUTATOR` defined, rather than assuming it does
		getField, ok = expr.Right().(*expression.GetField)
		if !ok {
			return indexBuilderExpr{isValid: false}
		}
		valueExpr = expr.Left()
		switch operator {
		case framework.Operator_BinaryGreaterThan:
			operator = framework.Operator_BinaryLessThan
		case framework.Operator_BinaryGreaterOrEqual:
			operator = framework.Operator_BinaryLessOrEqual
		case framework.Operator_BinaryLessThan:
			operator = framework.Operator_BinaryGreaterThan
		case framework.Operator_BinaryLessOrEqual:
			operator = framework.Operator_BinaryGreaterOrEqual
		}
	}
	valueType, ok := valueExpr.Type().(*pgtypes.DoltgresType)
	if !ok {
		return indexBuilderExpr{isValid: false}
	}
	// TODO: investigate whether index values must be literals or functions that can be evaluated ahead-of-time, also side-effects
	value, err := valueExpr.Eval(ctx, nil)
	if err != nil {
		return indexBuilderExpr{isValid: false}
	}
	// TODO: check that the operator for the field and value types belong to an operator class
	switch operator {
	case framework.Operator_BinaryEqual:
		return indexBuilderExpr{
			isValid:  true,
			strategy: OperatorStrategyNumber_Equals,
			column:   getField,
			literal:  pgexprs.NewUnsafeLiteral(value, valueType),
			original: expr,
		}
	case framework.Operator_BinaryGreaterThan:
		return indexBuilderExpr{
			isValid:  true,
			strategy: OperatorStrategyNumber_Greater,
			column:   getField,
			literal:  pgexprs.NewUnsafeLiteral(value, valueType),
			original: expr,
		}
	case framework.Operator_BinaryGreaterOrEqual:
		return indexBuilderExpr{
			isValid:  true,
			strategy: OperatorStrategyNumber_GreaterEquals,
			column:   getField,
			literal:  pgexprs.NewUnsafeLiteral(value, valueType),
			original: expr,
		}
	case framework.Operator_BinaryLessThan:
		return indexBuilderExpr{
			isValid:  true,
			strategy: OperatorStrategyNumber_Less,
			column:   getField,
			literal:  pgexprs.NewUnsafeLiteral(value, valueType),
			original: expr,
		}
	case framework.Operator_BinaryLessOrEqual:
		return indexBuilderExpr{
			isValid:  true,
			strategy: OperatorStrategyNumber_LessEquals,
			column:   getField,
			literal:  pgexprs.NewUnsafeLiteral(value, valueType),
			original: expr,
		}
	default:
		// We only support the above operator types, since those are the ones used for B-Tree indexing
		return indexBuilderExpr{isValid: false}
	}
}

// convertInExpression is called by convertExpression to handle IN expressions.
func (ib *IndexBuilder) convertInExpression(ctx *sql.Context, expr *pgexprs.InTuple) indexBuilderExpr {
	// TODO: implement this
	return indexBuilderExpr{isValid: false}
}

// IsMoreRestrictive returns whether the calling strategy is more restrictive than the given strategy. The more
// restrictive strategy is assumed to match fewer rows, however this may not be true depending on the values involved.
func (strategy OperatorStrategyNumber) IsMoreRestrictive(other OperatorStrategyNumber) bool {
	switch strategy {
	case OperatorStrategyNumber_Less, OperatorStrategyNumber_Greater:
		if other == OperatorStrategyNumber_LessEquals || other == OperatorStrategyNumber_GreaterEquals {
			return true
		}
	case OperatorStrategyNumber_LessEquals, OperatorStrategyNumber_GreaterEquals:
		// <= and >= are the least restrictive variants
		return false
	case OperatorStrategyNumber_Equals:
		if other != OperatorStrategyNumber_Equals {
			return true
		}
	}
	return false
}
