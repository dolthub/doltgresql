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

package ast

import (
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	pgnodes "github.com/dolthub/doltgresql/server/node"

	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
)

// nodeDropExtension handles *tree.DropExtension nodes.
func nodeDropExtension(ctx *Context, node *tree.DropExtension) (vitess.Statement, error) {
	return vitess.InjectedStatement{
		Statement: pgnodes.NewDropExtension(
			node.Names.ToStrings(),
			node.IfExists,
			node.DropBehavior == tree.DropCascade,
		),
		Children: nil,
	}, nil
}
