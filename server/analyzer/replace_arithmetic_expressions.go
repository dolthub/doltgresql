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
	gms_expression "github.com/dolthub/go-mysql-server/sql/expression"
	"github.com/dolthub/go-mysql-server/sql/plan"
	"github.com/dolthub/go-mysql-server/sql/transform"

	"github.com/dolthub/doltgresql/server/expression"
	"github.com/dolthub/doltgresql/server/functions/framework"
)

// ReplaceArithmeticExpressions replaces arithmetic expressions in the tree with the equivalent binary operators.
// This is a problem in certain nodes that do arithmetic, such as TopN, but all nodes are covered.
func ReplaceArithmeticExpressions(ctx *sql.Context, a *analyzer.Analyzer, node sql.Node, scope *plan.Scope, selector analyzer.RuleSelector, qFlags *sql.QueryFlags) (sql.Node, transform.TreeIdentity, error) {
	return transform.NodeExprsWithOpaque(node, func(e sql.Expression) (sql.Expression, transform.TreeIdentity, error) {
		switch e := e.(type) {
		case *gms_expression.Arithmetic:
			switch e.Operator() {
			case "+":
				expr, err := expression.NewBinaryOperator(framework.Operator_BinaryPlus).WithChildren(e.Children()...)
				if err != nil {
					return nil, transform.NewTree, err
				}
				return expr.(*expression.BinaryOperator), transform.NewTree, nil
			case "-":
				expr, err := expression.NewBinaryOperator(framework.Operator_BinaryMinus).WithChildren(e.Children()...)
				if err != nil {
					return nil, transform.NewTree, err
				}
				return expr.(*expression.BinaryOperator), transform.NewTree, nil
			case "*":
				expr, err := expression.NewBinaryOperator(framework.Operator_BinaryMultiply).WithChildren(e.Children()...)
				if err != nil {
					return nil, transform.NewTree, err
				}
				return expr.(*expression.BinaryOperator), transform.NewTree, nil
			case "/":
				expr, err := expression.NewBinaryOperator(framework.Operator_BinaryDivide).WithChildren(e.Children()...)
				if err != nil {
					return nil, transform.NewTree, err
				}
				return expr.(*expression.BinaryOperator), transform.NewTree, nil
			}
		}
		return e, transform.SameTree, nil
	})
}
