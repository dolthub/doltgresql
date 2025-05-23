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
	"github.com/cockroachdb/errors"

	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
	pgnodes "github.com/dolthub/doltgresql/server/node"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// nodeCreateType handles *tree.CreateType nodes.
func nodeCreateType(ctx *Context, node *tree.CreateType) (vitess.Statement, error) {
	if node == nil {
		return nil, nil
	}
	name, err := nodeUnresolvedObjectName(ctx, node.TypeName)
	if err != nil {
		return nil, err
	}
	schemaName := name.SchemaQualifier.String()
	typName := name.Name.String()
	var createTypeNode *pgnodes.CreateType
	switch node.Variety {
	case tree.Composite:
		typs := make([]pgnodes.CompositeAsType, len(node.Composite.Types))
		for i, t := range node.Composite.Types {
			_, dataType, err := nodeResolvableTypeReference(ctx, t.Type)
			if err != nil {
				return nil, err
			}

			if dataType == pgtypes.Record {
				return nil, errors.Errorf(`column "%s" has pseudo-type record`, t.AttrName)
			}

			typs[i] = pgnodes.CompositeAsType{
				AttrName:  t.AttrName,
				Typ:       dataType,
				Collation: t.Collate,
			}
		}
		createTypeNode = pgnodes.NewCreateCompositeType(schemaName, typName, typs)
	case tree.Enum:
		createTypeNode = pgnodes.NewCreateEnumType(schemaName, typName, node.Enum.Labels)
	case tree.Range:
		return nil, errors.Errorf("CREATE RANGE TYPE is not yet supported")
	case tree.Base:
		return nil, errors.Errorf("CREATE BASE TYPE is not yet supported")
	case tree.Shell:
		createTypeNode = pgnodes.NewCreateShellType(schemaName, typName)
	case tree.Domain:
		// NOT POSSIBLE
		return nil, errors.Errorf("use CREATE DOMAIN to create domain type")
	}

	return vitess.InjectedStatement{
		Statement: createTypeNode,
	}, nil
}
