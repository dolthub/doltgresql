package sessionstate

import (
	"errors"
	"testing"
)

func TestIdentitySelections(t *testing.T) {
	a := NewIdentity(11, true)
	b := NewIdentity(22, false)
	if !a.Initialized() || a.AuthenticatedRole() != 11 || a.SessionRole() != 11 || a.CurrentRole() != 11 || !a.AuthenticatedSuperuser() {
		t.Fatal("incorrect login identity")
	}
	a.SelectRole(33)
	if a.CurrentRole() != 33 || a.SessionRole() != 11 || b.CurrentRole() != 22 {
		t.Fatal("role selection changed session or another connection")
	}
	a.ResetRole()
	if a.CurrentRole() != 11 {
		t.Fatal("role reset did not restore session role")
	}
	a.SetSessionRole(44)
	a.SelectRole(55)
	a.ResetSessionRole()
	if a.SessionRole() != 11 || a.CurrentRole() != 11 {
		t.Fatal("session reset did not restore login defaults")
	}
}

func TestScopedExecutionRoleRestored(t *testing.T) {
	id := NewIdentity(11, true)
	id.SelectRole(22)
	want := errors.New("execution failed")
	err := id.WithExecutionRole(33, func() error {
		if !id.InScopedExecution() || id.CurrentRole() != 33 || id.SessionRole() != 11 {
			t.Fatal("scoped role was not active")
		}
		return id.WithExecutionRole(44, func() error {
			if id.CurrentRole() != 44 {
				t.Fatal("nested role was not active")
			}
			return want
		})
	})
	if !errors.Is(err, want) || id.InScopedExecution() || id.CurrentRole() != 22 {
		t.Fatalf("scoped role leaked after error: role %d, error %v", id.CurrentRole(), err)
	}
}
