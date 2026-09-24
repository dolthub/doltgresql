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
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/core/sessionstate"
	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
)

// ApplySessionAuthorization checks against the authenticated connection, not
// the currently selected role. The authentication-time superuser entitlement
// is intentionally preserved when the session role changes.
func ApplySessionAuthorization(ctx *sql.Context, name string, reset, local bool) error {
	identity, err := core.Identity(ctx)
	if err != nil {
		return err
	}
	if identity.InScopedExecution() {
		return pgerror.New(pgcode.InsufficientPrivilege, "cannot set session authorization within security-definer context")
	}
	var target sessionstate.RoleID
	LockRead(func() {
		if _, ok := LookupRoleByID(RoleID(identity.AuthenticatedRole())); !ok {
			err = errors.Errorf("authenticated role with ID %d no longer exists", identity.AuthenticatedRole())
			return
		}
		if reset {
			target = identity.AuthenticatedRole()
			return
		}
		role, ok := LookupRole(name)
		if !ok {
			err = pgerror.Newf(pgcode.UndefinedObject, `role "%s" does not exist`, name)
			return
		}
		if !identity.AuthenticatedSuperuser() && RoleID(identity.AuthenticatedRole()) != role.ID() {
			err = pgerror.New(pgcode.InsufficientPrivilege, "permission denied to set session authorization")
			return
		}
		target = sessionstate.RoleID(role.ID())
	})
	if err != nil {
		return err
	}
	return core.ApplyAuthorizedIdentityChange(ctx, local, func(next *sessionstate.Identity) error {
		if reset {
			next.ResetSessionRole()
		} else {
			next.SetSessionRole(target)
		}
		return nil
	})
}

func SessionAuthorizationSetting(ctx *sql.Context) (string, error) {
	identity, err := core.Identity(ctx)
	if err != nil {
		return "", err
	}
	role, err := ResolveRoleID(identity.SessionRole())
	if err != nil {
		return "", err
	}
	return role.Name, nil
}
