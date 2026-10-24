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

// ResolveSetOpTypes determines the common type for each column of a UNION, INTERSECT or EXCEPT, following
// PostgreSQL's type resolution rules. When the column types of the two sides differ, GMS converts both sides to
// char, so SELECT 5 UNION SELECT 5.0 returned two text rows instead of one numeric row. This replaces those
// conversions with implicit casts to the common type. Types that cannot be matched are still converted to text.
func ResolveSetOpTypes(ctx *sql.Context, a *analyzer.Analyzer, node sql.Node, scope *plan.Scope, selector analyzer.RuleSelector, qFlags *sql.QueryFlags) (sql.Node, transform.TreeIdentity, error) {
	node, same, err := transform.NodeWithOpaque(ctx, node, func(ctx *sql.Context, n sql.Node) (sql.Node, transform.TreeIdentity, error) {
		setOp, ok := n.(*plan.SetOp)
		if !ok {
			return n, transform.SameTree, nil
		}
		return transformSetOpNode(ctx, setOp)
	})
	if err != nil || same {
		return node, same, err
	}
	// Parent nodes that read the set operation's columns still have the char type, so we fix their GetField types
	// the same way as for VALUES.
	node, err = fixParentGetFieldTypes(ctx, node, nil)
	if err != nil {
		return nil, transform.SameTree, err
	}
	return node, transform.NewTree, nil
}

// transformSetOpNode replaces the char conversions that GMS adds to both sides of a set operation with implicit casts
// to the common type of each column.
func transformSetOpNode(ctx *sql.Context, setOp *plan.SetOp) (sql.Node, transform.TreeIdentity, error) {
	left, ok := setOp.Left().(*plan.Project)
	if !ok {
		return setOp, transform.SameTree, nil
	}
	right, ok := setOp.Right().(*plan.Project)
	if !ok {
		return setOp, transform.SameTree, nil
	}
	leftSch, rightSch := left.Child.Schema(ctx), right.Child.Schema(ctx)
	if len(left.Projections) != len(right.Projections) || len(left.Projections) != len(leftSch) || len(right.Projections) != len(rightSch) {
		return setOp, transform.SameTree, nil
	}

	var leftExprs, rightExprs []sql.Expression
	for i := range left.Projections {
		leftField, leftName, leftConverted := setOpField(left.Projections[i])
		rightField, rightName, rightConverted := setOpField(right.Projections[i])
		// A side that is already a char is not converted by GMS, but a set operation below it may have changed its
		// type, so we only need one of the sides to be converted
		if leftField == nil || rightField == nil || (!leftConverted && !rightConverted) {
			continue
		}
		leftType, ok := leftSch[i].Type.(*pgtypes.DoltgresType)
		if !ok {
			continue
		}
		rightType, ok := rightSch[i].Type.(*pgtypes.DoltgresType)
		if !ok {
			continue
		}
		explicit := false
		commonType, _, err := framework.FindCommonType(ctx, []*pgtypes.DoltgresType{leftType, rightType})
		if err != nil {
			// TODO: Postgres returns an error when the types cannot be matched, but GMS converts both sides to char,
			//  so for now we keep doing that with text
			commonType = pgtypes.Text
			explicit = true
		}
		// GMS only converts the column when the two types differ, and like Postgres, the common type only keeps a
		// type modifier that both sides share, so INT UNION NUMERIC(10,2) is NUMERIC rather than NUMERIC(10,2)
		commonType = commonType.WithAttTypMod(-1)
		if leftExprs == nil {
			leftExprs = append([]sql.Expression{}, left.Projections...)
			rightExprs = append([]sql.Expression{}, right.Projections...)
		}
		leftExprs[i] = expression.NewAlias(ctx, leftName, setOpCast(ctx, leftField, leftType, commonType, explicit))
		rightExprs[i] = expression.NewAlias(ctx, rightName, setOpCast(ctx, rightField, rightType, commonType, explicit))
	}
	if leftExprs == nil {
		return setOp, transform.SameTree, nil
	}

	newLeft, err := left.WithExpressions(ctx, leftExprs...)
	if err != nil {
		return nil, transform.NewTree, err
	}
	newRight, err := right.WithExpressions(ctx, rightExprs...)
	if err != nil {
		return nil, transform.NewTree, err
	}
	newSetOp, err := setOp.WithChildren(ctx, newLeft, newRight)
	if err != nil {
		return nil, transform.NewTree, err
	}
	return newSetOp, transform.NewTree, nil
}

// setOpField returns the field and the name of a set operation column, and whether GMS converted it to char. The field
// is nil if the expression is neither a field nor one of those conversions.
func setOpField(expr sql.Expression) (*expression.GetField, string, bool) {
	if field, ok := expr.(*expression.GetField); ok {
		return field, field.Name(), false
	}
	alias, ok := expr.(*expression.Alias)
	if !ok {
		return nil, "", false
	}
	child := alias.Child
	if field, ok := child.(*expression.GetField); ok {
		return field, alias.Name(), false
	}
	if gmsCast, ok := child.(*pgexprs.GMSCast); ok {
		child = gmsCast.Child()
	}
	convert, ok := child.(*expression.Convert)
	if !ok {
		return nil, "", false
	}
	field, ok := convert.Child.(*expression.GetField)
	if !ok {
		return nil, "", false
	}
	return field, alias.Name(), true
}

// setOpCast returns the field with its own type, cast to the common type if that is different.
func setOpCast(ctx *sql.Context, field *expression.GetField, fromType *pgtypes.DoltgresType, toType *pgtypes.DoltgresType, explicit bool) sql.Expression {
	var expr sql.Expression = expression.NewGetFieldWithTable(
		field.Index(), int(field.TableId()), fromType,
		field.Database(), field.Table(), field.Name(), field.IsNullable(ctx),
	)
	if fromType.Equals(toType) {
		return expr
	}
	if explicit {
		return pgexprs.NewExplicitCast(expr, toType)
	}
	return pgexprs.NewImplicitCast(expr, fromType, toType)
}
