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

	// Map the parser's object type to the auth package's PrivilegeObject.
	objectType, err := defaultPrivilegeObjectType(node.Target.TargetType)
	if err != nil {
		return nil, err
	}

	// Convert the privilege list.
	privileges, err := convertPrivilegeKinds(objectType, node.Privileges)
	if err != nil {
		return nil, err
	}

	return vitess.InjectedStatement{
		Statement: &pgnodes.AlterDefaultPrivileges{
			ForRoles:        node.TargetRoles,
			Schemas:         node.Target.InSchema,
			ObjectType:      objectType,
			Grantees:        node.Grantees,
			Privileges:      privileges,
			WithGrantOption: node.GrantOption,
			IsGrant:         node.Grant,
			Cascade:         node.DropBehavior == tree.DropCascade,
		},
		Children: nil,
	}, nil
}

// defaultPrivilegeObjectType maps the parser's privilege.ObjectType to the auth.PrivilegeObject
// used in default privilege storage.
func defaultPrivilegeObjectType(ot privilege.ObjectType) (auth.PrivilegeObject, error) {
	switch ot {
	case privilege.Table:
		return auth.PrivilegeObject_TABLE, nil
	case privilege.Sequence:
		return auth.PrivilegeObject_SEQUENCE, nil
	case privilege.Function, privilege.Procedure, privilege.Routine:
		return auth.PrivilegeObject_FUNCTION, nil
	case privilege.Type:
		return auth.PrivilegeObject_TYPE, nil
	case privilege.Schema:
		return auth.PrivilegeObject_SCHEMA, nil
	default:
		return 0, errors.Errorf("unsupported object type for ALTER DEFAULT PRIVILEGES: %s", ot)
	}
}
