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
	// Legacy grants derive SET and INHERIT from PostgreSQL 15 role attributes.
	// Explicit options require a versioned auth-file migration in a later phase.
}

// MembershipGrant is the logical view of a membership. Nil option pointers
// identify a PostgreSQL 15 legacy grant: INHERIT comes from role attributes,
// and SET is allowed through structural membership. Explicit options need an
// auth-file format migration before they can be written.
type MembershipGrant struct {
	Member        RoleID
	GrantedRole   RoleID
	Grantor       RoleID
	AdminOption   bool
	InheritOption *bool
	SetOption     *bool
}

// MembershipGrants returns the grants for a member without exposing the
// current pair-keyed storage. The caller must hold the auth lock.
func MembershipGrants(member RoleID) []MembershipGrant {
	values := globalDatabase.roleMembership.Data[member]
	grants := make([]MembershipGrant, 0, len(values))
	for _, value := range values {
		grants = append(grants, MembershipGrant{
			Member: value.Member, GrantedRole: value.Group,
			Grantor: value.GrantedBy, AdminOption: value.WithAdminOption,
		})
	}
	sort.Slice(grants, func(i, j int) bool { return grants[i].GrantedRole < grants[j].GrantedRole })
	return grants
}

// HasRoleMembership checks only stored membership edges. It does not grant
// virtual membership to superusers or apply INHERIT, SET, or ADMIN rules.
// The caller must hold the auth lock.
func HasRoleMembership(member, group RoleID) bool {
	if _, ok := LookupRoleByID(member); !ok {
		return false
	}
	if _, ok := LookupRoleByID(group); !ok {
		return false
	}
	return membershipPath(member, group, false)
}

// CanSetRole answers whether a session role may select a target. PostgreSQL 15
// permits SET through membership even when the member is NOINHERIT.
// The caller must hold the auth lock.
func CanSetRole(sessionRole, targetRole RoleID) bool {
	if _, ok := LookupRoleByID(targetRole); !ok {
		return false
	}
	actor, ok := LookupRoleByID(sessionRole)
	if !ok {
		return false
	}
	return actor.IsSuperUser || membershipPath(sessionRole, targetRole, false)
}

// InheritsPrivileges follows only paths whose member role has INHERIT. Role
// attributes, including SUPERUSER and BYPASSRLS, are never inherited here.
// The caller must hold the auth lock.
func InheritsPrivileges(effectiveRole, targetRole RoleID) bool {
	if _, ok := LookupRoleByID(targetRole); !ok {
		return false
	}
	if _, ok := LookupRoleByID(effectiveRole); !ok {
		return false
	}
	return membershipPath(effectiveRole, targetRole, true)
}

// CanAdministerRole requires an ADMIN grant on the target, reached from the
// actor through inherited memberships. It is distinct from SET authority.
// The caller must hold the auth lock.
func CanAdministerRole(actorRole, targetRole RoleID) bool {
	if _, ok := LookupRoleByID(targetRole); !ok {
		return false
	}
	actor, ok := LookupRoleByID(actorRole)
	if !ok {
		return false
	}
	if actor.IsSuperUser {
		return true
	}
	seen := make(map[RoleID]bool)
	var walk func(RoleID) bool
	walk = func(member RoleID) bool {
		if seen[member] {
			return false
		}
		seen[member] = true
		role, ok := LookupRoleByID(member)
		if !ok {
			return false
		}
		if grant, ok := globalDatabase.roleMembership.Data[member][targetRole]; ok && grant.WithAdminOption {
			return true
		}
		if !role.InheritPrivileges {
			return false
		}
		for group := range globalDatabase.roleMembership.Data[member] {
			if walk(group) {
				return true
			}
		}
		return false
	}
	return walk(actorRole)
}

func membershipPath(member, target RoleID, inheritOnly bool) bool {
	seen := make(map[RoleID]bool)
	var walk func(RoleID) bool
	walk = func(current RoleID) bool {
		if current == target {
			return true
		}
		if seen[current] {
			return false
		}
		seen[current] = true
		role, ok := LookupRoleByID(current)
		if !ok || (inheritOnly && !role.InheritPrivileges) {
			return false
		}
		for group := range globalDatabase.roleMembership.Data[current] {
			if walk(group) {
				return true
			}
		}
		return false
	}
	return walk(member)
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
	if HasRoleMembership(group, member) {
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

// GetAllGroupsWithMember returns every group that the role is a direct member of. This can also filter by groups that
// the member has privilege access on.
func GetAllGroupsWithMember(member RoleID, inheritsPrivilegesOnly bool) []RoleID {
	if _, ok := LookupRoleByID(member); !ok {
		return nil
	}
	groupMap := globalDatabase.roleMembership.Data[member]
	groups := make([]RoleID, 0, len(groupMap))
	for groupID := range groupMap {
		if inheritsPrivilegesOnly && !InheritsPrivileges(member, groupID) {
			continue
		}
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
