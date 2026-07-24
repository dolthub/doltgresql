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
	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"

	"github.com/dolthub/doltgresql/utils"
)

// DefaultPrivileges stores the default privileges automatically applied when objects are created.
type DefaultPrivileges struct {
	Data map[DefaultPrivilegeKey]DefaultPrivilegeValue
}

// DefaultPrivilegeKey identifies the context for a set of default privileges:
// the owner role, the optional schema scope, and the object type.
type DefaultPrivilegeKey struct {
	OwnerRole  RoleID
	Schema     string          // empty = applicable to any schema
	ObjectType PrivilegeObject // TABLE, SEQUENCE, FUNCTION, SCHEMA, TYPE
}

// DefaultPrivilegeValue stores the grantee ACL entries for a given DefaultPrivilegeKey.
type DefaultPrivilegeValue struct {
	Key      DefaultPrivilegeKey
	Grantees map[RoleID]DefaultPrivilegeGranteeValue
}

// DefaultPrivilegeGranteeValue stores the privileges granted to a specific role within a default ACL.
type DefaultPrivilegeGranteeValue struct {
	Grantee    RoleID
	Privileges map[Privilege]map[GrantedPrivilege]bool
}

// NewDefaultPrivileges returns a new *DefaultPrivileges.
func NewDefaultPrivileges() *DefaultPrivileges {
	return &DefaultPrivileges{make(map[DefaultPrivilegeKey]DefaultPrivilegeValue)}
}

// AddDefaultPrivilege adds a default privilege entry to the global database.
func AddDefaultPrivilege(key DefaultPrivilegeKey, grantee RoleID, privilege GrantedPrivilege, withGrantOption bool) {
	dpv, ok := globalDatabase.defaultPrivileges.Data[key]
	if !ok {
		dpv = DefaultPrivilegeValue{
			Key:      key,
			Grantees: make(map[RoleID]DefaultPrivilegeGranteeValue),
		}
	}
	granteeValue, ok := dpv.Grantees[grantee]
	if !ok {
		granteeValue = DefaultPrivilegeGranteeValue{
			Grantee:    grantee,
			Privileges: make(map[Privilege]map[GrantedPrivilege]bool),
		}
	}
	privilegeMap, ok := granteeValue.Privileges[privilege.Privilege]
	if !ok {
		privilegeMap = make(map[GrantedPrivilege]bool)
		granteeValue.Privileges[privilege.Privilege] = privilegeMap
	}
	privilegeMap[privilege] = withGrantOption
	dpv.Grantees[grantee] = granteeValue
	globalDatabase.defaultPrivileges.Data[key] = dpv
}

// RemoveDefaultPrivilege removes a default privilege entry from the global database.
// If grantOptionOnly is true, only the WITH GRANT OPTION flag is revoked.
func RemoveDefaultPrivilege(key DefaultPrivilegeKey, grantee RoleID, privilege GrantedPrivilege, grantOptionOnly bool) {
	dpv, ok := globalDatabase.defaultPrivileges.Data[key]
	if !ok {
		return
	}
	granteeValue, ok := dpv.Grantees[grantee]
	if !ok {
		return
	}
	privilegeMap, ok := granteeValue.Privileges[privilege.Privilege]
	if !ok {
		return
	}
	if grantOptionOnly {
		if privilege.GrantedBy.IsValid() {
			if _, ok = privilegeMap[privilege]; ok {
				privilegeMap[privilege] = false
			}
		} else {
			for k := range privilegeMap {
				privilegeMap[k] = false
			}
		}
	} else {
		if privilege.GrantedBy.IsValid() {
			delete(privilegeMap, privilege)
		} else {
			clear(privilegeMap)
		}
		if len(privilegeMap) == 0 {
			delete(granteeValue.Privileges, privilege.Privilege)
		}
	}
	if len(granteeValue.Privileges) == 0 {
		delete(dpv.Grantees, grantee)
	} else {
		dpv.Grantees[grantee] = granteeValue
	}
	if len(dpv.Grantees) == 0 {
		delete(globalDatabase.defaultPrivileges.Data, key)
	} else {
		globalDatabase.defaultPrivileges.Data[key] = dpv
	}
}

// GetAllDefaultPrivileges returns all default privilege entries.
func GetAllDefaultPrivileges() []DefaultPrivilegeValue {
	result := make([]DefaultPrivilegeValue, 0, len(globalDatabase.defaultPrivileges.Data))
	for _, v := range globalDatabase.defaultPrivileges.Data {
		result = append(result, v)
	}
	return result
}

// ApplyDefaultPrivilegesForNewTable applies any matching default privileges to a newly created table.
// Must be called under LockWrite.
func ApplyDefaultPrivilegesForNewTable(ownerRoleID RoleID, schemaName, tableName string) {
	for key, dpv := range globalDatabase.defaultPrivileges.Data {
		if key.OwnerRole != ownerRoleID || key.ObjectType != PrivilegeObject_TABLE {
			continue
		}
		if key.Schema != "" && key.Schema != schemaName {
			continue
		}
		for granteeID, granteeValue := range dpv.Grantees {
			for _, privilegeMap := range granteeValue.Privileges {
				for grantedPriv, withGrantOption := range privilegeMap {
					AddTablePrivilege(TablePrivilegeKey{
						Role:  granteeID,
						Table: doltdb.TableName{Name: tableName, Schema: schemaName},
					}, grantedPriv, withGrantOption)
				}
			}
		}
	}
}

