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

	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/server/types"
)

// ResolveType replaces types.ResolvableType to appropriate types.DoltgresType.
func ResolveType(ctx *sql.Context, a *analyzer.Analyzer, node sql.Node, scope *plan.Scope, selector analyzer.RuleSelector, qFlags *sql.QueryFlags) (sql.Node, transform.TreeIdentity, error) {
	return transform.Node(node, func(node sql.Node) (sql.Node, transform.TreeIdentity, error) {
		var same = transform.SameTree
		switch n := node.(type) {
		case *plan.CreateTable:
			for _, col := range n.TargetSchema() {
				if rt, ok := col.Type.(*types.DoltgresType); ok && !rt.IsResolvedType() {
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
			if rt, ok := col.Type.(*types.DoltgresType); ok && !rt.IsResolvedType() {
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
			if rt, ok := col.Type.(*types.DoltgresType); ok && !rt.IsResolvedType() {
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

// resolveType resolves any type that is unresolved yet. (e.g.: domain types)
func resolveType(ctx *sql.Context, typ *types.DoltgresType) (*types.DoltgresType, error) {
	schema, err := core.GetSchemaName(ctx, nil, typ.Schema)
	if err != nil {
		return nil, err
	}
	typs, err := core.GetTypesCollectionFromContext(ctx)
	if err != nil {
		return nil, err
	}
	resolvedTyp, exists := typs.GetType(schema, typ.Name)
	if !exists {
		return nil, types.ErrTypeDoesNotExist.New(typ.Name)
	}
	return resolvedTyp, nil
}
