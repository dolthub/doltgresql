// Copyright 2023 Dolthub, Inc.
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

package ast

import (
	"context"
	"fmt"
	"go/constant"
	"strings"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql/expression"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"
	"github.com/shopspring/decimal"
	"github.com/sirupsen/logrus"

	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
	"github.com/dolthub/doltgresql/postgres/parser/timeofday"
	"github.com/dolthub/doltgresql/postgres/parser/types"
	pgexprs "github.com/dolthub/doltgresql/server/expression"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgnodes "github.com/dolthub/doltgresql/server/node"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// nodeExprs handles tree.Exprs nodes.
func nodeExprs(ctx *Context, node tree.Exprs) (vitess.Exprs, error) {
	if len(node) == 0 {
		return nil, nil
	}
	exprs := make(vitess.Exprs, len(node))
	for i := range node {
		var err error
		if exprs[i], err = nodeExpr(ctx, node[i]); err != nil {
			return nil, err
		}
	}
	return exprs, nil
}

// nodeCompositeDatum handles tree.CompositeDatum nodes.
func nodeCompositeDatum(ctx *Context, node tree.CompositeDatum) (vitess.Expr, error) {
	return nodeExpr(ctx, node)
}

// nodeConstant handles tree.Constant nodes.
func nodeConstant(ctx *Context, node tree.Constant) (vitess.Expr, error) {
	return nodeExpr(ctx, node)
}

// nodeDatum handles tree.Datum nodes.
func nodeDatum(ctx *Context, node tree.Datum) (vitess.Expr, error) {
	return nodeExpr(ctx, node)
}

// nodeSubqueryExpr handles tree.SubqueryExpr nodes.
func nodeSubqueryExpr(ctx *Context, node tree.SubqueryExpr) (vitess.Expr, error) {
	return nodeExpr(ctx, node)
}

// nodeTypedExpr handles tree.TypedExpr nodes.
func nodeTypedExpr(ctx *Context, node tree.TypedExpr) (vitess.Expr, error) {
	return nodeExpr(ctx, node)
}

// nodeVariableExpr handles tree.VariableExpr nodes.
func nodeVariableExpr(ctx *Context, node tree.VariableExpr) (vitess.Expr, error) {
	return nodeExpr(ctx, node)
}

// nodeVarName handles tree.VarName nodes.
func nodeVarName(ctx *Context, node tree.VarName) (vitess.Expr, error) {
	return nodeExpr(ctx, node)
}

// nodeExpr handles tree.Expr nodes.
func nodeExpr(ctx *Context, node tree.Expr) (vitess.Expr, error) {
	switch node := node.(type) {
	case *tree.AllColumnsSelector:
		return nil, errors.Errorf("table.* syntax in this context is not yet supported")
	case *tree.AndExpr:
		left, err := nodeExpr(ctx, node.Left)
		if err != nil {
			return nil, err
		}
		right, err := nodeExpr(ctx, node.Right)
		if err != nil {
			return nil, err
		}
		return &vitess.AndExpr{
			Left:  left,
			Right: right,
		}, nil
	case *tree.AnnotateTypeExpr:
		return nil, errors.Errorf("ANNOTATE_TYPE is not yet supported")
	case *tree.Array:
		unresolvedChildren := make([]vitess.Expr, len(node.Exprs))
		var coercedType *pgtypes.DoltgresType
		if node.HasResolvedType() {
			_, resolvedType, err := nodeResolvableTypeReference(ctx, node.ResolvedType())
			if err != nil {
				return nil, err
			}
			if resolvedType.IsArrayType() {
				coercedType = resolvedType
			} else {
				return nil, errors.Errorf("array has invalid resolved type")
			}
		}
		for i, arrayExpr := range node.Exprs {
			var err error
			unresolvedChildren[i], err = nodeExpr(ctx, arrayExpr)
			if err != nil {
				return nil, err
			}
		}
		arrayExpr, err := pgexprs.NewArray(coercedType)
		if err != nil {
			return nil, err
		}
		return vitess.InjectedExpr{
			Expression: arrayExpr,
			Children:   unresolvedChildren,
		}, nil
	case *tree.ArrayFlatten:
		subquery, err := nodeExpr(ctx, node.Subquery)
		if err != nil {
			return nil, err
		}
		return vitess.InjectedExpr{
			Expression: pgexprs.ArrayFlatten{},
			Children:   vitess.Exprs{subquery},
		}, nil
	case *tree.BinaryExpr:
		// We will eventually support operators in other schemas, but for now we only can handle built-ins
		if len(node.Schema) > 0 && node.Schema != "pg_catalog" {
			return nil, errors.Errorf("schema %q not allowed in OPERATOR syntax", node.Schema)
		}

		left, err := nodeExpr(ctx, node.Left)
		if err != nil {
			return nil, err
		}
		right, err := nodeExpr(ctx, node.Right)
		if err != nil {
			return nil, err
		}
		var operator framework.Operator
		switch node.Operator {
		case tree.Bitand:
			operator = framework.Operator_BinaryBitAnd
		case tree.Bitor:
			operator = framework.Operator_BinaryBitOr
		case tree.Bitxor:
			operator = framework.Operator_BinaryBitXor
		case tree.Plus:
			operator = framework.Operator_BinaryPlus
		case tree.Minus:
			operator = framework.Operator_BinaryMinus
		case tree.Mult:
			operator = framework.Operator_BinaryMultiply
		case tree.Div:
			operator = framework.Operator_BinaryDivide
		case tree.FloorDiv:
			// TODO: replace with floor divide function
			return nil, errors.Errorf("the floor divide operator is not yet supported")
		case tree.Mod:
			operator = framework.Operator_BinaryMod
		case tree.Pow:
			// TODO: replace with power function
			return nil, errors.Errorf("the power operator is not yet supported")
		case tree.Concat:
			operator = framework.Operator_BinaryConcatenate
		case tree.LShift:
			operator = framework.Operator_BinaryShiftLeft
		case tree.RShift:
			operator = framework.Operator_BinaryShiftRight
		case tree.JSONFetchVal:
			operator = framework.Operator_BinaryJSONExtractJson
		case tree.JSONFetchText:
			operator = framework.Operator_BinaryJSONExtractText
		case tree.JSONFetchValPath:
			operator = framework.Operator_BinaryJSONExtractPathJson
		case tree.JSONFetchTextPath:
			operator = framework.Operator_BinaryJSONExtractPathText
		default:
			return nil, errors.Errorf("the binary operator used is not yet supported")
		}
		return vitess.InjectedExpr{
			Expression: pgexprs.NewBinaryOperator(operator),
			Children:   vitess.Exprs{left, right},
		}, nil
	case *tree.CaseExpr:
		expr, err := nodeExpr(ctx, node.Expr)
		if err != nil {
			return nil, err
		}
		whens := make([]*vitess.When, len(node.Whens))
		for i := range node.Whens {
			val, err := nodeExpr(ctx, node.Whens[i].Val)
			if err != nil {
				return nil, err
			}
			cond, err := nodeExpr(ctx, node.Whens[i].Cond)
			if err != nil {
				return nil, err
			}
			whens[i] = &vitess.When{
				Val:  val,
				Cond: cond,
			}
		}
		else_, err := nodeExpr(ctx, node.Else)
		if err != nil {
			return nil, err
		}
		return &vitess.CaseExpr{
			Expr:  expr,
			Whens: whens,
			Else:  else_,
		}, nil
	case *tree.CastExpr:
		expr, err := nodeExpr(ctx, node.Expr)
		if err != nil {
			return nil, err
		}

		switch node.SyntaxMode {
		case tree.CastExplicit, tree.CastShort:
			// Both of these are acceptable
		case tree.CastPrepend:
			// used for typed literals
			strVal, isStrVal := node.Expr.(*tree.StrVal)
			t, isT := node.Type.(*types.T)
			if isStrVal && isT {
				typedExpr, err := strVal.ResolveAsType(context.TODO(), nil, t)
				if err != nil {
					return nil, errors.Errorf("cannot resolve '%s' as type %s", strVal.String(), t.Name())
				}
				expr, err = nodeExpr(ctx, typedExpr)
				if err != nil {
					return nil, err
				}
			}
		default:
			return nil, errors.Errorf("unknown cast syntax")
		}

		convertType, resolvedType, err := nodeResolvableTypeReference(ctx, node.Type)
		if err != nil {
			return nil, err
		}

		// If we have the resolved type, then we've got a Doltgres type instead of a GMS type
		if !resolvedType.IsEmptyType() {
			cast, err := pgexprs.NewExplicitCastInjectable(resolvedType)
			if err != nil {
				return nil, err
			}
			return vitess.InjectedExpr{
				Expression: cast,
				Children:   vitess.Exprs{expr},
			}, nil
		} else {
			convertType, err = translateConvertType(convertType)
			if err != nil {
				return nil, err
			}
			return &vitess.ConvertExpr{
				Name: "CAST",
				Expr: expr,
				Type: convertType,
			}, nil
		}

	case *tree.CoalesceExpr:
		exprs, err := nodeExprsToSelectExprs(ctx, node.Exprs)
		if err != nil {
			return nil, err
		}

		return &vitess.FuncExpr{
			Name:  vitess.NewColIdent("COALESCE"),
			Exprs: exprs,
		}, nil
	case *tree.CollateExpr:
		logrus.Warnf("collate is not yet supported, ignoring")
		return nodeExpr(ctx, node.Expr)
	case *tree.ColumnAccessExpr:
		return nil, errors.Errorf("(E).x is not yet supported")
	case *tree.ColumnItem:
		var tableName vitess.TableName
		if node.TableName != nil {
			if node.TableName.NumParts > 2 {
				return nil, errors.Errorf("referencing items outside the database is not yet supported")
			}
			tableName.Name = vitess.NewTableIdent(node.TableName.Parts[0])
			tableName.SchemaQualifier = vitess.NewTableIdent(node.TableName.Parts[1])
		}
		return &vitess.ColName{
			Name:      vitess.NewColIdent(string(node.ColumnName)),
			Qualifier: tableName,
		}, nil
	case *tree.CommentOnColumn:
		return nil, errors.Errorf("comment on column is not yet supported")
	case *tree.ComparisonExpr:
		// We will eventually support operators in other schemas, but for now we only can handle built-ins
		if len(node.Schema) > 0 && node.Schema != "pg_catalog" {
			return nil, errors.Errorf("schema %q not allowed in OPERATOR syntax", node.Schema)
		}

		left, err := nodeExpr(ctx, node.Left)
		if err != nil {
			return nil, err
		}
		right, err := nodeExpr(ctx, node.Right)
		if err != nil {
			return nil, err
		}
		var operator string
		switch node.Operator {
		case tree.EQ:
			return vitess.InjectedExpr{
				Expression: pgexprs.NewBinaryOperator(framework.Operator_BinaryEqual),
				Children:   vitess.Exprs{left, right},
			}, nil
		case tree.LT:
			return vitess.InjectedExpr{
				Expression: pgexprs.NewBinaryOperator(framework.Operator_BinaryLessThan),
				Children:   vitess.Exprs{left, right},
			}, nil
		case tree.GT:
			return vitess.InjectedExpr{
				Expression: pgexprs.NewBinaryOperator(framework.Operator_BinaryGreaterThan),
				Children:   vitess.Exprs{left, right},
			}, nil
		case tree.LE:
			return vitess.InjectedExpr{
				Expression: pgexprs.NewBinaryOperator(framework.Operator_BinaryLessOrEqual),
				Children:   vitess.Exprs{left, right},
			}, nil
		case tree.GE:
			return vitess.InjectedExpr{
				Expression: pgexprs.NewBinaryOperator(framework.Operator_BinaryGreaterOrEqual),
				Children:   vitess.Exprs{left, right},
			}, nil
		case tree.NE:
			return vitess.InjectedExpr{
				Expression: pgexprs.NewBinaryOperator(framework.Operator_BinaryNotEqual),
				Children:   vitess.Exprs{left, right},
			}, nil
		case tree.In, tree.NotIn:
			var innerExpression vitess.InjectedExpr
			switch right := right.(type) {
			case vitess.ValTuple:
				innerExpression = vitess.InjectedExpr{
					Expression: pgexprs.NewInTuple(),
					Children:   vitess.Exprs{left, right},
				}
			case *vitess.Subquery:
				innerExpression = vitess.InjectedExpr{
					Expression: pgexprs.NewInSubquery(),
					Children:   vitess.Exprs{left, right},
				}
			case vitess.InjectedExpr:
				if _, ok := right.Expression.(*pgexprs.RecordExpr); ok {
					innerExpression = vitess.InjectedExpr{
						Expression: pgexprs.NewInTuple(),
						Children:   vitess.Exprs{left, vitess.ValTuple(right.Children)},
					}
				}
			}

			if innerExpression.Expression == nil {
				return nil, errors.Errorf("right side of IN expression is not a tuple or subquery, got %T", right)
			}

			switch node.Operator {
			case tree.In:
				return innerExpression, nil
			case tree.NotIn:
				return vitess.InjectedExpr{
					Expression: pgexprs.NewNot(),
					Children:   vitess.Exprs{innerExpression},
				}, nil
			default:
				return nil, errors.Errorf("unknown comparison operator used")
			}
		case tree.Like:
			operator = vitess.LikeStr
		case tree.NotLike:
			operator = vitess.NotLikeStr
		case tree.ILike:
			return nil, errors.Errorf("ILIKE is not yet supported")
		case tree.NotILike:
			return nil, errors.Errorf("ILIKE is not yet supported")
		case tree.SimilarTo:
			return nil, errors.Errorf("similar to is not yet supported")
		case tree.NotSimilarTo:
			return nil, errors.Errorf("similar to is not yet supported")
		case tree.RegMatch:
			operator = vitess.RegexpStr
		case tree.NotRegMatch:
			operator = vitess.NotRegexpStr
		case tree.RegIMatch:
			return nil, errors.Errorf("~* is not yet supported")
		case tree.NotRegIMatch:
			return nil, errors.Errorf("~* is not yet supported")
		case tree.TextSearchMatch:
			return nil, errors.Errorf("@@ is not yet supported")
		case tree.IsDistinctFrom:
			return nil, errors.Errorf("IS DISTINCT FROM is not yet supported")
		case tree.IsNotDistinctFrom:
			return nil, errors.Errorf("IS NOT DISTINCT FROM is not yet supported")
		case tree.Contains:
			return vitess.InjectedExpr{
				Expression: pgexprs.NewBinaryOperator(framework.Operator_BinaryJSONContainsRight),
				Children:   vitess.Exprs{left, right},
			}, nil
		case tree.ContainedBy:
			return vitess.InjectedExpr{
				Expression: pgexprs.NewBinaryOperator(framework.Operator_BinaryJSONContainsLeft),
				Children:   vitess.Exprs{left, right},
			}, nil
		case tree.JSONExists:
			return vitess.InjectedExpr{
				Expression: pgexprs.NewBinaryOperator(framework.Operator_BinaryJSONTopLevel),
				Children:   vitess.Exprs{left, right},
			}, nil
		case tree.JSONSomeExists:
			return vitess.InjectedExpr{
				Expression: pgexprs.NewBinaryOperator(framework.Operator_BinaryJSONTopLevelAny),
				Children:   vitess.Exprs{left, right},
			}, nil
		case tree.JSONAllExists:
			return vitess.InjectedExpr{
				Expression: pgexprs.NewBinaryOperator(framework.Operator_BinaryJSONTopLevelAll),
				Children:   vitess.Exprs{left, right},
			}, nil
		case tree.Overlaps:
			return nil, errors.Errorf("&& is not yet supported")
		case tree.Any:
			return vitess.InjectedExpr{
				Expression: pgexprs.NewAnyExpr(node.SubOperator.String()),
				Children:   vitess.Exprs{left, right},
			}, nil
		case tree.Some:
			return vitess.InjectedExpr{
				Expression: pgexprs.NewSomeExpr(node.SubOperator.String()),
				Children:   vitess.Exprs{left, right},
			}, nil
		case tree.All:
			return nil, errors.Errorf("ALL is not yet supported")
		default:
			return nil, errors.Errorf("unknown comparison operator used")
		}
		return &vitess.ComparisonExpr{
			Operator: operator,
			Left:     left,
			Right:    right,
			Escape:   nil, // TODO: is '\' the default in Postgres as well?
		}, nil
	case *tree.DArray:
		return nil, errors.Errorf("the statement is not yet supported")
	case *tree.DBitArray:
		return nil, errors.Errorf("the statement is not yet supported")
	case *tree.DBool:
		return vitess.InjectedExpr{
			Expression: pgexprs.NewRawLiteralBool(bool(*node)),
		}, nil
	case *tree.DBox2D:
		return nil, errors.Errorf("the statement is not yet supported")
	case *tree.DBytes:
		return nil, errors.Errorf("the statement is not yet supported")
	case *tree.DCollatedString:
		return nil, errors.Errorf("the statement is not yet supported")
	case *tree.DDate:
		t, err := node.Date.ToTime()
		if err != nil {
			return nil, err
		}
		return vitess.InjectedExpr{
			Expression: pgexprs.NewRawLiteralDate(t),
		}, nil
	case *tree.DDecimal:
		// TODO: should we use apd.Decimal for Numeric type values?
		// |Coeff| is always positive, so need to |Negative| to negate the big.Int
		bigInt := &node.Coeff
		if node.Negative {
			bigInt = bigInt.Neg(bigInt)
		}
		return vitess.InjectedExpr{
			Expression: pgexprs.NewRawLiteralNumeric(decimal.NewFromBigInt(bigInt, node.Exponent)),
		}, nil
	case *tree.DEnum:
		return nil, errors.Errorf("the statement is not yet supported")
	case *tree.DFloat:
		return vitess.InjectedExpr{
			Expression: pgexprs.NewRawLiteralFloat64(float64(*node)),
		}, nil
	case *tree.DGeography:
		return nil, errors.Errorf("the statement is not yet supported")
	case *tree.DGeometry:
		return nil, errors.Errorf("the statement is not yet supported")
	case *tree.DIPAddr:
		return nil, errors.Errorf("the statement is not yet supported")
	case *tree.DInt:
		return vitess.InjectedExpr{
			Expression: pgexprs.NewRawLiteralInt64(int64(*node)),
		}, nil
	case *tree.DInterval:
		cast, err := pgexprs.NewExplicitCastInjectable(pgtypes.Interval)
		if err != nil {
			return nil, err
		}
		expr := pgexprs.NewIntervalLiteral(node.Duration)
		return vitess.InjectedExpr{
			Expression: cast,
			Children:   vitess.Exprs{vitess.InjectedExpr{Expression: expr}},
		}, nil
	case *tree.DJSON:
		// JSON type is handled in string format
		return vitess.InjectedExpr{
			Expression: pgexprs.NewRawLiteralJSON(node.JSON.String()),
		}, nil
	case *tree.DOid:
		internalID := id.Cache().ToInternal(uint32(node.DInt))
		if !internalID.IsValid() {
			internalID = id.NewOID(uint32(node.DInt)).AsId()
		}
		return vitess.InjectedExpr{
			Expression: pgexprs.NewRawLiteralOid(internalID),
		}, nil
	case *tree.DOidWrapper:
		return nodeExpr(ctx, node.Wrapped)
	case *tree.DString:
		return vitess.InjectedExpr{
			Expression: pgexprs.NewUnknownLiteral(string(*node)),
		}, nil
	case *tree.DTime:
		return vitess.InjectedExpr{
			Expression: pgexprs.NewRawLiteralTime(timeofday.TimeOfDay(*node).ToTime()),
		}, nil
	case *tree.DTimeTZ:
		return vitess.InjectedExpr{
			Expression: pgexprs.NewRawLiteralTimeTZ(node.TimeTZ.ToTime()),
		}, nil
	case *tree.DTimestamp:
		return vitess.InjectedExpr{
			Expression: pgexprs.NewRawLiteralTimestamp(node.Time),
		}, nil
	case *tree.DTimestampTZ:
		return vitess.InjectedExpr{
			Expression: pgexprs.NewRawLiteralTimestampTZ(node.Time),
		}, nil
	case *tree.DTuple:
		return nil, errors.Errorf("the statement is not yet supported")
	case *tree.DUuid:
		return vitess.InjectedExpr{
			Expression: pgexprs.NewRawLiteralUuid(node.UUID),
		}, nil
	case tree.DefaultVal:
		// TODO: can we use this?
		defVal := &vitess.Default{ColName: ""}
		return defVal, nil
	case tree.DomainColumn:
		_, dataType, err := nodeResolvableTypeReference(ctx, node.Typ)
		if err != nil {
			return nil, err
		}
		return vitess.InjectedExpr{
			Expression: &pgnodes.DomainColumn{Typ: dataType},
		}, nil
	case *tree.FuncExpr:
		return nodeFuncExpr(ctx, node)
	case *tree.IfErrExpr:
		return nil, errors.Errorf("IFERROR is not yet supported")
	case *tree.IfExpr:
		cond, err := nodeExpr(ctx, node.Cond)
		if err != nil {
			return nil, err
		}
		trueVal, err := nodeExpr(ctx, node.True)
		if err != nil {
			return nil, err
		}
		falseVal, err := nodeExpr(ctx, node.Else)
		if err != nil {
			return nil, err
		}

		// TODO: this could be a postgres func, but postgres doesn't have an IF function, this is an extension from cockroach
		return &vitess.FuncExpr{
			Name: vitess.NewColIdent("IF"),
			Exprs: vitess.SelectExprs{
				&vitess.AliasedExpr{
					Expr: cond,
				},
				&vitess.AliasedExpr{
					Expr: trueVal,
				},
				&vitess.AliasedExpr{
					Expr: falseVal,
				},
			},
		}, nil
	case *tree.IndexedVar:
		// TODO: figure out if I can delete this
		return nil, errors.Errorf("this should probably be deleted (internal error, IndexedVar)")
	case *tree.IndirectionExpr:
		childExpr, err := nodeExpr(ctx, node.Expr)
		if err != nil {
			return nil, err
		}

		if len(node.Indirection) > 1 {
			return nil, errors.Errorf("multi dimensional array subscripts are not yet supported")
		} else if node.Indirection[0].Slice {
			return nil, errors.Errorf("slice subscripts are not yet supported")
		}

		indexExpr, err := nodeExpr(ctx, node.Indirection[0].Begin)
		if err != nil {
			return nil, err
		}

		return vitess.InjectedExpr{
			Expression: &pgexprs.Subscript{},
			Children:   vitess.Exprs{childExpr, indexExpr},
		}, nil
	case *tree.IsNotNullExpr:
		expr, err := nodeExpr(ctx, node.Expr)
		if err != nil {
			return nil, err
		}
		return &vitess.IsExpr{
			Operator: vitess.IsNotNullStr,
			Expr:     expr,
		}, nil
	case *tree.IsNullExpr:
		expr, err := nodeExpr(ctx, node.Expr)
		if err != nil {
			return nil, err
		}
		return &vitess.IsExpr{
			Operator: vitess.IsNullStr,
			Expr:     expr,
		}, nil
	case *tree.IsOfTypeExpr:
		return nil, errors.Errorf("IS OF is not yet supported")
	case *tree.NotExpr:
		expr, err := nodeExpr(ctx, node.Expr)
		if err != nil {
			return nil, err
		}
		return &vitess.NotExpr{
			Expr: expr,
		}, nil
	case *tree.NullIfExpr:
		expr1, err := nodeExprToSelectExpr(ctx, node.Expr1)
		if err != nil {
			return nil, err
		}

		expr2, err := nodeExprToSelectExpr(ctx, node.Expr2)
		if err != nil {
			return nil, err
		}

		return &vitess.FuncExpr{
			Name:  vitess.NewColIdent("NULLIF"),
			Exprs: vitess.SelectExprs{expr1, expr2},
		}, nil
	case tree.NullLiteral:
		return &vitess.NullVal{}, nil
	case *tree.NumVal:
		switch node.Kind() {
		case constant.Int:
			intLiteral, err := pgexprs.NewIntegerLiteral(node.FormattedString())
			return vitess.InjectedExpr{
				Expression: intLiteral,
			}, err
		case constant.Float:
			numericLiteral, err := pgexprs.NewNumericLiteral(node.FormattedString())
			return vitess.InjectedExpr{
				Expression: numericLiteral,
			}, err
		default:
			return nil, errors.Errorf("unknown number format")
		}
	case *tree.OrExpr:
		left, err := nodeExpr(ctx, node.Left)
		if err != nil {
			return nil, err
		}
		right, err := nodeExpr(ctx, node.Right)
		if err != nil {
			return nil, err
		}
		return &vitess.OrExpr{
			Left:  left,
			Right: right,
		}, nil
	case *tree.ParenExpr:
		expr, err := nodeExpr(ctx, node.Expr)
		if err != nil {
			return nil, err
		}
		return &vitess.ParenExpr{
			Expr: expr,
		}, nil
	case *tree.PartitionMaxVal:
		return nil, errors.Errorf("MAXVALUE is not yet supported")
	case *tree.PartitionMinVal:
		return nil, errors.Errorf("MINVALUE is not yet supported")
	case *tree.Placeholder:
		// TODO: deal with type annotation
		mysqlBindVarIdx := node.Idx + 1
		return vitess.NewValArg([]byte(fmt.Sprintf(":v%d", mysqlBindVarIdx))), nil
	case *tree.RangeCond:
		left, err := nodeExpr(ctx, node.Left)
		if err != nil {
			return nil, err
		}
		from, err := nodeExpr(ctx, node.From)
		if err != nil {
			return nil, err
		}
		to, err := nodeExpr(ctx, node.To)
		if err != nil {
			return nil, err
		}
		retExpr := vitess.Expr(&vitess.AndExpr{
			Left: vitess.InjectedExpr{
				Expression: pgexprs.NewBinaryOperator(framework.Operator_BinaryGreaterOrEqual),
				Children:   vitess.Exprs{left, from},
			},
			Right: vitess.InjectedExpr{
				Expression: pgexprs.NewBinaryOperator(framework.Operator_BinaryLessOrEqual),
				Children:   vitess.Exprs{left, to},
			},
		})
		if node.Symmetric {
			retExpr = &vitess.OrExpr{
				Left: retExpr,
				Right: &vitess.AndExpr{
					Left: vitess.InjectedExpr{
						Expression: pgexprs.NewBinaryOperator(framework.Operator_BinaryGreaterOrEqual),
						Children:   vitess.Exprs{left, to},
					},
					Right: vitess.InjectedExpr{
						Expression: pgexprs.NewBinaryOperator(framework.Operator_BinaryLessOrEqual),
						Children:   vitess.Exprs{left, from},
					},
				},
			}
		}
		if node.Not {
			retExpr = vitess.InjectedExpr{
				Expression: pgexprs.NewNot(),
				Children:   vitess.Exprs{retExpr},
			}
		}
		return retExpr, nil
	case *tree.StrVal:
		// TODO: determine what to do when node.WasScannedAsBytes() is true
		// For string literals, we mark the type as unknown, because Postgres has
		// more permissive implicit casting rules for literals than it does for strongly
		// typed values from a schema for example.
		unknownLiteral := pgexprs.NewUnknownLiteral(node.RawString())
		return vitess.InjectedExpr{
			Expression: unknownLiteral,
		}, nil
	case *tree.Subquery:
		return nodeSubqueryOrExists(ctx, node)
	case *tree.Tuple:
		if len(node.Labels) > 0 {
			return nil, errors.Errorf("tuple labels are not yet supported")
		}

		valTuple, err := nodeExprs(ctx, node.Exprs)
		if err != nil {
			return nil, err
		}

		return vitess.InjectedExpr{
			Expression: pgexprs.NewRecordExpr(),
			Children:   valTuple,
		}, nil
	case *tree.TupleStar:
		return nil, errors.Errorf("(E).* is not yet supported")
	case *tree.UnaryExpr:
		expr, err := nodeExpr(ctx, node.Expr)
		if err != nil {
			return nil, err
		}
		var operator framework.Operator
		switch node.Operator {
		// TODO: need to add UnaryPlus, it's like a no-op but Postgres actually implements it and it affects coercion
		case tree.UnaryMinus:
			operator = framework.Operator_UnaryMinus
		case tree.UnaryComplement:
			return &vitess.UnaryExpr{
				Operator: vitess.TildaStr,
				Expr:     expr,
			}, nil
		case tree.UnarySqrt:
			// TODO: replace with a function
			return nil, errors.Errorf("square root operator is not yet supported")
		case tree.UnaryCbrt:
			// TODO: replace with a function
			return nil, errors.Errorf("cube root operator is not yet supported")
		case tree.UnaryAbsolute:
			// TODO: replace with a function
			return nil, errors.Errorf("absolute operator is not yet supported")
		default:
			return nil, errors.Errorf("the unary operator used is not yet supported")
		}
		return vitess.InjectedExpr{
			Expression: pgexprs.NewUnaryOperator(operator),
			Children:   vitess.Exprs{expr},
		}, nil
	case tree.UnqualifiedStar:
		return nil, errors.Errorf("* syntax in this context is not yet supported")
	case *tree.UnresolvedName:
		if node.Star {
			return nil, errors.Errorf("* syntax in this context is not yet supported")
		}
		return unresolvedNameToColName(node)
	case nil:
		return nil, nil
	default:
		return nil, errors.Errorf("unknown expression: `%T`", node)
	}
}

// unresolvedNameToColName converts a tree.UnresolvedName to a vitess.ColName with the appropriate name qualifiers set.
func unresolvedNameToColName(name *tree.UnresolvedName) (*vitess.ColName, error) {
	var tableName vitess.TableName
	switch name.NumParts {
	case 4:
		tableName = vitess.TableName{
			Name:            vitess.NewTableIdent(name.Parts[1]),
			SchemaQualifier: vitess.NewTableIdent(name.Parts[2]),
			DbQualifier:     vitess.NewTableIdent(name.Parts[3]),
		}
	case 3:
		tableName = vitess.TableName{
			Name:            vitess.NewTableIdent(name.Parts[1]),
			SchemaQualifier: vitess.NewTableIdent(name.Parts[2]),
		}
	case 2:
		tableName = vitess.TableName{
			Name: vitess.NewTableIdent(name.Parts[1]),
		}
	case 1:
		// no table name
	default:
		return nil, errors.Errorf("invalid name: %s", name)
	}

	return &vitess.ColName{
		Name:      vitess.NewColIdent(name.Parts[0]),
		Qualifier: tableName,
	}, nil
}

// translateConvertType translates the *vitess.ConvertType expression given to a new one, substituting type names as
// appropriate. An error is returned if the type named cannot be supported.
func translateConvertType(convertType *vitess.ConvertType) (*vitess.ConvertType, error) {
	switch strings.ToLower(convertType.Type) {
	// passthrough types that need no conversion
	case expression.ConvertToBinary, expression.ConvertToChar, expression.ConvertToNChar, expression.ConvertToDate,
		expression.ConvertToDatetime, expression.ConvertToFloat, expression.ConvertToDouble, expression.ConvertToJSON,
		expression.ConvertToReal, expression.ConvertToSigned, expression.ConvertToTime, expression.ConvertToUnsigned:
		return convertType, nil
	case "text", "character varying", "varchar":
		return &vitess.ConvertType{
			Type: expression.ConvertToChar,
		}, nil
	case "integer", "bigint":
		return &vitess.ConvertType{
			Type: expression.ConvertToSigned,
		}, nil
	case "decimal", "numeric":
		return &vitess.ConvertType{
			Type: expression.ConvertToFloat,
		}, nil
	case "boolean":
		return &vitess.ConvertType{
			Type: expression.ConvertToSigned,
		}, nil
	case "timestamp", "timestamp with time zone", "timestamp without time zone":
		return &vitess.ConvertType{
			Type: expression.ConvertToDatetime,
		}, nil
	default:
		return nil, errors.Errorf("unknown convert type: `%T`", convertType.Type)
	}
}
