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
)

// nodeDropTable handles *tree.DropTable nodes.
func nodeDropTable(ctx *Context, node *tree.DropTable) (*vitess.DDL, error) {
	if node == nil || len(node.Names) == 0 {
		return nil, nil
	}
	switch node.DropBehavior {
	case tree.DropDefault:
		// Default behavior, nothing to do
	case tree.DropRestrict:
		return nil, fmt.Errorf("RESTRICT is not yet supported")
	case tree.DropCascade:
		return nil, fmt.Errorf("CASCADE is not yet supported")
	}
	tableNames := make([]vitess.TableName, len(node.Names))
	for i := range node.Names {
		var err error
		tableNames[i], err = nodeTableName(ctx, &node.Names[i])
		if err != nil {
			return nil, err
		}
	}
	return &vitess.DDL{
		Action:     vitess.DropStr,
		FromTables: tableNames,
		IfExists:   node.IfExists,
	}, nil
}
