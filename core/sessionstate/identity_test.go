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
