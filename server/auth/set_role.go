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

// ApplySetRole is the checked entry point for all supported SET ROLE forms.
// A selected role never becomes the authority for a subsequent switch.
func ApplySetRole(ctx *sql.Context, name string, none, reset, local bool) error {
	identity, err := core.Identity(ctx)
	if err != nil {
		return err
	}
	var target sessionstate.RoleID
	LockRead(func() {
		if _, ok := LookupRoleByID(RoleID(identity.SessionRole())); !ok {
			err = errors.Errorf("role with ID %d no longer exists", identity.SessionRole())
			return
		}
		if none || reset {
			return
		}
		role, ok := LookupRole(name)
		if !ok {
			err = pgerror.Newf(pgcode.UndefinedObject, `role "%s" does not exist`, name)
			return
		}
		if !CanSetRole(RoleID(identity.SessionRole()), role.ID()) {
			err = pgerror.Newf(pgcode.InsufficientPrivilege, `permission denied to set role "%s"`, name)
			return
		}
		target = sessionstate.RoleID(role.ID())
	})
	if err != nil {
		return err
	}
	return core.ApplyAuthorizedIdentityChange(ctx, local, func(next *sessionstate.Identity) error {
		if reset {
			next.ResetRole()
		} else {
			next.SelectRole(target)
		}
		return nil
	})
}

// SelectedRoleSetting projects the role GUC from typed identity state.
// NONE is a real selection state, distinct from selecting the session role.
func SelectedRoleSetting(ctx *sql.Context) (string, error) {
	identity, err := core.Identity(ctx)
	if err != nil {
		return "", err
	}
	id, selected := identity.SelectedRole()
	if !selected {
		return "none", nil
	}
	role, err := ResolveRoleID(id)
	if err != nil {
		return "", err
	}
	return role.Name, nil
}
