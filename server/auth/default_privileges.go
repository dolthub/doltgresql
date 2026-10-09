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
			Grantees: builtInDefaultGrantees(key),
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
	// A grant without the grant option does not remove an existing grant option
	privilegeMap[privilege] = privilegeMap[privilege] || withGrantOption
	dpv.Grantees[grantee] = granteeValue
	storeDefaultPrivilegeValue(dpv)
}

// RemoveDefaultPrivilege removes a default privilege entry from the global database.
// If grantOptionOnly is true, only the WITH GRANT OPTION flag is revoked.
func RemoveDefaultPrivilege(key DefaultPrivilegeKey, grantee RoleID, privilege GrantedPrivilege, grantOptionOnly bool) {
	dpv, ok := globalDatabase.defaultPrivileges.Data[key]
	if !ok {
		builtIn := builtInDefaultGrantees(key)
		if len(builtIn) == 0 {
			return
		}
		dpv = DefaultPrivilegeValue{
			Key:      key,
			Grantees: builtIn,
		}
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
	storeDefaultPrivilegeValue(dpv)
}

// storeDefaultPrivilegeValue stores the given value in the global database. An entry that matches the built-in defaults
// is removed, as is an empty schema-specific entry. An empty global entry is kept, since it overrides the built-in
// defaults (e.g. after revoking EXECUTE on functions from PUBLIC).
func storeDefaultPrivilegeValue(dpv DefaultPrivilegeValue) {
	builtIn := builtInDefaultGrantees(dpv.Key)
	if (len(builtIn) == 0 && len(dpv.Grantees) == 0) || (len(builtIn) > 0 && defaultGranteesEqual(dpv.Grantees, builtIn)) {
		delete(globalDatabase.defaultPrivileges.Data, dpv.Key)
	} else {
		globalDatabase.defaultPrivileges.Data[dpv.Key] = dpv
	}
}

// builtInDefaultGrantees returns the privileges that PostgreSQL grants to non-owners on newly created objects when no
// global default privilege entry exists for the owner. Only global (non-schema-specific) keys have built-in defaults,
// as schema-specific entries are applied in addition to the global ones.
func builtInDefaultGrantees(key DefaultPrivilegeKey) map[RoleID]DefaultPrivilegeGranteeValue {
	grantees := make(map[RoleID]DefaultPrivilegeGranteeValue)
	if !hasBuiltInDefaults(key) {
		return grantees
	}
	switch key.ObjectType {
	case PrivilegeObject_FUNCTION:
		public := GetRole("public")
		if !public.IsValid() {
			return grantees
		}
		grantees[public.ID()] = DefaultPrivilegeGranteeValue{
			Grantee: public.ID(),
			Privileges: map[Privilege]map[GrantedPrivilege]bool{
				Privilege_EXECUTE: {GrantedPrivilege{Privilege: Privilege_EXECUTE, GrantedBy: key.OwnerRole}: false},
			},
		}
	}
	return grantees
}

// hasBuiltInDefaults returns whether PostgreSQL grants privileges to non-owners on new objects for the given key when
// no default privilege entry exists. For such keys, an entry without any grantees is meaningful, as it removes the
// built-in defaults.
func hasBuiltInDefaults(key DefaultPrivilegeKey) bool {
	return key.Schema == "" && key.ObjectType == PrivilegeObject_FUNCTION
}

// removeRoles removes the default privileges owned by, granted to, or granted by any role for which |isRemoved|
// returns true.
func (dp *DefaultPrivileges) removeRoles(isRemoved func(RoleID) bool) {
	for key, value := range dp.Data {
		if isRemoved(key.OwnerRole) {
			delete(dp.Data, key)
			continue
		}
		for grantee, granteeValue := range value.Grantees {
			if isRemoved(grantee) {
				delete(value.Grantees, grantee)
				continue
			}
			for privilege, grants := range granteeValue.Privileges {
				for grant := range grants {
					if isRemoved(grant.GrantedBy) {
						delete(grants, grant)
					}
				}
				if len(grants) == 0 {
					delete(granteeValue.Privileges, privilege)
				}
			}
			if len(granteeValue.Privileges) == 0 {
				delete(value.Grantees, grantee)
			}
		}
		if len(value.Grantees) == 0 && !hasBuiltInDefaults(key) {
			delete(dp.Data, key)
		}
	}
}

// defaultGranteesEqual returns whether the two grantee maps contain the same privileges.
func defaultGranteesEqual(a, b map[RoleID]DefaultPrivilegeGranteeValue) bool {
	if len(a) != len(b) {
		return false
	}
	for grantee, aValue := range a {
		bValue, ok := b[grantee]
		if !ok || len(aValue.Privileges) != len(bValue.Privileges) {
			return false
		}
		for privilege, aMap := range aValue.Privileges {
			bMap, ok := bValue.Privileges[privilege]
			if !ok || len(aMap) != len(bMap) {
				return false
			}
			for grantedPrivilege, withGrantOption := range aMap {
				if bWithGrantOption, ok := bMap[grantedPrivilege]; !ok || bWithGrantOption != withGrantOption {
					return false
				}
			}
		}
	}
	return true
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
// Returns whether any privileges were added. Must be called under LockWrite.
func ApplyDefaultPrivilegesForNewTable(ownerRoleID RoleID, schemaName, tableName string) bool {
	applied := false
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
					applied = true
					AddTablePrivilege(TablePrivilegeKey{
						Role:  granteeID,
						Table: doltdb.TableName{Name: tableName, Schema: schemaName},
					}, grantedPriv, withGrantOption)
				}
			}
		}
	}
	return applied
}

// ApplyDefaultPrivilegesForNewSequence applies any matching default privileges to a newly created sequence.
// Returns whether any privileges were added. Must be called under LockWrite.
func ApplyDefaultPrivilegesForNewSequence(ownerRoleID RoleID, schemaName, seqName string) bool {
	applied := false
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
					applied = true
					AddSequencePrivilege(SequencePrivilegeKey{
						Role:   granteeID,
						Schema: schemaName,
						Name:   seqName,
					}, grantedPriv, withGrantOption)
				}
			}
		}
	}
	return applied
}

// ApplyDefaultPrivilegesForNewRoutine applies any matching default privileges to a newly created function or procedure.
// Returns whether any privileges were added. Must be called under LockWrite.
func ApplyDefaultPrivilegesForNewRoutine(ownerRoleID RoleID, schemaName, routineName string) bool {
	applied := false
	globalKey := DefaultPrivilegeKey{OwnerRole: ownerRoleID, ObjectType: PrivilegeObject_FUNCTION}
	if _, ok := globalDatabase.defaultPrivileges.Data[globalKey]; !ok {
		for granteeID, granteeValue := range builtInDefaultGrantees(globalKey) {
			for _, privilegeMap := range granteeValue.Privileges {
				for grantedPriv, withGrantOption := range privilegeMap {
					applied = true
					AddRoutinePrivilege(RoutinePrivilegeKey{
						Role:   granteeID,
						Schema: schemaName,
						Name:   routineName,
					}, grantedPriv, withGrantOption)
				}
			}
		}
	}
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
					applied = true
					AddRoutinePrivilege(RoutinePrivilegeKey{
						Role:   granteeID,
						Schema: schemaName,
						Name:   routineName,
					}, grantedPriv, withGrantOption)
				}
			}
		}
	}
	return applied
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
