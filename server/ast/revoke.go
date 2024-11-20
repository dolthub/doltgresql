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

	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/postgres/parser/privilege"
	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
	"github.com/dolthub/doltgresql/server/auth"
	pgnodes "github.com/dolthub/doltgresql/server/node"
)

// nodeRevoke handles *tree.Revoke nodes.
func nodeRevoke(ctx *Context, node *tree.Revoke) (vitess.Statement, error) {
	if node == nil {
		return nil, nil
	}
	var revokeTable *pgnodes.RevokeTable
	var revokeSchema *pgnodes.RevokeSchema
	var revokeDatabase *pgnodes.RevokeDatabase
	switch node.Targets.TargetType {
	case privilege.Table:
		tables := make([]doltdb.TableName, len(node.Targets.Tables)+len(node.Targets.InSchema))
		for i, table := range node.Targets.Tables {
			normalizedTable, err := table.NormalizeTablePattern()
			if err != nil {
				return nil, err
			}
			switch normalizedTable := normalizedTable.(type) {
			case *tree.TableName:
				if normalizedTable.ExplicitCatalog {
					return nil, fmt.Errorf("revoking privileges from other databases is not yet supported")
				}
				tables[i] = doltdb.TableName{
					Name:   string(normalizedTable.ObjectName),
					Schema: string(normalizedTable.SchemaName),
				}
			case *tree.AllTablesSelector:
				tables[i] = doltdb.TableName{
					Name:   "",
					Schema: string(normalizedTable.SchemaName),
				}
			default:
				return nil, fmt.Errorf(`unexpected table type in REVOKE: %T`, normalizedTable)
			}
		}
		for _, schema := range node.Targets.InSchema {
			tables = append(tables, doltdb.TableName{
				Name:   "",
				Schema: schema,
			})
		}
		privileges, err := convertPrivilegeKinds(auth.PrivilegeObject_TABLE, node.Privileges)
		if err != nil {
			return nil, err
		}
		revokeTable = &pgnodes.RevokeTable{
			Privileges: privileges,
			Tables:     tables,
		}
	case privilege.Schema:
		privileges, err := convertPrivilegeKinds(auth.PrivilegeObject_SCHEMA, node.Privileges)
		if err != nil {
			return nil, err
		}
		revokeSchema = &pgnodes.RevokeSchema{
			Privileges: privileges,
			Schemas:    node.Targets.Names,
		}
	case privilege.Database:
		privileges, err := convertPrivilegeKinds(auth.PrivilegeObject_DATABASE, node.Privileges)
		if err != nil {
			return nil, err
		}
		revokeDatabase = &pgnodes.RevokeDatabase{
			Privileges: privileges,
			Databases:  node.Targets.Databases.ToStrings(),
		}
	default:
		return nil, fmt.Errorf("this form of REVOKE is not yet supported")
	}
	return vitess.InjectedStatement{
		Statement: &pgnodes.Revoke{
			RevokeTable:    revokeTable,
			RevokeSchema:   revokeSchema,
			RevokeDatabase: revokeDatabase,
			RevokeRole:     nil,
			FromRoles:      node.Grantees,
			GrantedBy:      node.GrantedBy,
			GrantOptionFor: node.GrantOptionFor,
			Cascade:        node.DropBehavior == tree.DropCascade,
		},
		Children: nil,
	}, nil
}
