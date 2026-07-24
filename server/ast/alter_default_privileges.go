// Copyright 2026 Dolthub, Inc.
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

	"github.com/dolthub/doltgresql/postgres/parser/privilege"
	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
	"github.com/dolthub/doltgresql/server/auth"
	pgnodes "github.com/dolthub/doltgresql/server/node"
)

// nodeAlterDefaultPrivileges handles *tree.AlterDefaultPrivileges nodes.
func nodeAlterDefaultPrivileges(ctx *Context, node *tree.AlterDefaultPrivileges) (vitess.Statement, error) {
	if node == nil {
		return nil, nil
	}

	objType, err := convertDefaultPrivilegeObjectType(node.Target.TargetType)
	if err != nil {
		return nil, err
	}

	privileges, err := convertPrivilegeKinds(objType, node.Privileges)
	if err != nil {
		return nil, err
	}

	return vitess.InjectedStatement{
		Auth: vitess.AuthInformation{
			AuthType:    auth.AuthType_CREATE,
			TargetType:  auth.AuthTargetType_AlterDefaultPrivilegesIdentifiers,
			TargetNames: []string{node.TargetRole},
		},
		Statement: &pgnodes.AlterDefaultPrivileges{
			OwnerRole:   node.TargetRole,
			Schemas:     node.Target.InSchema,
			ObjectType:  objType,
			Privileges:  privileges,
			Grantees:    node.Grantees,
			Grant:       node.Grant,
			GrantOption: node.GrantOption,
			Cascade:     node.DropBehavior == tree.DropCascade,
		},
		Children: nil,
	}, nil
}

// convertDefaultPrivilegeObjectType converts a privilege.ObjectType to an auth.PrivilegeObject for use in default
// privileges. Only the object types valid for ALTER DEFAULT PRIVILEGES are accepted.
func convertDefaultPrivilegeObjectType(objType privilege.ObjectType) (auth.PrivilegeObject, error) {
	switch objType {
	case privilege.Table:
		return auth.PrivilegeObject_TABLE, nil
	case privilege.Sequence:
		return auth.PrivilegeObject_SEQUENCE, nil
	case privilege.Function, privilege.Procedure, privilege.Routine:
		return auth.PrivilegeObject_FUNCTION, nil
	case privilege.Schema:
		return auth.PrivilegeObject_SCHEMA, nil
	case privilege.Type:
		return auth.PrivilegeObject_TYPE, nil
	default:
		return 0, errors.Errorf("object type %q is not supported in ALTER DEFAULT PRIVILEGES", string(objType))
	}
}
