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
	"sort"

	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"

	"github.com/dolthub/doltgresql/utils"
)

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

// AllRoleMemberships returns every role membership in the database, sorted by the group's ID and then the member's
// ID. This does not handle locking, so callers should protect the call with LockRead.
func AllRoleMemberships() []RoleMembershipValue {
	var values []RoleMembershipValue
	for _, groupMap := range globalDatabase.roleMembership.Data {
		for _, value := range groupMap {
			values = append(values, value)
		}
	}
	sort.Slice(values, func(i, j int) bool {
		if values[i].Group != values[j].Group {
			return values[i].Group < values[j].Group
		}
		return values[i].Member < values[j].Member
	})
	return values
}

// AddMemberToGroup adds the member role to the group role.
func AddMemberToGroup(member RoleID, group RoleID, withAdminOption bool, grantedBy RoleID) {
	// We'll perform a sanity check for circular membership. This should be done before this call is made, but since we
	// make assumptions that circular relationships are forbidden (which could lead to infinite loops otherwise), we
	// enforce it here too.
	if isRoleMemberNoSuper(group, member) {
		panic("missing validation to prevent circular role relationships")
	}
	groupMap, ok := globalDatabase.roleMembership.Data[member]
	if !ok {
		groupMap = make(map[RoleID]RoleMembershipValue)
		globalDatabase.roleMembership.Data[member] = groupMap
	}
	if previous, exists := groupMap[group]; exists {
		// Repeating a grant does not remove an existing admin option or change
		// the original grantor.
		previous.WithAdminOption = previous.WithAdminOption || withAdminOption
		groupMap[group] = previous
		return
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
	if globalDatabase.rolesByID[member].id == 0 || globalDatabase.rolesByID[group].id == 0 {
		return 0, false, false
	}
	if IsSuperUser(member) {
		return group, true, true
	}
	if isRoleMemberNoSuper(member, group) {
		return group, HasRolePrivileges(member, group), HasRoleAdminOption(member, group)
	}
	return 0, false, false
}

// roleClosure returns reachable roles, including the starting role. Inherited
// privileges stop at NOINHERIT roles; membership itself does not.
func roleClosure(member RoleID, inheritsOnly bool) []RoleID {
	seen := make(map[RoleID]bool)
	roles := []RoleID{member}
	seen[member] = true
	for i := 0; i < len(roles); i++ {
		for _, group := range GetAllGroupsWithMember(roles[i], inheritsOnly) {
			if !seen[group] {
				seen[group] = true
				roles = append(roles, group)
			}
		}
	}
	return roles
}

// isRoleMemberNoSuper tests actual membership, including self, without the
// superuser shortcut. Cycle validation must use actual edges.
func isRoleMemberNoSuper(member, group RoleID) bool {
	for _, role := range roleClosure(member, false) {
		if role == group {
			return true
		}
	}
	return false
}

// HasRolePrivileges reports whether the member immediately has a role's
// privileges, traversing only inheritable membership paths. Call under a lock.
func HasRolePrivileges(member, group RoleID) bool {
	if IsSuperUser(member) {
		return true
	}
	for _, role := range roleClosure(member, true) {
		if role == group {
			return true
		}
	}
	return false
}

// HasRoleAdminOption implements PG15's admin-option traversal, independent of
// privilege inheritance. Call under a lock.
func HasRoleAdminOption(member, group RoleID) bool {
	if IsSuperUser(member) {
		return true
	}
	for _, role := range roleClosure(member, false) {
		if membership, ok := globalDatabase.roleMembership.Data[role][group]; ok && membership.WithAdminOption {
			return true
		}
	}
	return false
}

// CanAdministerRole permits superusers, CREATEROLE for nonsuperuser targets, or
// roles with ADMIN OPTION to grant and revoke membership, as in PostgreSQL 15.
func CanAdministerRole(member, group RoleID) bool {
	if IsSuperUser(member) {
		return true
	}
	return !IsSuperUser(group) && (globalDatabase.rolesByID[member].CanCreateRoles || HasRoleAdminOption(member, group))
}

// ValidateMembershipGrants rejects cycles for an entire proposed batch before
// any edge is installed. Existing state is never modified during validation.
func ValidateMembershipGrants(grants []RoleMembershipValue) error {
	edges := make(map[RoleID][]RoleID)
	for member, groups := range globalDatabase.roleMembership.Data {
		for group := range groups {
			edges[member] = append(edges[member], group)
		}
	}
	for _, grant := range grants {
		edges[grant.Member] = append(edges[grant.Member], grant.Group)
	}
	return validateMembershipEdges(edges)
}

// validateMembershipEdges detects cycles even in malformed persisted state.
func validateMembershipEdges(edges map[RoleID][]RoleID) error {
	state := make(map[RoleID]uint8)
	var visit func(RoleID) bool
	visit = func(role RoleID) bool {
		if state[role] == 1 {
			return false
		}
		if state[role] == 2 {
			return true
		}
		state[role] = 1
		for _, group := range edges[role] {
			if !visit(group) {
				return false
			}
		}
		state[role] = 2
		return true
	}
	for role := range edges {
		if !visit(role) {
			return pgerror.New(pgcode.InvalidGrantOperation, "role membership would create a cycle")
		}
	}
	return nil
}

// GetAllGroupsWithMember returns every group that the role is a direct member of. This can also filter by groups that
// the member has privilege access on.
func GetAllGroupsWithMember(member RoleID, inheritsPrivilegesOnly bool) []RoleID {
	memberRole, ok := globalDatabase.rolesByID[member]
	if !ok || (inheritsPrivilegesOnly && !memberRole.InheritPrivileges) {
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
	case 0, 1:
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
