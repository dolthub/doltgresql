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

import "github.com/dolthub/doltgresql/utils"

// Ownership holds all of the data related to the ownership of roles and database objects.
type Ownership struct {
	Data map[OwnershipKey]map[RoleID]struct{}
}

// OwnershipKey points to a specific database object.
type OwnershipKey struct {
	PrivilegeObject
	Schema string
	Name   string // TODO: this doesn't account for functions, which have: name(param_type1, param_type2, ...)
}

// NewOwnership returns a new *Ownership.
func NewOwnership() *Ownership {
	return &Ownership{
		Data: make(map[OwnershipKey]map[RoleID]struct{}),
	}
}

// AddOwner adds the given role as an owner to the global database.
func AddOwner(key OwnershipKey, role RoleID) {
	key = key.normalize()
	ownerMap, ok := globalDatabase.ownership.Data[key]
	if !ok {
		ownerMap = make(map[RoleID]struct{})
		globalDatabase.ownership.Data[key] = ownerMap
	}
	ownerMap[role] = struct{}{}
}

// GetOwners returns all owners matching the given key.
func GetOwners(key OwnershipKey) []RoleID {
	key = key.normalize()
	if ownerMap, ok := globalDatabase.ownership.Data[key]; ok {
		return utils.GetMapKeysSorted(ownerMap)
	}
	return nil
}

// IsOwner returns whether the given role is an owner for the key.
func IsOwner(key OwnershipKey, role RoleID) bool {
	key = key.normalize()
	if ownerMap, ok := globalDatabase.ownership.Data[key]; ok {
		_, ok = ownerMap[role]
		return ok
	}
	return false
}

// HasOwnerAccess returns whether the given role has access to the ownership of an object, along with the ID of the true
// owner (which may be the same as the given role).
func HasOwnerAccess(key OwnershipKey, role RoleID) RoleID {
	if IsSuperUser(role) {
		owners := GetOwners(key)
		if len(owners) == 0 {
			// This may happen if the privilege file is deleted
			return role
		}
		// Although there may be multiple owners, we'll only return the first one.
		// Postgres already allows for non-determinism with multiple membership paths, so this is fine.
		return owners[0]
	}
	if IsOwner(key, role) {
		return role
	}
	for _, group := range GetAllGroupsWithMember(role, true) {
		if returnedID := HasOwnerAccess(key, group); returnedID.IsValid() {
			return returnedID
		}
	}
	return 0
}

// RemoveOwner removes the role as an owner from the global database.
func RemoveOwner(key OwnershipKey, role RoleID) {
	key = key.normalize()
	if ownerMap, ok := globalDatabase.ownership.Data[key]; ok {
		delete(ownerMap, role)
		if len(ownerMap) == 0 {
			delete(globalDatabase.ownership.Data, key)
		}
	}
}

// normalize accounts for and corrects any potential variation for specific object types.
func (key OwnershipKey) normalize() OwnershipKey {
	if key.PrivilegeObject == PrivilegeObject_SCHEMA {
		if len(key.Schema) == 0 {
			return OwnershipKey{
				PrivilegeObject: PrivilegeObject_SCHEMA,
				Schema:          key.Name,
				Name:            key.Name,
			}
		} else if len(key.Name) == 0 {
			return OwnershipKey{
				PrivilegeObject: PrivilegeObject_SCHEMA,
				Schema:          key.Schema,
				Name:            key.Schema,
			}
		}
	}
	return key
}

// serialize writes the Ownership to the given writer.
func (ownership *Ownership) serialize(writer *utils.Writer) {
	// Version 0
	// Write the total number of values
	writer.Uint64(uint64(len(ownership.Data)))
	for key, roleMap := range ownership.Data {
		// Write the key
		writer.Byte(byte(key.PrivilegeObject))
		writer.String(key.Schema)
		writer.String(key.Name)
		// Write the total number of roles
		writer.Uint32(uint32(len(roleMap)))
		for role := range roleMap {
			writer.Uint64(uint64(role))
		}
	}
}

// deserialize reads the Ownership from the given reader.
func (ownership *Ownership) deserialize(version uint32, reader *utils.Reader) {
	ownership.Data = make(map[OwnershipKey]map[RoleID]struct{})
	switch version {
	case 0:
		// Read the total number of values
		dataCount := reader.Uint64()
		for dataIdx := uint64(0); dataIdx < dataCount; dataIdx++ {
			// Read the key
			key := OwnershipKey{}
			key.PrivilegeObject = PrivilegeObject(reader.Byte())
			key.Schema = reader.String()
			key.Name = reader.String()
			// Read the total number of roles
			roleCount := reader.Uint32()
			roleMap := make(map[RoleID]struct{})
			for roleIdx := uint32(0); roleIdx < roleCount; roleIdx++ {
				roleMap[RoleID(reader.Uint64())] = struct{}{}
			}
			ownership.Data[key] = roleMap
		}
	default:
		panic("unexpected version in Ownership")
	}
}
