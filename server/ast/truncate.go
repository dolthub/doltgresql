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

// nodeTruncate handles *tree.Truncate nodes.
func nodeTruncate(ctx *Context, node *tree.Truncate) (vitess.Statement, error) {
	if node == nil || len(node.Tables) == 0 {
		return nil, nil
	}
	switch node.DropBehavior {
	case tree.DropDefault:
		// Default behavior, nothing to do
	case tree.DropRestrict:
		// RESTRICT is the default behavior.
	case tree.DropCascade:
		return nil, errors.Errorf("CASCADE is not yet supported")
	}
	if len(node.Tables) > 1 {
		tables := make([]doltdb.TableName, 0, len(node.Tables))
		authNames := make([]string, 0, len(node.Tables)*3)
		for _, name := range node.Tables {
			if name.ExplicitCatalog {
				return nil, errors.New("cross-database multi-table TRUNCATE is not yet supported")
			}
			table, err := nodeTableName(ctx, &name)
			if err != nil {
				return nil, err
			}
			tables = append(tables, doltdb.TableName{Schema: table.SchemaQualifier.String(), Name: table.Name.String()})
			authNames = append(authNames, table.DbQualifier.String(), table.SchemaQualifier.String(), table.Name.String())
		}
		return vitess.InjectedStatement{
			Statement: &pgnodes.TruncateTables{Tables: tables},
			Auth:      vitess.AuthInformation{AuthType: auth.AuthType_TRUNCATE, TargetType: auth.AuthTargetType_TableIdentifiers, TargetNames: authNames},
		}, nil
	}
	tableName, err := nodeTableName(ctx, &node.Tables[0])
	if err != nil {
		return nil, err
	}
	return &vitess.DDL{
		Action: vitess.TruncateStr,
		Table:  tableName,
		Auth: vitess.AuthInformation{
			AuthType:    auth.AuthType_TRUNCATE,
			TargetType:  auth.AuthTargetType_TableIdentifiers,
			TargetNames: []string{tableName.DbQualifier.String(), tableName.SchemaQualifier.String(), tableName.Name.String()},
		},
	}, nil
}
