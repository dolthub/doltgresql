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
	"github.com/cockroachdb/errors"

	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
)

// nodeDropIndex handles *tree.DropIndex nodes.
func nodeDropIndex(ctx *Context, node *tree.DropIndex) (*vitess.AlterTable, error) {
	if node == nil || len(node.IndexList) == 0 {
		return nil, nil
	}
	switch node.DropBehavior {
	case tree.DropDefault:
		// Default behavior, nothing to do
	case tree.DropRestrict:
		return nil, errors.Errorf("RESTRICT is not yet supported")
	case tree.DropCascade:
		return nil, errors.Errorf("CASCADE is not yet supported")
	}
	if len(node.IndexList) > 1 {
		return nil, errors.Errorf("multi-index dropping is not yet supported")
	}
	if node.Concurrently {
		return nil, errors.Errorf("concurrent indexes are not yet supported")
	}
	var tableName vitess.TableName
	ddls := make([]*vitess.DDL, len(node.IndexList))
	for i, index := range node.IndexList {
		newTableName, err := nodeTableName(ctx, &index.Table)
		if err != nil {
			return nil, err
		}
		if !tableName.Name.IsEmpty() && tableName.String() != newTableName.String() {
			return nil, errors.Errorf("dropping indexes from different tables is not yet supported")
		}
		tableName = newTableName
		ddls[i] = &vitess.DDL{
			Action:   vitess.AlterStr,
			Table:    tableName,
			IfExists: node.IfExists,
			IndexSpec: &vitess.IndexSpec{
				Action:   vitess.DropStr,
				FromName: vitess.NewColIdent(string(index.Index)),
				ToName:   vitess.NewColIdent(string(index.Index)),
			},
		}
	}
	return &vitess.AlterTable{
		Table:      tableName,
		Statements: ddls,
	}, nil
}
