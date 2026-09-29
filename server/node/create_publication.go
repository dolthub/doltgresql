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
	"fmt"
	"strings"

	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/dsess"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/plan"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"
	"github.com/jackc/pgx/v5/pgproto3"

	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/core/publications"
	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
	"github.com/dolthub/doltgresql/server/auth"
)

// CreatePublication creates publication metadata. Doltgres does not yet publish a
// logical replication stream; storing a publication does not enable replication.
type CreatePublication struct {
	Name      string
	AllTables bool
	Options   tree.KVOptions
}

var _ sql.ExecSourceRel = (*CreatePublication)(nil)
var _ vitess.Injectable = (*CreatePublication)(nil)

// Children implements sql.ExecSourceRel.
func (c *CreatePublication) Children() []sql.Node { return nil }

// IsReadOnly implements sql.ExecSourceRel.
func (c *CreatePublication) IsReadOnly() bool { return false }

// Resolved implements sql.ExecSourceRel.
func (c *CreatePublication) Resolved() bool { return true }

// Schema implements sql.ExecSourceRel.
func (c *CreatePublication) Schema(ctx *sql.Context) sql.Schema { return nil }

// String implements sql.ExecSourceRel.
func (c *CreatePublication) String() string { return fmt.Sprintf("CREATE PUBLICATION %s", c.Name) }

// WithChildren implements sql.ExecSourceRel.
func (c *CreatePublication) WithChildren(ctx *sql.Context, children ...sql.Node) (sql.Node, error) {
	return plan.NillaryWithChildren(c, children...)
}

// WithResolvedChildren implements vitess.Injectable.
func (c *CreatePublication) WithResolvedChildren(ctx context.Context, children []any) (any, error) {
	if len(children) != 0 {
		return nil, ErrVitessChildCount.New(0, len(children))
	}
	return c, nil
}

// RowIter validates privileges and options before updating the versioned collection.
func (c *CreatePublication) RowIter(ctx *sql.Context, r sql.Row) (sql.RowIter, error) {
	if tx := ctx.GetTransaction(); tx != nil && tx.IsReadOnly() {
		return nil, pgerror.New(pgcode.ReadOnlySQLTransaction, "cannot execute CREATE PUBLICATION in a read-only transaction")
	}
	var role auth.Role
	var canCreate bool
	auth.LockRead(func() {
		role = auth.GetRole(ctx.Client().User)
		public := auth.GetRole("public")
		canCreate = auth.HasDatabasePrivilege(auth.DatabasePrivilegeKey{Role: role.ID(), Name: ctx.GetCurrentDatabase()}, auth.Privilege_CREATE) ||
			auth.HasDatabasePrivilege(auth.DatabasePrivilegeKey{Role: public.ID(), Name: ctx.GetCurrentDatabase()}, auth.Privilege_CREATE)
	})
	if !canCreate {
		return nil, pgerror.Newf(pgcode.InsufficientPrivilege, "permission denied for database %s", ctx.GetCurrentDatabase())
	}
	if c.AllTables && !role.IsSuperUser {
		return nil, pgerror.New(pgcode.InsufficientPrivilege, "must be superuser to create FOR ALL TABLES publication")
	}
	collection, err := core.GetPublicationsCollectionFromContext(ctx, "")
	if err != nil {
		return nil, err
	}
	publicationID := id.NewPublication(c.Name)
	if collection.HasPublication(ctx, publicationID) {
		return nil, pgerror.Newf(pgcode.DuplicateObject, `publication "%s" already exists`, c.Name)
	}
	publication, err := publicationOptions(c.Options)
	if err != nil {
		return nil, err
	}
	publication.ID = publicationID
	publication.OwnerRoleID = uint64(role.ID())
	publication.AllTables = c.AllTables
	if err = collection.CreatePublication(ctx, publication); err != nil {
		return nil, err
	}
	// Be explicit about the current capability, even when wal_level is set to logical.
	dsess.DSessFromSess(ctx.Session).Notice(&pgproto3.NoticeResponse{
		Severity: "WARNING", Code: pgcode.FeatureNotSupported.String(),
		Message: "publications are stored as metadata; logical replication publishing is not supported",
	})
	return sql.RowsToRowIter(), nil
}

