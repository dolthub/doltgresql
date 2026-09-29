// Copyright 2026 Dolthub, Inc.
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

	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
	pgnodes "github.com/dolthub/doltgresql/server/node"
)

// nodeCreatePublication converts the supported publication forms to an executable node.
func nodeCreatePublication(ctx *Context, n *tree.CreatePublication) (vitess.Statement, error) {
	return vitess.InjectedStatement{Statement: &pgnodes.CreatePublication{Name: string(n.Name), AllTables: n.AllTables, Options: n.Options}}, nil
}

// nodeDropPublication converts DROP PUBLICATION. CASCADE and RESTRICT are equivalent.
func nodeDropPublication(ctx *Context, n *tree.DropPublication) (vitess.Statement, error) {
	names := make([]string, len(n.Names))
	for i, name := range n.Names {
		names[i] = string(name)
	}
	return vitess.InjectedStatement{Statement: &pgnodes.DropPublication{Names: names, IfExists: n.IfExists}}, nil
}
