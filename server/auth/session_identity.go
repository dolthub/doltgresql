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
	"github.com/cockroachdb/errors"
	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/dsess"
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/core/sessionstate"
)

// InitializeSessionIdentity is the explicit connection/bootstrap entry point.
// It snapshots only the authentication-time superuser entitlement. Call it
// after authentication and before executing any SQL on the session.
func InitializeSessionIdentity(sess sql.Session, name string) error {
	var role Role
	var ok bool
	LockRead(func() { role, ok = LookupRole(name) })
	if !ok {
		return errors.Errorf("authenticated role %q does not exist", name)
	}
	if err := core.InitializeIdentityOnSession(sess, sessionstate.RoleID(role.ID()), role.IsSuperUser); err != nil {
		return err
	}
	InstallDoltPrincipalProvider(sess)
	return nil
}

// InstallDoltPrincipalProvider makes Dolt's branch-control user follow the
// effective SQL role. The connection's Client.User remains the login principal.
func InstallDoltPrincipalProvider(sess sql.Session) {
	doltSess, ok := sess.(*dsess.DoltSession)
	if !ok {
		return
	}
	doltSess.DoltgresPrincipalProvider = func() string {
		identity, err := core.IdentityFromSession(sess)
		if err != nil {
			return ""
		}
		name, _ := RoleNameForSession(RoleID(identity.CurrentRole()))
		return name
	}
}

// CurrentRoleLocked resolves the acting role by stable ID. The caller must
// already hold the auth read or write lock.
func CurrentRoleLocked(ctx *sql.Context) (Role, error) {
	identity, err := core.Identity(ctx)
	if err != nil {
		return Role{}, err
	}
	role, ok := LookupRoleByID(RoleID(identity.CurrentRole()))
	if !ok {
		return Role{}, errors.Errorf("role with ID %d no longer exists", identity.CurrentRole())
	}
	return role, nil
}

// CurrentRole resolves the acting role for an ordinary SQL authorization
// decision. Call CurrentRoleLocked instead when already holding the auth lock.
func CurrentRole(ctx *sql.Context) (Role, error) {
	var role Role
	var err error
	LockRead(func() { role, err = CurrentRoleLocked(ctx) })
	return role, err
}

// ResolveRoleID reads the current role record. It acquires the auth read lock;
// callers already inside LockRead/LockWrite must use LookupRoleByID directly.
func ResolveRoleID(id sessionstate.RoleID) (Role, error) {
	var role Role
	var ok bool
	LockRead(func() { role, ok = LookupRoleByID(RoleID(id)) })
	if !ok {
		return Role{}, errors.Errorf("role with ID %d no longer exists", id)
	}
	return role, nil
}
