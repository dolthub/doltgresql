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
	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"

	"github.com/dolthub/doltgresql/utils"
)

// TablePrivileges contains the privileges given to a role on a table.
type TablePrivileges struct {
	Data map[TablePrivilegeKey]TablePrivilegeValue
}

// TablePrivilegeKey points to a specific table object.
type TablePrivilegeKey struct {
	Role  RoleID
	Table doltdb.TableName
}

// TablePrivilegeValue is the value associated with the TablePrivilegeKey.
type TablePrivilegeValue struct {
	Key        TablePrivilegeKey
	Privileges map[Privilege]map[GrantedPrivilege]bool
}

// NewTablePrivileges returns a new *TablePrivileges.
func NewTablePrivileges() *TablePrivileges {
	return &TablePrivileges{make(map[TablePrivilegeKey]TablePrivilegeValue)}
}

// AddTablePrivilege adds the given table privilege to the global database.
func AddTablePrivilege(key TablePrivilegeKey, privilege GrantedPrivilege, withGrantOption bool) {
	tablePrivilegeValue, ok := globalDatabase.tablePrivileges.Data[key]
	if !ok {
		tablePrivilegeValue = TablePrivilegeValue{
			Key:        key,
			Privileges: make(map[Privilege]map[GrantedPrivilege]bool),
		}
		globalDatabase.tablePrivileges.Data[key] = tablePrivilegeValue
	}
	privilegeMap, ok := tablePrivilegeValue.Privileges[privilege.Privilege]
	if !ok {
		privilegeMap = make(map[GrantedPrivilege]bool)
		tablePrivilegeValue.Privileges[privilege.Privilege] = privilegeMap
	}
	privilegeMap[privilege] = withGrantOption
}

// HasTablePrivilege checks whether the user has the given privilege on the associated table.
func HasTablePrivilege(key TablePrivilegeKey, privilege Privilege) bool {
	if IsSuperUser(key.Role) || IsOwner(OwnershipKey{
		PrivilegeObject: PrivilegeObject_TABLE,
		Schema:          key.Table.Schema,
		Name:            key.Table.Name,
	}, key.Role) {
		return true
	}
	// If a table name was provided, then we also want to search for privileges provided to all tables in the schema
	// space. Since those are saved with an empty table name, we can easily do another search by removing the table.
	if len(key.Table.Name) > 0 {
		if ok := HasTablePrivilege(TablePrivilegeKey{
			Role:  key.Role,
			Table: doltdb.TableName{Name: "", Schema: key.Table.Schema},
		}, privilege); ok {
			return true
		}
	}
	if tablePrivilegeValue, ok := globalDatabase.tablePrivileges.Data[key]; ok {
		if privilegeMap, ok := tablePrivilegeValue.Privileges[privilege]; ok && len(privilegeMap) > 0 {
			return true
		}
	}
	for _, group := range GetAllGroupsWithMember(key.Role, true) {
		if HasTablePrivilege(TablePrivilegeKey{
			Role:  group,
			Table: key.Table,
		}, privilege) {
			return true
		}
	}
	return false
}

// HasTablePrivilegeGrantOption checks whether the user has WITH GRANT OPTION for the given privilege on the associated
// table. Returns the role that has WITH GRANT OPTION, or an invalid role if WITH GRANT OPTION is not available.
func HasTablePrivilegeGrantOption(key TablePrivilegeKey, privilege Privilege) RoleID {
	ownershipKey := OwnershipKey{
		PrivilegeObject: PrivilegeObject_TABLE,
		Schema:          key.Table.Schema,
		Name:            key.Table.Name,
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
	// If a table name was provided, then we also want to search for privileges provided to all tables in the schema
	// space. Since those are saved with an empty table name, we can easily do another search by removing the table.
	if len(key.Table.Name) > 0 {
		if returnedID := HasTablePrivilegeGrantOption(TablePrivilegeKey{
			Role:  key.Role,
			Table: doltdb.TableName{Name: "", Schema: key.Table.Schema},
		}, privilege); returnedID.IsValid() {
			return returnedID
		}
	}
	if tablePrivilegeValue, ok := globalDatabase.tablePrivileges.Data[key]; ok {
		if privilegeMap, ok := tablePrivilegeValue.Privileges[privilege]; ok {
			for _, withGrantOption := range privilegeMap {
				if withGrantOption {
					return key.Role
				}
			}
		}
	}
	for _, group := range GetAllGroupsWithMember(key.Role, true) {
		if returnedID := HasTablePrivilegeGrantOption(TablePrivilegeKey{
			Role:  group,
			Table: key.Table,
		}, privilege); returnedID.IsValid() {
			return returnedID
		}
	}
	return 0
}

// RemoveTablePrivilege removes the privilege from the global database. If `grantOptionOnly` is true, then only the WITH
// GRANT OPTION portion is revoked. If `grantOptionOnly` is false, then the full privilege is removed. If the GrantedBy
// field contains a valid RoleID, then only the privilege associated with that granter is removed. Otherwise, the
// privilege is completely removed for the grantee.
func RemoveTablePrivilege(key TablePrivilegeKey, privilege GrantedPrivilege, grantOptionOnly bool) {
	if tablePrivilegeValue, ok := globalDatabase.tablePrivileges.Data[key]; ok {
		if privilegeMap, ok := tablePrivilegeValue.Privileges[privilege.Privilege]; ok {
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
					delete(tablePrivilegeValue.Privileges, privilege.Privilege)
				}
			}
		}
		if len(tablePrivilegeValue.Privileges) == 0 {
			delete(globalDatabase.tablePrivileges.Data, key)
		}
	}
}

// serialize writes the TablePrivileges to the given writer.
func (tp *TablePrivileges) serialize(writer *utils.Writer) {
	// Version 0
	// Write the total number of values
	writer.Uint64(uint64(len(tp.Data)))
	for _, value := range tp.Data {
		// Write the key
		writer.Uint64(uint64(value.Key.Role))
		writer.String(value.Key.Table.Name)
		writer.String(value.Key.Table.Schema)
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

// deserialize reads the TablePrivileges from the given reader.
func (tp *TablePrivileges) deserialize(version uint32, reader *utils.Reader) {
	tp.Data = make(map[TablePrivilegeKey]TablePrivilegeValue)
	switch version {
	case 0:
		// Read the total number of values
		dataCount := reader.Uint64()
		for dataIdx := uint64(0); dataIdx < dataCount; dataIdx++ {
			// Read the key
			tpv := TablePrivilegeValue{Privileges: make(map[Privilege]map[GrantedPrivilege]bool)}
			tpv.Key.Role = RoleID(reader.Uint64())
			tpv.Key.Table.Name = reader.String()
			tpv.Key.Table.Schema = reader.String()
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
				tpv.Privileges[privilege] = grantedMap
			}
			tp.Data[tpv.Key] = tpv
		}
	default:
		panic("unexpected version in TablePrivileges")
	}
}
