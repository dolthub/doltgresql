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
	"github.com/dolthub/go-mysql-server/sql/expression"
	"github.com/dolthub/go-mysql-server/sql/plan"
	"github.com/dolthub/go-mysql-server/sql/transform"

	pgexpression "github.com/dolthub/doltgresql/server/expression"
	pgtransform "github.com/dolthub/doltgresql/server/transform"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// resolveScalarFunctionAliases makes a bare alias of a scalar function in FROM denote its value,
// even when its named OUT parameter has a different name. Resolve this before operators and casts
// consume its type, and refresh references to the value through enclosing subqueries.
func resolveScalarFunctionAliases(ctx *sql.Context, node sql.Node) (sql.Node, transform.TreeIdentity, error) {
	return pgtransform.NodeWithOpaque(ctx, node, func(ctx *sql.Context, n sql.Node) (sql.Node, transform.TreeIdentity, error) {
		rewrite := func(ctx *sql.Context, _ sql.Node, expr sql.Expression) (sql.Expression, transform.TreeIdentity, error) {
			if row, ok := expr.(*pgexpression.TableToComposite); ok && len(row.Children()) == 1 {
				if field, ok := row.Children()[0].(*expression.GetField); ok && isScalarTableFunction(ctx, n, field.TableId()) {
					return field.WithName(field.Table()), transform.NewTree, nil
				}
			}
			if field, ok := expr.(*expression.GetField); ok {
				if typ, ok := field.Type(ctx).(*pgtypes.DoltgresType); ok && typ.IsCompositeType() {
					if scalar := scalarSubqueryColumnType(ctx, n, field); scalar != nil {
						return expression.NewGetFieldWithTable(field.Index(), int(field.TableId()), scalar,
							field.Database(), field.Table(), field.Name(), field.IsNullable(ctx)).WithId(field.Id()), transform.NewTree, nil
					}
				}
			}
			return expr, transform.SameTree, nil
		}
		rewritten, same, err := transform.OneNodeExprsWithNode(ctx, n, rewrite)
		if err != nil {
			return nil, transform.SameTree, err
		}
		// GMS keeps a second copy of subquery projections for predicate pushdown.
		if alias, ok := rewritten.(*plan.SubqueryAlias); ok && alias.ScopeMapping != nil {
			mappings := make(map[sql.ColumnId]sql.Expression, len(alias.ScopeMapping))
			for id, expr := range alias.ScopeMapping {
				expr, unchanged, err := transform.ExprWithNode(ctx, alias, expr, rewrite)
				if err != nil {
					return nil, transform.SameTree, err
				}
				mappings[id] = expr
				same = same && unchanged
			}
			if !same {
				rewritten = alias.WithScopeMapping(mappings)
			}
		}
		return rewritten, same, nil
	})
}

func isScalarTableFunction(ctx *sql.Context, node sql.Node, tableID sql.TableId) bool {
	return transform.InspectUp(ctx, node, func(ctx *sql.Context, n sql.Node) bool {
		alias, ok := n.(*plan.TableAlias)
		if !ok || alias.Id() != tableID || len(alias.Schema(ctx)) != 1 {
			return false
		}
		if _, ok := alias.Child.(sql.TableFunction); !ok {
			return false
		}
		expressions, ok := alias.Child.(sql.Expressioner)
		if !ok || len(expressions.Expressions()) != 1 {
			return false
		}
		typ, ok := expressions.Expressions()[0].Type(ctx).(*pgtypes.DoltgresType)
		return ok && !typ.IsRecordType() && !typ.IsCompositeType()
	})
}

// scalarSubqueryColumnType obtains the updated type when a function alias projected by a
// subquery was initially treated as a record. Ordinary table records retain their composite type.
func scalarSubqueryColumnType(ctx *sql.Context, node sql.Node, field *expression.GetField) sql.Type {
	var typ sql.Type
	transform.InspectUp(ctx, node, func(ctx *sql.Context, n sql.Node) bool {
		alias, ok := n.(*plan.SubqueryAlias)
		if !ok || alias.Id() != field.TableId() || !alias.Columns().Contains(field.Id()) {
			return false
		}
		schema := alias.Schema(ctx)
		index := 0
		alias.Columns().ForEach(func(id sql.ColumnId) {
			if id == field.Id() && index < len(schema) {
				if scalar, ok := schema[index].Type.(*pgtypes.DoltgresType); ok && !scalar.IsCompositeType() && !scalar.IsRecordType() {
					typ = scalar
				}
			}
			index++
		})
		return true
	})
	return typ
}
