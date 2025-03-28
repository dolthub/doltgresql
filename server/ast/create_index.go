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
	"strings"

	"github.com/cockroachdb/errors"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
)

// nodeCreateIndex handles *tree.CreateIndex nodes.
func nodeCreateIndex(ctx *Context, node *tree.CreateIndex) (*vitess.AlterTable, error) {
	if node == nil {
		return nil, nil
	}
	if node.Concurrently {
		return nil, errors.Errorf("concurrent index creation is not yet supported")
	}
	if node.Using != "" && strings.ToLower(node.Using) != "btree" {
		return nil, errors.Errorf("index method %s is not yet supported", node.Using)
	}
	if node.Predicate != nil {
		return nil, errors.Errorf("WHERE is not yet supported")
	}
	indexDef, err := nodeIndexTableDef(ctx, &tree.IndexTableDef{
		Name:        node.Name,
		Columns:     node.Columns,
		IndexParams: node.IndexParams,
	})
	if err != nil {
		return nil, err
	}
	tableName, err := nodeTableName(ctx, &node.Table)
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
