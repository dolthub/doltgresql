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
	"fmt"
	"go/constant"
	"strings"

	"github.com/dolthub/go-mysql-server/sql/expression"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
)

// nodeExprs handles tree.Exprs nodes.
func nodeExprs(node tree.Exprs) (vitess.Exprs, error) {
	if len(node) == 0 {
		return nil, nil
	}
	exprs := make(vitess.Exprs, len(node))
	for i := range node {
		var err error
		if exprs[i], err = nodeExpr(node[i]); err != nil {
			return nil, err
		}
	}
	return exprs, nil
}

// nodeCompositeDatum handles tree.CompositeDatum nodes.
func nodeCompositeDatum(node tree.CompositeDatum) (vitess.Expr, error) {
	return nodeExpr(node)
}

// nodeConstant handles tree.Constant nodes.
func nodeConstant(node tree.Constant) (vitess.Expr, error) {
	return nodeExpr(node)
}

// nodeDatum handles tree.Datum nodes.
func nodeDatum(node tree.Datum) (vitess.Expr, error) {
	return nodeExpr(node)
}

// nodeSubqueryExpr handles tree.SubqueryExpr nodes.
func nodeSubqueryExpr(node tree.SubqueryExpr) (vitess.Expr, error) {
	return nodeExpr(node)
}

// nodeTypedExpr handles tree.TypedExpr nodes.
func nodeTypedExpr(node tree.TypedExpr) (vitess.Expr, error) {
	return nodeExpr(node)
}

// nodeVariableExpr handles tree.VariableExpr nodes.
func nodeVariableExpr(node tree.VariableExpr) (vitess.Expr, error) {
	return nodeExpr(node)
}

// nodeVarName handles tree.VarName nodes.
func nodeVarName(node tree.VarName) (vitess.Expr, error) {
	return nodeExpr(node)
}

