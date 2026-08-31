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

package expression

import (
	"context"
	"fmt"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/expression"
	"github.com/dolthub/go-mysql-server/sql/expression/function/vector"
	"github.com/dolthub/go-mysql-server/sql/plan"
	"github.com/dolthub/go-mysql-server/sql/procedures"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/server/extensions"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// BinaryOperator represents a VALUE OPERATOR VALUE expression.
type BinaryOperator struct {
	operator     framework.Operator
	compiledFunc framework.Function
}

var _ vitess.Injectable = (*BinaryOperator)(nil)
var _ sql.Expression = (*BinaryOperator)(nil)
var _ expression.BinaryExpression = (*BinaryOperator)(nil)
var _ expression.Equality = (*BinaryOperator)(nil)
var _ sql.IndexComparisonExpression = (*BinaryOperator)(nil)
var _ procedures.InterpreterExpr = (*BinaryOperator)(nil)
var _ vector.OrderableDistance = (*BinaryOperator)(nil)

// NewBinaryOperator returns a new *BinaryOperator.
func NewBinaryOperator(operator framework.Operator) *BinaryOperator {
	return &BinaryOperator{operator: operator}
}

// Children implements the sql.Expression interface.
func (b *BinaryOperator) Children() []sql.Expression {
	return b.compiledFunc.Children()
}

// Eval implements the sql.Expression interface.
func (b *BinaryOperator) Eval(ctx *sql.Context, row sql.Row) (any, error) {
	return b.compiledFunc.Eval(ctx, row)
}

// IsNullable implements the sql.Expression interface.
func (b *BinaryOperator) IsNullable(ctx *sql.Context) bool {
	return b.compiledFunc.IsNullable(ctx)
}

// RepresentsEquality implements the expression.Equality interface.
func (b *BinaryOperator) RepresentsEquality() bool {
	return b.operator == framework.Operator_BinaryEqual
}

// Resolved implements the sql.Expression interface.
func (b *BinaryOperator) Resolved() bool {
	return b.compiledFunc.Resolved()
}

// SetStatementRunner implements the procedures.InterpreterExpr interface.
func (b *BinaryOperator) SetStatementRunner(ctx *sql.Context, runner sql.StatementRunner) sql.Expression {
	interpreterExpr, ok := b.compiledFunc.(procedures.InterpreterExpr)
	if !ok {
		return b
	}
	return &BinaryOperator{
		operator:     b.operator,
		compiledFunc: interpreterExpr.SetStatementRunner(ctx, runner).(framework.Function),
	}
}

// String implements the sql.Expression interface.
func (b *BinaryOperator) String() string {
	if b.compiledFunc == nil {
		return fmt.Sprintf("? %s ?", b.operator.String())
	}
	// We know that we'll always have two parameters here
	switch f := b.compiledFunc.(type) {
	case *framework.CompiledFunction:
		return fmt.Sprintf("%s %s %s",
			f.Arguments[0].String(), b.operator.String(), f.Arguments[1].String())
	case *framework.QuickFunction2:
		return fmt.Sprintf("%s %s %s",
			f.Arguments[0].String(), b.operator.String(), f.Arguments[1].String())
	default:
		return fmt.Sprintf("unexpected binary operator function type: %T", b.compiledFunc)
	}
}

// SwapParameters implements the expression.Equality interface.
func (b *BinaryOperator) SwapParameters(ctx *sql.Context) (expression.Equality, error) {
	// TODO: for now we'll assume this is valid, but we should check for the `COMMUTATOR` property on the operator
	f, err := b.WithResolvedChildren(ctx, []any{b.Right(), b.Left()})
	if err != nil {
		return nil, err
	}
	return f.(expression.Equality), nil
}

// ToComparer implements the expression.Equality interface.
func (b *BinaryOperator) ToComparer(ctx *sql.Context) (expression.Comparer, error) {
	return NewJoinComparator(ctx, b)
}

// Type implements the sql.Expression interface.
func (b *BinaryOperator) Type(ctx *sql.Context) sql.Type {
	return b.compiledFunc.Type(ctx)
}

// WithChildren implements the sql.Expression interface.
func (b *BinaryOperator) WithChildren(ctx *sql.Context, children ...sql.Expression) (sql.Expression, error) {
	if len(children) != 2 {
		return nil, sql.ErrInvalidChildrenNumber.New(b, len(children), 2)
	}
	if b.compiledFunc != nil {
		compiledFunc, err := b.compiledFunc.WithChildren(ctx, children...)
		if err != nil {
			return nil, err
		}
		return &BinaryOperator{
			operator:     b.operator,
			compiledFunc: compiledFunc.(framework.Function),
		}, nil
	} else {
		binOp, err := b.WithResolvedChildren(ctx, []any{children[0], children[1]})
		if err != nil {
			return nil, err
		}
		return binOp.(sql.Expression), nil
	}
}

// WithResolvedChildren implements the vitess.InjectableExpression interface.
func (b *BinaryOperator) WithResolvedChildren(ctx context.Context, children []any) (any, error) {
	if len(children) != 2 {
		return nil, errors.Errorf("invalid vitess child count, expected `2` but got `%d`", len(children))
	}
	sqlCtx := ctx.(*sql.Context)
	left, ok := children[0].(sql.Expression)
	if !ok {
		return nil, errors.Errorf("expected vitess child to be an expression but has type `%T`", children[0])
	}
	right, ok := children[1].(sql.Expression)
	if !ok {
		return nil, errors.Errorf("expected vitess child to be an expression but has type `%T`", children[1])
	}
	funcName := "internal_binary_operator_func_" + b.operator.String()
	var compiledFunc framework.Function
	builtInFunc := framework.GetBinaryFunction(b.operator).Compile(sqlCtx, funcName, left, right)
	if builtInFunc != nil && builtInFunc.StashedError() == nil {
		compiledFunc = builtInFunc
	} else {
		userFunc, err := getUserDefinedOperator(sqlCtx, b.operator, left, right)
		if err != nil {
			return nil, err
		}
		if userFunc != nil {
			compiledFunc = userFunc
		} else if builtInFunc != nil {
			compiledFunc = builtInFunc
		}
	}
	if compiledFunc == nil {
		return nil, errors.Errorf("operator does not exist: %s %s %s",
			left.Type(sqlCtx).String(), b.operator.String(), right.Type(sqlCtx).String())
	}
	return &BinaryOperator{
		operator:     b.operator,
		compiledFunc: compiledFunc,
	}, nil
}

// Operator returns the operator that is used.
func (b *BinaryOperator) Operator() framework.Operator {
	return b.operator
}

// Left implements the expression.BinaryExpression interface.
func (b *BinaryOperator) Left() sql.Expression {
	// We know that we'll always have two parameters here
	switch f := b.compiledFunc.(type) {
	case *framework.CompiledFunction:
		return f.Arguments[0]
	case *framework.QuickFunction2:
		return f.Arguments[0]
	default:
		return nil
	}
}

// Right implements the expression.BinaryExpression interface.
func (b *BinaryOperator) Right() sql.Expression {
	// We know that we'll always have two parameters here
	switch f := b.compiledFunc.(type) {
	case *framework.CompiledFunction:
		return f.Arguments[1]
	case *framework.QuickFunction2:
		return f.Arguments[1]
	default:
		return nil
	}
}

// IndexScanOperation implements the sql.IndexComparisonExpression interface.
func (b *BinaryOperator) IndexScanOperation() (sql.IndexScanOp, sql.Expression, sql.Expression, bool) {
	left := unwrapIndexScanTarget(b.Left())
	right := unwrapIndexScanTarget(b.Right())
	switch b.operator {
	case framework.Operator_BinaryEqual:
		return sql.IndexScanOpEq, left, right, true
	case framework.Operator_BinaryLessThan:
		return sql.IndexScanOpLt, left, right, true
	case framework.Operator_BinaryLessOrEqual:
		return sql.IndexScanOpLte, left, right, true
	case framework.Operator_BinaryGreaterThan:
		return sql.IndexScanOpGt, left, right, true
	case framework.Operator_BinaryGreaterOrEqual:
		return sql.IndexScanOpGte, left, right, true
	case framework.Operator_BinaryNotEqual:
		return sql.IndexScanOpNotEq, left, right, true
	}
	return 0, nil, nil, false
}

// DistanceMetric implements the vector.OrderableDistance interface.
func (b *BinaryOperator) DistanceMetric() sql.DistanceType {
	compiledFunc, ok := b.compiledFunc.(*framework.CompiledFunction)
	if !ok {
		return nil
	}
	extensionName, symbol, ok := compiledFunc.ResolvedExtensionRoutine()
	if !ok {
		return nil
	}
	return extensions.GetDistanceType(extensionName, symbol)
}

// TargetAndQuery implements the vector.OrderableDistance interface.
func (b *BinaryOperator) TargetAndQuery() (sql.Expression, sql.Expression, bool) {
	if b.DistanceMetric() == nil {
		return nil, nil, false
	}
	left := unwrapIndexScanTarget(b.Left())
	right := unwrapIndexScanTarget(b.Right())
	if _, ok := left.(*expression.GetField); ok {
		if isRowIndependentQueryVector(right) {
			return left, right, true
		}
	} else if _, ok = right.(*expression.GetField); ok {
		if isRowIndependentQueryVector(left) {
			return right, left, true
		}
	}
	return nil, nil, false
}

// isRowIndependentQueryVector returns whether the expression can serve as the query vector of a vector index scan.
func isRowIndependentQueryVector(expr sql.Expression) bool {
	if !expr.Resolved() {
		return false
	}
	switch e := expr.(type) {
	case *expression.GetField, *plan.Subquery:
		return false
	case *expression.Literal:
		return e.Value() != nil
	}
	for _, child := range expr.Children() {
		if !isRowIndependentQueryVector(child) {
			return false
		}
	}
	return true
}

// unwrapIndexScanTarget removes a GMSCast wrapper from an expression when the cast wraps a GetField.
// The analyzer wraps GMS-typed columns (e.g. the to_commit/from_commit columns of the dolt_diff_*
// system tables) in a GMSCast so that they may be used by the Doltgres function framework, but index
// scan costing only recognizes bare GetFields as index targets.
func unwrapIndexScanTarget(expr sql.Expression) sql.Expression {
	if cast, ok := expr.(*GMSCast); ok {
		if gf, ok := cast.Child().(*expression.GetField); ok {
			return gf
		}
	}
	return expr
}

// getUserDefinedOperator returns the function for the user-defined operator matching the given operands. Returns nil if
// an operator is not found.
func getUserDefinedOperator(ctx *sql.Context, operator framework.Operator, left sql.Expression, right sql.Expression) (framework.Function, error) {
	leftType, ok := left.Type(ctx).(*pgtypes.DoltgresType)
	if !ok {
		if leftType, _ = pgtypes.FromGmsTypeToDoltgresType(left.Type(ctx)); leftType == nil {
			return nil, nil
		}
	}
	rightType, ok := right.Type(ctx).(*pgtypes.DoltgresType)
	if !ok {
		if rightType, _ = pgtypes.FromGmsTypeToDoltgresType(right.Type(ctx)); rightType == nil {
			return nil, nil
		}
	}
	operatorCollection, err := core.GetOperatorsCollectionFromContext(ctx, "")
	if err != nil {
		return nil, err
	}
	searchPath, err := core.SearchPath(ctx)
	if err != nil {
		return nil, err
	}
	op, ok, err := operatorCollection.ResolveOperator(ctx, searchPath, operator.String(), leftType.ID, rightType.ID)
	if err != nil || !ok {
		return nil, err
	}
	userFunc, err := framework.GetUserFunction(ctx, op.Function.SchemaName(), op.Function.FunctionName(), left, right)
	if err != nil {
		return nil, err
	}
	if userFunc == nil {
		return nil, sql.ErrFunctionNotFound.New(op.Function.FunctionName())
	}
	return userFunc, nil
}
