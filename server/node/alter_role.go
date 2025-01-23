// Copyright 2024 Dolthub, Inc.
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

package node

import (
	"time"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/plan"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"
	gmserrors "gopkg.in/src-d/go-errors.v1"

	"github.com/dolthub/doltgresql/server/auth"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// AlterRole handles the ALTER ROLE and ALTER USER statements (ALTER USER is an alias).
type AlterRole struct {
	Name    string
	Options map[string]any
}

// ErrVitessChildCount is returned by WithResolvedChildren to indicate that the expected child count is incorrect.
var ErrVitessChildCount = gmserrors.NewKind("invalid vitess child count, expected `%d` but got `%d`")

var _ sql.ExecSourceRel = (*AlterRole)(nil)
var _ vitess.Injectable = (*AlterRole)(nil)

// Children implements the interface sql.ExecSourceRel.
func (c *AlterRole) Children() []sql.Node {
	return nil
}

// IsReadOnly implements the interface sql.ExecSourceRel.
func (c *AlterRole) IsReadOnly() bool {
	return false
}

// Resolved implements the interface sql.ExecSourceRel.
func (c *AlterRole) Resolved() bool {
	return true
}

// RowIter implements the interface sql.ExecSourceRel.
func (c *AlterRole) RowIter(ctx *sql.Context, r sql.Row) (sql.RowIter, error) {
	var userRole auth.Role
	var role auth.Role
	auth.LockRead(func() {
		userRole = auth.GetRole(ctx.Client().User)
		role = auth.GetRole(c.Name)
	})
	if !userRole.IsValid() {
		return nil, errors.Errorf(`role "%s" does not exist`, userRole.Name)
	}
	if !role.IsValid() {
		return nil, errors.Errorf(`role "%s" does not exist`, c.Name)
	}

	if role.IsSuperUser && !userRole.IsSuperUser {
		// Only superusers can modify other superusers
		// TODO: grab the actual error message
		return nil, errors.Errorf(`role "%s" does not have permission to alter role "%s"`, userRole.Name, role.Name)
	} else if !userRole.IsSuperUser && !userRole.CanCreateRoles && role.ID() != userRole.ID() {
		// A role may only modify itself if it doesn't have the ability to create roles
		// TODO: allow non-role-creating roles to only modify their own password, and grab actual error message
		return nil, errors.Errorf(`role "%s" does not have permission to alter role "%s"`, userRole.Name, role.Name)
	}
	for optionName, optionValue := range c.Options {
		switch optionName {
		case "BYPASSRLS":
			if !userRole.IsSuperUser {
				// TODO: grab the actual error message
				return nil, errors.Errorf(`role "%s" does not have permission to alter role "%s"`, userRole.Name, role.Name)
			}
			role.CanBypassRowLevelSecurity = true
		case "CONNECTION_LIMIT":
			role.ConnectionLimit = optionValue.(int32)
		case "CREATEDB":
			role.CanCreateDB = true
		case "CREATEROLE":
			role.CanCreateRoles = true
		case "INHERIT":
			role.InheritPrivileges = true
		case "LOGIN":
			role.CanLogin = true
		case "NOBYPASSRLS":
			if !userRole.IsSuperUser {
				// TODO: grab the actual error message
				return nil, errors.Errorf(`role "%s" does not have permission to alter role "%s"`, userRole.Name, role.Name)
			}
			role.CanBypassRowLevelSecurity = false
		case "NOCREATEDB":
			role.CanCreateDB = false
		case "NOCREATEROLE":
			role.CanCreateRoles = false
		case "NOINHERIT":
			role.InheritPrivileges = false
		case "NOLOGIN":
			role.CanLogin = false
		case "NOREPLICATION":
			if !userRole.IsSuperUser {
				// TODO: grab the actual error message
				return nil, errors.Errorf(`role "%s" does not have permission to alter role "%s"`, userRole.Name, role.Name)
			}
			role.IsReplicationRole = false
		case "NOSUPERUSER":
			if !userRole.IsSuperUser {
				// TODO: grab the actual error message
				return nil, errors.Errorf(`role "%s" does not have permission to alter role "%s"`, userRole.Name, role.Name)
			}
			role.IsSuperUser = false
		case "PASSWORD":
			password, _ := optionValue.(*string)
			if password == nil {
				role.Password = nil
			} else {
				var err error
				role.Password, err = auth.NewScramSha256Password(*password)
				if err != nil {
					return nil, err
				}
			}
		case "REPLICATION":
			if !userRole.IsSuperUser {
				// TODO: grab the actual error message
				return nil, errors.Errorf(`role "%s" does not have permission to alter role "%s"`, userRole.Name, role.Name)
			}
			role.IsReplicationRole = true
		case "SUPERUSER":
			if !userRole.IsSuperUser {
				// TODO: grab the actual error message
				return nil, errors.Errorf(`role "%s" does not have permission to alter role "%s"`, userRole.Name, role.Name)
			}
			role.IsSuperUser = true
		case "VALID_UNTIL":
			timeString, _ := optionValue.(*string)
			if timeString == nil {
				role.ValidUntil = nil
			} else {
				validUntilAny, err := pgtypes.TimestampTZ.IoInput(ctx, *timeString)
				if err != nil {
					return nil, err
				}
				validUntilTime := validUntilAny.(time.Time)
				role.ValidUntil = &validUntilTime
			}
		default:
			return nil, errors.Errorf(`unknown role option "%s"`, optionName)
		}
	}
	var err error
	auth.LockWrite(func() {
		auth.SetRole(role)
		err = auth.PersistChanges()
	})
	if err != nil {
		return nil, err
	}
	return sql.RowsToRowIter(), nil
}

// Schema implements the interface sql.ExecSourceRel.
func (c *AlterRole) Schema() sql.Schema {
	return nil
}

// String implements the interface sql.ExecSourceRel.
func (c *AlterRole) String() string {
	return "ALTER ROLE"
}

// WithChildren implements the interface sql.ExecSourceRel.
func (c *AlterRole) WithChildren(children ...sql.Node) (sql.Node, error) {
	return plan.NillaryWithChildren(c, children...)
}

// WithResolvedChildren implements the interface vitess.Injectable.
func (c *AlterRole) WithResolvedChildren(children []any) (any, error) {
	if len(children) != 0 {
		return nil, ErrVitessChildCount.New(0, len(children))
	}
	return c, nil
}
