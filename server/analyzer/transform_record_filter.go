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

// TransformRecordFilter finds Filter nodes over Indexable tables with composite indexes over Record columns and
// decomposes the Record comparison into an equivalent filter expression over each column joined by ANDs and ORs.
func TransformRecordFilter(
	ctx *sql.Context,
	_ *analyzer.Analyzer,
	node sql.Node,
	_ *plan.Scope,
	_ analyzer.RuleSelector,
	_ *sql.QueryFlags,
) (sql.Node, transform.TreeIdentity, error) {
	return transform.Node(ctx, node, func(ctx *sql.Context, n sql.Node) (sql.Node, transform.TreeIdentity, error) {
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
		// TODO: should only convert expressions when there's applicable index
		if _, ok = tblNode.UnderlyingTable().(sql.IndexAddressableTable); !ok {
			return n, transform.SameTree, nil
		}

		newExpr, same, err := decomposeRecordFilter(ctx, filter.Expression)
		if err != nil {
			return nil, transform.SameTree, err
		}
		if same {
			return n, transform.SameTree, nil
		}
		return plan.NewFilter(ctx, newExpr, filter.Child), transform.NewTree, nil
	})
}

// decomposeRecordFilter is the helper function to TransformRecordFilter that decomposes the Record comparison.
func decomposeRecordFilter(ctx *sql.Context, expr sql.Expression) (sql.Expression, transform.TreeIdentity, error) {
	return transform.Expr(ctx, expr,
		func(ctx *sql.Context, e sql.Expression) (sql.Expression, transform.TreeIdentity, error) {
			binExpr, ok := e.(*expression.BinaryOperator)
			if !ok {
				return e, transform.SameTree, nil
			}
			// TODO: possible for Literal to be on left
			recExpr, ok := binExpr.Left().(*expression.RecordExpr)
			if !ok {
				return e, transform.SameTree, nil
			}
			recExprs := recExpr.Expressions()

			// TODO: possible for RecordExpr to be on right
			litExpr, ok := binExpr.Right().(*gmsexpr.Literal)
			if !ok {
				return e, transform.SameTree, nil
			}
			recVals, ok := litExpr.Val.([]types.RecordValue)
			if !ok {
				return e, transform.SameTree, nil
			}

			var newExpr sql.Expression
			var err error
			switch binExpr.Operator() {
			case framework.Operator_BinaryEqual:
				newExpr, err = decomposeRecordFilterEquals(ctx, recExprs, recVals)
			case framework.Operator_BinaryLessThan:
				newExpr, err = decomposeRecordFilterLessThan(ctx, recExprs, recVals)
			case framework.Operator_BinaryGreaterThan:
				newExpr, err = decomposeRecordFilterGreaterThan(ctx, recExprs, recVals)
			case framework.Operator_BinaryLessOrEqual:
				newExpr, err = decomposeRecordFilterLessThanEquals(ctx, recExprs, recVals)
			case framework.Operator_BinaryGreaterOrEqual:
				newExpr, err = decomposeRecordFilterGreaterThanEquals(ctx, recExprs, recVals)
			default:
				return e, transform.SameTree, nil
			}
			if err != nil {
				return nil, transform.SameTree, err
			}
			return newExpr, transform.NewTree, nil
		})
}

func decomposeRecordFilterEquals(
	ctx *sql.Context,
	recExprs []sql.Expression,
	recVals []types.RecordValue,
) (sql.Expression, error) {
	n := len(recExprs)
	exprs := make([]sql.Expression, n)
	for i := 0; i < n; i++ {
		newLit := gmsexpr.NewLiteral(recVals[i].Value, recVals[i].Type)
		expr, err := expression.NewBinaryOperator(framework.Operator_BinaryEqual).WithChildren(ctx, recExprs[i], newLit)
		if err != nil {
			return nil, err
		}
		exprs[i] = expr
	}
	return gmsexpr.JoinAnd(exprs...), nil
}

func decomposeRecordFilterLessThan(
	ctx *sql.Context,
	recExprs []sql.Expression,
	recVals []types.RecordValue,
) (sql.Expression, error) {
	n := len(recExprs)
	orExprs := make([]sql.Expression, n)
	for i := 0; i < n; i++ {
		andExprs := make([]sql.Expression, n-i)
		for j := 0; j < n-i; j++ {
			newLit := gmsexpr.NewLiteral(recVals[j].Value, recVals[j].Type)
			expr, err := expression.NewBinaryOperator(framework.Operator_BinaryEqual).WithChildren(
				ctx,
				recExprs[j],
				newLit,
			)
			if err != nil {
				return nil, err
			}
			andExprs[j] = expr
		}
		newLit := gmsexpr.NewLiteral(recVals[n-i-1].Value, recVals[n-i-1].Type)
		expr, err := expression.NewBinaryOperator(framework.Operator_BinaryLessThan).WithChildren(
			ctx,
			recExprs[i],
			newLit,
		)
		if err != nil {
			return nil, err
		}
		andExprs[n-i-1] = expr
		orExprs[i] = gmsexpr.JoinAnd(andExprs...)
	}
	return gmsexpr.JoinOr(orExprs...), nil
}

