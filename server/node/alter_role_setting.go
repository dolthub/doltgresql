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

package node

import (
	"context"
	"strings"

	"github.com/dolthub/dolt/go/libraries/doltcore/doltdb"
	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/dsess"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/plan"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/core/sessionstate"
	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
	"github.com/dolthub/doltgresql/server/auth"
	"github.com/dolthub/doltgresql/server/config"
)

// AlterRoleSetting handles ALTER ROLE ... SET/RESET and ALTER DATABASE ... SET/RESET, which set the configuration
// parameter defaults that are applied when a session starts.
type AlterRoleSetting struct {
	// Role is the role that the settings apply to. This may also be current_user, current_role, or session_user. This
	// is ignored when AllRoles is true.
	Role string
	// AllRoles applies the settings to every role. This is true for ALTER ROLE ALL and ALTER DATABASE, as ALTER
	// DATABASE d is equivalent to ALTER ROLE ALL IN DATABASE d.
	AllRoles bool
	// Database is the database that the settings apply to. An empty string applies the settings to every database.
	Database string
	// Name is the configuration parameter's name. This is ignored when ResetAll is true.
	Name string
	// Value is the configuration parameter's value. This is ignored when Reset, ResetAll, or FromCurrent is true.
	Value string
	// FromCurrent uses the session's current value of the parameter.
	FromCurrent bool
	// Reset removes the parameter.
	Reset bool
	// ResetAll removes every parameter.
	ResetAll bool
}

var _ sql.ExecSourceRel = (*AlterRoleSetting)(nil)
var _ vitess.Injectable = (*AlterRoleSetting)(nil)

// Children implements the interface sql.ExecSourceRel.
func (c *AlterRoleSetting) Children() []sql.Node {
	return nil
}

// IsReadOnly implements the interface sql.ExecSourceRel.
func (c *AlterRoleSetting) IsReadOnly() bool {
	return false
}

// Resolved implements the interface sql.ExecSourceRel.
func (c *AlterRoleSetting) Resolved() bool {
	return true
}

// RowIter implements the interface sql.ExecSourceRel.
func (c *AlterRoleSetting) RowIter(ctx *sql.Context, r sql.Row) (sql.RowIter, error) {
	var userRole auth.Role
	var role auth.Role
	var roleExists bool
	var roleErr error
	auth.LockRead(func() {
		userRole, roleErr = auth.CurrentRoleLocked(ctx)
		if roleErr != nil || c.AllRoles {
			return
		}
		switch c.Role {
		case "current_user", "current_role":
			role, roleExists = userRole, true
		case "session_user":
			var identity sessionstate.IdentitySnapshot
			if identity, roleErr = core.Identity(ctx); roleErr == nil {
				role, roleExists = auth.LookupRoleByID(auth.RoleID(identity.SessionRole()))
			}
		default:
			role, roleExists = auth.LookupRole(c.Role)
		}
	})
	if roleErr != nil {
		return nil, roleErr
	}
	if !c.AllRoles && !roleExists {
		return nil, pgerror.Newf(pgcode.UndefinedObject, `role "%s" does not exist`, c.Role)
	}
	if len(c.Database) > 0 && !dsess.DSessFromSess(ctx.Session).Provider().HasDatabase(ctx, c.Database) {
		return nil, pgerror.Newf(pgcode.InvalidCatalogName, `database "%s" does not exist`, c.Database)
	}
	if err := c.checkPermissions(userRole, role); err != nil {
		return nil, err
	}

	key := auth.RoleSettingKey{Role: role.ID(), Database: c.Database}
	name := c.Name
	value := c.Value
	if !c.ResetAll {
		var superUserOnly bool
		var err error
		name, value, superUserOnly, err = c.validateParameter(ctx)
		if err != nil {
			return nil, err
		}
		// Superusers may set any parameter, and others may only set parameters that do not require a superuser
		if superUserOnly && !userRole.IsSuperUser {
			return nil, pgerror.Newf(pgcode.InsufficientPrivilege, `permission denied to set parameter "%s"`, name)
		}
	}

	var err error
	var rsc doltdb.ReplicationStatusController
	auth.LockWrite(func() {
		switch {
		case c.ResetAll:
			auth.ResetAllRoleSettings(key)
		case c.Reset:
			auth.ResetRoleSetting(key, name)
		default:
			auth.SetRoleSetting(key, name, value)
		}
		err = auth.PersistChanges(ctx, &rsc)
	})
	if err != nil {
		return nil, err
	}
	auth.WaitForReplication(ctx, rsc)
	return sql.RowsToRowIter(), nil
}

