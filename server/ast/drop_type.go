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
	"fmt"

	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
	pgnodes "github.com/dolthub/doltgresql/server/node"
)

// nodeDropType handles *tree.DropType nodes.
func nodeDropType(ctx *Context, node *tree.DropType) (vitess.Statement, error) {
	if node == nil {
		return nil, nil
	}
	if len(node.Names) != 1 {
		return nil, fmt.Errorf("dropping multiple types in DROP TYPE is not yet supported")
	}
	tn := node.Names[0].ToTableName()
	return vitess.InjectedStatement{
		Statement: pgnodes.NewDropType(
			node.IfExists,
			tn.Catalog(),
			tn.Schema(),
			tn.Object(),
			node.DropBehavior == tree.DropCascade),
		Children: nil,
	}, nil
}
