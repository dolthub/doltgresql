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
	ownerMap, ok := globalDatabase.ownership.Data[key]
	if !ok {
		ownerMap = make(map[RoleID]struct{})
		globalDatabase.ownership.Data[key] = ownerMap
	}
	ownerMap[role] = struct{}{}
}

// GetOwners returns all owners matching the given key.
func GetOwners(key OwnershipKey) []RoleID {
	if ownerMap, ok := globalDatabase.ownership.Data[key]; ok {
		return utils.GetMapKeysSorted(ownerMap)
	}
	return nil
}

// IsOwner returns whether the given owner has an entry for the key.
func IsOwner(key OwnershipKey, role RoleID) bool {
	if ownerMap, ok := globalDatabase.ownership.Data[key]; ok {
		_, ok = ownerMap[role]
		return ok
	}
	return false
}

// RemoveOwner removes the role as an owner from the global database.
func RemoveOwner(key OwnershipKey, role RoleID) {
	if ownerMap, ok := globalDatabase.ownership.Data[key]; ok {
		delete(ownerMap, role)
		if len(ownerMap) == 0 {
			delete(globalDatabase.ownership.Data, key)
		}
	}
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
