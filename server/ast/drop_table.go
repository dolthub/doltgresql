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
	"github.com/dolthub/doltgresql/server/auth"
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
		return nil, errors.Errorf("RESTRICT is not yet supported")
	case tree.DropCascade:
		return nil, errors.Errorf("CASCADE is not yet supported")
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
	return &vitess.DDL{
		Action:     vitess.DropStr,
		FromTables: tableNames,
		IfExists:   node.IfExists,
		Auth: vitess.AuthInformation{
			AuthType:    auth.AuthType_DROPTABLE,
			TargetType:  auth.AuthTargetType_TableIdentifiers,
			TargetNames: authTableNames,
		},
	}, nil
}
