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

// nodeDropRole handles *tree.DropRole nodes.
func nodeDropRole(node *tree.DropRole) (vitess.Statement, error) {
	if node == nil {
		return nil, nil
	}
	var names []string
	for _, name := range node.Names {
		switch name := name.(type) {
		case *tree.StrVal:
			names = append(names, name.RawString())
		default:
			return nil, fmt.Errorf("unknown type `%T` for DROP ROLE name", name)
		}
	}
	return vitess.InjectedStatement{
		Statement: &pgnodes.DropRole{
			Names:    names,
			IfExists: node.IfExists,
		},
		Children: nil,
	}, nil
}
