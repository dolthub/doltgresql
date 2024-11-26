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

// RoleMembership contains all roles that have been granted to other roles.
type RoleMembership struct {
	Data map[RoleID]map[RoleID]RoleMembershipValue
}

// RoleMembershipValue contains specific membership information between two roles.
type RoleMembershipValue struct {
	Member          RoleID
	Group           RoleID
	WithAdminOption bool
	GrantedBy       RoleID
}

// NewRoleMembership returns a new *RoleMembership.
func NewRoleMembership() *RoleMembership {
	return &RoleMembership{
		Data: make(map[RoleID]map[RoleID]RoleMembershipValue),
	}
}

// AddMemberToGroup adds the member role to the group role.
func AddMemberToGroup(member RoleID, group RoleID, withAdminOption bool, grantedBy RoleID) {
	// We'll perform a sanity check for circular membership. This should be done before this call is made, but since we
	// make assumptions that circular relationships are forbidden (which could lead to infinite loops otherwise), we
	// enforce it here too.
	if groupID, _, _ := IsRoleAMember(group, member); (groupID.IsValid() || member == group) && !globalDatabase.rolesByID[group].IsSuperUser {
		panic("missing validation to prevent circular role relationships")
	}
	groupMap, ok := globalDatabase.roleMembership.Data[member]
	if !ok {
		groupMap = make(map[RoleID]RoleMembershipValue)
		globalDatabase.roleMembership.Data[member] = groupMap
	}
	groupMap[group] = RoleMembershipValue{
		Member:          member,
		Group:           group,
		WithAdminOption: withAdminOption,
		GrantedBy:       grantedBy,
	}
}

// IsRoleAMember returns whether the given role is a member of the group by returning the group's ID. Also returns
// whether the member was granted WITH ADMIN OPTION, allowing it to grant membership to the group to other roles. A
// member does not automatically have ADMIN OPTION on itself, therefore this check must be performed.
func IsRoleAMember(member RoleID, group RoleID) (groupID RoleID, inheritsPrivileges bool, hasWithAdminOption bool) {
	// If the member and group are the same, then we only check for SUPERUSER status to allow WITH ADMIN OPTION
	if member == group {
		return group, true, globalDatabase.rolesByID[member].IsSuperUser
	}
	// Postgres does not allow for circular role membership, so we can recursively check without worry:
	// https://www.postgresql.org/docs/15/catalog-pg-auth-members.html
	if groupMap, ok := globalDatabase.roleMembership.Data[member]; ok {
		for _, value := range groupMap {
			if value.Group == group {
				return group, globalDatabase.rolesByID[member].InheritPrivileges, value.WithAdminOption
			}
			// This recursively walks through memberships
			if groupID, _, hasWithAdminOption = IsRoleAMember(value.Group, group); groupID.IsValid() {
				return groupID, globalDatabase.rolesByID[member].InheritPrivileges, hasWithAdminOption
			}
		}
	}
	// A SUPERUSER has access to everything, and therefore functions as though it's a member of every group
	if globalDatabase.rolesByID[member].IsSuperUser {
		return group, true, true
	}
	return 0, false, false
}

// GetAllGroupsWithMember returns every group that the role is a direct member of. This can also filter by groups that
// the member has privilege access on.
func GetAllGroupsWithMember(member RoleID, inheritsPrivilegesOnly bool) []RoleID {
	memberRole, ok := globalDatabase.rolesByID[member]
	if !ok || !memberRole.InheritPrivileges {
		return nil
	}
	groupMap := globalDatabase.roleMembership.Data[member]
	groups := make([]RoleID, 0, len(groupMap))
	for groupID := range groupMap {
		groups = append(groups, groupID)
	}
	return groups
}

// RemoveMemberFromGroup removes the member from the group. If `adminOptionOnly` is true, then only the WITH ADMIN
// OPTION portion is revoked. If `adminOptionOnly` is false, then the member is fully is removed.
func RemoveMemberFromGroup(member RoleID, group RoleID, adminOptionOnly bool) {
	if groupMap, ok := globalDatabase.roleMembership.Data[member]; ok {
		if adminOptionOnly {
			value := groupMap[group]
			value.WithAdminOption = false
			groupMap[group] = value
		} else {
			delete(groupMap, group)
		}
		if len(groupMap) == 0 {
			delete(globalDatabase.roleMembership.Data, member)
		}
	}
}

// serialize writes the RoleMembership to the given writer.
func (membership *RoleMembership) serialize(writer *utils.Writer) {
	// Version 0
	// Write the total number of members
	writer.Uint64(uint64(len(membership.Data)))
	for _, groupMap := range membership.Data {
		// Write the number of groups
		writer.Uint64(uint64(len(groupMap)))
		for _, mapValue := range groupMap {
			// Write the membership information
			writer.Uint64(uint64(mapValue.Member))
			writer.Uint64(uint64(mapValue.Group))
			writer.Bool(mapValue.WithAdminOption)
			writer.Uint64(uint64(mapValue.GrantedBy))
		}
	}
}

// deserialize reads the RoleMembership from the given reader.
func (membership *RoleMembership) deserialize(version uint32, reader *utils.Reader) {
	membership.Data = make(map[RoleID]map[RoleID]RoleMembershipValue)
	switch version {
	case 0:
		// Read the total number of members
		memberCount := reader.Uint64()
		for memberIdx := uint64(0); memberIdx < memberCount; memberIdx++ {
			// Read the number of groups
			groupCount := reader.Uint64()
			groupMap := make(map[RoleID]RoleMembershipValue)
			var member RoleID
			for groupIdx := uint64(0); groupIdx < groupCount; groupIdx++ {
				// Read the membership information
				value := RoleMembershipValue{}
				value.Member = RoleID(reader.Uint64())
				value.Group = RoleID(reader.Uint64())
				value.WithAdminOption = reader.Bool()
				value.GrantedBy = RoleID(reader.Uint64())
				// Add the information to the map
				groupMap[value.Group] = value
				member = value.Member
			}
			// Add the group map to the data
			membership.Data[member] = groupMap
		}
	default:
		panic("unexpected version in RoleMembership")
	}
}