// ApplyDefaultPrivilegesForNewSequence applies any matching default privileges to a newly created sequence.
// Must be called under LockWrite.
func ApplyDefaultPrivilegesForNewSequence(ownerRoleID RoleID, schemaName, seqName string) {
	for key, dpv := range globalDatabase.defaultPrivileges.Data {
		if key.OwnerRole != ownerRoleID || key.ObjectType != PrivilegeObject_SEQUENCE {
			continue
		}
		if key.Schema != "" && key.Schema != schemaName {
			continue
		}
		for granteeID, granteeValue := range dpv.Grantees {
			for _, privilegeMap := range granteeValue.Privileges {
				for grantedPriv, withGrantOption := range privilegeMap {
					AddSequencePrivilege(SequencePrivilegeKey{
						Role:   granteeID,
						Schema: schemaName,
						Name:   seqName,
					}, grantedPriv, withGrantOption)
				}
			}
		}
	}
}

// ApplyDefaultPrivilegesForNewRoutine applies any matching default privileges to a newly created function or procedure.
// Must be called under LockWrite.
func ApplyDefaultPrivilegesForNewRoutine(ownerRoleID RoleID, schemaName, routineName string) {
	for key, dpv := range globalDatabase.defaultPrivileges.Data {
		if key.OwnerRole != ownerRoleID || key.ObjectType != PrivilegeObject_FUNCTION {
			continue
		}
		if key.Schema != "" && key.Schema != schemaName {
			continue
		}
		for granteeID, granteeValue := range dpv.Grantees {
			for _, privilegeMap := range granteeValue.Privileges {
				for grantedPriv, withGrantOption := range privilegeMap {
					AddRoutinePrivilege(RoutinePrivilegeKey{
						Role:   granteeID,
						Schema: schemaName,
						Name:   routineName,
					}, grantedPriv, withGrantOption)
				}
			}
		}
	}
}

// DefaultPrivilegeObjTypeChar returns the PostgreSQL pg_default_acl defaclobjtype character for a PrivilegeObject.
func DefaultPrivilegeObjTypeChar(objType PrivilegeObject) string {
	switch objType {
	case PrivilegeObject_TABLE:
		return "r"
	case PrivilegeObject_SEQUENCE:
		return "S"
	case PrivilegeObject_FUNCTION:
		return "f"
	case PrivilegeObject_TYPE:
		return "T"
	case PrivilegeObject_SCHEMA:
		return "n"
	default:
		return "?"
	}
}

// serialize writes the DefaultPrivileges to the given writer.
func (dp *DefaultPrivileges) serialize(writer *utils.Writer) {
	// Version 2
	// Write the total number of values
	writer.Uint64(uint64(len(dp.Data)))
	for _, value := range dp.Data {
		writer.Uint64(uint64(value.Key.OwnerRole))
		writer.String(value.Key.Schema)
		writer.Uint8(uint8(value.Key.ObjectType))
		writer.Uint64(uint64(len(value.Grantees)))
		for _, granteeValue := range value.Grantees {
			writer.Uint64(uint64(granteeValue.Grantee))
			writer.Uint64(uint64(len(granteeValue.Privileges)))
			for priv, privilegeMap := range granteeValue.Privileges {
				writer.String(string(priv))
				writer.Uint32(uint32(len(privilegeMap)))
				for grantedPrivilege, withGrantOption := range privilegeMap {
					writer.Uint64(uint64(grantedPrivilege.GrantedBy))
					writer.Bool(withGrantOption)
				}
			}
		}
	}
}

// deserialize reads the DefaultPrivileges from the given reader.
func (dp *DefaultPrivileges) deserialize(version uint32, reader *utils.Reader) {
	dp.Data = make(map[DefaultPrivilegeKey]DefaultPrivilegeValue)
	switch version {
	case 0:
	case 1:
	case 2:
		dataCount := reader.Uint64()
		for i := uint64(0); i < dataCount; i++ {
			dpv := DefaultPrivilegeValue{
				Grantees: make(map[RoleID]DefaultPrivilegeGranteeValue),
			}
			dpv.Key.OwnerRole = RoleID(reader.Uint64())
			dpv.Key.Schema = reader.String()
			dpv.Key.ObjectType = PrivilegeObject(reader.Uint8())
			granteeCount := reader.Uint64()
			for j := uint64(0); j < granteeCount; j++ {
				granteeValue := DefaultPrivilegeGranteeValue{
					Grantee:    RoleID(reader.Uint64()),
					Privileges: make(map[Privilege]map[GrantedPrivilege]bool),
				}
				privCount := reader.Uint64()
				for k := uint64(0); k < privCount; k++ {
					priv := Privilege(reader.String())
					grantedCount := reader.Uint32()
					grantedMap := make(map[GrantedPrivilege]bool)
					for l := uint32(0); l < grantedCount; l++ {
						gp := GrantedPrivilege{
							Privilege: priv,
							GrantedBy: RoleID(reader.Uint64()),
						}
						grantedMap[gp] = reader.Bool()
					}
					granteeValue.Privileges[priv] = grantedMap
				}
				dpv.Grantees[granteeValue.Grantee] = granteeValue
			}
			dp.Data[dpv.Key] = dpv
		}
	default:
		panic("unexpected version in SequencePrivileges")
	}
}
