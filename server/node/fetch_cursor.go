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
	"github.com/dolthub/go-mysql-server/sql/types"
	vitess "github.com/dolthub/vitess/go/vt/sqlparser"

	"github.com/dolthub/doltgresql/core"
	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
)

// FetchCursor moves through a cursor created by DECLARE. FETCH returns the rows that it moves over, while MOVE only
// returns how many rows it moved over.
type FetchCursor struct {
	Name      string
	Direction tree.FetchDirection
	Count     int64
	IsMove    bool
}

var _ sql.ExecSourceRel = (*FetchCursor)(nil)
var _ vitess.Injectable = (*FetchCursor)(nil)

// Children implements the interface sql.ExecSourceRel.
func (f *FetchCursor) Children() []sql.Node {
	return nil
}

// IsReadOnly implements the interface sql.ExecSourceRel.
func (f *FetchCursor) IsReadOnly() bool {
	return true
}

// Resolved implements the interface sql.ExecSourceRel.
func (f *FetchCursor) Resolved() bool {
	return true
}

// RowIter implements the interface sql.ExecSourceRel.
func (f *FetchCursor) RowIter(ctx *sql.Context, r sql.Row) (sql.RowIter, error) {
	cursor, ok, err := core.GetCursor(ctx, f.Name)
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, pgerror.Newf(pgcode.InvalidCursorName, `cursor "%s" does not exist`, f.Name)
	}
	if f.IsMove {
		moved, err := cursor.Move(f.Direction, f.Count)
		if err != nil {
			return nil, err
		}
		return sql.RowsToRowIter(sql.NewRow(types.NewOkResult(moved))), nil
	}
	rows, err := cursor.Fetch(f.Direction, f.Count)
	if err != nil {
		return nil, err
	}
	return sql.RowsToRowIter(rows...), nil
}

// Schema implements the interface sql.ExecSourceRel.
func (f *FetchCursor) Schema(ctx *sql.Context) sql.Schema {
	if f.IsMove {
		return types.OkResultSchema
	}
	cursor, ok, err := core.GetCursor(ctx, f.Name)
	if err != nil || !ok {
		return nil
	}
	return cursor.Schema
}

// String implements the interface sql.ExecSourceRel.
func (f *FetchCursor) String() string {
	if f.IsMove {
		return "MOVE"
	}
	return "FETCH"
}

// WithChildren implements the interface sql.ExecSourceRel.
func (f *FetchCursor) WithChildren(ctx *sql.Context, children ...sql.Node) (sql.Node, error) {
	return plan.NillaryWithChildren(f, children...)
}

// WithResolvedChildren implements the interface vitess.Injectable.
func (f *FetchCursor) WithResolvedChildren(ctx context.Context, children []any) (any, error) {
	if len(children) != 0 {
		return nil, ErrVitessChildCount.New(0, len(children))
	}
	return f, nil
}
