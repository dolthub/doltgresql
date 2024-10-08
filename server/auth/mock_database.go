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

// TODO: this will actually hold whatever needs to be held for proper authentication and role management.
//  For now though, this just exists to test that passwords for users are being stored and retrieved correctly.

var (
	globalMockDatabase = MockDatabase{make(map[string]Role)}
)

// MockDatabase is a temporary database to hold role passwords.
type MockDatabase struct {
	Roles map[string]Role
}

// ClearDatabase clears the internal database, leaving only the default users. This is primarily for use by tests.
func ClearDatabase() {
	clear(globalMockDatabase.Roles)
	initDefault()
}

// DropRole removes the given role from the database. If the role does not exist, then this is a no-op.
func DropRole(name string) {
	delete(globalMockDatabase.Roles, name)
}

// GetRole returns the role with the given name. Use RoleExists to determine if the role exists, as this will return a
// role with the default values set if it does not exist.
func GetRole(name string) Role {
	role, ok := globalMockDatabase.Roles[name]
	if !ok {
		return CreateDefaultRole(name)
	}
	return role
}

// RoleExists returns whether the given role exists.
func RoleExists(name string) bool {
	_, ok := globalMockDatabase.Roles[name]
	return ok
}

// SetRole sets the role matching the given name. This will add a role that does not yet exist, and overwrite an
// existing role.
func SetRole(role Role) {
	// TODO: figure something out for concurrency
	globalMockDatabase.Roles[role.Name] = role
}

// init simply calls initDefault during program initialization.
func init() {
	initDefault()
}

// initDefault initializes the mock database and fills it with default users for testing.
func initDefault() {
	var err error
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
