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
	"strings"

	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
)

// nodeCreateIndex handles *tree.CreateIndex nodes.
func nodeCreateIndex(node *tree.CreateIndex) (*vitess.AlterTable, error) {
	if node == nil {
		return nil, nil
	}
	if node.Concurrently {
		return nil, fmt.Errorf("concurrent indexes are not yet supported")
	}
	if node.Using != "" && strings.ToLower(node.Using) != "btree" {
		return nil, fmt.Errorf("index tablespace is not yet supported")
	}
	if node.Predicate != nil {
		return nil, fmt.Errorf("WHERE is not yet supported")
	}
	indexDef, err := nodeIndexTableDef(&tree.IndexTableDef{
		Name:        node.Name,
		Columns:     node.Columns,
		IndexParams: node.IndexParams,
	})
	if err != nil {
		return nil, err
	}
	tableName, err := nodeTableName(&node.Table)
	if err != nil {
		return nil, err
	}
	var indexType string
	if node.Unique {
		indexType = vitess.UniqueStr
	}
	return &vitess.AlterTable{
		Table: tableName,
		Statements: []*vitess.DDL{
			{
				Action:      vitess.AlterStr,
				Table:       tableName,
				IfNotExists: node.IfNotExists,
				IndexSpec: &vitess.IndexSpec{
					Action:   vitess.CreateStr,
					FromName: indexDef.Info.Name,
					ToName:   indexDef.Info.Name,
					Type:     indexType,
					Columns:  indexDef.Columns,
					Options:  indexDef.Options,
				},
			},
		},
	}, nil
}
