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

// Package sessionstate holds durable, auth-independent PostgreSQL session state.
package sessionstate

// RoleID is a stable role identifier within the running server. Auth owns the
// mapping between IDs and role records.
type RoleID uint64

// Identity keeps the connection principal separate from the session and
// effective principals. Its zero value is uninitialized.
type Identity struct {
	authenticated          RoleID
	authenticatedSuperuser bool
	session                RoleID
	selected               RoleID // zero means NONE: use the session role
	resetSession           RoleID
	resetSelected          RoleID
	executionRole          RoleID
	executionDepth         int
}

// IdentitySnapshot is a read-only copy of a session's identity at one point in
// time. Callers must read the session again to observe a later role change.
type IdentitySnapshot struct{ identity Identity }

func (i Identity) Snapshot() IdentitySnapshot { return IdentitySnapshot{identity: i} }

func (s IdentitySnapshot) Initialized() bool            { return s.identity.Initialized() }
func (s IdentitySnapshot) AuthenticatedRole() RoleID    { return s.identity.AuthenticatedRole() }
func (s IdentitySnapshot) AuthenticatedSuperuser() bool { return s.identity.AuthenticatedSuperuser() }
func (s IdentitySnapshot) SessionRole() RoleID          { return s.identity.SessionRole() }
func (s IdentitySnapshot) SelectedRole() (RoleID, bool) { return s.identity.SelectedRole() }
func (s IdentitySnapshot) CurrentRole() RoleID          { return s.identity.CurrentRole() }
func (s IdentitySnapshot) InScopedExecution() bool      { return s.identity.InScopedExecution() }

func NewIdentity(authenticated RoleID, superuser bool) Identity {
	return Identity{authenticated: authenticated, authenticatedSuperuser: superuser,
		session: authenticated, resetSession: authenticated}
}

func (i Identity) Initialized() bool            { return i.authenticated != 0 }
func (i Identity) AuthenticatedRole() RoleID    { return i.authenticated }
func (i Identity) AuthenticatedSuperuser() bool { return i.authenticatedSuperuser }
func (i Identity) SessionRole() RoleID          { return i.session }
func (i Identity) SelectedRole() (RoleID, bool) { return i.selected, i.selected != 0 }
func (i Identity) CurrentRole() RoleID {
	if i.executionDepth != 0 {
		return i.executionRole
	}
	if i.selected != 0 {
		return i.selected
	}
	return i.session
}

// InScopedExecution identifies a future SECURITY DEFINER style override.
// Identity-changing SQL is prohibited while such an override is active.
func (i Identity) InScopedExecution() bool { return i.executionDepth != 0 }

// WithExecutionRole is an internal execution boundary. The override is restored
// on normal return, error, cancellation, or panic. Callers must first resolve
// and authorize the target role.
func (i *Identity) WithExecutionRole(target RoleID, run func() error) error {
	previous, depth := i.executionRole, i.executionDepth
	i.executionRole, i.executionDepth = target, depth+1
	defer func() { i.executionRole, i.executionDepth = previous, depth }()
	return run()
}

// SelectRole applies a selection after the caller has checked SET permission.
// A zero target selects NONE. SQL entry points must use an auth checked adapter.
func (i *Identity) SelectRole(target RoleID) { i.selected = target }

// ResetRole restores the connection's role-selection default.
func (i *Identity) ResetRole() { i.selected = i.resetSelected }

// SetSessionRole applies a session-authorization change after the caller has
// checked authority. A session change also clears an explicit role selection.
func (i *Identity) SetSessionRole(target RoleID) {
	i.session = target
	i.selected = 0
}

// ResetSessionRole restores the connection's initial session authorization.
func (i *Identity) ResetSessionRole() {
	i.session = i.resetSession
	i.selected = i.resetSelected
}
