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

// TODO: doc
type Ownership struct {
	Data map[OwnershipKey]RoleID
}

// TODO: doc
type OwnershipKey struct {
	PrivilegeObject
	Schema string
	Name   string // TODO: this doesn't account for functions, which have: name(param_type1, param_type2, ...)
}

// TODO: doc
func NewOwnership() *Ownership {
	return &Ownership{
		Data: make(map[OwnershipKey]RoleID),
	}
}

// TODO: doc
func AddOwner(key OwnershipKey, role RoleID) {
	globalMockDatabase.ownership.Data[key] = role
}

// TODO: doc
func GetOwner(key OwnershipKey) RoleID {
	if roleID, ok := globalMockDatabase.ownership.Data[key]; ok {
		return roleID
	}
	return RoleID(0)
}

// TODO: doc
func IsOwner(key OwnershipKey, role RoleID) bool {
	return GetOwner(key) == role
}
