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
	"github.com/dolthub/doltgresql/utils"
)

// DatabasePrivileges contains the privileges given to a role on a database.
type DatabasePrivileges struct {
	Data map[DatabasePrivilegeKey]DatabasePrivilegeValue
}

// DatabasePrivilegeKey points to a specific database object.
type DatabasePrivilegeKey struct {
	Role RoleID
	Name string
}

// DatabasePrivilegeValue is the value associated with the DatabasePrivilegeKey.
type DatabasePrivilegeValue struct {
	Key        DatabasePrivilegeKey
	Privileges map[Privilege]map[GrantedPrivilege]bool
}

// NewDatabasePrivileges returns a new *DatabasePrivileges.
func NewDatabasePrivileges() *DatabasePrivileges {
	return &DatabasePrivileges{make(map[DatabasePrivilegeKey]DatabasePrivilegeValue)}
}

// AddDatabasePrivilege adds the given database privilege to the global database.
func AddDatabasePrivilege(key DatabasePrivilegeKey, privilege GrantedPrivilege, withGrantOption bool) {
	databasePrivilegeValue, ok := globalDatabase.databasePrivileges.Data[key]
	if !ok {
		databasePrivilegeValue = DatabasePrivilegeValue{
			Key:        key,
			Privileges: make(map[Privilege]map[GrantedPrivilege]bool),
		}
		globalDatabase.databasePrivileges.Data[key] = databasePrivilegeValue
	}
	privilegeMap, ok := databasePrivilegeValue.Privileges[privilege.Privilege]
	if !ok {
		privilegeMap = make(map[GrantedPrivilege]bool)
		databasePrivilegeValue.Privileges[privilege.Privilege] = privilegeMap
	}
	privilegeMap[privilege] = withGrantOption
}

// HasDatabasePrivilege checks whether the user has the given privilege on the associated database.
func HasDatabasePrivilege(key DatabasePrivilegeKey, privilege Privilege) bool {
	if IsSuperUser(key.Role) || IsOwner(OwnershipKey{
		PrivilegeObject: PrivilegeObject_DATABASE,
		Name:            key.Name,
	}, key.Role) {
		return true
	}
	if databasePrivilegeValue, ok := globalDatabase.databasePrivileges.Data[key]; ok {
		if privilegeMap, ok := databasePrivilegeValue.Privileges[privilege]; ok && len(privilegeMap) > 0 {
			return true
		}
	}
	for _, group := range GetAllGroupsWithMember(key.Role, true) {
		if HasDatabasePrivilege(DatabasePrivilegeKey{
			Role: group,
			Name: key.Name,
		}, privilege) {
			return true
		}
	}
	return false
}

// HasDatabasePrivilegeGrantOption checks whether the user has WITH GRANT OPTION for the given privilege on the associated
// database. Returns the role that has WITH GRANT OPTION, or an invalid role if WITH GRANT OPTION is not available.
func HasDatabasePrivilegeGrantOption(key DatabasePrivilegeKey, privilege Privilege) RoleID {
	ownershipKey := OwnershipKey{
		PrivilegeObject: PrivilegeObject_DATABASE,
		Name:            key.Name,
	}
	if IsSuperUser(key.Role) {
		owners := GetOwners(ownershipKey)
		if len(owners) == 0 {
			// This may happen if the privilege file is deleted
			return key.Role
		}
		// Although there may be multiple owners, we'll only return the first one.
		// Postgres already allows for non-determinism with multiple membership paths, so this is fine.
		return owners[0]
	} else if IsOwner(ownershipKey, key.Role) {
		return key.Role
	}
	if databasePrivilegeValue, ok := globalDatabase.databasePrivileges.Data[key]; ok {
		if privilegeMap, ok := databasePrivilegeValue.Privileges[privilege]; ok {
			for _, withGrantOption := range privilegeMap {
				if withGrantOption {
					return key.Role
				}
			}
		}
	}
	for _, group := range GetAllGroupsWithMember(key.Role, true) {
		if returnedID := HasDatabasePrivilegeGrantOption(DatabasePrivilegeKey{
			Role: group,
			Name: key.Name,
		}, privilege); returnedID.IsValid() {
			return returnedID
		}
	}
	return 0
}

