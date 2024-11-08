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

// nodeGrant handles *tree.Grant nodes.
func nodeGrant(ctx *Context, node *tree.Grant) (vitess.Statement, error) {
	if node == nil {
		return nil, nil
	}
	var grantTable *pgnodes.GrantTable
	switch node.Targets.TargetType {
	case privilege.Table:
		tables := make([]doltdb.TableName, len(node.Targets.Tables))
		for i, table := range node.Targets.Tables {
			normalizedTable, err := table.NormalizeTablePattern()
			if err != nil {
				return nil, err
			}
			switch normalizedTable := normalizedTable.(type) {
			case *tree.TableName:
				if normalizedTable.ExplicitCatalog {
					return nil, fmt.Errorf("granting privileges to other databases is not yet supported")
				}
				tables[i] = doltdb.TableName{
					Name:   string(normalizedTable.ObjectName),
					Schema: string(normalizedTable.SchemaName),
				}
			case *tree.AllTablesSelector:
				return nil, fmt.Errorf("selecting all tables in a schema is not yet supported")
			default:
				return nil, fmt.Errorf(`unexpected table type in GRANT: %T`, normalizedTable)
			}
		}
		privileges, err := convertPrivilegeKinds(auth.PrivilegeObject_TABLE, node.Privileges)
		if err != nil {
			return nil, err
		}
		grantTable = &pgnodes.GrantTable{
			Privileges:         privileges,
			Tables:             tables,
			AllTablesInSchemas: nil,
		}
	default:
		return nil, fmt.Errorf("this form of GRANT is not yet supported")
	}
	return vitess.InjectedStatement{
		Statement: &pgnodes.Grant{
			GrantTable:      grantTable,
			ToRoles:         node.Grantees,
			WithGrantOption: node.WithGrantOption,
			GrantedBy:       node.GrantedBy,
		},
		Children: nil,
	}, nil
}

// convertPrivilegeKind converts a privilege from its parser representation to the server representation.
func convertPrivilegeKinds(object auth.PrivilegeObject, kinds []privilege.Kind) ([]auth.Privilege, error) {
	privileges := make([]auth.Privilege, len(kinds))
	for i, kind := range kinds {
		switch kind {
		case privilege.ALL:
			// If we encounter ALL, then we know to return all privileges for this object
			return object.AllPrivileges(), nil
		case privilege.ALTERSYSTEM:
			privileges[i] = auth.Privilege_ALTER_SYSTEM
		case privilege.CONNECT:
			privileges[i] = auth.Privilege_CONNECT
		case privilege.CREATE:
			privileges[i] = auth.Privilege_CREATE
		case privilege.DELETE:
			privileges[i] = auth.Privilege_DELETE
		case privilege.EXECUTE:
			privileges[i] = auth.Privilege_EXECUTE
		case privilege.INSERT:
			privileges[i] = auth.Privilege_INSERT
		case privilege.REFERENCES:
			privileges[i] = auth.Privilege_REFERENCES
		case privilege.SELECT:
			privileges[i] = auth.Privilege_SELECT
		case privilege.SET:
			privileges[i] = auth.Privilege_SET
		case privilege.TEMPORARY:
			privileges[i] = auth.Privilege_TEMPORARY
		case privilege.TRIGGER:
			privileges[i] = auth.Privilege_TRIGGER
		case privilege.TRUNCATE:
			privileges[i] = auth.Privilege_TRUNCATE
		case privilege.UPDATE:
			privileges[i] = auth.Privilege_UPDATE
		case privilege.USAGE:
			privileges[i] = auth.Privilege_USAGE
		default:
			// This shouldn't be possible unless we update our list of supported privileges
			return nil, fmt.Errorf("unknown privilege kind: %v", kind)
		}
	}
	for _, p := range privileges {
		if !object.IsValid(p) {
			return nil, fmt.Errorf("invalid privilege type %s for relation", p.String())
		}
	}
	return privileges, nil
}
