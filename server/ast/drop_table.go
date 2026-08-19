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
	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"

	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
	"github.com/dolthub/doltgresql/server/auth"
	pgnodes "github.com/dolthub/doltgresql/server/node"
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
	switch node.DropBehavior {
	case tree.DropDefault, tree.DropRestrict:
		// RESTRICT is the default behavior in Postgres, so both are handled by the standard DROP TABLE path, which
		// refuses to drop a table that other objects depend on.
	case tree.DropCascade:
		// CASCADE also drops the objects that depend on the dropped tables, which requires Doltgres-specific handling.
		var databaseName string
		dropTables := make([]doltdb.TableName, len(tableNames))
		for i, tableName := range tableNames {
			dbQualifier := tableName.DbQualifier.String()
			if len(dbQualifier) > 0 {
				if len(databaseName) > 0 && databaseName != dbQualifier {
					return nil, errors.Errorf("DROP TABLE CASCADE is currently only supported for a single database")
				}
				databaseName = dbQualifier
			}
			dropTables[i] = doltdb.TableName{
				Schema: tableName.SchemaQualifier.String(),
				Name:   tableName.Name.String(),
			}
		}
		return vitess.InjectedStatement{
			Statement: pgnodes.NewDropTableCascade(node.IfExists, databaseName, dropTables),
			Children:  nil,
			Auth:      authInformation,
		}, nil
	}
	return &vitess.DDL{
		Action:     vitess.DropStr,
		FromTables: tableNames,
		IfExists:   node.IfExists,
		Auth:       authInformation,
	}, nil
}
