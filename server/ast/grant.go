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
	var grantSchema *pgnodes.GrantSchema
	var grantDatabase *pgnodes.GrantDatabase
	var grantSequence *pgnodes.GrantSequence
	var grantRoutine *pgnodes.GrantRoutine
	switch node.Targets.TargetType {
	case privilege.Table:
		tables, err := tablePrivilegeTargets(node.Targets, "GRANT")
		if err != nil {
			return nil, err
		}
		privileges, err := convertPrivilegeKinds(auth.PrivilegeObject_TABLE, node.Privileges)
		if err != nil {
			return nil, err
		}
		grantTable = &pgnodes.GrantTable{
			Privileges: privileges,
			Tables:     tables,
		}
	case privilege.Schema:
		privileges, err := convertPrivilegeKinds(auth.PrivilegeObject_SCHEMA, node.Privileges)
		if err != nil {
			return nil, err
		}
		grantSchema = &pgnodes.GrantSchema{
			Privileges: privileges,
			Schemas:    node.Targets.Names,
		}
	case privilege.Database:
		privileges, err := convertPrivilegeKinds(auth.PrivilegeObject_DATABASE, node.Privileges)
		if err != nil {
			return nil, err
		}
		grantDatabase = &pgnodes.GrantDatabase{
			Privileges: privileges,
			Databases:  node.Targets.Databases.ToStrings(),
		}
	case privilege.Sequence:
		privileges, err := convertPrivilegeKinds(auth.PrivilegeObject_SEQUENCE, node.Privileges)
		if err != nil {
			return nil, err
		}
		grantSequence = &pgnodes.GrantSequence{
			Privileges: privileges,
			Sequences:  sequencePrivilegeTargets(node.Targets),
		}
	case privilege.Function, privilege.Procedure, privilege.Routine:
		privileges, err := convertPrivilegeKinds(auth.PrivilegeObject_FUNCTION, node.Privileges)
		if err != nil {
			return nil, err
		}
		grantRoutine = &pgnodes.GrantRoutine{
			Privileges: privileges,
			Routines:   routinePrivilegeTargets(node.Targets),
		}
	default:
		return nil, errors.Errorf("this form of GRANT is not yet supported")
	}
	return vitess.InjectedStatement{
		Statement: &pgnodes.Grant{
			GrantTable:      grantTable,
			GrantSchema:     grantSchema,
			GrantDatabase:   grantDatabase,
			GrantSequence:   grantSequence,
			GrantRoutine:    grantRoutine,
			GrantRole:       nil,
			ToRoles:         node.Grantees,
			WithGrantOption: node.WithGrantOption,
			GrantedBy:       node.GrantedBy,
		},
		Children: nil,
	}, nil
}

// tablePrivilegeTargets returns the tables named by a GRANT or REVOKE ON TABLE statement.
func tablePrivilegeTargets(targets tree.TargetList, statement string) ([]pgnodes.PrivilegeTarget, error) {
	tables := make([]pgnodes.PrivilegeTarget, 0, len(targets.Tables)+len(targets.InSchema))
	for _, table := range targets.Tables {
		normalizedTable, err := table.NormalizeTablePattern()
		if err != nil {
			return nil, err
		}
		switch normalizedTable := normalizedTable.(type) {
		case *tree.TableName:
			tables = append(tables, pgnodes.PrivilegeTarget{
				Database: string(normalizedTable.CatalogName),
				Schema:   string(normalizedTable.SchemaName),
				Name:     string(normalizedTable.ObjectName),
			})
		case *tree.AllTablesSelector:
			tables = append(tables, pgnodes.PrivilegeTarget{
				Database: string(normalizedTable.CatalogName),
				Schema:   string(normalizedTable.SchemaName),
				Name:     "",
			})
		default:
			return nil, errors.Errorf(`unexpected table type in %s: %T`, statement, normalizedTable)
		}
	}
	for _, schema := range targets.InSchema {
		tables = append(tables, pgnodes.PrivilegeTarget{
			Schema: schema,
			Name:   "",
		})
	}
	return tables, nil
}

// sequencePrivilegeTargets returns the sequences named by a GRANT or REVOKE ON SEQUENCE statement.
func sequencePrivilegeTargets(targets tree.TargetList) []pgnodes.PrivilegeTarget {
	sequences := make([]pgnodes.PrivilegeTarget, 0, len(targets.Sequences)+len(targets.InSchema))
	for _, seq := range targets.Sequences {
		sequences = append(sequences, objectPrivilegeTarget(seq))
	}
	for _, schema := range targets.InSchema {
		sequences = append(sequences, pgnodes.PrivilegeTarget{
			Schema: schema,
			Name:   "",
		})
	}
	return sequences
}

// routinePrivilegeTargets returns the routines named by a GRANT or REVOKE ON FUNCTION/PROCEDURE/ROUTINE statement.
func routinePrivilegeTargets(targets tree.TargetList) []pgnodes.PrivilegeTarget {
	routines := make([]pgnodes.PrivilegeTarget, 0, len(targets.Routines)+len(targets.InSchema))
	for _, r := range targets.Routines {
		routine := objectPrivilegeTarget(r.Name)
		// TODO: there can be 2 routines with the same name but different argument types
		//  need a fix for getting argument types from parsing CALL statement, which would set routine.ArgTypes
		routine.Args = routineArgTypesKey(r.Args)
		routines = append(routines, routine)
	}
	for _, schema := range targets.InSchema {
		routines = append(routines, pgnodes.PrivilegeTarget{
			Schema: schema,
			Name:   "",
		})
	}
	return routines
}

// objectPrivilegeTarget returns the privilege target for the given sequence or routine name.
func objectPrivilegeTarget(name *tree.UnresolvedObjectName) pgnodes.PrivilegeTarget {
	target := pgnodes.PrivilegeTarget{Name: name.Parts[0]}
	if name.NumParts >= 2 {
		target.Schema = name.Parts[1]
	}
	if name.NumParts >= 3 {
		target.Database = name.Parts[2]
	}
	return target
}

// routineArgTypesKey builds a canonical string key from a RoutineArgs list using only the argument types.
func routineArgTypesKey(args tree.RoutineArgs) string {
	parts := make([]string, len(args))
	for i, arg := range args {
		parts[i] = arg.Type.SQLString()
	}
	return strings.Join(parts, ",")
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
		case privilege.DROP:
			privileges[i] = auth.Privilege_DROP
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
			return nil, errors.Errorf("unknown privilege kind: %v", kind)
		}
	}
	for _, p := range privileges {
		if !object.IsValid(p) {
			return nil, errors.Errorf("invalid privilege type %s for relation", p.String())
		}
	}
	return privileges, nil
}