// publicationOptions follows PostgreSQL 15's publication option defaults and
// defGetBoolean semantics, which are stricter than a SQL boolean cast.
func publicationOptions(options tree.KVOptions) (publications.Publication, error) {
	p := publications.Publication{PublishInsert: true, PublishUpdate: true, PublishDelete: true, PublishTruncate: true}
	seen := make(map[tree.Name]bool)
	for _, option := range options {
		if seen[option.Key] {
			return p, pgerror.New(pgcode.Syntax, "conflicting or redundant options")
		}
		seen[option.Key] = true
		switch option.Key {
		case "publish":
			if option.Value == nil {
				return p, pgerror.New(pgcode.Syntax, "publish requires a parameter")
			}
			var value string
			switch v := option.Value.(type) {
			case *tree.StrVal:
				value = v.RawString()
			case *tree.NumVal:
				value = v.FormattedString()
			case *tree.DBool:
				value = v.String()
			}
			actions, err := publicationActions(value)
			if err != nil {
				return p, err
			}
			p.PublishInsert, p.PublishUpdate, p.PublishDelete, p.PublishTruncate = false, false, false, false
			for _, action := range actions {
				switch action {
				case "insert":
					p.PublishInsert = true
				case "update":
					p.PublishUpdate = true
				case "delete":
					p.PublishDelete = true
				case "truncate":
					p.PublishTruncate = true
				default:
					return p, pgerror.Newf(pgcode.Syntax, `unrecognized value for publication option "publish": "%s"`, action)
				}
			}
		case "publish_via_partition_root":
			var valid bool
			if option.Value == nil {
				p.PublishViaRoot, valid = true, true
			}
			switch v := option.Value.(type) {
			case *tree.DBool:
				p.PublishViaRoot, valid = bool(*v), true
			case *tree.NumVal:
				// A float literal such as 1.0 is not an integer def_arg.
				if !strings.ContainsAny(v.OrigString(), ".eE") {
					value, err := v.AsInt64()
					if err == nil && (value == 0 || value == 1) {
						p.PublishViaRoot, valid = value == 1, true
					}
				}
			case *tree.StrVal:
				switch strings.ToLower(v.RawString()) {
				case "true", "on":
					p.PublishViaRoot, valid = true, true
				case "false", "off":
					p.PublishViaRoot, valid = false, true
				}
			}
			if !valid {
				return p, pgerror.New(pgcode.Syntax, "publish_via_partition_root requires a Boolean value")
			}
		default:
			return p, pgerror.Newf(pgcode.Syntax, `unrecognized publication parameter: "%s"`, option.Key)
		}
	}
	return p, nil
}

// publicationActions parses a comma-separated SQL identifier list, as PostgreSQL's
// SplitIdentifierString does. Unquoted identifiers are folded; quoted ones retain case.
func publicationActions(value string) ([]string, error) {
	invalid := func() ([]string, error) {
		return nil, pgerror.New(pgcode.Syntax, `invalid list syntax in parameter "publish"`)
	}
	value = strings.TrimSpace(value)
	var actions []string
	for len(value) > 0 {
		var action string
		if value[0] == '"' {
			var b strings.Builder
			value = value[1:]
			closed := false
			for len(value) > 0 {
				if value[0] == '"' {
					if len(value) > 1 && value[1] == '"' {
						b.WriteByte('"')
						value = value[2:]
						continue
					}
					value = value[1:]
					closed = true
					break
				}
				b.WriteByte(value[0])
				value = value[1:]
			}
			if !closed || b.Len() == 0 {
				return invalid()
			}
			action = b.String()
		} else {
			end := strings.IndexAny(value, ", \t\n\r\f\v")
			if end < 0 {
				end = len(value)
			}
			if end == 0 {
				return invalid()
			}
			action, value = strings.ToLower(value[:end]), value[end:]
		}
		actions = append(actions, action)
		value = strings.TrimSpace(value)
		if len(value) == 0 {
			break
		}
		if value[0] != ',' {
			return invalid()
		}
		value = strings.TrimSpace(value[1:])
		if len(value) == 0 {
			return invalid()
		}
	}
	return actions, nil
}
