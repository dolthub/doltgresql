package sessionstate

import "testing"

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
