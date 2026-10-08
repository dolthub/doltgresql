// Copyright 2025 Dolthub, Inc.
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
	coreextensions "github.com/dolthub/doltgresql/core/extensions"
	"github.com/dolthub/doltgresql/core/id"
	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
	"github.com/dolthub/doltgresql/server/extensions"
	"github.com/dolthub/doltgresql/server/extensions/extdef"
)

// DropExtension implements DROP EXTENSION.
type DropExtension struct {
	Names    []string
	IfExists bool
	Cascade  bool
}

var _ sql.ExecSourceRel = (*DropExtension)(nil)
var _ vitess.Injectable = (*DropExtension)(nil)

// NewDropExtension returns a new *DropExtension.
func NewDropExtension(names []string, ifExists bool, cascade bool) *DropExtension {
	return &DropExtension{
		Names:    names,
		IfExists: ifExists,
		Cascade:  cascade,
	}
}

// Children implements the interface sql.ExecSourceRel.
func (c *DropExtension) Children() []sql.Node {
	return nil
}

// IsReadOnly implements the interface sql.ExecSourceRel.
func (c *DropExtension) IsReadOnly() bool {
	return false
}

// Resolved implements the interface sql.ExecSourceRel.
func (c *DropExtension) Resolved() bool {
	return true
}

// RowIter implements the interface sql.ExecSourceRel.
func (c *DropExtension) RowIter(ctx *sql.Context, r sql.Row) (sql.RowIter, error) {
	extCollection, err := core.GetExtensionsCollectionFromContext(ctx, "")
	if err != nil {
		return nil, err
	}
	var loaded []coreextensions.Extension
	for _, name := range c.Names {
		ext, err := extCollection.GetLoadedExtension(ctx, id.NewExtension(name))
		if err != nil {
			return nil, err
		}
		if ext.ExtName.IsValid() {
			loaded = append(loaded, ext)
		} else if c.IfExists {
			dsess.DSessFromSess(ctx.Session).Notice(&pgproto3.NoticeResponse{
				Severity: "NOTICE",
				Message:  fmt.Sprintf(`extension "%s" does not exist, skipping`, name),
			})
		} else {
			return nil, pgerror.Newf(pgcode.UndefinedObject, `extension "%s" does not exist`, name)
		}
	}
	declarations := make([]*extdef.Extension, len(loaded))
	for i, ext := range loaded {
		if declarations[i], err = extensions.Get(ext.ExtName.Name()); err != nil {
			return nil, err
		}
		//TODO: drop the objects that depend on the extension when CASCADE is given
		if err = extensions.CheckDependents(ctx, declarations[i], ext.Namespace.SchemaName()); err != nil {
			return nil, err
		}
	}
	for i, ext := range loaded {
		if err = extensions.DropObjects(ctx, declarations[i], ext.Namespace.SchemaName()); err != nil {
			return nil, err
		}
		if err = extCollection.DropLoadedExtension(ctx, ext.ExtName); err != nil {
			return nil, err
		}
	}
	return sql.RowsToRowIter(), nil
}

// Schema implements the interface sql.ExecSourceRel.
func (c *DropExtension) Schema(ctx *sql.Context) sql.Schema {
	return nil
}

// String implements the interface sql.ExecSourceRel.
func (c *DropExtension) String() string {
	return "DROP EXTENSION"
}

// WithChildren implements the interface sql.ExecSourceRel.
func (c *DropExtension) WithChildren(ctx *sql.Context, children ...sql.Node) (sql.Node, error) {
	return plan.NillaryWithChildren(c, children...)
}

// WithResolvedChildren implements the interface vitess.Injectable.
func (c *DropExtension) WithResolvedChildren(ctx context.Context, children []any) (any, error) {
	if len(children) != 0 {
		return nil, ErrVitessChildCount.New(0, len(children))
	}
	return c, nil
}
