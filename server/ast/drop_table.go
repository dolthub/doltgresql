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
	"github.com/dolthub/doltgresql/server/auth"
)

// nodeDropTable handles *tree.DropTable nodes.
func nodeDropTable(ctx *Context, node *tree.DropTable) (vitess.Statement, error) {
	if node == nil || len(node.Names) == 0 {
		return nil, nil
	}
	tableNames := make([]vitess.TableName, len(node.Names))
	authTableNames := make([]string, 0, len(node.Names)*3)
	for i := range node.Names {
		var err error
		tableNames[i], err = nodeTableName(ctx, &node.Names[i])
		if err != nil {
			return nil, err
		}
		authTableNames = append(authTableNames,
			tableNames[i].DbQualifier.String(), tableNames[i].SchemaQualifier.String(), tableNames[i].Name.String())
	}
	authInformation := vitess.AuthInformation{
		AuthType:    auth.AuthType_DROPTABLE,
		TargetType:  auth.AuthTargetType_TableIdentifiers,
		TargetNames: authTableNames,
	}
	return &vitess.DDL{
		Action:     vitess.DropStr,
		FromTables: tableNames,
		IfExists:   node.IfExists,
		// RESTRICT is the default behavior in Postgres, so both are handled by the standard DROP TABLE path, which
		// refuses to drop a table that other objects depend on. CASCADE also drops the dependent objects, which the
		// DROP TABLE pre-execution hook handles.
		Cascade: node.DropBehavior == tree.DropCascade,
		Auth:    authInformation,
	}, nil
}
