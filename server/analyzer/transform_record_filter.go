package analyzer

import (
	"github.com/dolthub/doltgresql/server/expression"
	"github.com/dolthub/doltgresql/server/functions/framework"
	"github.com/dolthub/doltgresql/server/types"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/analyzer"
	gmsexpr "github.com/dolthub/go-mysql-server/sql/expression"
	"github.com/dolthub/go-mysql-server/sql/plan"
	"github.com/dolthub/go-mysql-server/sql/transform"
)

func TransformRecordFilter(ctx *sql.Context, a *analyzer.Analyzer, n sql.Node, scope *plan.Scope, selector analyzer.RuleSelector, qFlags *sql.QueryFlags) (sql.Node, transform.TreeIdentity, error) {
	// Does this table have primary key with composite index
	newNode, same, err := transform.Node(ctx, n, func(ctx *sql.Context, n sql.Node) (sql.Node, transform.TreeIdentity, error) {
		filter, ok := n.(*plan.Filter)
		if !ok {
			return n, transform.SameTree, nil
		}

		var tblNode sql.TableNode
		switch child := filter.Child.(type) {
		case *plan.ResolvedTable:
			tblNode = child
		case *plan.TableAlias:
			tblNode, _ = child.Child.(sql.TableNode)
		default:
			return n, transform.SameTree, nil
		}

		idxAddrTbl, ok := tblNode.UnderlyingTable().(sql.IndexAddressableTable)
		if !ok {
			return n, transform.SameTree, nil
		}

		indexes, err := idxAddrTbl.GetIndexes(ctx)
		if err != nil {
			return n, transform.SameTree, err
		}

		// TODO: only do this if there's a matching primary key?
		if indexes == nil {
		}

		binExpr, ok := filter.Expression.(*expression.BinaryOperator)
		if !ok {
			return n, transform.SameTree, nil
		}
		// TODO: possible arguments have swapped sides?
		lRec, ok := binExpr.Left().(*expression.RecordExpr)
		if !ok {
			return n, transform.SameTree, nil
		}
		rLit, ok := binExpr.Right().(*gmsexpr.Literal)
		if !ok {
			return n, transform.SameTree, nil
		}
		lExprs := lRec.Expressions()
		rExprs, ok := rLit.Val.([]types.RecordValue)
		if !ok {
			return n, transform.SameTree, nil
		}
		if len(lExprs) != len(rExprs) {
			return n, transform.SameTree, nil
		}

		switch binExpr.Operator() {
		case framework.Operator_BinaryEqual:
			exprs := make([]sql.Expression, len(lExprs))
			for i := 0; i < len(lExprs); i++ {
				newLit := gmsexpr.NewLiteral(rExprs[i].Value, rExprs[i].Type)
				expr, err := expression.NewBinaryOperator(framework.Operator_BinaryEqual).WithChildren(ctx, lExprs[i], newLit)
				if err != nil {
					return nil, transform.NewTree, err
				}
				exprs[i] = expr
			}
			newExpr := gmsexpr.JoinAnd(exprs...)
			newFilter := plan.NewFilter(ctx, newExpr, filter.Child)
			return newFilter, transform.NewTree, nil
		//case framework.Operator_BinaryLessThan:
		//case framework.Operator_BinaryLessOrEqual:
		//case framework.Operator_BinaryGreaterThan:
		//case framework.Operator_BinaryGreaterOrEqual:
		//case framework.Operator_BinaryNotEqual:
		default:
			return n, transform.SameTree, nil
		}
	})
	return newNode, same, err
}
