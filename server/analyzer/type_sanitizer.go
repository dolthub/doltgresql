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
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/analyzer"
	"github.com/dolthub/go-mysql-server/sql/expression"
	"github.com/dolthub/go-mysql-server/sql/plan"
	"github.com/dolthub/go-mysql-server/sql/transform"
	"github.com/dolthub/go-mysql-server/sql/types"
	"github.com/dolthub/vitess/go/vt/proto/query"
	"github.com/shopspring/decimal"

	pgexprs "github.com/dolthub/doltgresql/server/expression"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtransform "github.com/dolthub/doltgresql/server/transform"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// TypeSanitizer converts all GMS types into Doltgres types. Some places, such as parameter binding, will always default
// to GMS types, so by taking care of all conversions here, we can ensure that Doltgres only needs to worry about its
// own types.
func TypeSanitizer(ctx *sql.Context, a *analyzer.Analyzer, node sql.Node, scope *plan.Scope, selector analyzer.RuleSelector, qFlags *sql.QueryFlags) (sql.Node, transform.TreeIdentity, error) {
	return pgtransform.NodeExprsWithNodeWithOpaque(node, func(n sql.Node, expr sql.Expression) (sql.Expression, transform.TreeIdentity, error) {
		// This can be updated if we find more expressions that return GMS types.
		// These should eventually be replaced with Doltgres-equivalents over time, rendering this function unnecessary.
		switch expr := expr.(type) {
		case *expression.GetField:
			switch n := n.(type) {
			case *plan.Project, *plan.Filter, *plan.GroupBy:
				child := n.Children()[0]
				// Some dolt_ tables do not have doltgres types for their columns, so we convert them here
				if rt, ok := child.(*plan.ResolvedTable); ok && strings.HasPrefix(rt.Name(), "dolt_") {
					// This is a projection on a table, so we can safely convert the type
					if _, ok := expr.Type().(*pgtypes.DoltgresType); !ok {
						return pgexprs.NewGMSCast(expr), transform.NewTree, nil
					}
				}
			}
			return expr, transform.SameTree, nil
		case *expression.Literal:
			return typeSanitizerLiterals(ctx, expr)
		case *expression.Not, *expression.And, *expression.Or, *expression.Like:
			return pgexprs.NewGMSCast(expr), transform.NewTree, nil
		case sql.FunctionExpression:
			// Compiled functions are Doltgres functions. We're only concerned with GMS functions.
			if _, ok := expr.(framework.Function); !ok {
				// Some aggregation functions cannot be wrapped due to expectations in the analyzer, so we exclude them here.
				switch expr.FunctionName() {
				case "Count", "CountDistinct", "group_concat", "JSONObjectAgg", "Sum":
				default:
					// Some GMS functions wrap Doltgres parameters, so we'll only handle those that return GMS types
					if _, ok := expr.Type().(*pgtypes.DoltgresType); !ok {
						return pgexprs.NewGMSCast(expr), transform.NewTree, nil
					}
				}
			}
		case *sql.ColumnDefaultValue:
			// Due to how interfaces work, we sometimes pass (*ColumnDefaultValue)(nil), so we have to check for it
			if expr != nil && expr.Expr != nil {
				defaultExpr := expr.Expr
				if _, ok := defaultExpr.Type().(*pgtypes.DoltgresType); !ok {
					defaultExpr = pgexprs.NewGMSCast(defaultExpr)
				}
				defaultExprType := defaultExpr.Type().(*pgtypes.DoltgresType)
				outType, ok := expr.OutType.(*pgtypes.DoltgresType)
				if !ok {
					return nil, transform.NewTree, errors.Errorf("default values must have a non-GMS OutType: `%s`", expr.OutType.String())
				}
				if !outType.Equals(defaultExprType) {
					defaultExpr = pgexprs.NewAssignmentCast(defaultExpr, defaultExprType, outType)
				}
				newDefault, err := sql.NewColumnDefaultValue(defaultExpr, outType, expr.Literal, expr.Parenthesized, expr.ReturnNil)
				return newDefault, transform.NewTree, err
			}
		}
		return expr, transform.SameTree, nil
	})
}

