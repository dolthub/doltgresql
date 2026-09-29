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
	"encoding/binary"
	"testing"

	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/dolthub/dolt/go/libraries/doltcore/env"
	"github.com/dolthub/dolt/go/libraries/utils/filesys"
	"github.com/stretchr/testify/require"

	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/utils"
)

// initPredefinedRoleTest isolates the process-wide auth state for unit tests.
func initPredefinedRoleTest(t *testing.T) {
	t.Helper()
	db, lock, fs, fileName, counter := globalDatabase, globalLock, fileSystem, authFileName, userIDCounter.Load()
	t.Cleanup(func() {
		globalDatabase, globalLock, fileSystem, authFileName = db, lock, fs, fileName
		userIDCounter.Store(counter)
	})
	Init(nil, nil)
}

// TestPredefinedRolePrivileges exercises inheritance barriers, additive grants,
// revocation, and the exact data-role capability boundary.
func TestPredefinedRolePrivileges(t *testing.T) {
	initPredefinedRoleTest(t)
	reader, writer := GetRole("pg_read_all_data"), GetRole("pg_write_all_data")
	for _, role := range []Role{reader, writer} {
		require.True(t, role.IsPredefined())
		require.True(t, role.InheritPrivileges)
		require.False(t, role.CanLogin || role.IsSuperUser || role.CanCreateRoles || role.CanCreateDB || role.IsReplicationRole || role.CanBypassRowLevelSecurity)
		require.Nil(t, role.Password)
		require.Nil(t, role.ValidUntil)
		require.EqualValues(t, -1, role.ConnectionLimit)
	}
	user, intermediate := CreateDefaultRole("reader"), CreateDefaultRole("intermediate")
	require.NoError(t, SetRole(user))
	require.NoError(t, SetRole(intermediate))
	grantor := GetRole("postgres").ID()
	AddMemberToGroup(user.ID(), intermediate.ID(), false, grantor)
	AddMemberToGroup(intermediate.ID(), reader.ID(), false, grantor)
	table := TablePrivilegeKey{Role: user.ID(), Table: doltdb.TableName{Schema: "future_schema", Name: "future_table"}}
	sequence := SequencePrivilegeKey{Role: user.ID(), Schema: "future_schema", Name: "future_seq"}
	schema := SchemaPrivilegeKey{Role: user.ID(), Schema: "future_schema"}
	require.True(t, HasTablePrivilege(table, Privilege_SELECT))
	require.True(t, HasSequencePrivilege(sequence, Privilege_SELECT))
	require.True(t, HasSchemaPrivilege(schema, Privilege_USAGE))
	for _, privilege := range []Privilege{Privilege_INSERT, Privilege_UPDATE, Privilege_DELETE, Privilege_TRUNCATE, Privilege_DROP, Privilege_REFERENCES, Privilege_TRIGGER} {
		require.False(t, HasTablePrivilege(table, privilege), privilege)
	}
	require.False(t, HasSequencePrivilege(sequence, Privilege_USAGE))
	require.False(t, HasSequencePrivilege(sequence, Privilege_UPDATE))
	require.False(t, HasSchemaPrivilege(schema, Privilege_CREATE))
	require.False(t, HasDatabasePrivilege(DatabasePrivilegeKey{Role: user.ID(), Name: "future_db"}, Privilege_CONNECT))
	require.Zero(t, HasTablePrivilegeGrantOption(table, Privilege_SELECT))
	require.Zero(t, HasSchemaPrivilegeGrantOption(schema, Privilege_USAGE))
	require.Zero(t, HasSequencePrivilegeGrantOption(sequence, Privilege_SELECT))
	require.False(t, CanAdministerRole(user.ID(), reader.ID()))

	intermediate.InheritPrivileges = false
	require.NoError(t, SetRole(intermediate))
	require.False(t, HasTablePrivilege(table, Privilege_SELECT))
	group, inherits, _ := IsRoleAMember(user.ID(), reader.ID())
	require.Equal(t, reader.ID(), group)
	require.False(t, inherits)
	AddMemberToGroup(user.ID(), reader.ID(), false, grantor)
	require.True(t, HasTablePrivilege(table, Privilege_SELECT), "a second valid path must remain effective")
	user.InheritPrivileges = false
	require.NoError(t, SetRole(user))
	require.False(t, HasTablePrivilege(table, Privilege_SELECT))
	user.InheritPrivileges = true
	require.NoError(t, SetRole(user))
	RemoveMemberFromGroup(user.ID(), reader.ID(), false)
	require.False(t, HasTablePrivilege(table, Privilege_SELECT))
	AddTablePrivilege(table, GrantedPrivilege{Privilege: Privilege_SELECT, GrantedBy: grantor}, false)
	require.True(t, HasTablePrivilege(table, Privilege_SELECT), "an independent grant survives membership revocation")

	AddMemberToGroup(user.ID(), writer.ID(), false, grantor)
	for _, privilege := range []Privilege{Privilege_INSERT, Privilege_UPDATE, Privilege_DELETE} {
		require.True(t, HasTablePrivilege(table, privilege))
	}
	require.True(t, HasSequencePrivilege(sequence, Privilege_UPDATE))
	require.False(t, HasSequencePrivilege(sequence, Privilege_USAGE))
	require.False(t, HasSequencePrivilege(sequence, Privilege_SELECT))
	require.Zero(t, HasSequencePrivilegeGrantOption(sequence, Privilege_UPDATE))
	require.EqualValues(t, 6181, id.Cache().ToOID(id.NewId(id.Section_User, reader.Name)))
	require.EqualValues(t, 6182, id.Cache().ToOID(id.NewId(id.Section_User, writer.Name)))
}

