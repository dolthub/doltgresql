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
	"sync"
	"testing"

	"github.com/dolthub/doltgresql/core/sessionstate"
)

func TestResolveRoleIDAfterRenameAndReuse(t *testing.T) {
	oldDB, oldLock := globalDatabase, globalLock
	globalDatabase, globalLock = newEmptyDatabase(), &sync.RWMutex{}
	t.Cleanup(func() {
		globalDatabase, globalLock = oldDB, oldLock
		publishRoleNames()
	})
	LockWrite(func() { SetRole(Role{Name: "first", id: 101}) })
	role, err := ResolveRoleID(sessionstate.RoleID(101))
	if err != nil || role.Name != "first" {
		t.Fatalf("initial lookup: %v, %v", role.Name, err)
	}
	LockWrite(func() { RenameRole("first", "renamed") })
	role, err = ResolveRoleID(sessionstate.RoleID(101))
	if err != nil || role.Name != "renamed" {
		t.Fatalf("renamed lookup: %v, %v", role.Name, err)
	}
	if name, ok := RoleNameForSession(101); !ok || name != "renamed" {
		t.Fatalf("role-name view after rename = %q, %t", name, ok)
	}
	LockWrite(func() {
		DropRole("renamed")
		SetRole(Role{Name: "renamed", id: 102})
	})
	if _, err = ResolveRoleID(sessionstate.RoleID(101)); err == nil {
		t.Fatal("dropped role ID resolved to replacement")
	}
	if _, ok := RoleNameForSession(101); ok {
		t.Fatal("role-name view resolved a dropped ID")
	}
}
