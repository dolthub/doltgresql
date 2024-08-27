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
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/expression"

	pgexprs "github.com/dolthub/doltgresql/server/expression"
)

// indexBuilderExpr breaks an expression down such that a range can be constructed from it. In addition, also carries
// the original expression which is used for validation within an index iterator.
type indexBuilderExpr struct {
	isValid  bool
	strategy OperatorStrategyNumber
	column   *expression.GetField
	literal  *pgexprs.Literal
	original sql.Expression
}

// withIndex returns a new expression with the given index for the GetField expression. This is due to how GMS handles
// GetField differently when working with indexes.
func (expr indexBuilderExpr) withIndex(columnIndex int) indexBuilderExpr {
	if expr.column.Index() == columnIndex {
		return expr
	}
	newIdxExpr := expr
	newIdxExpr.column = newIdxExpr.column.WithIndex(columnIndex).(*expression.GetField)
	originalExprChildren := expr.original.Children()
	newOriginalExprChildren := make([]sql.Expression, len(originalExprChildren))
	copy(newOriginalExprChildren, originalExprChildren)
	for i, child := range originalExprChildren {
		if getField, ok := child.(*expression.GetField); ok {
			newOriginalExprChildren[i] = getField.WithIndex(columnIndex)
			break
		}
	}
	newIdxExpr.original, _ = expr.original.WithChildren(newOriginalExprChildren...)
	return newIdxExpr
}