// RemoveDatabasePrivilege removes the privilege from the global database. If `grantOptionOnly` is true, then only the WITH
// GRANT OPTION portion is revoked. If `grantOptionOnly` is false, then the full privilege is removed. If the GrantedBy
// field contains a valid RoleID, then only the privilege associated with that granter is removed. Otherwise, the
// privilege is completely removed for the grantee.
func RemoveDatabasePrivilege(key DatabasePrivilegeKey, privilege GrantedPrivilege, grantOptionOnly bool) {
	if databasePrivilegeValue, ok := globalDatabase.databasePrivileges.Data[key]; ok {
		if privilegeMap, ok := databasePrivilegeValue.Privileges[privilege.Privilege]; ok {
			if grantOptionOnly {
				// This is provided when we only want to revoke the WITH GRANT OPTION, and not the privilege itself.
				// If a role is provided in GRANTED BY, then we specifically delete the option associated with that role.
				// If no role was given, then we'll remove WITH GRANT OPTION from all of the associated roles.
				if privilege.GrantedBy.IsValid() {
					if _, ok = privilegeMap[privilege]; ok {
						privilegeMap[privilege] = false
					}
				} else {
					for privilegeMapKey := range privilegeMap {
						privilegeMap[privilegeMapKey] = false
					}
				}
			} else {
				// If a role is provided in GRANTED BY, then we specifically delete the privilege associated with that role.
				// If no role was given, then we'll delete the privileges granted by all roles.
				if privilege.GrantedBy.IsValid() {
					delete(privilegeMap, privilege)
				} else {
					privilegeMap = nil
				}
				if len(privilegeMap) == 0 {
					delete(databasePrivilegeValue.Privileges, privilege.Privilege)
				}
			}
		}
		if len(databasePrivilegeValue.Privileges) == 0 {
			delete(globalDatabase.databasePrivileges.Data, key)
		}
	}
}

// serialize writes the DatabasePrivileges to the given writer.
func (sp *DatabasePrivileges) serialize(writer *utils.Writer) {
	// Version 0
	// Write the total number of values
	writer.Uint64(uint64(len(sp.Data)))
	for _, value := range sp.Data {
		// Write the key
		writer.Uint64(uint64(value.Key.Role))
		writer.String(value.Key.Name)
		// Write the total number of privileges
		writer.Uint64(uint64(len(value.Privileges)))
		for privilege, privilegeMap := range value.Privileges {
			writer.String(string(privilege))
			// Write the number of granted privileges
			writer.Uint32(uint32(len(privilegeMap)))
			for grantedPrivilege, withGrantOption := range privilegeMap {
				writer.Uint64(uint64(grantedPrivilege.GrantedBy))
				writer.Bool(withGrantOption)
			}
		}
	}
}

// deserialize reads the DatabasePrivileges from the given reader.
func (sp *DatabasePrivileges) deserialize(version uint32, reader *utils.Reader) {
	sp.Data = make(map[DatabasePrivilegeKey]DatabasePrivilegeValue)
	switch version {
	case 0:
		// Read the total number of values
		dataCount := reader.Uint64()
		for dataIdx := uint64(0); dataIdx < dataCount; dataIdx++ {
			// Read the key
			spv := DatabasePrivilegeValue{Privileges: make(map[Privilege]map[GrantedPrivilege]bool)}
			spv.Key.Role = RoleID(reader.Uint64())
			spv.Key.Name = reader.String()
			// Read the total number of privileges
			privilegeCount := reader.Uint64()
			for privilegeIdx := uint64(0); privilegeIdx < privilegeCount; privilegeIdx++ {
				privilege := Privilege(reader.String())
				// Read the number of granted privileges
				grantedCount := reader.Uint32()
				grantedMap := make(map[GrantedPrivilege]bool)
				for grantedIdx := uint32(0); grantedIdx < grantedCount; grantedIdx++ {
					grantedPrivilege := GrantedPrivilege{}
					grantedPrivilege.Privilege = privilege
					grantedPrivilege.GrantedBy = RoleID(reader.Uint64())
					grantedMap[grantedPrivilege] = reader.Bool()
				}
				spv.Privileges[privilege] = grantedMap
			}
			sp.Data[spv.Key] = spv
		}
	default:
		panic("unexpected version in DatabasePrivileges")
	}
}
