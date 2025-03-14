// Copyright 2025 Dolthub, Inc.
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

// nodeDropFunction handles *tree.DropFunction nodes.
func nodeDropFunction(_ *Context, node *tree.DropFunction) (vitess.Statement, error) {
	if node == nil {
		return nil, nil
	}

	if node.DropBehavior == tree.DropCascade {
		return nil, fmt.Errorf("DROP FUNCTION with CASCADE is not supported yet")
	}

	if len(node.Functions) == 0 {
		return nil, fmt.Errorf("no function name specified for DROP FUNCTION")
	}

	return vitess.InjectedStatement{
		Statement: pgnodes.NewDropFunction(
			node.IfExists,
			node.Functions,
			node.DropBehavior == tree.DropCascade),
		Children: nil,
	}, nil
}
