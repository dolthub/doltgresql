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
	"errors"
	"fmt"
	"time"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/plan"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/server/auth"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

// CreateRole handles the CREATE ROLE and CREATE USER statements (CREATE USER is an alias).
type CreateRole struct {
	Name                      string
	IfNotExists               bool
	Password                  string // PASSWORD 'password'
	IsPasswordNull            bool   // PASSWORD NULL
	IsSuperUser               bool   // SUPERUSER | NOSUPERUSER
	CanCreateDB               bool   // CREATEDB | NOCREATEDB
	CanCreateRoles            bool   // CREATEROLE | NOCREATEROLE
	InheritPrivileges         bool   // INHERIT | NOINHERIT
	CanLogin                  bool   // LOGIN | NOLOGIN
	IsReplicationRole         bool   // REPLICATION | NOREPLICATION
	CanBypassRowLevelSecurity bool   // BYPASSRLS | NOBYPASSRLS
	ConnectionLimit           int32  // CONNECTION LIMIT connlimit
	ValidUntil                string // VALID UNTIL 'timestamp'
	IsValidUntilSet           bool
	AddToRoles                []string // IN ROLE role_name [, ...]
	AddAsMembers              []string // ROLE role_name [, ...]
	AddAsAdminMembers         []string // ADMIN role_name [, ...]
}

var _ sql.ExecSourceRel = (*CreateRole)(nil)
var _ vitess.Injectable = (*CreateRole)(nil)

// Children implements the interface sql.ExecSourceRel.
func (c *CreateRole) Children() []sql.Node {
	return nil
}

// IsReadOnly implements the interface sql.ExecSourceRel.
func (c *CreateRole) IsReadOnly() bool {
	return false
}

// Resolved implements the interface sql.ExecSourceRel.
func (c *CreateRole) Resolved() bool {
	return true
}

// RowIter implements the interface sql.ExecSourceRel.
func (c *CreateRole) RowIter(ctx *sql.Context, r sql.Row) (sql.RowIter, error) {
	if auth.RoleExists(c.Name) {
		if c.IfNotExists {
			return sql.RowsToRowIter(), nil
		}
		return nil, fmt.Errorf(`role "%s" already exists`, c.Name)
	}

	role := auth.CreateDefaultRole(c.Name)
	if !c.IsPasswordNull {
		password, err := auth.NewScramSha256Password(c.Password)
		if err != nil {
			return nil, err
		}
		role.Password = password
	}
	role.IsSuperUser = c.IsSuperUser
	role.CanCreateDB = c.CanCreateDB
	role.CanCreateRoles = c.CanCreateRoles
	role.InheritPrivileges = c.InheritPrivileges
	role.CanLogin = c.CanLogin
	role.IsReplicationRole = c.IsReplicationRole
	role.CanBypassRowLevelSecurity = c.CanBypassRowLevelSecurity
	role.ConnectionLimit = c.ConnectionLimit
	if c.IsValidUntilSet {
		validUntilAny, err := framework.IoInput(ctx, pgtypes.TimestampTZ, c.ValidUntil)
		if err != nil {
			return nil, err
		}
		validUntilTime := validUntilAny.(time.Time)
		role.ValidUntil = &validUntilTime
	}
	if len(c.AddToRoles) > 0 {
		return nil, errors.New("IN ROLE is not yet supported")
	}
	if len(c.AddAsMembers) > 0 {
		return nil, errors.New("ROLE is not yet supported")
	}
	if len(c.AddAsAdminMembers) > 0 {
		return nil, errors.New("ADMIN is not yet supported")
	}

	auth.SetRole(role)
	return sql.RowsToRowIter(), nil
}

// Schema implements the interface sql.ExecSourceRel.
func (c *CreateRole) Schema() sql.Schema {
	return nil
}

// String implements the interface sql.ExecSourceRel.
func (c *CreateRole) String() string {
	if c.CanLogin {
		return "CREATE USER"
	} else {
		return "CREATE ROLE"
	}
}

// WithChildren implements the interface sql.ExecSourceRel.
func (c *CreateRole) WithChildren(children ...sql.Node) (sql.Node, error) {
	return plan.NillaryWithChildren(c, children...)
}

// WithResolvedChildren implements the interface vitess.Injectable.
func (c *CreateRole) WithResolvedChildren(children []any) (any, error) {
	if len(children) != 0 {
		return nil, ErrVitessChildCount.New(0, len(children))
	}
	return c, nil
}
