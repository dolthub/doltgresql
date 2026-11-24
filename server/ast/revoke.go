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
	var revokeSequence *pgnodes.RevokeSequence
	var revokeRoutine *pgnodes.RevokeRoutine
	switch node.Targets.TargetType {
	case privilege.Table:
		tables, err := tablePrivilegeTargets(node.Targets, "REVOKE")
		if err != nil {
			return nil, err
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
	case privilege.Sequence:
		privileges, err := convertPrivilegeKinds(auth.PrivilegeObject_SEQUENCE, node.Privileges)
		if err != nil {
			return nil, err
		}
		revokeSequence = &pgnodes.RevokeSequence{
			Privileges: privileges,
			Sequences:  sequencePrivilegeTargets(node.Targets),
		}
	case privilege.Function, privilege.Procedure, privilege.Routine:
		privileges, err := convertPrivilegeKinds(auth.PrivilegeObject_FUNCTION, node.Privileges)
		if err != nil {
			return nil, err
		}
		revokeRoutine = &pgnodes.RevokeRoutine{
			Privileges: privileges,
			Routines:   routinePrivilegeTargets(node.Targets),
			Kind:       routineKind(node.Targets.TargetType),
		}
	default:
		return nil, errors.Errorf("this form of REVOKE is not yet supported")
	}
	return vitess.InjectedStatement{
		Statement: &pgnodes.Revoke{
			RevokeTable:    revokeTable,
			RevokeSchema:   revokeSchema,
			RevokeDatabase: revokeDatabase,
			RevokeSequence: revokeSequence,
			RevokeRoutine:  revokeRoutine,
			RevokeRole:     nil,
			FromRoles:      node.Grantees,
			GrantedBy:      node.GrantedBy,
			GrantOptionFor: node.GrantOptionFor,
			Cascade:        node.DropBehavior == tree.DropCascade,
		},
		Children: nil,
	}, nil
}
