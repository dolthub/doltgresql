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
	"github.com/dolthub/go-mysql-server/sql/plan"
	"github.com/dolthub/go-mysql-server/sql/transform"

	pgnodes "github.com/dolthub/doltgresql/server/node"

	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/server/expression"
	pgtransform "github.com/dolthub/doltgresql/server/transform"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// ResolveType replaces types.ResolvableType to appropriate pgtypes.DoltgresType.
func ResolveType(ctx *sql.Context, a *analyzer.Analyzer, node sql.Node, scope *plan.Scope, selector analyzer.RuleSelector, qFlags *sql.QueryFlags) (sql.Node, transform.TreeIdentity, error) {
	n, sameExpr, err := ResolveTypeForExprs(ctx, a, node, scope, selector, qFlags)
	if err != nil {
		return nil, transform.NewTree, err
	}

	n, sameNode, err := ResolveTypeForNodes(ctx, a, n, scope, selector, qFlags)
	if err != nil {
		return nil, transform.NewTree, err
	}

	return n, sameExpr && sameNode, nil
}

// ResolveTypeForNodes replaces types.ResolvableType to appropriate pgtypes.DoltgresType.
func ResolveTypeForNodes(ctx *sql.Context, a *analyzer.Analyzer, node sql.Node, scope *plan.Scope, selector analyzer.RuleSelector, qFlags *sql.QueryFlags) (sql.Node, transform.TreeIdentity, error) {
	return transform.Node(node, func(node sql.Node) (sql.Node, transform.TreeIdentity, error) {
		var same = transform.SameTree
		switch n := node.(type) {
		case *plan.AddColumn:
			col := n.Column()
			if rt, ok := col.Type.(*pgtypes.DoltgresType); ok && !rt.IsResolvedType() {
				dt, err := resolveType(ctx, rt)
				if err != nil {
					return nil, transform.NewTree, err
				}
				same = transform.NewTree
				col.Type = dt
			}
			return node, same, nil
		case *pgnodes.CreateFunction:
			retType, err := resolveType(ctx, n.ReturnType)
			if err != nil {
				return nil, transform.NewTree, err
			}
			paramTypes := make([]*pgtypes.DoltgresType, len(n.ParameterTypes))
			for i := range n.ParameterTypes {
				paramTypes[i], err = resolveType(ctx, n.ParameterTypes[i])
				if err != nil {
					return nil, transform.NewTree, err
				}
			}
			n.ReturnType = retType
			n.ParameterTypes = paramTypes
			return node, transform.NewTree, nil
		case *pgnodes.CreateProcedure:
			paramTypes := make([]*pgtypes.DoltgresType, len(n.ParameterTypes))
			for i := range n.ParameterTypes {
				var err error
				paramTypes[i], err = resolveType(ctx, n.ParameterTypes[i])
				if err != nil {
					return nil, transform.NewTree, err
				}
			}
			n.ParameterTypes = paramTypes
			return node, transform.NewTree, nil
		case *plan.CreateTable:
			for _, col := range n.TargetSchema() {
				if rt, ok := col.Type.(*pgtypes.DoltgresType); ok && !rt.IsResolvedType() {
					dt, err := resolveType(ctx, rt)
					if err != nil {
						return nil, transform.NewTree, err
					}
					same = transform.NewTree
					col.Type = dt
				}
			}
			return node, same, nil
		case *plan.ModifyColumn:
			col := n.NewColumn()
			if rt, ok := col.Type.(*pgtypes.DoltgresType); ok && !rt.IsResolvedType() {
				dt, err := resolveType(ctx, rt)
				if err != nil {
					return nil, transform.NewTree, err
				}
				same = transform.NewTree
				col.Type = dt
			}
			return node, same, nil
		default:
			// TODO: add nodes that use unresolved types like domain
			return node, transform.SameTree, nil
		}
	})
}

// ResolveTypeForExprs replaces types.ResolvableType to appropriate pgtypes.DoltgresType.
func ResolveTypeForExprs(ctx *sql.Context, a *analyzer.Analyzer, node sql.Node, scope *plan.Scope, selector analyzer.RuleSelector, qFlags *sql.QueryFlags) (sql.Node, transform.TreeIdentity, error) {
	return pgtransform.NodeExprsWithOpaque(node, func(expr sql.Expression) (sql.Expression, transform.TreeIdentity, error) {
		var same = transform.SameTree
		switch e := expr.(type) {
		case *expression.ExplicitCast:
			if rt, ok := e.Type().(*pgtypes.DoltgresType); ok && !rt.IsResolvedType() {
				dt, err := resolveType(ctx, rt)
				if err != nil {
					return nil, transform.NewTree, err
				}
				same = transform.NewTree
				if !dt.IsDefined {
					return nil, transform.NewTree, pgtypes.ErrTypeIsOnlyAShell.New(dt.Name())
				} else {
					expr = e.WithCastToType(dt)
				}
			}
			return expr, same, nil
		default:
			// TODO: add expressions that use unresolved types like domain
			return e, transform.SameTree, nil
		}
	})
}

// resolveType resolves any type that is unresolved yet. (e.g.: domain types, built-in types that schema specified, etc.)
func resolveType(ctx *sql.Context, typ *pgtypes.DoltgresType) (*pgtypes.DoltgresType, error) {
	if typ.IsResolvedType() {
		return typ, nil
	}
	schema, err := core.GetSchemaName(ctx, nil, typ.ID.SchemaName())
	if err != nil {
		return nil, err
	}
	typs, err := core.GetTypesCollectionFromContext(ctx)
	if err != nil {
		return nil, err
	}
	resolvedTyp, err := typs.GetType(ctx, id.NewType(schema, typ.ID.TypeName()))
	if err != nil {
		return nil, err
	}
	if resolvedTyp == nil {
		// If a blank schema is provided, then we'll also try the pg_catalog, since a type is most likely to be there
		if typ.ID.SchemaName() == "" {
			resolvedTyp, err = typs.GetType(ctx, id.NewType("pg_catalog", typ.ID.TypeName()))
			if err != nil {
				return nil, err
			}
			if resolvedTyp != nil && (typ.ID.TypeName() == "unknown" || resolvedTyp.ID != pgtypes.Unknown.ID) {
				return resolvedTyp, nil
			}
		}
		return nil, pgtypes.ErrTypeDoesNotExist.New(typ.Name())
	}
	return resolvedTyp, nil
}
