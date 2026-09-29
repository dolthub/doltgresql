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

package auth

import (
	"strings"

	"github.com/cockroachdb/errors"

	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
)

// predefinedRoleKind is persisted independently of the role name, so upgrades cannot
// accidentally give a legacy user-created role a built-in capability.
type predefinedRoleKind uint8

const (
	predefinedRoleNone predefinedRoleKind = iota
	predefinedRoleReadAllData
	predefinedRoleWriteAllData
)

var predefinedRoles = []struct {
	kind predefinedRoleKind
	name string
}{
	{predefinedRoleReadAllData, "pg_read_all_data"},
	{predefinedRoleWriteAllData, "pg_write_all_data"},
}

// IsPredefined reports whether this is a registered built-in role.
func (r Role) IsPredefined() bool {
	return r.predefined != predefinedRoleNone
}

// CheckRoleName rejects SQL attempts to use PostgreSQL's reserved role namespace.
func CheckRoleName(name string) error {
	if strings.HasPrefix(name, "pg_") {
		return pgerror.Newf(pgcode.ReservedName, `role name "%s" is reserved`, name)
	}
	return nil
}

// CheckRoleCanBeDropped protects predefined roles even from superusers, matching
// PostgreSQL's dependency error for database-system roles.
func CheckRoleCanBeDropped(role Role) error {
	if role.IsPredefined() {
		return pgerror.Newf(pgcode.DependentObjectsStillExist, "cannot drop role %s because it is required by the database system", role.Name)
	}
	return nil
}

// ensurePredefinedRoles validates existing built-ins and adds missing definitions to
// this database. Callers must first advance userIDCounter past all persisted IDs.
func (db *Database) ensurePredefinedRoles() error {
	for _, role := range db.rolesByID {
		if !role.IsPredefined() {
			continue
		}
		valid := false
		for _, definition := range predefinedRoles {
			if role.predefined == definition.kind && role.Name == definition.name {
				expected := createDefaultRoleWithoutID(definition.name)
				expected.id, expected.predefined = role.id, definition.kind
				valid = role == expected
				break
			}
		}
		if !valid {
			return errors.Errorf(`invalid predefined role "%s" in authorization database`, role.Name)
		}
	}
	// Validate all names before allocating or inserting anything.
	for _, definition := range predefinedRoles {
		if roleID, ok := db.rolesByName[definition.name]; ok && db.rolesByID[roleID].predefined != definition.kind {
			return errors.Errorf(`role "%s" conflicts with a predefined role; rename or remove the existing role using the previous Doltgres version before upgrading`, definition.name)
		}
	}
	for _, definition := range predefinedRoles {
		if _, ok := db.rolesByName[definition.name]; ok {
			continue
		}
		role := CreateDefaultRole(definition.name)
		role.predefined = definition.kind
		db.rolesByName[role.Name] = role.id
		db.rolesByID[role.id] = role
	}
	return nil
}

// hasPredefinedPrivilege evaluates capabilities locally. The ordinary privilege
// checks traverse inheritable memberships, so these rights apply to future objects
// without inserting ACL entries or providing object grant options.
func hasPredefinedPrivilege(roleID RoleID, object PrivilegeObject, privilege Privilege) bool {
	switch globalDatabase.rolesByID[roleID].predefined {
	case predefinedRoleReadAllData:
		return (object == PrivilegeObject_TABLE || object == PrivilegeObject_SEQUENCE) && privilege == Privilege_SELECT ||
			object == PrivilegeObject_SCHEMA && privilege == Privilege_USAGE
	case predefinedRoleWriteAllData:
		return object == PrivilegeObject_TABLE && (privilege == Privilege_INSERT || privilege == Privilege_UPDATE || privilege == Privilege_DELETE) ||
			object == PrivilegeObject_SEQUENCE && privilege == Privilege_UPDATE ||
			object == PrivilegeObject_SCHEMA && privilege == Privilege_USAGE
	}
	return false
}
