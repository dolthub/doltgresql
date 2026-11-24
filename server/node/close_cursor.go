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

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/plan"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
)

// CloseCursor closes a cursor created by DECLARE. An empty name closes every cursor.
type CloseCursor struct {
	Name string
}

var _ sql.ExecSourceRel = (*CloseCursor)(nil)
var _ vitess.Injectable = (*CloseCursor)(nil)

// Children implements the interface sql.ExecSourceRel.
func (c *CloseCursor) Children() []sql.Node {
	return nil
}

// IsReadOnly implements the interface sql.ExecSourceRel.
func (c *CloseCursor) IsReadOnly() bool {
	return true
}

// Resolved implements the interface sql.ExecSourceRel.
func (c *CloseCursor) Resolved() bool {
	return true
}

// RowIter implements the interface sql.ExecSourceRel.
func (c *CloseCursor) RowIter(ctx *sql.Context, r sql.Row) (sql.RowIter, error) {
	if c.Name == "" {
		if err := core.CloseAllCursors(ctx); err != nil {
			return nil, err
		}
		return sql.RowsToRowIter(), nil
	}
	cursor, ok, err := core.GetCursor(ctx, c.Name)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, pgerror.Newf(pgcode.InvalidCursorName, `cursor "%s" does not exist`, c.Name)
	}
	if err = cursor.Close(ctx); err != nil {
		return nil, err
	}
	return sql.RowsToRowIter(), nil
}

// Schema implements the interface sql.ExecSourceRel.
func (c *CloseCursor) Schema(ctx *sql.Context) sql.Schema {
	return nil
}

// String implements the interface sql.ExecSourceRel.
func (c *CloseCursor) String() string {
	if c.Name == "" {
		return "CLOSE ALL"
	}
	return "CLOSE"
}

// WithChildren implements the interface sql.ExecSourceRel.
func (c *CloseCursor) WithChildren(ctx *sql.Context, children ...sql.Node) (sql.Node, error) {
	return plan.NillaryWithChildren(c, children...)
}

// WithResolvedChildren implements the interface vitess.Injectable.
func (c *CloseCursor) WithResolvedChildren(ctx context.Context, children []any) (any, error) {
	if len(children) != 0 {
		return nil, ErrVitessChildCount.New(0, len(children))
	}
	return c, nil
}
