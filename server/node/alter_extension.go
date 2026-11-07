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

	"github.com/dolthub/dolt/go/libraries/doltcore/sqle/dsess"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/plan"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"
	"github.com/jackc/pgx/v5/pgproto3"

	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
	pgexprs "github.com/dolthub/doltgresql/server/expression"
	"github.com/dolthub/doltgresql/server/extensions"
)

// AlterExtension implements ALTER EXTENSION ... UPDATE, which rewrites an installed extension's objects to match the
// version of the extension that this server emulates. Version is empty when no version is given.
type AlterExtension struct {
	Name    string
	Version string
	Runner  pgexprs.StatementRunner
}

var _ sql.ExecSourceRel = (*AlterExtension)(nil)
var _ sql.Expressioner = (*AlterExtension)(nil)
var _ vitess.Injectable = (*AlterExtension)(nil)

// NewAlterExtension returns a new *AlterExtension.
func NewAlterExtension(name string, version string) *AlterExtension {
	return &AlterExtension{
		Name:    name,
		Version: version,
	}
}

// Children implements the interface sql.ExecSourceRel.
func (c *AlterExtension) Children() []sql.Node {
	return nil
}

// Expressions implements the interface sql.Expressioner.
func (c *AlterExtension) Expressions() []sql.Expression {
	return []sql.Expression{c.Runner}
}

// IsReadOnly implements the interface sql.ExecSourceRel.
func (c *AlterExtension) IsReadOnly() bool {
	return false
}

// Resolved implements the interface sql.ExecSourceRel.
func (c *AlterExtension) Resolved() bool {
	return true
}

// RowIter implements the interface sql.ExecSourceRel.
func (c *AlterExtension) RowIter(ctx *sql.Context, r sql.Row) (sql.RowIter, error) {
	extCollection, err := core.GetExtensionsCollectionFromContext(ctx, "")
	if err != nil {
		return nil, err
	}
	installed, err := extCollection.GetLoadedExtension(ctx, id.NewExtension(c.Name))
	if err != nil {
		return nil, err
	}
	if !installed.ExtName.IsValid() {
		return nil, pgerror.Newf(pgcode.UndefinedObject, `extension "%s" does not exist`, c.Name)
	}
	ext, err := extensions.Get(c.Name)
	if err != nil {
		return nil, err
	}
	version := ext.Control.DefaultVersion
	if c.Version != "" && c.Version != version {
		return nil, pgerror.Newf(pgcode.InvalidParameterValue, `extension "%s" has no update path from version "%s" to version "%s"`,
			c.Name, installed.Version, c.Version)
	}
	if err = extensions.UpdateObjects(ctx, c.Runner.Runner, ext, installed.Namespace.SchemaName()); err != nil {
		return nil, err
	}
	if installed.Version == version {
		dsess.DSessFromSess(ctx.Session).Notice(&pgproto3.NoticeResponse{
			Severity: "NOTICE",
			Message:  fmt.Sprintf(`version "%s" of extension "%s" is already installed`, version, c.Name),
		})
		return sql.RowsToRowIter(), nil
	}
	if err = extCollection.DropLoadedExtension(ctx, installed.ExtName); err != nil {
		return nil, err
	}
	installed.Version = version
	if err = extCollection.AddLoadedExtension(ctx, installed); err != nil {
		return nil, err
	}
	return sql.RowsToRowIter(), nil
}

// Schema implements the interface sql.ExecSourceRel.
func (c *AlterExtension) Schema(ctx *sql.Context) sql.Schema {
	return nil
}

// String implements the interface sql.ExecSourceRel.
func (c *AlterExtension) String() string {
	return "ALTER EXTENSION"
}

// WithChildren implements the interface sql.ExecSourceRel.
func (c *AlterExtension) WithChildren(ctx *sql.Context, children ...sql.Node) (sql.Node, error) {
	return plan.NillaryWithChildren(c, children...)
}

// WithResolvedChildren implements the interface vitess.Injectable.
func (c *AlterExtension) WithResolvedChildren(ctx context.Context, children []any) (any, error) {
	if len(children) != 0 {
		return nil, ErrVitessChildCount.New(0, len(children))
	}
	return c, nil
}

// WithExpressions implements the interface sql.Expressioner.
func (c *AlterExtension) WithExpressions(ctx *sql.Context, expressions ...sql.Expression) (sql.Node, error) {
	if len(expressions) != 1 {
		return nil, sql.ErrInvalidChildrenNumber.New(c, len(expressions), 1)
	}
	newC := *c
	newC.Runner = expressions[0].(pgexprs.StatementRunner)
	return &newC, nil
}
