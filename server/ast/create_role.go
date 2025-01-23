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
	pgnodes "github.com/dolthub/doltgresql/server/node"
)

// nodeCreateRole handles *tree.CreateRole nodes.
func nodeCreateRole(ctx *Context, node *tree.CreateRole) (vitess.Statement, error) {
	if node == nil {
		return nil, nil
	}
	if len(node.Name) == 0 {
		// The parser should make this impossible, but extra error checking is never bad
		return nil, errors.New(`role name cannot be empty`)
	}
	switch node.Name {
	case `public`:
		return nil, errors.New(`role name "public" is reserved`)
	case `current_role`, `current_user`, `session_user`:
		return nil, errors.Errorf(`%s cannot be used as a role name here`, strings.ToUpper(node.Name))
	}
	createRole := &pgnodes.CreateRole{
		Name:                      node.Name,
		IfNotExists:               node.IfNotExists,
		Password:                  "",
		IsPasswordNull:            true,
		IsSuperUser:               false,
		CanCreateDB:               false,
		CanCreateRoles:            false,
		InheritPrivileges:         true,
		CanLogin:                  !node.IsRole,
		IsReplicationRole:         false,
		CanBypassRowLevelSecurity: false,
		ConnectionLimit:           -1,
		ValidUntil:                "",
		IsValidUntilSet:           false,
		AddToRoles:                nil,
		AddAsMembers:              nil,
		AddAsAdminMembers:         nil,
	}
	for _, kvOption := range node.KVOptions {
		switch strings.ToUpper(string(kvOption.Key)) {
		case "BYPASSRLS":
			createRole.CanBypassRowLevelSecurity = true
		case "CONNECTION_LIMIT":
			switch value := kvOption.Value.(type) {
			case *tree.DInt:
				if value == nil {
					createRole.ConnectionLimit = -1
				} else {
					// We enforce that only int32 values will fit here in the parser
					createRole.ConnectionLimit = int32(*value)
				}
			case tree.NullLiteral:
				createRole.ConnectionLimit = -1
			default:
				return nil, errors.Errorf(`unknown role option value (%T) for option "%s"`, kvOption.Value, kvOption.Key)
			}
		case "CREATEDB":
			createRole.CanCreateDB = true
		case "CREATEROLE":
			createRole.CanCreateRoles = true
		case "INHERIT":
			createRole.InheritPrivileges = true
		case "LOGIN":
			createRole.CanLogin = true
		case "NOBYPASSRLS":
			createRole.CanBypassRowLevelSecurity = false
		case "NOCREATEDB":
			createRole.CanCreateDB = false
		case "NOCREATEROLE":
			createRole.CanCreateRoles = false
		case "NOINHERIT":
			createRole.InheritPrivileges = false
		case "NOLOGIN":
			createRole.CanLogin = false
		case "NOREPLICATION":
			createRole.IsReplicationRole = false
		case "NOSUPERUSER":
			createRole.IsSuperUser = false
		case "PASSWORD":
			switch value := kvOption.Value.(type) {
			case *tree.DString:
				if value == nil {
					createRole.Password = ""
					createRole.IsPasswordNull = true
				} else {
					createRole.Password = string(*value)
					createRole.IsPasswordNull = false
				}
			case tree.NullLiteral:
				createRole.Password = ""
				createRole.IsPasswordNull = true
			default:
				return nil, errors.Errorf(`unknown role option value (%T) for option "%s"`, kvOption.Value, kvOption.Key)
			}
		case "REPLICATION":
			createRole.IsReplicationRole = true
		case "SUPERUSER":
			createRole.IsSuperUser = true
		case "SYSID":
			// This is an option that is ignored by Postgres. Assuming it used to be relevant, but not any longer.
		case "VALID_UNTIL":
			strVal, ok := kvOption.Value.(*tree.DString)
			if !ok {
				return nil, errors.Errorf(`unknown role option value (%T) for option "%s"`, kvOption.Value, kvOption.Key)
			}
			if strVal == nil {
				createRole.ValidUntil = ""
				createRole.IsValidUntilSet = false
			} else {
				createRole.ValidUntil = string(*strVal)
				createRole.IsValidUntilSet = true
			}
		default:
			return nil, errors.Errorf(`unknown role option "%s"`, kvOption.Key)
		}
	}
	return vitess.InjectedStatement{
		Statement: createRole,
		Children:  nil,
	}, nil
}
