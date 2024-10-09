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
)

// nodeShowTables handles *tree.ShowTables nodes.
func nodeShowTables(node *tree.ShowTables) (vitess.Statement, error) {
	if node == nil {
		return nil, nil
	}
	
	var opts *vitess.ShowTablesOpt
	
	if node.ExplicitCatalog {
		opts = &vitess.ShowTablesOpt{
			DbName:     node.CatalogName.String(),
			SchemaName: node.SchemaName.String(),
		}
	} else if node.ExplicitSchema {
		opts = &vitess.ShowTablesOpt{
			SchemaName: node.SchemaName.String(),
		}
	}

	return &vitess.Show{
		Type:          "tables",
		ShowTablesOpt: opts,
	}, nil
}