// typeSanitizerLiterals handles literal expressions for TypeSanitizer.
func typeSanitizerLiterals(ctx context.Context, gmsLiteral *expression.Literal) (sql.Expression, transform.TreeIdentity, error) {
	// GMS may resolve Doltgres literals and then stick them in GMS literals, so we have to account for that here
	if doltgresType, ok := gmsLiteral.Type().(*pgtypes.DoltgresType); ok {
		return pgexprs.NewUnsafeLiteral(gmsLiteral.Value(), doltgresType), transform.NewTree, nil
	}
	switch gmsLiteral.Type().Type() {
	case query.Type_INT8, query.Type_INT16, query.Type_YEAR:
		newVal, _, err := types.Int16.Convert(ctx, gmsLiteral.Value())
		if err != nil {
			return nil, transform.NewTree, err
		}
		if newVal == nil {
			return pgexprs.NewNullLiteral(), transform.NewTree, nil
		}
		return pgexprs.NewRawLiteralInt16(newVal.(int16)), transform.NewTree, nil
	case query.Type_INT24, query.Type_INT32:
		newVal, _, err := types.Int32.Convert(ctx, gmsLiteral.Value())
		if err != nil {
			return nil, transform.NewTree, err
		}
		if newVal == nil {
			return pgexprs.NewNullLiteral(), transform.NewTree, nil
		}
		return pgexprs.NewRawLiteralInt32(newVal.(int32)), transform.NewTree, nil
	case query.Type_INT64, query.Type_ENUM:
		newVal, _, err := types.Int64.Convert(ctx, gmsLiteral.Value())
		if err != nil {
			return nil, transform.NewTree, err
		}
		if newVal == nil {
			return pgexprs.NewNullLiteral(), transform.NewTree, nil
		}
		return pgexprs.NewRawLiteralInt64(newVal.(int64)), transform.NewTree, nil
	case query.Type_UINT8, query.Type_UINT16, query.Type_UINT24, query.Type_UINT32:
		newVal, _, err := types.Uint32.Convert(ctx, gmsLiteral.Value())
		if err != nil {
			return nil, transform.NewTree, err
		}
		if newVal == nil {
			return pgexprs.NewNullLiteral(), transform.NewTree, nil
		}
		return pgexprs.NewRawLiteralInt64(int64(newVal.(uint32))), transform.NewTree, nil
	case query.Type_UINT64, query.Type_SET:
		newVal, _, err := types.Uint64.Convert(ctx, gmsLiteral.Value())
		if err != nil {
			return nil, transform.NewTree, err
		}
		if newVal == nil {
			return pgexprs.NewNullLiteral(), transform.NewTree, nil
		}
		newLiteral, err := pgexprs.NewNumericLiteral(strconv.FormatUint(newVal.(uint64), 10))
		return newLiteral, transform.NewTree, err
	case query.Type_FLOAT32:
		newVal, _, err := types.Float32.Convert(ctx, gmsLiteral.Value())
		if err != nil {
			return nil, transform.NewTree, err
		}
		if newVal == nil {
			return pgexprs.NewNullLiteral(), transform.NewTree, nil
		}
		return pgexprs.NewRawLiteralFloat32(newVal.(float32)), transform.NewTree, nil
	case query.Type_FLOAT64:
		newVal, _, err := types.Float64.Convert(ctx, gmsLiteral.Value())
		if err != nil {
			return nil, transform.NewTree, err
		}
		if newVal == nil {
			return pgexprs.NewNullLiteral(), transform.NewTree, nil
		}
		return pgexprs.NewRawLiteralFloat64(newVal.(float64)), transform.NewTree, nil
	case query.Type_DECIMAL:
		dec, ok := gmsLiteral.Value().(decimal.Decimal)
		if !ok {
			return nil, transform.NewTree, errors.Errorf("SANITIZER: expected decimal type: %T", gmsLiteral.Value())
		}
		return pgexprs.NewRawLiteralNumeric(dec), transform.NewTree, nil
	case query.Type_DATE, query.Type_DATETIME, query.Type_TIMESTAMP:
		newVal, _, err := types.Datetime.Convert(ctx, gmsLiteral.Value())
		if err != nil {
			return nil, transform.NewTree, err
		}
		if newVal == nil {
			return pgexprs.NewNullLiteral(), transform.NewTree, nil
		}
		return pgexprs.NewRawLiteralTimestamp(newVal.(time.Time)), transform.NewTree, nil
	case query.Type_CHAR, query.Type_VARCHAR, query.Type_TEXT:
		str, ok := gmsLiteral.Value().(string)
		if !ok {
			return nil, transform.NewTree, errors.Errorf("SANITIZER: expected string type: %T", gmsLiteral.Value())
		}
		return pgexprs.NewUnknownLiteral(str), transform.NewTree, nil
	case query.Type_BINARY, query.Type_VARBINARY, query.Type_BLOB:
		newVal := gmsLiteral.Value()
		if newVal == nil {
			return pgexprs.NewNullLiteral(), transform.NewTree, nil
		} else if str, ok := newVal.(string); ok {
			return pgexprs.NewUnknownLiteral(str), transform.NewTree, nil
		} else if b, ok := newVal.([]byte); ok {
			return pgexprs.NewUnknownLiteral(string(b)), transform.NewTree, nil
		}
		return nil, transform.NewTree, errors.Errorf("SANITIZER: invalid binary type: %T", gmsLiteral.Value())
	case query.Type_JSON:
		newVal := gmsLiteral.Value()
		if newVal == nil {
			return pgexprs.NewNullLiteral(), transform.NewTree, nil
		}
		str, ok := newVal.(string)
		if !ok {
			return nil, transform.NewTree, errors.Errorf("SANITIZER: expected string type: %T", gmsLiteral.Value())
		}
		return pgexprs.NewUnknownLiteral(str), transform.NewTree, nil
	case query.Type_NULL_TYPE:
		return pgexprs.NewNullLiteral(), transform.NewTree, nil
	default:
		return nil, transform.NewTree, errors.Errorf("SANITIZER: encountered a GMS type that cannot be handled: %s", gmsLiteral.Type().String())
	}
}
