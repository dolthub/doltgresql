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

// DefaultPrivileges stores all default privilege settings.
type DefaultPrivileges struct {
	Data map[DefaultPrivilegeKey]DefaultPrivilegeValue
}

// DefaultPrivilegeKey uniquely identifies a default privilege entry. It encodes the combination of:
// which role is creating objects (ForRole), in which schema (Schema, empty = all schemas),
// for which object type (ObjectType), and which role receives the privilege (Grantee).
type DefaultPrivilegeKey struct {
	ForRole    RoleID
	Schema     string
	ObjectType PrivilegeObject
	Grantee    RoleID
}

// DefaultPrivilegeValue stores the privileges associated with a DefaultPrivilegeKey.
type DefaultPrivilegeValue struct {
	Key        DefaultPrivilegeKey
	Privileges map[Privilege]bool // privilege -> withGrantOption
}

// NewDefaultPrivileges returns a new *DefaultPrivileges.
func NewDefaultPrivileges() *DefaultPrivileges {
	return &DefaultPrivileges{make(map[DefaultPrivilegeKey]DefaultPrivilegeValue)}
}

// AddDefaultPrivilege adds (or updates) a default privilege in the global database.
func AddDefaultPrivilege(key DefaultPrivilegeKey, privilege Privilege, withGrantOption bool) {
	val, ok := globalDatabase.defaultPrivileges.Data[key]
	if !ok {
		val = DefaultPrivilegeValue{
			Key:        key,
			Privileges: make(map[Privilege]bool),
		}
	}
	// Preserve a true (withGrantOption) once set; only update to false if explicitly requested.
	if existing, exists := val.Privileges[privilege]; !exists || withGrantOption || !existing {
		val.Privileges[privilege] = withGrantOption
	}
	globalDatabase.defaultPrivileges.Data[key] = val
}

// RemoveDefaultPrivilege removes a default privilege from the global database.
// If grantOptionOnly is true, only the WITH GRANT OPTION flag is cleared.
func RemoveDefaultPrivilege(key DefaultPrivilegeKey, privilege Privilege, grantOptionOnly bool) {
	val, ok := globalDatabase.defaultPrivileges.Data[key]
	if !ok {
		return
	}
	if grantOptionOnly {
		if _, exists := val.Privileges[privilege]; exists {
			val.Privileges[privilege] = false
			globalDatabase.defaultPrivileges.Data[key] = val
		}
	} else {
		delete(val.Privileges, privilege)
		if len(val.Privileges) == 0 {
			delete(globalDatabase.defaultPrivileges.Data, key)
		} else {
			globalDatabase.defaultPrivileges.Data[key] = val
		}
	}
}

// GetDefaultPrivilegesForCreator returns all default privilege entries that apply when the given
// creator role creates an object of the given type in the given schema. Entries with an empty
// Schema field match all schemas.
func GetDefaultPrivilegesForCreator(creatorRole RoleID, objectType PrivilegeObject, schema string) []DefaultPrivilegeValue {
	var result []DefaultPrivilegeValue
	for _, val := range globalDatabase.defaultPrivileges.Data {
		if val.Key.ForRole != creatorRole || val.Key.ObjectType != objectType {
			continue
		}
		if val.Key.Schema == "" || val.Key.Schema == schema {
			result = append(result, val)
		}
	}
	return result
}

// serialize writes the DefaultPrivileges to the given writer.
func (dp *DefaultPrivileges) serialize(writer *utils.Writer) {
	// Version 0
	writer.Uint64(uint64(len(dp.Data)))
	for _, value := range dp.Data {
		writer.Uint64(uint64(value.Key.ForRole))
		writer.String(value.Key.Schema)
		writer.Uint8(uint8(value.Key.ObjectType))
		writer.Uint64(uint64(value.Key.Grantee))
		writer.Uint64(uint64(len(value.Privileges)))
		for priv, withGrantOption := range value.Privileges {
			writer.String(string(priv))
			writer.Bool(withGrantOption)
		}
	}
}

// deserialize reads the DefaultPrivileges from the given reader.
func (dp *DefaultPrivileges) deserialize(version uint32, reader *utils.Reader) {
	dp.Data = make(map[DefaultPrivilegeKey]DefaultPrivilegeValue)
	switch version {
	case 0:
		dataCount := reader.Uint64()
		for dataIdx := uint64(0); dataIdx < dataCount; dataIdx++ {
			dpv := DefaultPrivilegeValue{Privileges: make(map[Privilege]bool)}
			dpv.Key.ForRole = RoleID(reader.Uint64())
			dpv.Key.Schema = reader.String()
			dpv.Key.ObjectType = PrivilegeObject(reader.Uint8())
			dpv.Key.Grantee = RoleID(reader.Uint64())
			privilegeCount := reader.Uint64()
			for privIdx := uint64(0); privIdx < privilegeCount; privIdx++ {
				priv := Privilege(reader.String())
				withGrantOption := reader.Bool()
				dpv.Privileges[priv] = withGrantOption
			}
			dp.Data[dpv.Key] = dpv
		}
	default:
		panic("unexpected version in DefaultPrivileges")
	}
}
