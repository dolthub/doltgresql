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
	"github.com/dolthub/doltgresql/utils"
)

// RoutinePrivileges contains the privileges given to a role on a routine (function or procedure).
type RoutinePrivileges struct {
	Data map[RoutinePrivilegeKey]RoutinePrivilegeValue
}

// RoutinePrivilegeKey points to a specific routine object. An empty Name represents all routines in the schema.
// ArgTypes is a comma-separated list of argument type SQL strings, e.g. "integer,text".
type RoutinePrivilegeKey struct {
	Role     RoleID
	Schema   string
	Name     string
	ArgTypes string
}

// RoutinePrivilegeValue is the value associated with the RoutinePrivilegeKey.
type RoutinePrivilegeValue struct {
	Key        RoutinePrivilegeKey
	Privileges map[Privilege]map[GrantedPrivilege]bool
}

// NewRoutinePrivileges returns a new *RoutinePrivileges.
func NewRoutinePrivileges() *RoutinePrivileges {
	return &RoutinePrivileges{make(map[RoutinePrivilegeKey]RoutinePrivilegeValue)}
}

// AddRoutinePrivilege adds the given routine privilege to the global database.
func AddRoutinePrivilege(key RoutinePrivilegeKey, privilege GrantedPrivilege, withGrantOption bool) {
	routinePrivilegeValue, ok := globalDatabase.routinePrivileges.Data[key]
	if !ok {
		routinePrivilegeValue = RoutinePrivilegeValue{
			Key:        key,
			Privileges: make(map[Privilege]map[GrantedPrivilege]bool),
		}
		globalDatabase.routinePrivileges.Data[key] = routinePrivilegeValue
	}
	privilegeMap, ok := routinePrivilegeValue.Privileges[privilege.Privilege]
	if !ok {
		privilegeMap = make(map[GrantedPrivilege]bool)
		routinePrivilegeValue.Privileges[privilege.Privilege] = privilegeMap
	}
	privilegeMap[privilege] = withGrantOption
}

// HasRoutinePrivilege checks whether the user has the given privilege on the associated routine.
func HasRoutinePrivilege(key RoutinePrivilegeKey, privilege Privilege) bool {
	if IsSuperUser(key.Role) {
		return true
	}
	// If a routine name was provided, also check for privileges on all routines in the schema.
	if len(key.Name) > 0 {
		if HasRoutinePrivilege(RoutinePrivilegeKey{
			Role:   key.Role,
			Schema: key.Schema,
			Name:   "",
		}, privilege) {
			return true
		}
	}
	if routinePrivilegeValue, ok := globalDatabase.routinePrivileges.Data[key]; ok {
		if privilegeMap, ok := routinePrivilegeValue.Privileges[privilege]; ok && len(privilegeMap) > 0 {
			return true
		}
	}
	for _, group := range GetAllGroupsWithMember(key.Role, true) {
		if HasRoutinePrivilege(RoutinePrivilegeKey{
			Role:     group,
			Schema:   key.Schema,
			Name:     key.Name,
			ArgTypes: key.ArgTypes,
		}, privilege) {
			return true
		}
	}
	return false
}

// HasRoutinePrivilegeGrantOption checks whether the user has WITH GRANT OPTION for the given privilege on the
// associated routine. Returns the role that has WITH GRANT OPTION, or an invalid role if not available.
func HasRoutinePrivilegeGrantOption(key RoutinePrivilegeKey, privilege Privilege) RoleID {
	if IsSuperUser(key.Role) {
		return key.Role
	}
	// If a routine name was provided, also check for grant option on all routines in the schema.
	if len(key.Name) > 0 {
		if returnedID := HasRoutinePrivilegeGrantOption(RoutinePrivilegeKey{
			Role:   key.Role,
			Schema: key.Schema,
			Name:   "",
		}, privilege); returnedID.IsValid() {
			return returnedID
		}
	}
	if routinePrivilegeValue, ok := globalDatabase.routinePrivileges.Data[key]; ok {
		if privilegeMap, ok := routinePrivilegeValue.Privileges[privilege]; ok {
			for _, withGrantOption := range privilegeMap {
				if withGrantOption {
					return key.Role
				}
			}
		}
	}
	for _, group := range GetAllGroupsWithMember(key.Role, true) {
		if returnedID := HasRoutinePrivilegeGrantOption(RoutinePrivilegeKey{
			Role:     group,
			Schema:   key.Schema,
			Name:     key.Name,
			ArgTypes: key.ArgTypes,
		}, privilege); returnedID.IsValid() {
			return returnedID
		}
	}
	return 0
}

