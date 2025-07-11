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

// nodeCreateExtension handles *tree.CreateExtension nodes.
func nodeCreateExtension(ctx *Context, node *tree.CreateExtension) (vitess.Statement, error) {
	if len(node.Schema) > 0 {
		return NotYetSupportedError("SCHEMA is not yet supported")
	}
	if len(node.Version) > 0 {
		return NotYetSupportedError("VERSION is not yet supported")
	}
	if node.Cascade {
		return NotYetSupportedError("CASCADE is not yet supported")
	}
	return vitess.InjectedStatement{
		Statement: pgnodes.NewCreateExtension(
			string(node.Name),
			node.IfNotExists,
			node.Schema,
			node.Version,
			node.Cascade,
		),
		Children: nil,
	}, nil
}
