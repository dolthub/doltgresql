// Copyright 2024 Dolthub, Inc.
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

import "sync/atomic"

// TODO: this will actually hold whatever needs to be held for proper authentication and role management.
//  For now though, this just exists to test that passwords for users are being stored and retrieved correctly.

var (
	globalMockDatabase MockDatabase
	userIDCounter      atomic.Uint64
)

// MockDatabase is a temporary database to hold role passwords.
type MockDatabase struct {
	rolesByName     map[string]RoleID
	rolesByID       map[RoleID]Role
	tablePrivileges *TablePrivileges
	ownership       *Ownership
}

// ClearDatabase clears the internal database, leaving only the default users. This is primarily for use by tests.
func ClearDatabase() {
	clear(globalMockDatabase.rolesByName)
	clear(globalMockDatabase.rolesByID)
	initDefault()
}

// DropRole removes the given role from the database. If the role does not exist, then this is a no-op.
func DropRole(name string) {
	if roleID, ok := globalMockDatabase.rolesByName[name]; ok {
		delete(globalMockDatabase.rolesByName, name)
		delete(globalMockDatabase.rolesByID, roleID)
	}
}

// GetRole returns the role with the given name. Use RoleExists to determine if the role exists, as this will return a
// role with the default values set if it does not exist.
func GetRole(name string) Role {
	roleID, ok := globalMockDatabase.rolesByName[name]
	if !ok {
		return createDefaultRoleWithoutID(name)
	}
	return globalMockDatabase.rolesByID[roleID]
}

// RenameRole renames the role with the old name to the new name. If the role does not exist, then this is a no-op.
func RenameRole(oldName string, newName string) {
	if roleID, ok := globalMockDatabase.rolesByName[oldName]; ok {
		delete(globalMockDatabase.rolesByName, oldName)
		globalMockDatabase.rolesByName[newName] = roleID
		role := globalMockDatabase.rolesByID[roleID]
		role.Name = newName
		globalMockDatabase.rolesByID[roleID] = role
	}
}

// RoleExists returns whether the given role exists.
func RoleExists(name string) bool {
	_, ok := globalMockDatabase.rolesByName[name]
	return ok
}

// SetRole sets the role matching the given name. This will add a role that does not yet exist, and overwrite an
// existing role.
func SetRole(role Role) {
	// We want to ignore invalid roles, which should not exist outside specific circumstances (like during login)
	if role.id == 0 {
		return
	}
	// TODO: figure something out for concurrency
	if existingRole, ok := globalMockDatabase.rolesByID[role.id]; ok {
		delete(globalMockDatabase.rolesByName, existingRole.Name)
	}
	globalMockDatabase.rolesByName[role.Name] = role.id
	globalMockDatabase.rolesByID[role.ID()] = role
}

// init simply calls initDefault during program initialization.
func init() {
	globalMockDatabase = MockDatabase{
		rolesByName:     make(map[string]RoleID),
		rolesByID:       make(map[RoleID]Role),
		tablePrivileges: NewTablePrivileges(),
	}
	initDefault()
}

// initDefault initializes the mock database and fills it with default users for testing.
func initDefault() {
	var err error
	SetRole(CreateDefaultRole("PUBLIC"))
	doltgres := CreateDefaultRole("doltgres")
	doltgres.CanLogin = true
	doltgres.Password, err = NewScramSha256Password("")
	if err != nil {
		panic(err)
	}
	SetRole(doltgres)
	postgres := CreateDefaultRole("postgres")
	postgres.CanLogin = true
	postgres.Password, err = NewScramSha256Password("password")
	if err != nil {
		panic(err)
	}
	SetRole(postgres)
}
