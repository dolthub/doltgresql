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
	"github.com/cockroachdb/errors"
	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/utils"
)

// PersistChanges will save the state of the global database to disk (assuming we are not using the pure in-memory
// implementation). When cluster replication is enabled, the new state is also offered to the standby replicas, and
// the replication-ack waiters are appended to |rsc|. Callers must hold the write lock, and should pass |rsc| to
// WaitForReplication once the lock is released.
func PersistChanges(ctx *sql.Context, rsc *doltdb.ReplicationStatusController) error {
	if clusterReplicator != nil {
		return clusterReplicator.SendToReplicas(ctx, globalDatabase.serialize(), rsc)
	}
	if fileSystem != nil {
		return WriteSerializedDatabase(globalDatabase.serialize())
	}
	return nil
}

// serialize returns the Database as a byte slice.
func (db *Database) serialize() []byte {
	writer := utils.NewWriter(16384)
	// Write the version
	writer.Uint32(2)
	// Write the roles
	writer.Uint32(uint32(len(db.rolesByID)))
	for _, role := range db.rolesByID {
		role.serialize(writer)
	}
	// Write the database privileges
	db.databasePrivileges.serialize(writer)
	// Write the schema privileges
	db.schemaPrivileges.serialize(writer)
	// Write the table privileges
	db.tablePrivileges.serialize(writer)
	// Write the sequence privileges
	db.sequencePrivileges.serialize(writer)
	// Write the routine privileges
	db.routinePrivileges.serialize(writer)
	// Write the role chain
	db.roleMembership.serialize(writer)
	return writer.Data()
}

// deserialize creates a Database from a byte slice.
func (db *Database) deserialize(data []byte) (err error) {
	// Reader uses slice indexing. Treat truncated snapshots as a failed load,
	// rather than crashing a standby or replacing its live authorization state.
	defer func() {
		if recover() != nil {
			err = errors.New("invalid auth database format")
		}
	}()
	if len(data) < 4 {
		return errors.New("invalid auth database format")
	}
	reader := utils.NewReader(data)
	version := reader.Uint32()
	switch version {
	case 0, 1, 2:
		err = db.deserializeVersion(reader, version)
	default:
		return errors.Errorf("Authorization database format %d is not supported, please upgrade Doltgres", version)
	}
	if err != nil {
		return err
	}
	if !reader.IsEmpty() {
		return errors.New("invalid trailing data in auth database")
	}
	db.removeInvalidRoleReferences()
	edges := make(map[RoleID][]RoleID)
	for member, groups := range db.roleMembership.Data {
		for group := range groups {
			edges[member] = append(edges[member], group)
		}
	}
	if err := validateMembershipEdges(edges); err != nil {
		return err
	}
	// Advance the role ID counter past every persisted role. Without this, IDs minted after loading serialized
	// state collide with existing roles, which SetRole then silently replaces.
	var maxID uint64
	for id := range db.rolesByID {
		if uint64(id) > maxID {
			maxID = uint64(id)
		}
	}
	for {
		current := userIDCounter.Load()
		if current >= maxID || userIDCounter.CompareAndSwap(current, maxID) {
			break
		}
	}
	return db.ensurePredefinedRoles()
}

// removeInvalidRoleReferences removes authorization records that refer to roles absent from the database.
func (db *Database) removeInvalidRoleReferences() {
	for key, value := range db.databasePrivileges.Data {
		if _, ok := db.rolesByID[key.Role]; !ok || removeInvalidPrivilegeGrants(db.rolesByID, value.Privileges) {
			delete(db.databasePrivileges.Data, key)
		}
	}
	for key, value := range db.schemaPrivileges.Data {
		if _, ok := db.rolesByID[key.Role]; !ok || removeInvalidPrivilegeGrants(db.rolesByID, value.Privileges) {
			delete(db.schemaPrivileges.Data, key)
		}
	}
	for key, value := range db.tablePrivileges.Data {
		if _, ok := db.rolesByID[key.Role]; !ok || removeInvalidPrivilegeGrants(db.rolesByID, value.Privileges) {
			delete(db.tablePrivileges.Data, key)
		}
	}
	for key, value := range db.sequencePrivileges.Data {
		if _, ok := db.rolesByID[key.Role]; !ok || removeInvalidPrivilegeGrants(db.rolesByID, value.Privileges) {
			delete(db.sequencePrivileges.Data, key)
		}
	}
	for key, value := range db.routinePrivileges.Data {
		if _, ok := db.rolesByID[key.Role]; !ok || removeInvalidPrivilegeGrants(db.rolesByID, value.Privileges) {
			delete(db.routinePrivileges.Data, key)
		}
	}
	for member, groups := range db.roleMembership.Data {
		if _, ok := db.rolesByID[member]; !ok {
			delete(db.roleMembership.Data, member)
			continue
		}
		for group, membership := range groups {
			_, groupExists := db.rolesByID[group]
			_, grantorExists := db.rolesByID[membership.GrantedBy]
			if !groupExists || !grantorExists {
				delete(groups, group)
			}
		}
		if len(groups) == 0 {
			delete(db.roleMembership.Data, member)
		}
	}
}

// removeInvalidPrivilegeGrants removes grants made by nonexistent roles and reports whether the map is empty.
func removeInvalidPrivilegeGrants(roles map[RoleID]Role, privileges map[Privilege]map[GrantedPrivilege]bool) bool {
	for privilege, grants := range privileges {
		for grant := range grants {
			if _, ok := roles[grant.GrantedBy]; !ok {
				delete(grants, grant)
			}
		}
		if len(grants) == 0 {
			delete(privileges, privilege)
		}
	}
	return len(privileges) == 0
}

// deserializeVersion reads auth state, retaining older privilege layouts while
// reading the persisted built-in marker only from version 2 role records.
func (db *Database) deserializeVersion(reader *utils.Reader, version uint32) error {
	// Read the roles
	clear(db.rolesByName)
	clear(db.rolesByID)
	roleCount := reader.Uint32()
	for i := uint32(0); i < roleCount; i++ {
		r := Role{}
		r.deserialize(version, reader)
		if !r.IsValid() {
			return errors.New("invalid role ID in auth database")
		}
		if _, exists := db.rolesByName[r.Name]; exists {
			return errors.Errorf(`duplicate role name "%s" in auth database`, r.Name)
		}
		if _, exists := db.rolesByID[r.id]; exists {
			return errors.Errorf("duplicate role ID %d in auth database", r.id)
		}
		db.rolesByName[r.Name] = r.id
		db.rolesByID[r.id] = r
	}
	// Read the database privileges
	// Only role records changed in version 2.
	privilegeVersion := version
	if privilegeVersion > 1 {
		privilegeVersion = 1
	}
	db.databasePrivileges.deserialize(privilegeVersion, reader)
	// Read the schema privileges
	db.schemaPrivileges.deserialize(privilegeVersion, reader)
	// Read the table privileges
	db.tablePrivileges.deserialize(privilegeVersion, reader)
	// Read the sequence privileges
	db.sequencePrivileges.deserialize(privilegeVersion, reader)
	// Read the routine privileges
	db.routinePrivileges.deserialize(privilegeVersion, reader)
	// Read the role membership
	db.roleMembership.deserialize(privilegeVersion, reader)
	return nil
}