// legacyAuthData encodes the old role record layout for migration fixtures.
func legacyAuthData(db Database, version uint32) []byte {
	w := utils.NewWriter(1024)
	w.Uint32(version)
	w.Uint32(uint32(len(db.rolesByID)))
	for _, role := range db.rolesByID {
		rw := utils.NewWriter(128)
		role.serialize(rw)
		// The version 2 marker is the final byte of a role record.
		for _, b := range rw.Data()[:len(rw.Data())-1] {
			w.Uint8(b)
		}
	}
	db.databasePrivileges.serialize(w)
	db.schemaPrivileges.serialize(w)
	db.tablePrivileges.serialize(w)
	if version > 0 {
		db.sequencePrivileges.serialize(w)
		db.routinePrivileges.serialize(w)
	}
	db.roleMembership.serialize(w)
	return w.Data()
}

// TestPredefinedRoleUpgrade preserves user identities and grants, persists new
// built-ins before any query, and retains a recovery copy of the old auth file.
func TestPredefinedRoleUpgrade(t *testing.T) {
	for _, version := range []uint32{0, 1} {
		t.Run(string(rune('0'+version)), func(t *testing.T) {
			initPredefinedRoleTest(t)
			old := newEmptyDatabase()
			role := createDefaultRoleWithoutID("legacy")
			role.id = 10000
			role.CanLogin = true
			role.Password, _ = NewScramSha256Password("legacy-password")
			old.rolesByName[role.Name], old.rolesByID[role.id] = role.id, role
			member := createDefaultRoleWithoutID("legacy_member")
			member.id = 9999
			old.rolesByName[member.Name], old.rolesByID[member.id] = member.id, member
			old.roleMembership.Data[member.id] = map[RoleID]RoleMembershipValue{
				role.id: {Member: member.id, Group: role.id, GrantedBy: role.id, WithAdminOption: true},
			}
			key := TablePrivilegeKey{Role: role.id, Table: doltdb.TableName{Schema: "private", Name: "existing"}}
			old.tablePrivileges.Data[key] = TablePrivilegeValue{Key: key, Privileges: map[Privilege]map[GrantedPrivilege]bool{
				Privilege_SELECT: {{Privilege: Privilege_SELECT, GrantedBy: role.id}: true},
			}}
			data := legacyAuthData(old, version)
			fs, err := filesys.LocalFilesysWithWorkingDir(t.TempDir())
			require.NoError(t, err)
			require.NoError(t, fs.WriteFile("auth.db", data, 0600))
			authFileName = "auth.db"
			Init(&env.DoltEnv{FS: fs}, nil)
			loaded := GetRole(role.Name)
			require.Equal(t, role.id, loaded.ID())
			require.True(t, loaded.CanLogin)
			require.Equal(t, role.Password, loaded.Password)
			require.Equal(t, role.id, HasTablePrivilegeGrantOption(key, Privilege_SELECT))
			require.True(t, HasRoleAdminOption(member.id, role.id))
			reader := GetRole("pg_read_all_data")
			require.Greater(t, uint64(reader.ID()), uint64(role.ID()))
			AddMemberToGroup(role.ID(), reader.ID(), false, role.ID())
			require.NoError(t, WriteSerializedDatabase(globalDatabase.serialize()))
			Init(&env.DoltEnv{FS: fs}, nil)
			require.Equal(t, reader.ID(), GetRole(reader.Name).ID())
			require.True(t, HasTablePrivilege(TablePrivilegeKey{Role: role.ID(), Table: doltdb.TableName{Schema: "private", Name: "new_table"}}, Privilege_SELECT))
			persisted, err := fs.ReadFile("auth.db")
			require.NoError(t, err)
			require.EqualValues(t, 2, binary.BigEndian.Uint32(persisted))
			backup, err := fs.ReadFile("auth.db.v1.bak")
			require.NoError(t, err)
			require.Equal(t, data, backup)
		})
	}
}