// checkPermissions returns an error if the user role may not alter the settings of the target role or database. These
// checks match Postgres.
func (c *AlterRoleSetting) checkPermissions(userRole auth.Role, role auth.Role) error {
	if userRole.IsSuperUser {
		return nil
	}
	if c.AllRoles {
		if len(c.Database) > 0 {
			// Doltgres does not support database ownership, so only superusers are treated as owners
			return pgerror.Newf(pgcode.InsufficientPrivilege, "must be owner of database %s", c.Database)
		}
		return pgerror.New(pgcode.InsufficientPrivilege, "permission denied to alter setting")
	}
	if role.IsSuperUser {
		return pgerror.New(pgcode.InsufficientPrivilege, "must be superuser to alter superusers")
	}
	if !userRole.CanCreateRoles && role.ID() != userRole.ID() {
		return pgerror.Newf(pgcode.InsufficientPrivilege, `permission denied to alter role "%s"`, role.Name)
	}
	return nil
}

// validateParameter validates the parameter's name and value, returning the name and value to store, and whether
// only superusers may set the parameter.
func (c *AlterRoleSetting) validateParameter(ctx *sql.Context) (name string, value string, superUserOnly bool, err error) {
	name = c.Name
	value = c.Value
	if c.FromCurrent {
		value, err = currentSettingText(ctx, name)
		if err != nil {
			return "", "", false, err
		}
	}
	if config.IsValidPostgresConfigParameter(name) {
		if c.Reset {
			// Stored names are matched case-insensitively, so there's nothing else to validate when removing
			return name, "", false, nil
		}
		validatedName, superUserOnly, err := config.ValidatePostgresParameterValue(ctx, name, value)
		if err != nil {
			return "", "", false, err
		}
		return validatedName, value, superUserOnly, nil
	}
	if config.IsValidCustomParameterName(name) {
		return strings.ToLower(name), value, false, nil
	}
	if strings.Contains(name, ".") {
		return "", "", false, pgerror.Newf(pgcode.InvalidName, `invalid configuration parameter name "%s"`, name)
	}
	return "", "", false, pgerror.Newf(pgcode.UndefinedObject, `unrecognized configuration parameter "%s"`, name)
}

// currentSettingText returns the session's current value of the parameter, as text.
func currentSettingText(ctx *sql.Context, name string) (string, error) {
	if value, ok, err := core.Setting(ctx, name); err != nil {
		return "", err
	} else if ok {
		return config.FormatPostgresParameterValue(name, value), nil
	}
	if config.IsValidPostgresConfigParameter(name) {
		value, err := ctx.GetSessionVariable(ctx, name)
		if err != nil {
			return "", err
		}
		return config.FormatPostgresParameterValue(name, value), nil
	}
	return "", pgerror.Newf(pgcode.UndefinedObject, `unrecognized configuration parameter "%s"`, name)
}

// Schema implements the interface sql.ExecSourceRel.
func (c *AlterRoleSetting) Schema(ctx *sql.Context) sql.Schema {
	return nil
}

// String implements the interface sql.ExecSourceRel.
func (c *AlterRoleSetting) String() string {
	if c.AllRoles && len(c.Database) > 0 {
		return "ALTER DATABASE"
	}
	return "ALTER ROLE"
}

// WithChildren implements the interface sql.ExecSourceRel.
func (c *AlterRoleSetting) WithChildren(ctx *sql.Context, children ...sql.Node) (sql.Node, error) {
	return plan.NillaryWithChildren(c, children...)
}

// WithResolvedChildren implements the interface vitess.Injectable.
func (c *AlterRoleSetting) WithResolvedChildren(ctx context.Context, children []any) (any, error) {
	if len(children) != 0 {
		return nil, ErrVitessChildCount.New(0, len(children))
	}
	return c, nil
}
