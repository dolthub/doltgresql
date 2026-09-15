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
	"testing"

	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
)

// TestDeserializeRemovesInvalidRoleReferences verifies that upgrades sanitize authorization data written by affected releases.
func TestDeserializeRemovesInvalidRoleReferences(t *testing.T) {
	originalDatabase := globalDatabase
	t.Cleanup(func() { globalDatabase = originalDatabase })
	globalDatabase = newEmptyDatabase()

	survivor := Role{Name: "survivor", InheritPrivileges: true, id: 1}
	orphan := Role{Name: "orphan", InheritPrivileges: true, id: 2}
	otherGroup := Role{Name: "other_group", InheritPrivileges: true, id: 3}
	SetRole(survivor)
	SetRole(orphan)
	SetRole(otherGroup)

	orphanGrant := GrantedPrivilege{Privilege: Privilege_SELECT, GrantedBy: orphan.ID()}
	survivorGrant := GrantedPrivilege{Privilege: Privilege_SELECT, GrantedBy: survivor.ID()}
	databaseKey := DatabasePrivilegeKey{Role: orphan.ID(), Name: "database"}
	schemaKey := SchemaPrivilegeKey{Role: orphan.ID(), Schema: "schema"}
	tableKey := TablePrivilegeKey{Role: orphan.ID(), Table: doltdb.TableName{Schema: "schema", Name: "table"}}
	sequenceKey := SequencePrivilegeKey{Role: orphan.ID(), Schema: "schema", Name: "sequence"}
	routineKey := RoutinePrivilegeKey{Role: orphan.ID(), Schema: "schema", Name: "routine"}
	AddDatabasePrivilege(databaseKey, survivorGrant, false)
	AddSchemaPrivilege(schemaKey, survivorGrant, false)
	AddTablePrivilege(tableKey, survivorGrant, false)
	AddSequencePrivilege(sequenceKey, survivorGrant, false)
	AddRoutinePrivilege(routineKey, survivorGrant, false)

	survivorDatabaseKey := DatabasePrivilegeKey{Role: survivor.ID(), Name: "granted_database"}
	survivorSchemaKey := SchemaPrivilegeKey{Role: survivor.ID(), Schema: "granted_schema"}
	survivorTableKey := TablePrivilegeKey{Role: survivor.ID(), Table: doltdb.TableName{Schema: "schema", Name: "granted_table"}}
	survivorSequenceKey := SequencePrivilegeKey{Role: survivor.ID(), Schema: "schema", Name: "granted_sequence"}
	survivorRoutineKey := RoutinePrivilegeKey{Role: survivor.ID(), Schema: "schema", Name: "granted_routine"}
	mixedDatabaseKey := DatabasePrivilegeKey{Role: survivor.ID(), Name: "mixed_database"}
	AddDatabasePrivilege(survivorDatabaseKey, orphanGrant, false)
	AddSchemaPrivilege(survivorSchemaKey, orphanGrant, false)
	AddTablePrivilege(survivorTableKey, orphanGrant, false)
	AddSequencePrivilege(survivorSequenceKey, orphanGrant, false)
	AddRoutinePrivilege(survivorRoutineKey, orphanGrant, false)
	AddDatabasePrivilege(mixedDatabaseKey, orphanGrant, false)
	AddDatabasePrivilege(mixedDatabaseKey, survivorGrant, false)

	AddMemberToGroup(orphan.ID(), otherGroup.ID(), false, survivor.ID())
	AddMemberToGroup(survivor.ID(), orphan.ID(), false, survivor.ID())
	AddMemberToGroup(survivor.ID(), otherGroup.ID(), false, orphan.ID())
	delete(globalDatabase.rolesByName, orphan.Name)
	delete(globalDatabase.rolesByID, orphan.ID())

	loaded := newEmptyDatabase()
	if err := loaded.deserialize(globalDatabase.serialize()); err != nil {
		t.Fatal(err)
	}
	globalDatabase = loaded

	assertNoPrivilege := func(name string, hasPrivilege bool) {
		t.Helper()
		if hasPrivilege {
			t.Errorf("orphaned role reference retained %s privilege", name)
		}
	}
	assertNoPrivilege("database grantee", HasDatabasePrivilege(databaseKey, Privilege_SELECT))
	assertNoPrivilege("schema grantee", HasSchemaPrivilege(schemaKey, Privilege_SELECT))
	assertNoPrivilege("table grantee", HasTablePrivilege(tableKey, Privilege_SELECT))
	assertNoPrivilege("sequence grantee", HasSequencePrivilege(sequenceKey, Privilege_SELECT))
	assertNoPrivilege("routine grantee", HasRoutinePrivilege(routineKey, Privilege_SELECT))
	assertNoPrivilege("database grantor", HasDatabasePrivilege(survivorDatabaseKey, Privilege_SELECT))
	assertNoPrivilege("schema grantor", HasSchemaPrivilege(survivorSchemaKey, Privilege_SELECT))
	assertNoPrivilege("table grantor", HasTablePrivilege(survivorTableKey, Privilege_SELECT))
	assertNoPrivilege("sequence grantor", HasSequencePrivilege(survivorSequenceKey, Privilege_SELECT))
	assertNoPrivilege("routine grantor", HasRoutinePrivilege(survivorRoutineKey, Privilege_SELECT))
	if !HasDatabasePrivilege(mixedDatabaseKey, Privilege_SELECT) {
		t.Error("sanitizing one grantor removed the surviving grant")
	}
	if group, _, _ := IsRoleAMember(orphan.ID(), otherGroup.ID()); group.IsValid() {
		t.Error("orphaned role retained membership as member")
	}
	if group, _, _ := IsRoleAMember(survivor.ID(), orphan.ID()); group.IsValid() {
		t.Error("orphaned role retained membership as group")
	}
	if group, _, _ := IsRoleAMember(survivor.ID(), otherGroup.ID()); group.IsValid() {
		t.Error("orphaned role retained membership as grantor")
	}
}
