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
	"strconv"
	"strings"
	"time"

	"github.com/cockroachdb/apd/v3"
	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/analyzer"
	"github.com/dolthub/go-mysql-server/sql/expression"
	"github.com/dolthub/go-mysql-server/sql/plan"
	"github.com/dolthub/go-mysql-server/sql/transform"
	"github.com/dolthub/go-mysql-server/sql/types"
	"github.com/dolthub/vitess/go/vt/proto/query"

	pgexprs "github.com/dolthub/doltgresql/server/expression"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtransform "github.com/dolthub/doltgresql/server/transform"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// TypeSanitizer converts all GMS types into Doltgres types. Some places, such as parameter binding, will always default
// to GMS types, so by taking care of all conversions here, we can ensure that Doltgres only needs to worry about its
// own types.
func TypeSanitizer(ctx *sql.Context, a *analyzer.Analyzer, node sql.Node, scope *plan.Scope, selector analyzer.RuleSelector, qFlags *sql.QueryFlags) (sql.Node, transform.TreeIdentity, error) {
	// TODO: this probably should not be opaque, we should let the analyzer dig into subqueries and analyze them when
	//  it chooses. Doing all type transformations upfront like this masks bugs where certain tyupe conversion errors
	//  only manifest in a subquery
	return pgtransform.NodeExprsWithNodeWithOpaque(ctx, node, func(ctx *sql.Context, n sql.Node, expr sql.Expression) (sql.Expression, transform.TreeIdentity, error) {
		// This can be updated if we find more expressions that return GMS types.
		// These should eventually be replaced with Doltgres-equivalents over time, rendering this function unnecessary.
		switch expr := expr.(type) {
		case *expression.GetField:
			switch n := n.(type) {
			case *plan.Project, *plan.Filter, *plan.GroupBy, *plan.Having:
				child := n.Children()[0]
				// Some dolt_ tables do not have doltgres types for their columns, so we convert them here
				if rt, ok := child.(*plan.ResolvedTable); ok && strings.HasPrefix(rt.Name(), "dolt_") {
					// This is a projection on a table, so we can safely convert the type
					if _, ok := expr.Type(ctx).(*pgtypes.DoltgresType); !ok {
						return pgexprs.NewGMSCast(expr), transform.NewTree, nil
					}
				}
				// Window function outputs carry GMS types in the Window schema; wrap with GMSCast
				// so the type is properly converted to a Doltgres type for the wire protocol.
				if isWindowOrWrapsWindow(child) {
					if _, ok := expr.Type(ctx).(*pgtypes.DoltgresType); !ok {
						// GMS ranking functions (rank, dense_rank, ntile) use Uint64 internally,
						// but Postgres returns bigint. ExplicitCast overrides the Numeric mapping.
						if expr.Type(ctx).Type() == query.Type_UINT64 {
							return pgexprs.NewExplicitCast(expr, pgtypes.Int64), transform.NewTree, nil
						}
						return pgexprs.NewGMSCast(expr), transform.NewTree, nil
					}
					// After this point expr IS a DoltgresType.
					doltType := expr.Type(ctx).(*pgtypes.DoltgresType)
					// GMS window SUM propagates the child column's DoltgresType via Sum.Type(),
					// but always accumulates float64 at runtime. Match by ColumnId in
					// Window.SelectExprs; for integer aggregates apply Float64→Int64.
					if winNode := findWindowNode(child); winNode != nil {
						for _, selectExpr := range winNode.SelectExprs {
							ide, ok := selectExpr.(sql.IdExpression)
							if !ok || ide.Id() != expr.Id() {
								continue
							}
							if _, isAgg := selectExpr.(sql.Aggregation); isAgg {
								if doltType.Equals(pgtypes.Int32) || doltType.Equals(pgtypes.Int16) || doltType.Equals(pgtypes.Int64) {
									return pgexprs.NewAssignmentCast(expr, pgtypes.Float64, pgtypes.Int64), transform.NewTree, nil
								}
							}
							break
						}
					}
					// An outer Project may carry a stale declared type; child.Schema() holds
					// the corrected type after bottom-up transform. Re-annotate on mismatch.
					if innerType := windowSchemaTypeByName(child, ctx, expr.Name()); innerType != nil {
						if !doltType.Equals(innerType) {
							return pgexprs.NewAssignmentCast(expr, innerType, innerType), transform.NewTree, nil
						}
					}
				}
				// If a GetField references a SubqueryAlias and has a GMS type but the SubqueryAlias
				// schema already has a DoltgresType (e.g. from AggCast propagation through a CTE),
				// rebuild the GetField with the correct type so downstream operators compile for the
				// actual runtime type rather than the planbuilder-assigned GMS type.
				if _, ok := child.(*plan.SubqueryAlias); ok {
					if _, ok := expr.Type(ctx).(*pgtypes.DoltgresType); !ok {
						if actualType := windowSchemaTypeByName(child, ctx, expr.Name()); actualType != nil {
							return expression.NewGetField(expr.Index(), actualType, expr.Name(), expr.IsNullable(ctx)), transform.NewTree, nil
						}
					}
				}
				// When a GroupBy child (possibly through a Having wrapper) has AggCast applied
				// (e.g. SUM over integers), its schema reports the corrected DoltgresType, but the
				// GetField may have been constructed earlier with a stale type: either a GMS type
				// (float64 from GMS's internal Sum type promotion), or the pre-AggCast DoltgresType
				// (e.g. Int32, from Sum.Type() before AggCast overrode it to Int64). Re-annotate so
				// rowToBytes uses the correct Doltgres type for wire serialization. The runtime value
				// is already correct from AggCast.
				if gb := findGroupByChild(child); gb != nil {
					if doltType := groupBySchemaTypeById(gb, ctx, expr.Id()); doltType != nil {
						if currentType, ok := expr.Type(ctx).(*pgtypes.DoltgresType); !ok || !currentType.Equals(doltType) {
							return pgexprs.NewAssignmentCast(expr, doltType, doltType), transform.NewTree, nil
						}
					}
				}
			}
			return expr, transform.SameTree, nil
		case *expression.Literal:
			// We want to leave limit literals alone, as they are expected to be GMS types when they appear in certain
			// parts of the query (subqueries in particular)
			// TODO: fix the limit and offset validation analysis to handle doltgres types
			if _, isLimit := n.(*plan.Limit); isLimit {
				break
			}
			if _, isOffset := n.(*plan.Offset); isOffset {
				break
			}
			return typeSanitizerLiterals(ctx, expr)
		case *expression.Not, *expression.And, *expression.Or, *expression.Like:
			return pgexprs.NewGMSCast(expr), transform.NewTree, nil
		case sql.FunctionExpression:
			// Compiled functions are Doltgres functions. We're only concerned with GMS functions.
			if _, ok := expr.(framework.Function); !ok {
				// Window functions (and GMS Sum) implement sql.WindowAdaptableExpression and must
				// NOT be wrapped in GMSCast — that hides the interface from windowToIter.
				if _, ok := expr.(sql.WindowAdaptableExpression); ok {
					// GMS SUM always accumulates float64 internally, but Postgres SUM(int2/int4/int8)
					// returns bigint. In GroupBy context, wrap with AggCast to convert the float64
					// result to int64 while preserving the sql.Aggregation interface.
					if expr.FunctionName() == "Sum" {
						if _, isGroupBy := n.(*plan.GroupBy); isGroupBy {
							if doltType, ok := expr.Type(ctx).(*pgtypes.DoltgresType); ok {
								if doltType.Equals(pgtypes.Int32) || doltType.Equals(pgtypes.Int16) || doltType.Equals(pgtypes.Int64) {
									return pgexprs.NewAggCast(expr.(sql.Aggregation), pgtypes.Int64), transform.NewTree, nil
								}
							}
						}
					}
					return expr, transform.SameTree, nil
				}
				// Some aggregation functions cannot be wrapped due to expectations in the analyzer, so we exclude them here.
				switch expr.FunctionName() {
				case "Count", "CountDistinct", "group_concat", "JSONObjectAgg", "Sum":
				case "coalesce":
					// Replace GMS Coalesce with a Doltgres-native implementation that uses
					// Postgres type-resolution rules (FindCommonType) to infer the result type.
					// GMS's Coalesce.Type() falls back to LongText when its arguments are
					// DoltgresTypes because they don't satisfy GMS's IsNumber/IsText checks.
					if _, isPgCoalesce := expr.(*pgexprs.PgCoalesce); !isPgCoalesce {
						children := expr.Children()
						allDoltgresTypes := true
						for _, child := range children {
							if _, ok := child.Type(ctx).(*pgtypes.DoltgresType); !ok {
								allDoltgresTypes = false
								break
							}
						}
						if allDoltgresTypes {
							pgCoalesce, err := pgexprs.NewPgCoalesce(ctx, children...)
							if err != nil {
								return nil, transform.NewTree, err
							}
							return pgCoalesce, transform.NewTree, nil
						}
					}
					// Fall through to GMSCast if children aren't DoltgresTypes yet.
					if _, ok := expr.Type(ctx).(*pgtypes.DoltgresType); !ok {
						return pgexprs.NewGMSCast(expr), transform.NewTree, nil
					}
				default:
					// Some GMS functions wrap Doltgres parameters, so we'll only handle those that return GMS types
					if _, ok := expr.Type(ctx).(*pgtypes.DoltgresType); !ok {
						return pgexprs.NewGMSCast(expr), transform.NewTree, nil
					}
				}
			}
		case *plan.ExistsSubquery:
			return pgexprs.NewExplicitCast(pgexprs.NewGMSCast(expr), pgtypes.Bool), transform.NewTree, nil
		case *sql.ColumnDefaultValue:
			// Due to how interfaces work, we sometimes pass (*ColumnDefaultValue)(nil), so we have to check for it
			if expr != nil && expr.Expr != nil {
				defaultExpr := expr.Expr
				if _, ok := defaultExpr.Type(ctx).(*pgtypes.DoltgresType); !ok {
					defaultExpr = pgexprs.NewGMSCast(defaultExpr)
				}
				defaultExprType := defaultExpr.Type(ctx).(*pgtypes.DoltgresType)
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
func typeSanitizerLiterals(ctx *sql.Context, gmsLiteral *expression.Literal) (sql.Expression, transform.TreeIdentity, error) {
	// GMS may resolve Doltgres literals and then stick them in GMS literals, so we have to account for that here
	if doltgresType, ok := gmsLiteral.Type(ctx).(*pgtypes.DoltgresType); ok {
		return pgexprs.NewUnsafeLiteral(gmsLiteral.Value(), doltgresType), transform.NewTree, nil
	}
	switch gmsLiteral.Type(ctx).Type() {
	case query.Type_INT8, query.Type_INT16, query.Type_YEAR, query.Type_INT24, query.Type_INT32:
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
		dec, ok := gmsLiteral.Value().(*apd.Decimal)
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
		return nil, transform.NewTree, errors.Errorf("SANITIZER: encountered a GMS type that cannot be handled: %s", gmsLiteral.Type(ctx).String())
	}
}

// isWindowOrWrapsWindow reports whether n is, or transitively wraps, a *plan.Window.
func isWindowOrWrapsWindow(n sql.Node) bool {
	return findWindowNode(n) != nil
}

// findWindowNode traverses Sort/Limit/Offset/Distinct/Filter/Project wrappers to find
// the underlying *plan.Window, or nil if n does not wrap a Window.
func findWindowNode(n sql.Node) *plan.Window {
	if w, ok := n.(*plan.Window); ok {
		return w
	}
	switch n.(type) {
	case *plan.Sort, *plan.Limit, *plan.Offset, *plan.Distinct, *plan.Filter, *plan.Project:
		children := n.Children()
		if len(children) == 1 {
			return findWindowNode(children[0])
		}
	}
	return nil
}

// findGroupByChild returns the GroupBy if n is a GroupBy or transitively wraps one
// through Having nodes. Returns nil if n does not expose a GroupBy this way.
func findGroupByChild(n sql.Node) *plan.GroupBy {
	switch typed := n.(type) {
	case *plan.GroupBy:
		return typed
	case *plan.Having:
		children := typed.Children()
		if len(children) == 1 {
			return findGroupByChild(children[0])
		}
	}
	return nil
}

// windowSchemaTypeByName returns the DoltgresType for a named column in n's schema,
// or nil if no match is found.
func windowSchemaTypeByName(n sql.Node, ctx *sql.Context, name string) *pgtypes.DoltgresType {
	for _, col := range n.Schema(ctx) {
		if strings.EqualFold(col.Name, name) {
			if t, ok := col.Type.(*pgtypes.DoltgresType); ok {
				return t
			}
		}
	}
	return nil
}

// groupBySchemaTypeById returns the DoltgresType of the SelectDeps expression in gb whose
// ColumnId matches id, or nil if no such expression exists or its type isn't a DoltgresType.
func groupBySchemaTypeById(gb *plan.GroupBy, ctx *sql.Context, id sql.ColumnId) *pgtypes.DoltgresType {
	for _, e := range gb.SelectDeps {
		ide, ok := e.(sql.IdExpression)
		if !ok || ide.Id() != id {
			continue
		}
		t, _ := e.Type(ctx).(*pgtypes.DoltgresType)
		return t
	}
	return nil
}
