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

// SchemaPrivileges contains the privileges given to a role on a schema.
type SchemaPrivileges struct {
	Data map[SchemaPrivilegeKey]SchemaPrivilegeValue
}

// SchemaPrivilegeKey points to a specific schema object.
type SchemaPrivilegeKey struct {
	Role   RoleID
	Schema string
}

// SchemaPrivilegeValue is the value associated with the SchemaPrivilegeKey.
type SchemaPrivilegeValue struct {
	Key        SchemaPrivilegeKey
	Privileges map[Privilege]map[GrantedPrivilege]bool
}

// NewSchemaPrivileges returns a new *SchemaPrivileges.
func NewSchemaPrivileges() *SchemaPrivileges {
	return &SchemaPrivileges{make(map[SchemaPrivilegeKey]SchemaPrivilegeValue)}
}

// AddSchemaPrivilege adds the given schema privilege to the global database.
func AddSchemaPrivilege(key SchemaPrivilegeKey, privilege GrantedPrivilege, withGrantOption bool) {
	schemaPrivilegeValue, ok := globalDatabase.schemaPrivileges.Data[key]
	if !ok {
		schemaPrivilegeValue = SchemaPrivilegeValue{
			Key:        key,
			Privileges: make(map[Privilege]map[GrantedPrivilege]bool),
		}
		globalDatabase.schemaPrivileges.Data[key] = schemaPrivilegeValue
	}
	privilegeMap, ok := schemaPrivilegeValue.Privileges[privilege.Privilege]
	if !ok {
		privilegeMap = make(map[GrantedPrivilege]bool)
		schemaPrivilegeValue.Privileges[privilege.Privilege] = privilegeMap
	}
	privilegeMap[privilege] = withGrantOption
}

// HasSchemaPrivilege checks whether the user has the given privilege on the associated schema.
func HasSchemaPrivilege(key SchemaPrivilegeKey, privilege Privilege) bool {
	if IsSuperUser(key.Role) || IsOwner(OwnershipKey{
		PrivilegeObject: PrivilegeObject_SCHEMA,
		Schema:          key.Schema,
	}, key.Role) {
		return true
	}
	if schemaPrivilegeValue, ok := globalDatabase.schemaPrivileges.Data[key]; ok {
		if privilegeMap, ok := schemaPrivilegeValue.Privileges[privilege]; ok && len(privilegeMap) > 0 {
			return true
		}
	}
	for _, group := range GetAllGroupsWithMember(key.Role, true) {
		if HasSchemaPrivilege(SchemaPrivilegeKey{
			Role:   group,
			Schema: key.Schema,
		}, privilege) {
			return true
		}
	}
	return false
}

// HasSchemaPrivilegeGrantOption checks whether the user has WITH GRANT OPTION for the given privilege on the associated
// schema. Returns the role that has WITH GRANT OPTION, or an invalid role if WITH GRANT OPTION is not available.
func HasSchemaPrivilegeGrantOption(key SchemaPrivilegeKey, privilege Privilege) RoleID {
	ownershipKey := OwnershipKey{
		PrivilegeObject: PrivilegeObject_SCHEMA,
		Schema:          key.Schema,
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
	if schemaPrivilegeValue, ok := globalDatabase.schemaPrivileges.Data[key]; ok {
		if privilegeMap, ok := schemaPrivilegeValue.Privileges[privilege]; ok {
			for _, withGrantOption := range privilegeMap {
				if withGrantOption {
					return key.Role
				}
			}
		}
	}
	for _, group := range GetAllGroupsWithMember(key.Role, true) {
		if returnedID := HasSchemaPrivilegeGrantOption(SchemaPrivilegeKey{
			Role:   group,
			Schema: key.Schema,
		}, privilege); returnedID.IsValid() {
			return returnedID
		}
	}
	return 0
}

// RemoveSchemaPrivilege removes the privilege from the global database. If `grantOptionOnly` is true, then only the WITH
// GRANT OPTION portion is revoked. If `grantOptionOnly` is false, then the full privilege is removed. If the GrantedBy
// field contains a valid RoleID, then only the privilege associated with that granter is removed. Otherwise, the
// privilege is completely removed for the grantee.
func RemoveSchemaPrivilege(key SchemaPrivilegeKey, privilege GrantedPrivilege, grantOptionOnly bool) {
	if schemaPrivilegeValue, ok := globalDatabase.schemaPrivileges.Data[key]; ok {
		if privilegeMap, ok := schemaPrivilegeValue.Privileges[privilege.Privilege]; ok {
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
					delete(schemaPrivilegeValue.Privileges, privilege.Privilege)
				}
			}
		}
		if len(schemaPrivilegeValue.Privileges) == 0 {
			delete(globalDatabase.schemaPrivileges.Data, key)
		}
	}
}

// serialize writes the SchemaPrivileges to the given writer.
func (sp *SchemaPrivileges) serialize(writer *utils.Writer) {
	// Version 0
	// Write the total number of values
	writer.Uint64(uint64(len(sp.Data)))
	for _, value := range sp.Data {
		// Write the key
		writer.Uint64(uint64(value.Key.Role))
		writer.String(value.Key.Schema)
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

// deserialize reads the SchemaPrivileges from the given reader.
func (sp *SchemaPrivileges) deserialize(version uint32, reader *utils.Reader) {
	sp.Data = make(map[SchemaPrivilegeKey]SchemaPrivilegeValue)
	switch version {
	case 0:
		// Read the total number of values
		dataCount := reader.Uint64()
		for dataIdx := uint64(0); dataIdx < dataCount; dataIdx++ {
			// Read the key
			spv := SchemaPrivilegeValue{Privileges: make(map[Privilege]map[GrantedPrivilege]bool)}
			spv.Key.Role = RoleID(reader.Uint64())
			spv.Key.Schema = reader.String()
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
		panic("unexpected version in SchemaPrivileges")
	}
}
