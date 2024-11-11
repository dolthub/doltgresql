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

import (
	"os"
	"sync"
	"sync/atomic"

	"github.com/dolthub/dolt/go/libraries/doltcore/env"
	"github.com/dolthub/dolt/go/libraries/utils/filesys"
)

// authFileName is the name of the file that contains all authorization-related data.
const authFileName = "auth.db"

var (
	globalDatabase Database
	globalLock     *sync.RWMutex
	userIDCounter  atomic.Uint64
	fileSystem     filesys.Filesys
)

// Database contains all information pertaining to authorization and privileges. This is a global structure that is
// shared between all branches.
type Database struct {
	rolesByName     map[string]RoleID
	rolesByID       map[RoleID]Role
	ownership       *Ownership
	tablePrivileges *TablePrivileges
}

// ClearDatabase clears the internal database, leaving only the default users. This is primarily for use by tests.
func ClearDatabase() {
	clear(globalDatabase.rolesByName)
	clear(globalDatabase.rolesByID)
	dbInitDefault()
}

// DropRole removes the given role from the database. If the role does not exist, then this is a no-op.
func DropRole(name string) {
	if roleID, ok := globalDatabase.rolesByName[name]; ok {
		delete(globalDatabase.rolesByName, name)
		delete(globalDatabase.rolesByID, roleID)

	}
}

// GetRole returns the role with the given name. Use RoleExists to determine if the role exists, as this will return a
// role with the default values set if it does not exist.
func GetRole(name string) Role {
	roleID, ok := globalDatabase.rolesByName[name]
	if !ok {
		return createDefaultRoleWithoutID(name)
	}
	return globalDatabase.rolesByID[roleID]
}

// RenameRole renames the role with the old name to the new name. If the role does not exist, then this is a no-op.
func RenameRole(oldName string, newName string) {
	if roleID, ok := globalDatabase.rolesByName[oldName]; ok {
		delete(globalDatabase.rolesByName, oldName)
		globalDatabase.rolesByName[newName] = roleID
		role := globalDatabase.rolesByID[roleID]
		role.Name = newName
		globalDatabase.rolesByID[roleID] = role
	}
}

// RoleExists returns whether the given role exists.
func RoleExists(name string) bool {
	_, ok := globalDatabase.rolesByName[name]
	return ok
}

// SetRole sets the role matching the given name. This will add a role that does not yet exist, and overwrite an
// existing role.
func SetRole(role Role) {
	// We want to ignore invalid roles, which should not exist outside specific circumstances (like during login)
	if role.id == 0 {
		return
	}
	if existingRole, ok := globalDatabase.rolesByID[role.id]; ok {
		delete(globalDatabase.rolesByName, existingRole.Name)
	}
	globalDatabase.rolesByName[role.Name] = role.id
	globalDatabase.rolesByID[role.ID()] = role
}

// LockRead takes an anonymous function and runs it while using a read lock. This ensures that the lock is automatically
// released once the function finishes.
func LockRead(f func()) {
	globalLock.RLock()
	defer globalLock.RUnlock()
	f()
}

// LockWrite takes an anonymous function and runs it while using a write lock. This ensures that the lock is
// automatically released once the function finishes.
func LockWrite(f func()) {
	globalLock.Lock()
	defer globalLock.Unlock()
	f()
}

// dbInit handle the global database initialization. Panics if an error occurs, since it points to something going
// terribly wrong.
func dbInit(dEnv *env.DoltEnv) {
	globalDatabase = Database{
		rolesByName:     make(map[string]RoleID),
		rolesByID:       make(map[RoleID]Role),
		ownership:       NewOwnership(),
		tablePrivileges: NewTablePrivileges(),
	}
	globalLock = &sync.RWMutex{}
	if dEnv != nil {
		if _, ok := dEnv.FS.(*filesys.InMemFS); !ok {
			fileSystem = dEnv.FS
			authData, err := fileSystem.ReadFile(authFileName)
			if os.IsNotExist(err) {
				dbInitDefault()
				if err = fileSystem.WriteFile(authFileName, globalDatabase.serialize(), 0644); err != nil {
					panic(err)
				}
			} else if err != nil {
				panic(err)
			} else if err = globalDatabase.deserialize(authData); err != nil {
				panic(err)
			}
		} else {
			dbInitDefault()
		}
	} else {
		dbInitDefault()
	}
}

// dbInitDefault initializes the database and fills it with default users for testing.
func dbInitDefault() {
	var err error
	public := CreateDefaultRole("public")
	SetRole(public)
	doltgres := CreateDefaultRole("doltgres")
	doltgres.IsSuperUser = true
	doltgres.CanCreateRoles = true
	doltgres.CanCreateDB = true
	doltgres.CanLogin = true
	doltgres.Password, err = NewScramSha256Password("")
	if err != nil {
		panic(err)
	}
	SetRole(doltgres)
	postgres := CreateDefaultRole("postgres")
	postgres.IsSuperUser = true
	postgres.CanCreateRoles = true
	postgres.CanCreateDB = true
	postgres.CanLogin = true
	postgres.Password, err = NewScramSha256Password("password")
	if err != nil {
		panic(err)
	}
	SetRole(postgres)
}
