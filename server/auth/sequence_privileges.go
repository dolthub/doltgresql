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

// SequencePrivileges contains the privileges given to a role on a sequence.
type SequencePrivileges struct {
	Data map[SequencePrivilegeKey]SequencePrivilegeValue
}

// SequencePrivilegeKey points to a specific sequence object. An empty Name represents all sequences in the schema.
type SequencePrivilegeKey struct {
	Role   RoleID
	Schema string
	Name   string
}

// SequencePrivilegeValue is the value associated with the SequencePrivilegeKey.
type SequencePrivilegeValue struct {
	Key        SequencePrivilegeKey
	Privileges map[Privilege]map[GrantedPrivilege]bool
}

// NewSequencePrivileges returns a new *SequencePrivileges.
func NewSequencePrivileges() *SequencePrivileges {
	return &SequencePrivileges{make(map[SequencePrivilegeKey]SequencePrivilegeValue)}
}

// AddSequencePrivilege adds the given sequence privilege to the global database.
func AddSequencePrivilege(key SequencePrivilegeKey, privilege GrantedPrivilege, withGrantOption bool) {
	seqPrivilegeValue, ok := globalDatabase.sequencePrivileges.Data[key]
	if !ok {
		seqPrivilegeValue = SequencePrivilegeValue{
			Key:        key,
			Privileges: make(map[Privilege]map[GrantedPrivilege]bool),
		}
		globalDatabase.sequencePrivileges.Data[key] = seqPrivilegeValue
	}
	privilegeMap, ok := seqPrivilegeValue.Privileges[privilege.Privilege]
	if !ok {
		privilegeMap = make(map[GrantedPrivilege]bool)
		seqPrivilegeValue.Privileges[privilege.Privilege] = privilegeMap
	}
	privilegeMap[privilege] = withGrantOption
}

// HasSequencePrivilege checks whether the user has the given privilege on the associated sequence.
func HasSequencePrivilege(key SequencePrivilegeKey, privilege Privilege) bool {
	if IsSuperUser(key.Role) {
		return true
	}
	// If a sequence name was provided, also check for privileges on all sequences in the schema.
	if len(key.Name) > 0 {
		if HasSequencePrivilege(SequencePrivilegeKey{
			Role:   key.Role,
			Schema: key.Schema,
			Name:   "",
		}, privilege) {
			return true
		}
	}
	if seqPrivilegeValue, ok := globalDatabase.sequencePrivileges.Data[key]; ok {
		if privilegeMap, ok := seqPrivilegeValue.Privileges[privilege]; ok && len(privilegeMap) > 0 {
			return true
		}
	}
	for _, group := range GetAllGroupsWithMember(key.Role, true) {
		if HasSequencePrivilege(SequencePrivilegeKey{
			Role:   group,
			Schema: key.Schema,
			Name:   key.Name,
		}, privilege) {
			return true
		}
	}
	return false
}

// HasSequencePrivilegeGrantOption checks whether the user has WITH GRANT OPTION for the given privilege on the
// associated sequence. Returns the role that has WITH GRANT OPTION, or an invalid role if not available.
func HasSequencePrivilegeGrantOption(key SequencePrivilegeKey, privilege Privilege) RoleID {
	if IsSuperUser(key.Role) {
		return key.Role
	}
	// If a sequence name was provided, also check for grant option on all sequences in the schema.
	if len(key.Name) > 0 {
		if returnedID := HasSequencePrivilegeGrantOption(SequencePrivilegeKey{
			Role:   key.Role,
			Schema: key.Schema,
			Name:   "",
		}, privilege); returnedID.IsValid() {
			return returnedID
		}
	}
	if seqPrivilegeValue, ok := globalDatabase.sequencePrivileges.Data[key]; ok {
		if privilegeMap, ok := seqPrivilegeValue.Privileges[privilege]; ok {
			for _, withGrantOption := range privilegeMap {
				if withGrantOption {
					return key.Role
				}
			}
		}
	}
	for _, group := range GetAllGroupsWithMember(key.Role, true) {
		if returnedID := HasSequencePrivilegeGrantOption(SequencePrivilegeKey{
			Role:   group,
			Schema: key.Schema,
			Name:   key.Name,
		}, privilege); returnedID.IsValid() {
			return returnedID
		}
	}
	return 0
}

// RemoveSequencePrivilege removes the privilege from the global database. If `grantOptionOnly` is true, then only the
// WITH GRANT OPTION portion is revoked. If `grantOptionOnly` is false, then the full privilege is removed.
func RemoveSequencePrivilege(key SequencePrivilegeKey, privilege GrantedPrivilege, grantOptionOnly bool) {
	if seqPrivilegeValue, ok := globalDatabase.sequencePrivileges.Data[key]; ok {
		if privilegeMap, ok := seqPrivilegeValue.Privileges[privilege.Privilege]; ok {
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
					delete(seqPrivilegeValue.Privileges, privilege.Privilege)
				}
			}
		}
		if len(seqPrivilegeValue.Privileges) == 0 {
			delete(globalDatabase.sequencePrivileges.Data, key)
		}
	}
}

// serialize writes the SequencePrivileges to the given writer.
func (sp *SequencePrivileges) serialize(writer *utils.Writer) {
	// Version 0
	// Write the total number of values
	writer.Uint64(uint64(len(sp.Data)))
	for _, value := range sp.Data {
		writer.Uint64(uint64(value.Key.Role))
		writer.String(value.Key.Schema)
		writer.String(value.Key.Name)
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

// deserialize reads the SequencePrivileges from the given reader.
func (sp *SequencePrivileges) deserialize(version uint32, reader *utils.Reader) {
	sp.Data = make(map[SequencePrivilegeKey]SequencePrivilegeValue)
	switch version {
	case 0:
		// Read the total number of values
		dataCount := reader.Uint64()
		for dataIdx := uint64(0); dataIdx < dataCount; dataIdx++ {
			spv := SequencePrivilegeValue{Privileges: make(map[Privilege]map[GrantedPrivilege]bool)}
			spv.Key.Role = RoleID(reader.Uint64())
			spv.Key.Schema = reader.String()
			spv.Key.Name = reader.String()
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
				spv.Privileges[privilege] = grantedMap
			}
			sp.Data[spv.Key] = spv
		}
	default:
		panic("unexpected version in SequencePrivileges")
	}
}
