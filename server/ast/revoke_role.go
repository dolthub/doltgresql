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
	pgnodes "github.com/dolthub/doltgresql/server/node"

	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
)

// nodeRevokeRole handles *tree.RevokeRole nodes.
func nodeRevokeRole(ctx *Context, node *tree.RevokeRole) (vitess.Statement, error) {
	if node == nil {
		return nil, nil
	}
	return vitess.InjectedStatement{
		Statement: &pgnodes.Revoke{
			RevokeRole: &pgnodes.RevokeRole{
				Groups: node.Roles.ToStrings(),
			},
			FromRoles:      node.Members,
			GrantedBy:      node.GrantedBy,
			GrantOptionFor: len(node.Option) > 0,
			Cascade:        node.DropBehavior == tree.DropCascade,
		},
		Children: nil,
	}, nil
}
