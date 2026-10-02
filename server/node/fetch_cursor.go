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
	"io"

	"github.com/cockroachdb/errors"
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
	cursors, err := core.GetCursors(ctx)
	if err != nil {
		return nil, err
	}
	cursor, ok := cursors[f.Name]
	if !ok {
		return nil, pgerror.Newf(pgcode.InvalidCursorName, `cursor "%s" does not exist`, f.Name)
	}
	rows, moved, err := f.fetch(ctx, cursor)
	if err != nil {
		return nil, err
	}
	if f.IsMove {
		return sql.RowsToRowIter(sql.NewRow(types.NewOkResult(moved))), nil
	}
	return sql.RowsToRowIter(rows...), nil
}

// Schema implements the interface sql.ExecSourceRel.
func (f *FetchCursor) Schema(ctx *sql.Context) sql.Schema {
	if f.IsMove {
		return types.OkResultSchema
	}
	cursors, err := core.GetCursors(ctx)
	if err != nil {
		return nil
	}
	if cursor, ok := cursors[f.Name]; ok {
		return cursor.Schema
	}
	return nil
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

// fetch moves the cursor, returning the rows that it moved over along with their count.
func (f *FetchCursor) fetch(ctx *sql.Context, cursor *core.Cursor) (rows []sql.Row, moved int, err error) {
	keepRows := !f.IsMove
	count := f.Count
	forward := true
	switch f.Direction {
	case tree.FetchDirectionForward:
		forward = count >= 0
	case tree.FetchDirectionBackward:
		forward = count < 0
	case tree.FetchDirectionAbsolute:
		return fetchAbsolute(ctx, cursor, count, keepRows)
	case tree.FetchDirectionRelative:
		if count != 0 {
			return fetchRelative(ctx, cursor, count, keepRows)
		}
	}
	if count < 0 {
		count = -count
	}
	if count == 0 {
		isOnRow := cursor.Position > 0 && !isAfterLastRow(cursor)
		if f.IsMove {
			if isOnRow {
				return nil, 1, nil
			}
			return nil, 0, nil
		}
		if isOnRow {
			if _, _, err = runSelect(ctx, cursor, false, 1, false); err != nil {
				return nil, 0, err
			}
			count, forward = 1, true
		}
	}
	if f.IsMove && !forward && count == tree.FetchAll {
		if isAfterLastRow(cursor) {
			moved = cursor.RowsRead
		} else if cursor.Position > 0 {
			moved = cursor.Position - 1
		}
		return nil, moved, rewind(cursor)
	}
	return runSelect(ctx, cursor, forward, count, keepRows)
}

// fetchAbsolute moves the cursor to the row at `count`, counting backward from the last row when `count` is
// negative, and returns that row if it exists. A `count` of zero moves the cursor before the first row.
func fetchAbsolute(ctx *sql.Context, cursor *core.Cursor, count int64, keepRows bool) ([]sql.Row, int, error) {
	if count > 0 {
		position := int64(cursor.Position)
		if count <= position {
			if _, _, err := runSelect(ctx, cursor, false, position-count+1, false); err != nil {
				return nil, 0, err
			}
		} else if count > position+1 {
			if _, _, err := runSelect(ctx, cursor, true, count-position-1, false); err != nil {
				return nil, 0, err
			}
		}
		return runSelect(ctx, cursor, true, 1, keepRows)
	}
	if count < 0 {
		if _, _, err := runSelect(ctx, cursor, true, tree.FetchAll, false); err != nil {
			return nil, 0, err
		}
		if count < -1 {
			if _, _, err := runSelect(ctx, cursor, false, -count-1, false); err != nil {
				return nil, 0, err
			}
		}
		return runSelect(ctx, cursor, false, 1, keepRows)
	}
	return nil, 0, rewind(cursor)
}

// fetchRelative moves the cursor to the row that is `count` rows away from the current row, and returns that row if it
// exists.
func fetchRelative(ctx *sql.Context, cursor *core.Cursor, count int64, keepRows bool) ([]sql.Row, int, error) {
	if count > 0 {
		if count > 1 {
			if _, _, err := runSelect(ctx, cursor, true, count-1, false); err != nil {
				return nil, 0, err
			}
		}
		return runSelect(ctx, cursor, true, 1, keepRows)
	}
	if count < -1 {
		if _, _, err := runSelect(ctx, cursor, false, -count-1, false); err != nil {
			return nil, 0, err
		}
	}
	return runSelect(ctx, cursor, false, 1, keepRows)
}

// runSelect moves the cursor up to `count` rows forward or backward, stopping once it passes either end, and returns
// how many rows it moved over. When `keepRows` is true, it also returns those rows in the order that it moved over
// them. This mirrors PostgreSQL's PortalRunSelect.
func runSelect(ctx *sql.Context, cursor *core.Cursor, forward bool, count int64, keepRows bool) ([]sql.Row, int, error) {
	var rows []sql.Row
	moved := 0
	if forward {
		if count <= 0 || isAfterLastRow(cursor) {
			return nil, 0, nil
		}
		for int64(moved) < count {
			row, ok, err := readNextRow(ctx, cursor)
			if err != nil {
				return nil, 0, err
			}
			if !ok {
				cursor.Position = cursor.RowsRead + 1
				break
			}
			cursor.Position++
			moved++
			if keepRows {
				rows = append(rows, row)
			}
		}
		return rows, moved, nil
	}
	if !cursor.IsScrollable {
		return nil, 0, errScanForwardOnly()
	}
	if count <= 0 || cursor.Position == 0 {
		return nil, 0, nil
	}
	moved = int(min(int64(cursor.Position-1), count))
	if keepRows {
		rows = make([]sql.Row, moved)
		for i := range rows {
			rows[i] = cursor.Rows[cursor.Position-2-i]
		}
	}
	if count == tree.FetchAll || int64(moved) < count {
		cursor.Position = 0
	} else {
		cursor.Position -= moved
	}
	return rows, moved, nil
}

// readNextRow returns the row after the cursor's position, reading it from the cursor's query when the cursor has not
// read it yet. It returns false once there are no more rows.
func readNextRow(ctx *sql.Context, cursor *core.Cursor) (sql.Row, bool, error) {
	if cursor.Position < cursor.RowsRead {
		return cursor.Rows[cursor.Position], true, nil
	}
	if cursor.Iter == nil {
		return nil, false, nil
	}
	row, err := cursor.Iter.Next(ctx)
	if err == io.EOF {
		return nil, false, cursor.Close(ctx)
	}
	if err != nil {
		return nil, false, err
	}
	cursor.RowsRead++
	if cursor.IsScrollable {
		cursor.Rows = append(cursor.Rows, row)
	}
	return row, true, nil
}

// isAfterLastRow returns whether the cursor has moved past its last row.
func isAfterLastRow(cursor *core.Cursor) bool {
	return cursor.Iter == nil && cursor.Position > cursor.RowsRead
}

// rewind moves the cursor to before its first row.
func rewind(cursor *core.Cursor) error {
	if cursor.Position == 0 {
		return nil
	}
	if !cursor.IsScrollable {
		return errScanForwardOnly()
	}
	cursor.Position = 0
	return nil
}

// errScanForwardOnly returns the error for moving backward through a cursor that is not scrollable.
func errScanForwardOnly() error {
	return errors.WithHint(pgerror.New(pgcode.ObjectNotInPrerequisiteState, "cursor can only scan forward"),
		"Declare it with SCROLL option to enable backward scan.")
}
