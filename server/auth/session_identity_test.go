package auth

import (
	"sync"
	"testing"

	"github.com/dolthub/doltgresql/core/sessionstate"
)

func TestResolveRoleIDAfterRenameAndReuse(t *testing.T) {
	oldDB, oldLock := globalDatabase, globalLock
	globalDatabase, globalLock = newEmptyDatabase(), &sync.RWMutex{}
	t.Cleanup(func() { globalDatabase, globalLock = oldDB, oldLock })
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
	LockWrite(func() {
		DropRole("renamed")
		SetRole(Role{Name: "renamed", id: 102})
	})
	if _, err = ResolveRoleID(sessionstate.RoleID(101)); err == nil {
		t.Fatal("dropped role ID resolved to replacement")
	}
}