// TestPredefinedRoleRejectsInvalidLoads ensures collisions cannot promote a
// user-created role and that malformed snapshots do not replace live state.
func TestPredefinedRoleRejectsInvalidLoads(t *testing.T) {
	initPredefinedRoleTest(t)
	reader := GetRole("pg_read_all_data")
	collision := newEmptyDatabase()
	role := CreateDefaultRole(reader.Name)
	collision.rolesByName[role.Name], collision.rolesByID[role.id] = role.id, role
	for _, data := range [][]byte{legacyAuthData(collision, 0), legacyAuthData(collision, 1), collision.serialize()} {
		require.ErrorContains(t, OverwriteDatabase(data), "conflicts with a predefined role")
		require.Equal(t, reader.ID(), GetRole(reader.Name).ID())
	}
	invalid := newEmptyDatabase()
	role.predefined = predefinedRoleReadAllData
	role.CanLogin = true
	invalid.rolesByName[role.Name], invalid.rolesByID[role.id] = role.id, role
	require.ErrorContains(t, OverwriteDatabase(invalid.serialize()), "invalid predefined role")
	data := globalDatabase.serialize()
	require.Error(t, OverwriteDatabase(data[:len(data)-1]))
	require.Equal(t, reader.ID(), GetRole(reader.Name).ID())
	require.NoError(t, OverwriteDatabase(data))
	require.Equal(t, reader.ID(), GetRole(reader.Name).ID())
}

// TestPredefinedRoleLifecycle protects canonical definitions even through the
// internal mutation API, while allowing grants and preventing membership cycles.
func TestPredefinedRoleLifecycle(t *testing.T) {
	initPredefinedRoleTest(t)
	reader := GetRole("pg_read_all_data")
	require.Error(t, DropRole(reader.Name))
	require.Error(t, RenameRole(reader.Name, "renamed"))
	changed := reader
	changed.CanLogin = true
	require.Error(t, SetRole(changed))
	require.Equal(t, reader, GetRole(reader.Name))
	a, b := CreateDefaultRole("a"), CreateDefaultRole("b")
	require.NoError(t, SetRole(a))
	require.NoError(t, SetRole(b))
	overwrite := a
	overwrite.Name = reader.Name
	require.Error(t, SetRole(overwrite))
	require.Equal(t, a, GetRole(a.Name), "a rejected overwrite must preserve the old name")
	require.Error(t, ValidateMembershipGrants([]RoleMembershipValue{{Member: a.ID(), Group: b.ID()}, {Member: b.ID(), Group: a.ID()}}))
	require.Empty(t, globalDatabase.roleMembership.Data)
	AddMemberToGroup(a.ID(), reader.ID(), true, b.ID())
	AddMemberToGroup(a.ID(), reader.ID(), false, reader.ID())
	require.True(t, HasRoleAdminOption(a.ID(), reader.ID()))
	require.Equal(t, b.ID(), globalDatabase.roleMembership.Data[a.ID()][reader.ID()].GrantedBy)
	RemoveMemberFromGroup(a.ID(), reader.ID(), true)
	require.False(t, HasRoleAdminOption(a.ID(), reader.ID()))
	require.True(t, HasRolePrivileges(a.ID(), reader.ID()))
}