// RemoveRoutinePrivilege removes the privilege from the global database. If `grantOptionOnly` is true, then only the
// WITH GRANT OPTION portion is revoked. If `grantOptionOnly` is false, then the full privilege is removed.
func RemoveRoutinePrivilege(key RoutinePrivilegeKey, privilege GrantedPrivilege, grantOptionOnly bool) {
	if routinePrivilegeValue, ok := globalDatabase.routinePrivileges.Data[key]; ok {
		if privilegeMap, ok := routinePrivilegeValue.Privileges[privilege.Privilege]; ok {
			if grantOptionOnly {
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
				if privilege.GrantedBy.IsValid() {
					delete(privilegeMap, privilege)
				} else {
					privilegeMap = nil
				}
				if len(privilegeMap) == 0 {
					delete(routinePrivilegeValue.Privileges, privilege.Privilege)
				}
			}
		}
		if len(routinePrivilegeValue.Privileges) == 0 {
			delete(globalDatabase.routinePrivileges.Data, key)
		}
	}
}

// serialize writes the RoutinePrivileges to the given writer.
func (rp *RoutinePrivileges) serialize(writer *utils.Writer) {
	// Version 0
	// Write the total number of values
	writer.Uint64(uint64(len(rp.Data)))
	for _, value := range rp.Data {
		writer.Uint64(uint64(value.Key.Role))
		writer.String(value.Key.Schema)
		writer.String(value.Key.Name)
		writer.String(value.Key.ArgTypes)
		// Write the total number of privileges
		writer.Uint64(uint64(len(value.Privileges)))
		for privilege, privilegeMap := range value.Privileges {
			writer.String(string(privilege))
			writer.Uint32(uint32(len(privilegeMap)))
			for grantedPrivilege, withGrantOption := range privilegeMap {
				writer.Uint64(uint64(grantedPrivilege.GrantedBy))
				writer.Bool(withGrantOption)
			}
		}
	}
}

// deserialize reads the RoutinePrivileges from the given reader.
func (rp *RoutinePrivileges) deserialize(version uint32, reader *utils.Reader) {
	rp.Data = make(map[RoutinePrivilegeKey]RoutinePrivilegeValue)
	switch version {
	case 2:
		// Read the total number of values
		dataCount := reader.Uint64()
		for dataIdx := uint64(0); dataIdx < dataCount; dataIdx++ {
			rpv := RoutinePrivilegeValue{Privileges: make(map[Privilege]map[GrantedPrivilege]bool)}
			rpv.Key.Role = RoleID(reader.Uint64())
			rpv.Key.Schema = reader.String()
			rpv.Key.Name = reader.String()
			rpv.Key.ArgTypes = reader.String()
			// Read the total number of privileges
			privilegeCount := reader.Uint64()
			for privilegeIdx := uint64(0); privilegeIdx < privilegeCount; privilegeIdx++ {
				privilege := Privilege(reader.String())
				grantedCount := reader.Uint32()
				grantedMap := make(map[GrantedPrivilege]bool)
				for grantedIdx := uint32(0); grantedIdx < grantedCount; grantedIdx++ {
					grantedPrivilege := GrantedPrivilege{}
					grantedPrivilege.Privilege = privilege
					grantedPrivilege.GrantedBy = RoleID(reader.Uint64())
					grantedMap[grantedPrivilege] = reader.Bool()
				}
				rpv.Privileges[privilege] = grantedMap
			}
			rp.Data[rpv.Key] = rpv
		}
	default:
		panic("unexpected version in RoutinePrivileges")
	}
}
