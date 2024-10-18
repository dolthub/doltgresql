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
	"github.com/dolthub/doltgresql/server/types"
)

// ResolveType replaces types.ResolvableType to appropriate types.DoltgresType.
func ResolveType(ctx *sql.Context, a *analyzer.Analyzer, node sql.Node, scope *plan.Scope, selector analyzer.RuleSelector, qFlags *sql.QueryFlags) (sql.Node, transform.TreeIdentity, error) {
	return transform.Node(node, func(node sql.Node) (sql.Node, transform.TreeIdentity, error) {
		switch n := node.(type) {
		case sql.SchemaTarget:
			switch n.(type) {
			case *plan.AlterPK, *plan.AddColumn, *plan.ModifyColumn, *plan.CreateTable, *plan.DropColumn:
				// DDL nodes must resolve any new column type, continue to logic below
				// TODO: types can be used in casting in SELECT, etc. add those nodes
			default:
				// other node types are not altering the schema and therefore don't need resolution of column type
				return node, transform.SameTree, nil
			}

			var same = transform.SameTree
			for _, col := range n.TargetSchema() {
				if rt, ok := col.Type.(types.ResolvableType); ok {
					dt, err := resolveResolvableType(ctx, rt.Typ)
					if err != nil {
						return nil, transform.SameTree, err
					}
					same = transform.NewTree
					col.Type = dt
				}
			}
			return node, same, nil
		default:
			return node, transform.SameTree, nil
		}
	})
}

// resolveResolvableType resolves any type that is unresolved yet.
// TODO: add other types that need resolution at analyzer stage.
func resolveResolvableType(ctx *sql.Context, typ tree.ResolvableTypeReference) (types.DoltgresType, error) {
	switch t := typ.(type) {
	case *tree.UnresolvedObjectName:
		domain := t.ToTableName()
		return resolveDomainType(ctx, string(domain.SchemaName), string(domain.ObjectName))
	}
	return nil, fmt.Errorf("the given type %T is not yet supported", typ)
}

// resolveDomainType resolves DomainType from given schema and domain name.
func resolveDomainType(ctx *sql.Context, schema, domainName string) (types.DoltgresType, error) {
	schema, err := core.GetSchemaName(ctx, nil, schema)
	if err != nil {
		return nil, err
	}
	domains, err := core.GetTypesCollectionFromContext(ctx)
	if err != nil {
		return nil, err
	}
	domain, exists := domains.GetDomainType(schema, domainName)
	if !exists {
		return nil, types.ErrTypeDoesNotExist.New(domainName)
	}

	// TODO: need to account for non build-in type as base type
	asType, ok := types.OidToBuildInDoltgresType[domain.BaseTypeOID]
	if !ok {
		return nil, fmt.Errorf(`cannot resolve base type for "%s" domain type`, domainName)
	}

	return types.DomainType{
		Schema:      schema,
		Name:        domainName,
		AsType:      asType,
		DefaultExpr: domain.Default,
		NotNull:     domain.NotNull,
		Checks:      domain.Checks,
	}, nil
}
