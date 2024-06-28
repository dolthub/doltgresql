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
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/analyzer"
	"github.com/dolthub/go-mysql-server/sql/expression"

	pgexprs "github.com/dolthub/doltgresql/server/expression"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// IndexLeafChildren overrides the IndexLeafChildren function in GMS to use Doltgres types.
func IndexLeafChildren(e sql.Expression) (analyzer.IndexScanOp, sql.Expression, sql.Expression, bool) {
	var op analyzer.IndexScanOp
	var left sql.Expression
	var right sql.Expression
	switch expr := e.(type) {
	case *pgexprs.BinaryOperator:
		switch expr.Operator() {
		case framework.Operator_BinaryEqual:
			op = analyzer.IndexScanOpEq
			left = expr.Left()
			right = expr.Right()
		case framework.Operator_BinaryGreaterOrEqual:
			op = analyzer.IndexScanOpGte
			left = expr.Left()
			right = expr.Right()
		case framework.Operator_BinaryGreaterThan:
			op = analyzer.IndexScanOpGt
			left = expr.Left()
			right = expr.Right()
		case framework.Operator_BinaryLessOrEqual:
			op = analyzer.IndexScanOpLte
			left = expr.Left()
			right = expr.Right()
		case framework.Operator_BinaryLessThan:
			op = analyzer.IndexScanOpLt
			left = expr.Left()
			right = expr.Right()
		default:
			return 0, nil, nil, false
		}
	default:
		return 0, nil, nil, false
	}
	// GMS indexes do not use our operator functions, so they fail for comparisons between different types.
	// As a workaround, we'll cast the value to the same type as the GetField.
	// GMS only uses indexes when one of the two expressions is a GetField, so we can make use of that restriction here.
	// We need to transition indexes to use our operator functions, which will require a complete overhaul of indexing.
	if _, ok := left.(*expression.GetField); !ok {
		left, right = right, left
		op = op.Swap()
	}
	getField, ok := left.(*expression.GetField)
	if !ok {
		return 0, nil, nil, false
	}
	leftType, ok := getField.Type().(pgtypes.DoltgresType)
	if !ok {
		return 0, nil, nil, false
	}
	rightType, ok := right.Type().(pgtypes.DoltgresType)
	if !ok {
		return 0, nil, nil, false
	}
	if !leftType.Equals(rightType) {
		right = pgexprs.NewExplicitCast(right, leftType)
	}
	return op, left, right, true
}
