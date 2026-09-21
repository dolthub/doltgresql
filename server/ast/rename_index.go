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
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
	pgnodes "github.com/dolthub/doltgresql/server/node"
)

// nodeRenameIndex handles *tree.RenameIndex nodes.
func nodeRenameIndex(ctx *Context, node *tree.RenameIndex) (vitess.Statement, error) {
	if node == nil {
		return nil, nil
	}
	var dbName, schemaName, tableName string
	if node.Index != nil {
		if node.Index.Table.ExplicitCatalog {
			dbName = string(node.Index.Table.CatalogName)
		}
		if node.Index.Table.ExplicitSchema {
			schemaName = string(node.Index.Table.SchemaName)
		}
		// The table name is only set when using the CockroachDB `table@index` syntax
		tableName = string(node.Index.Table.ObjectName)
	}
	return vitess.InjectedStatement{
		Statement: pgnodes.NewRenameIndex(
			node.IfExists,
			dbName,
			schemaName,
			tableName,
			string(node.Index.Index),
			string(node.NewName),
		),
		Children: nil,
	}, nil
}