// nodeExpr handles tree.Expr nodes.
func nodeExpr(node tree.Expr) (vitess.Expr, error) {
	switch node := node.(type) {
	case *tree.AllColumnsSelector:
		return nil, fmt.Errorf("table.* syntax is not yet supported in this context")
	case *tree.AndExpr:
		left, err := nodeExpr(node.Left)
		if err != nil {
			return nil, err
		}
		right, err := nodeExpr(node.Right)
		if err != nil {
			return nil, err
		}
		return &vitess.AndExpr{
			Left:  left,
			Right: right,
		}, nil
	case *tree.AnnotateTypeExpr:
		return nil, fmt.Errorf("ANNOTATE_TYPE is not yet supported")
	case *tree.Array:
		return nil, fmt.Errorf("arrays are not yet supported")
	case *tree.ArrayFlatten:
		return nil, fmt.Errorf("flattening arrays is not yet supported")
	case *tree.BinaryExpr:
		left, err := nodeExpr(node.Left)
		if err != nil {
			return nil, err
		}
		right, err := nodeExpr(node.Right)
		if err != nil {
			return nil, err
		}
		var operator string
		switch node.Operator {
		case tree.Bitand:
			operator = vitess.BitAndStr
		case tree.Bitor:
			operator = vitess.BitOrStr
		case tree.Bitxor:
			operator = vitess.BitXorStr
		case tree.Plus:
			operator = vitess.PlusStr
		case tree.Minus:
			operator = vitess.MinusStr
		case tree.Mult:
			operator = vitess.MultStr
		case tree.Div:
			operator = vitess.DivStr
		case tree.FloorDiv:
			//TODO: replace with floor divide function
			return nil, fmt.Errorf("the floor divide operator is not yet supported")
		case tree.Mod:
			operator = vitess.ModStr
		case tree.Pow:
			//TODO: replace with power function
			return nil, fmt.Errorf("the power operator is not yet supported")
		case tree.Concat:
			//TODO: replace with concat function
			return nil, fmt.Errorf("the concat operator is not yet supported")
		case tree.LShift:
			operator = vitess.ShiftLeftStr
		case tree.RShift:
			operator = vitess.ShiftRightStr
		case tree.JSONFetchVal:
			return nil, fmt.Errorf("JSON operators are not yet supported")
		case tree.JSONFetchText:
			return nil, fmt.Errorf("JSON operators are not yet supported")
		case tree.JSONFetchValPath:
			return nil, fmt.Errorf("JSON operators are not yet supported")
		case tree.JSONFetchTextPath:
			return nil, fmt.Errorf("JSON operators are not yet supported")
		default:
			return nil, fmt.Errorf("the binary operator used is not yet supported")
		}
		return &vitess.BinaryExpr{
			Operator: operator,
			Left:     left,
			Right:    right,
		}, nil
	case *tree.CaseExpr:
		expr, err := nodeExpr(node.Expr)
		if err != nil {
			return nil, err
		}
		whens := make([]*vitess.When, len(node.Whens))
		for i := range node.Whens {
			val, err := nodeExpr(node.Whens[i].Val)
			if err != nil {
				return nil, err
			}
			cond, err := nodeExpr(node.Whens[i].Cond)
			if err != nil {
				return nil, err
			}
			whens[i] = &vitess.When{
				Val:  val,
				Cond: cond,
			}
		}
		else_, err := nodeExpr(node.Else)
		if err != nil {
			return nil, err
		}
		return &vitess.CaseExpr{
			Expr:  expr,
			Whens: whens,
			Else:  else_,
		}, nil
	case *tree.CastExpr:
		expr, err := nodeExpr(node.Expr)
		if err != nil {
			return nil, err
		}

		switch node.SyntaxMode {
		case tree.CastExplicit, tree.CastShort:
			// Both of these are acceptable
		case tree.CastPrepend:
			return nil, fmt.Errorf("typed literals are not yet supported")
		default:
			return nil, fmt.Errorf("unknown cast syntax")
		}

		convertType, _, err := nodeResolvableTypeReference(node.Type)
		if err != nil {
			return nil, err
		}

		convertType, err = translateConvertType(convertType)
		if err != nil {
			return nil, err
		}

		return &vitess.ConvertExpr{
			Name: "CAST",
			Expr: expr,
			Type: convertType,
		}, nil
	case *tree.CoalesceExpr:
		exprs, err := nodeExprsToSelectExprs(node.Exprs)
		if err != nil {
			return nil, err
		}

		return &vitess.FuncExpr{
			Name:  vitess.NewColIdent("COALESCE"),
			Exprs: exprs,
		}, nil
	case *tree.CollateExpr:
		return nil, fmt.Errorf("collations are not yet supported")
	case *tree.ColumnAccessExpr:
		return nil, fmt.Errorf("(E).x is not yet supported")
	case *tree.ColumnItem:
		var tableName vitess.TableName
		if node.TableName != nil {
			if node.TableName.NumParts > 1 {
				return nil, fmt.Errorf("referencing items outside the schema or database is not yet supported")
			}
			tableName.Name = vitess.NewTableIdent(node.TableName.Parts[0])
		}
		return &vitess.ColName{
			Name:      vitess.NewColIdent(string(node.ColumnName)),
			Qualifier: tableName,
		}, nil
	case *tree.CommentOnColumn:
		return nil, fmt.Errorf("comment on column is not yet supported")
	case *tree.ComparisonExpr:
		left, err := nodeExpr(node.Left)
		if err != nil {
			return nil, err
		}
		right, err := nodeExpr(node.Right)
		if err != nil {
			return nil, err
		}
		var operator string
		switch node.Operator {
		case tree.EQ:
			operator = vitess.EqualStr
		case tree.LT:
			operator = vitess.LessThanStr
		case tree.GT:
			operator = vitess.GreaterThanStr
		case tree.LE:
			operator = vitess.LessEqualStr
		case tree.GE:
			operator = vitess.GreaterEqualStr
		case tree.NE:
			operator = vitess.NotEqualStr
		case tree.In:
			operator = vitess.InStr
		case tree.NotIn:
			operator = vitess.NotInStr
		case tree.Like:
			operator = vitess.LikeStr
		case tree.NotLike:
			operator = vitess.NotLikeStr
		case tree.ILike:
			return nil, fmt.Errorf("ILIKE is not yet supported")
		case tree.NotILike:
			return nil, fmt.Errorf("ILIKE is not yet supported")
		case tree.SimilarTo:
			return nil, fmt.Errorf("similar to is not yet supported")
		case tree.NotSimilarTo:
			return nil, fmt.Errorf("similar to is not yet supported")
		case tree.RegMatch:
			operator = vitess.RegexpStr
		case tree.NotRegMatch:
			operator = vitess.NotRegexpStr
		case tree.RegIMatch:
			return nil, fmt.Errorf("~* is not yet supported")
		case tree.NotRegIMatch:
			return nil, fmt.Errorf("~* is not yet supported")
		case tree.IsDistinctFrom:
			return nil, fmt.Errorf("IS DISTINCT FROM is not yet supported")
		case tree.IsNotDistinctFrom:
			return nil, fmt.Errorf("IS NOT DISTINCT FROM is not yet supported")
		case tree.Contains:
			return nil, fmt.Errorf("@> is not yet supported")
		case tree.ContainedBy:
			return nil, fmt.Errorf("<@ is not yet supported")
		case tree.JSONExists:
			return nil, fmt.Errorf("? is not yet supported")
		case tree.JSONSomeExists:
			return nil, fmt.Errorf("?| is not yet supported")
		case tree.JSONAllExists:
			return nil, fmt.Errorf("?& is not yet supported")
		case tree.Overlaps:
			return nil, fmt.Errorf("&& is not yet supported")
		case tree.Any:
			return nil, fmt.Errorf("ANY is not yet supported")
		case tree.Some:
			return nil, fmt.Errorf("SOME is not yet supported")
		case tree.All:
			return nil, fmt.Errorf("ALL is not yet supported")
		default:
			return nil, fmt.Errorf("unknown comparison operator used")
		}
		return &vitess.ComparisonExpr{
			Operator: operator,
			Left:     left,
			Right:    right,
			Escape:   nil, //TODO: is '\' the default in Postgres as well?
		}, nil
	case *tree.DArray:
		return nil, fmt.Errorf("the statement is not yet supported")
	case *tree.DBitArray:
		return nil, fmt.Errorf("the statement is not yet supported")
	case *tree.DBool:
		return nil, fmt.Errorf("the statement is not yet supported")
	case *tree.DBox2D:
		return nil, fmt.Errorf("the statement is not yet supported")
	case *tree.DBytes:
		return nil, fmt.Errorf("the statement is not yet supported")
	case *tree.DCollatedString:
		return nil, fmt.Errorf("the statement is not yet supported")
	case *tree.DDate:
		return nil, fmt.Errorf("the statement is not yet supported")
	case *tree.DDecimal:
		return nil, fmt.Errorf("the statement is not yet supported")
	case *tree.DEnum:
		return nil, fmt.Errorf("the statement is not yet supported")
	case *tree.DFloat:
		return nil, fmt.Errorf("the statement is not yet supported")
	case *tree.DGeography:
		return nil, fmt.Errorf("the statement is not yet supported")
	case *tree.DGeometry:
		return nil, fmt.Errorf("the statement is not yet supported")
	case *tree.DIPAddr:
		return nil, fmt.Errorf("the statement is not yet supported")
	case *tree.DInt:
		return nil, fmt.Errorf("the statement is not yet supported")
	case *tree.DInterval:
		return nil, fmt.Errorf("the statement is not yet supported")
	case *tree.DJSON:
		return nil, fmt.Errorf("the statement is not yet supported")
	case *tree.DOid:
		return nil, fmt.Errorf("the statement is not yet supported")
	case *tree.DOidWrapper:
		return nil, fmt.Errorf("the statement is not yet supported")
	case *tree.DString:
		return nil, fmt.Errorf("the statement is not yet supported")
	case *tree.DTime:
		return nil, fmt.Errorf("the statement is not yet supported")
	case *tree.DTimeTZ:
		return nil, fmt.Errorf("the statement is not yet supported")
	case *tree.DTimestamp:
		return nil, fmt.Errorf("the statement is not yet supported")
	case *tree.DTimestampTZ:
		return nil, fmt.Errorf("the statement is not yet supported")
	case *tree.DTuple:
		return nil, fmt.Errorf("the statement is not yet supported")
	case *tree.DUuid:
		return nil, fmt.Errorf("the statement is not yet supported")
	case *tree.DefaultVal:
		return nil, fmt.Errorf("default values are not yet supported")
	case *tree.FuncExpr:
		return nodeFuncExpr(node)
	case *tree.IfErrExpr:
		return nil, fmt.Errorf("IFERROR is not yet supported")
	case *tree.IfExpr:
		//TODO: probably should be the IF function
		return nil, fmt.Errorf("IF is not yet supported")
	case *tree.IndexedVar:
		//TODO: figure out if I can delete this
		return nil, fmt.Errorf("this should probably be deleted (internal error, IndexedVar)")
	case *tree.IndirectionExpr:
		return nil, fmt.Errorf("subscripts are not yet supported")
	case *tree.IsNotNullExpr:
		expr, err := nodeExpr(node.Expr)
		if err != nil {
			return nil, err
		}
		return &vitess.IsExpr{
			Operator: vitess.IsNotNullStr,
			Expr:     expr,
		}, nil
	case *tree.IsNullExpr:
		expr, err := nodeExpr(node.Expr)
		if err != nil {
			return nil, err
		}
		return &vitess.IsExpr{
			Operator: vitess.IsNullStr,
			Expr:     expr,
		}, nil
	case *tree.IsOfTypeExpr:
		return nil, fmt.Errorf("IS OF is not yet supported")
	case *tree.NotExpr:
		expr, err := nodeExpr(node.Expr)
		if err != nil {
			return nil, err
		}
		return &vitess.NotExpr{
			Expr: expr,
		}, nil
	case *tree.NullIfExpr:
		expr1, err := nodeExprToSelectExpr(node.Expr1)
		if err != nil {
			return nil, err
		}

		expr2, err := nodeExprToSelectExpr(node.Expr2)
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
			return vitess.NewIntVal([]byte(node.FormattedString())), nil
		case constant.Float:
			return vitess.NewFloatVal([]byte(node.FormattedString())), nil
		default:
			return nil, fmt.Errorf("unknown number format")
		}
	case *tree.OrExpr:
		left, err := nodeExpr(node.Left)
		if err != nil {
			return nil, err
		}
		right, err := nodeExpr(node.Right)
		if err != nil {
			return nil, err
		}
		return &vitess.OrExpr{
			Left:  left,
			Right: right,
		}, nil
	case *tree.ParenExpr:
		expr, err := nodeExpr(node.Expr)
		if err != nil {
			return nil, err
		}
		return &vitess.ParenExpr{
			Expr: expr,
		}, nil
	case *tree.PartitionMaxVal:
		return nil, fmt.Errorf("MAXVALUE is not yet supported")
	case *tree.PartitionMinVal:
		return nil, fmt.Errorf("MINVALUE is not yet supported")
	case *tree.Placeholder:
		// TODO: deal with type annotation
		mysqlBindVarIdx := node.Idx + 1
		return vitess.NewValArg([]byte(fmt.Sprintf(":v%d", mysqlBindVarIdx))), nil
	case *tree.RangeCond:
		operator := vitess.BetweenStr
		if node.Not {
			operator = vitess.NotBetweenStr
		}
		left, err := nodeExpr(node.Left)
		if err != nil {
			return nil, err
		}
		from, err := nodeExpr(node.From)
		if err != nil {
			return nil, err
		}
		to, err := nodeExpr(node.To)
		if err != nil {
			return nil, err
		}
		if node.Symmetric {
			return &vitess.OrExpr{
				Left: &vitess.RangeCond{
					Operator: operator,
					Left:     left,
					From:     from,
					To:       to,
				},
				Right: &vitess.RangeCond{
					Operator: operator,
					Left:     left,
					From:     to,
					To:       from,
				},
			}, nil
		} else {
			return &vitess.RangeCond{
				Operator: operator,
				Left:     left,
				From:     from,
				To:       to,
			}, nil
		}
	case *tree.StrVal:
		//TODO: determine what to do when node.WasScannedAsBytes() is true
		return &vitess.SQLVal{
			Type: vitess.StrVal,
			Val:  []byte(node.RawString()),
		}, nil
	case *tree.Subquery:
		return nodeSubquery(node)
	case *tree.Tuple:
		if len(node.Labels) > 0 {
			return nil, fmt.Errorf("tuple labels are not yet supported")
		}
		if node.Row {
			return nil, fmt.Errorf("ROW keyword for tuples not yet supported")
		}

		valTuple, err := nodeExprs(node.Exprs)
		if err != nil {
			return nil, err
		}
		return vitess.ValTuple(valTuple), nil
	case *tree.TupleStar:
		return nil, fmt.Errorf("(E).* is not yet supported")
	case *tree.UnaryExpr:
		expr, err := nodeExpr(node.Expr)
		if err != nil {
			return nil, err
		}
		switch node.Operator {
		case tree.UnaryMinus:
			return &vitess.UnaryExpr{
				Operator: vitess.UMinusStr,
				Expr:     expr,
			}, nil
		case tree.UnaryComplement:
			return &vitess.UnaryExpr{
				Operator: vitess.TildaStr,
				Expr:     expr,
			}, nil
		case tree.UnarySqrt:
			//TODO: replace with a function
			return nil, fmt.Errorf("square root operator is not yet supported")
		case tree.UnaryCbrt:
			//TODO: replace with a function
			return nil, fmt.Errorf("cube root operator is not yet supported")
		default:
			return nil, fmt.Errorf("the unary operator used is not yet supported")
		}
	case tree.UnqualifiedStar:
		return nil, fmt.Errorf("* syntax is not yet supported in this context")
	case *tree.UnresolvedName:
		if node.NumParts > 2 {
			return nil, fmt.Errorf("referencing items outside the schema or database is not yet supported")
		}
		if node.Star {
			return nil, fmt.Errorf("name resolution on this statement is not yet supported")
		}
		var tableName vitess.TableName
		if node.NumParts == 2 {
			tableName.Name = vitess.NewTableIdent(node.Parts[1])
		}
		return &vitess.ColName{
			Name:      vitess.NewColIdent(node.Parts[0]),
			Qualifier: tableName,
		}, nil
	case nil:
		return nil, nil
	default:
		return nil, fmt.Errorf("unknown expression: `%T`", node)
	}
}

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
		return nil, fmt.Errorf("unknown convert type: `%T`", convertType.Type)
	}
}