func decomposeRecordFilterGreaterThan(
	ctx *sql.Context,
	recExprs []sql.Expression,
	recVals []types.RecordValue,
) (sql.Expression, error) {
	n := len(recExprs)
	orExprs := make([]sql.Expression, n)
	for i := 0; i < n; i++ {
		andExprs := make([]sql.Expression, n-i)
		for j := 0; j < n-i; j++ {
			newLit := gmsexpr.NewLiteral(recVals[j].Value, recVals[j].Type)
			expr, err := expression.NewBinaryOperator(framework.Operator_BinaryEqual).WithChildren(
				ctx,
				recExprs[j],
				newLit,
			)
			if err != nil {
				return nil, err
			}
			andExprs[j] = expr
		}
		newLit := gmsexpr.NewLiteral(recVals[n-i-1].Value, recVals[n-i-1].Type)
		expr, err := expression.NewBinaryOperator(framework.Operator_BinaryGreaterThan).WithChildren(
			ctx,
			recExprs[i],
			newLit,
		)
		if err != nil {
			return nil, err
		}
		andExprs[n-i-1] = expr
		orExprs[i] = gmsexpr.JoinAnd(andExprs...)
	}
	return gmsexpr.JoinOr(orExprs...), nil
}

func decomposeRecordFilterLessThanEquals(
	ctx *sql.Context,
	recExprs []sql.Expression,
	recVals []types.RecordValue,
) (sql.Expression, error) {
	n := len(recExprs)
	orExprs := make([]sql.Expression, n)
	andExprs := make([]sql.Expression, n)
	for i := 0; i < n-1; i++ {
		newLit := gmsexpr.NewLiteral(recVals[i].Value, recVals[i].Type)
		expr, err := expression.NewBinaryOperator(framework.Operator_BinaryEqual).WithChildren(
			ctx,
			recExprs[i],
			newLit,
		)
		if err != nil {
			return nil, err
		}
		andExprs[i] = expr
	}
	newLit := gmsexpr.NewLiteral(recVals[n-1].Value, recVals[n-1].Type)
	expr, err := expression.NewBinaryOperator(framework.Operator_BinaryLessOrEqual).WithChildren(
		ctx,
		recExprs[n-1],
		newLit,
	)
	if err != nil {
		return nil, err
	}
	andExprs[n-1] = expr
	orExprs[0] = gmsexpr.JoinAnd(andExprs...)

	for i := 1; i < n; i++ {
		andExprs = make([]sql.Expression, n-i)
		for j := 0; j < n-i; j++ {
			newLit = gmsexpr.NewLiteral(recVals[j].Value, recVals[j].Type)
			expr, err = expression.NewBinaryOperator(framework.Operator_BinaryEqual).WithChildren(
				ctx,
				recExprs[j],
				newLit,
			)
			if err != nil {
				return nil, err
			}
			andExprs[j] = expr
		}
		newLit = gmsexpr.NewLiteral(recVals[n-i-1].Value, recVals[n-i-1].Type)
		expr, err = expression.NewBinaryOperator(framework.Operator_BinaryLessThan).WithChildren(
			ctx,
			recExprs[n-i-1],
			newLit,
		)
		if err != nil {
			return nil, err
		}
		andExprs[n-i-1] = expr
		orExprs[i] = gmsexpr.JoinAnd(andExprs...)
	}
	return gmsexpr.JoinOr(orExprs...), nil
}

func decomposeRecordFilterGreaterThanEquals(
	ctx *sql.Context,
	recExprs []sql.Expression,
	recVals []types.RecordValue,
) (sql.Expression, error) {
	n := len(recExprs)
	orExprs := make([]sql.Expression, n)
	andExprs := make([]sql.Expression, n)
	for i := 0; i < n-1; i++ {
		newLit := gmsexpr.NewLiteral(recVals[i].Value, recVals[i].Type)
		expr, err := expression.NewBinaryOperator(framework.Operator_BinaryEqual).WithChildren(
			ctx,
			recExprs[i],
			newLit,
		)
		if err != nil {
			return nil, err
		}
		andExprs[i] = expr
	}

	newLit := gmsexpr.NewLiteral(recVals[n-1].Value, recVals[n-1].Type)
	expr, err := expression.NewBinaryOperator(framework.Operator_BinaryGreaterOrEqual).WithChildren(
		ctx,
		recExprs[n-1],
		newLit,
	)
	if err != nil {
		return nil, err
	}
	andExprs[n-1] = expr
	orExprs[0] = gmsexpr.JoinAnd(andExprs...)

	for i := 1; i < n; i++ {
		andExprs = make([]sql.Expression, n-i)
		for j := 0; j < n-i; j++ {
			newLit = gmsexpr.NewLiteral(recVals[j].Value, recVals[j].Type)
			expr, err = expression.NewBinaryOperator(framework.Operator_BinaryEqual).WithChildren(
				ctx,
				recExprs[j],
				newLit,
			)
			if err != nil {
				return nil, err
			}
			andExprs[j] = expr
		}
		newLit = gmsexpr.NewLiteral(recVals[n-i-1].Value, recVals[n-i-1].Type)
		expr, err = expression.NewBinaryOperator(framework.Operator_BinaryGreaterThan).WithChildren(
			ctx,
			recExprs[n-i-1],
			newLit,
		)
		if err != nil {
			return nil, err
		}
		andExprs[n-i-1] = expr
		orExprs[i] = gmsexpr.JoinAnd(andExprs...)
	}
	return gmsexpr.JoinOr(orExprs...), nil
}
