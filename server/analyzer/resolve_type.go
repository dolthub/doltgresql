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
	"fmt"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/analyzer"
	"github.com/dolthub/go-mysql-server/sql/plan"
	"github.com/dolthub/go-mysql-server/sql/transform"

	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
	"github.com/dolthub/doltgresql/server/expression"
	"github.com/dolthub/doltgresql/server/node"
	pgtransform "github.com/dolthub/doltgresql/server/transform"
	"github.com/dolthub/doltgresql/server/types"
)

// ResolveType replaces types.ResolvableType to appropriate types.DoltgresType.
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

// ResolveTypeForNodes replaces types.ResolvableType to appropriate types.DoltgresType.
func ResolveTypeForNodes(ctx *sql.Context, a *analyzer.Analyzer, node sql.Node, scope *plan.Scope, selector analyzer.RuleSelector, qFlags *sql.QueryFlags) (sql.Node, transform.TreeIdentity, error) {
	return transform.Node(node, func(node sql.Node) (sql.Node, transform.TreeIdentity, error) {
		var same = transform.SameTree
		switch n := node.(type) {
		case *plan.CreateTable:
			for _, col := range n.TargetSchema() {
				if rt, ok := col.Type.(types.DoltgresType); ok && !rt.IsResolvedType() {
					dt, err := resolveType(ctx, rt)
					if err != nil {
						return nil, transform.NewTree, err
					}
					same = transform.NewTree
					col.Type = dt
				}
			}
			return node, same, nil
		case *plan.AddColumn:
			col := n.Column()
			if rt, ok := col.Type.(types.DoltgresType); ok && !rt.IsResolvedType() {
				dt, err := resolveType(ctx, rt)
				if err != nil {
					return nil, transform.NewTree, err
				}
				same = transform.NewTree
				col.Type = dt
			}
			return node, same, nil
		case *plan.ModifyColumn:
			col := n.NewColumn()
			if rt, ok := col.Type.(types.DoltgresType); ok && !rt.IsResolvedType() {
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

// ResolveTypeForExprs replaces types.ResolvableType to appropriate types.DoltgresType.
func ResolveTypeForExprs(ctx *sql.Context, a *analyzer.Analyzer, node sql.Node, scope *plan.Scope, selector analyzer.RuleSelector, qFlags *sql.QueryFlags) (sql.Node, transform.TreeIdentity, error) {
	return pgtransform.NodeExprsWithOpaque(node, func(expr sql.Expression) (sql.Expression, transform.TreeIdentity, error) {
		var same = transform.SameTree
		switch e := expr.(type) {
		case *expression.ExplicitCast:
			if rt, ok := e.Type().(types.DoltgresType); ok && !rt.IsResolvedType() {
				dt, err := resolveType(ctx, rt)
				if err != nil {
					return nil, transform.NewTree, err
				}
				same = transform.NewTree
				if dt.TypType == types.TypeType_Domain {
					nullable := !dt.NotNull
					colChecks, err := getDomainCheckConstraintsForCast(ctx, a, dt.Checks, e.Child())
					if err != nil {
						return nil, transform.NewTree, err
					}
					expr = e.WithDomainCastToType(dt, nullable, colChecks)
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

// resolveType resolves any type that is unresolved yet. (e.g.: domain types)
func resolveType(ctx *sql.Context, typ types.DoltgresType) (types.DoltgresType, error) {
	schema, err := core.GetSchemaName(ctx, nil, typ.Schema)
	if err != nil {
		return types.DoltgresType{}, err
	}
	typs, err := core.GetTypesCollectionFromContext(ctx)
	if err != nil {
		return types.DoltgresType{}, err
	}
	resolvedTyp, exists := typs.GetType(schema, typ.Name)
	if !exists {
		return types.DoltgresType{}, types.ErrTypeDoesNotExist.New(typ.Name)
	}
	return resolvedTyp, nil
}

// getDomainCheckConstraintsForCast takes the check constraint definitions, parses, builds and returns sql.CheckConstraints.
func getDomainCheckConstraintsForCast(ctx *sql.Context, a *analyzer.Analyzer, checkDefs []*sql.CheckDefinition, value sql.Expression) (sql.CheckConstraints, error) {
	checks := make(sql.CheckConstraints, len(checkDefs))
	for i, check := range checkDefs {
		q := fmt.Sprintf("select %s", check.CheckExpression)
		checkExpr, err := parseAndReplaceDomainCheckConstraint(ctx, a, check.CheckExpression, q, tree.DomainColumn{})
		if err != nil {
			return nil, err
		}

		// replace DomainColumn with given sql.Expression
		checkExpr, _, _ = transform.Expr(checkExpr, func(expr sql.Expression) (sql.Expression, transform.TreeIdentity, error) {
			switch e := expr.(type) {
			case *node.DomainColumn:
				expr = value
				return expr, transform.NewTree, nil
			default:
				return e, transform.SameTree, nil
			}
		})
		checks[i] = &sql.CheckConstraint{
			Name:     check.Name,
			Expr:     checkExpr,
			Enforced: true,
		}
	}
	return checks, nil
}
