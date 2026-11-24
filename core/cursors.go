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

package core

import (
	"time"

	"github.com/cockroachdb/errors"
	"github.com/dolthub/go-mysql-server/sql"

	"github.com/dolthub/doltgresql/postgres/parser/pgcode"
	"github.com/dolthub/doltgresql/postgres/parser/pgerror"
	"github.com/dolthub/doltgresql/postgres/parser/sem/tree"
)

// Cursor is a cursor created by DECLARE, which holds every row of its query.
type Cursor struct {
	Name         string
	Statement    string
	Schema       sql.Schema
	IsHoldable   bool
	IsScrollable bool
	CreationTime time.Time
	rows         []sql.Row
	// position is 0 before the first row, the 1-based index of the current row, or one past the last row.
	position int
}

// NewCursor returns a cursor that holds the given rows.
func NewCursor(name string, statement string, schema sql.Schema, rows []sql.Row, isHoldable bool, isScrollable bool) *Cursor {
	return &Cursor{
		Name:         name,
		Statement:    statement,
		Schema:       schema,
		IsHoldable:   isHoldable,
		IsScrollable: isScrollable,
		CreationTime: time.Now(),
		rows:         rows,
	}
}

// AddCursor adds the cursor to the session's open cursors.
func AddCursor(ctx *sql.Context, cursor *Cursor) error {
	cv, err := getContextValues(ctx)
	if err != nil {
		return err
	}
	if _, ok := cv.cursors[cursor.Name]; ok {
		return pgerror.Newf(pgcode.DuplicateCursor, `cursor "%s" already exists`, cursor.Name)
	}
	if cv.cursors == nil {
		cv.cursors = make(map[string]*Cursor)
	}
	cv.cursors[cursor.Name] = cursor
	return nil
}

// GetCursor returns the session's open cursor with the given name, if one exists.
func GetCursor(ctx *sql.Context, name string) (*Cursor, bool, error) {
	cv, err := getContextValues(ctx)
	if err != nil {
		return nil, false, err
	}
	cursor, ok := cv.cursors[name]
	return cursor, ok, nil
}

// GetCursors returns every open cursor in the session.
func GetCursors(ctx *sql.Context) ([]*Cursor, error) {
	cv, err := getContextValues(ctx)
	if err != nil {
		return nil, err
	}
	cursors := make([]*Cursor, 0, len(cv.cursors))
	for _, cursor := range cv.cursors {
		cursors = append(cursors, cursor)
	}
	return cursors, nil
}

// CloseAllCursors closes every open cursor in the session.
func CloseAllCursors(ctx *sql.Context) error {
	cv, err := getContextValues(ctx)
	if err != nil {
		return err
	}
	clear(cv.cursors)
	return nil
}

// Close removes the cursor from the session's open cursors.
func (cursor *Cursor) Close(ctx *sql.Context) error {
	cv, err := getContextValues(ctx)
	if err != nil {
		return err
	}
	if cv.cursors[cursor.Name] == cursor {
		delete(cv.cursors, cursor.Name)
	}
	return nil
}

// Fetch moves the cursor as FETCH does, returning the rows that it moved over.
func (cursor *Cursor) Fetch(direction tree.FetchDirection, count int64) ([]sql.Row, error) {
	rows, _, err := cursor.fetch(direction, count, false)
	return rows, err
}

// Move moves the cursor as MOVE does, returning how many rows it moved over.
func (cursor *Cursor) Move(direction tree.FetchDirection, count int64) (int, error) {
	_, moved, err := cursor.fetch(direction, count, true)
	return moved, err
}

// fetch moves the cursor, returning the rows that it moved over along with their count, or only their count when
// `isMove` is true.
func (cursor *Cursor) fetch(direction tree.FetchDirection, count int64, isMove bool) (rows []sql.Row, moved int, err error) {
	returnRows := !isMove
	forward := true
	switch direction {
	case tree.FetchDirectionForward:
		forward = count >= 0
	case tree.FetchDirectionBackward:
		forward = count < 0
	case tree.FetchDirectionAbsolute:
		return cursor.fetchAbsolute(count, returnRows)
	case tree.FetchDirectionRelative:
		// RELATIVE 0 returns the current row, the same as FORWARD 0, so a count of 0 falls through to the code below.
		if count != 0 {
			return cursor.fetchRelative(count, returnRows)
		}
	}
	if count < 0 {
		count = -count
	}
	// A count of 0 returns the current row without moving, which is done by stepping back one row and then reading it
	// again. MOVE only reports whether there is a current row.
	if count == 0 {
		isOnRow := cursor.position > 0 && !cursor.isAfterLastRow()
		if isMove {
			if isOnRow {
				return nil, 1, nil
			}
			return nil, 0, nil
		}
		if isOnRow {
			if _, _, err = cursor.step(false, 1, false); err != nil {
				return nil, 0, err
			}
			count, forward = 1, true
		}
	}
	// MOVE BACKWARD ALL moves before the first row and reports how many rows it passed: every row when the cursor is
	// past the last row, or the rows before the current row otherwise. Like PostgreSQL, it rewinds the cursor, so a cursor
	// that is not scrollable may still do this when it is already before its first row.
	if isMove && !forward && count == tree.FetchAll {
		if cursor.isAfterLastRow() {
			moved = len(cursor.rows)
		} else if cursor.position > 0 {
			moved = cursor.position - 1
		}
		return nil, moved, cursor.rewind()
	}
	return cursor.step(forward, count, returnRows)
}

// fetchAbsolute moves the cursor to the row at `count`, counting backward from the last row when `count` is
// negative, and returns that row if it exists. A `count` of zero moves the cursor before the first row.
func (cursor *Cursor) fetchAbsolute(count int64, returnRows bool) ([]sql.Row, int, error) {
	if count > 0 {
		position := int64(cursor.position)
		if count <= position {
			if _, _, err := cursor.step(false, position-count+1, false); err != nil {
				return nil, 0, err
			}
		} else if count > position+1 {
			if _, _, err := cursor.step(true, count-position-1, false); err != nil {
				return nil, 0, err
			}
		}
		return cursor.step(true, 1, returnRows)
	}
	if count < 0 {
		if _, _, err := cursor.step(true, tree.FetchAll, false); err != nil {
			return nil, 0, err
		}
		if count < -1 {
			if _, _, err := cursor.step(false, -count-1, false); err != nil {
				return nil, 0, err
			}
		}
		return cursor.step(false, 1, returnRows)
	}
	return nil, 0, cursor.rewind()
}

// fetchRelative moves the cursor to the row that is `count` rows away from the current row, and returns that row if it
// exists.
func (cursor *Cursor) fetchRelative(count int64, returnRows bool) ([]sql.Row, int, error) {
	if count > 0 {
		if count > 1 {
			if _, _, err := cursor.step(true, count-1, false); err != nil {
				return nil, 0, err
			}
		}
		return cursor.step(true, 1, returnRows)
	}
	if count < -1 {
		if _, _, err := cursor.step(false, -count-1, false); err != nil {
			return nil, 0, err
		}
	}
	return cursor.step(false, 1, returnRows)
}

// step moves the cursor up to `count` rows forward or backward, one row at a time, stopping once it passes either end.
// It returns how many rows it moved over, along with those rows in the order that it moved over them when `returnRows`
// is true. Only scrollable cursors may move backward.
func (cursor *Cursor) step(forward bool, count int64, returnRows bool) ([]sql.Row, int, error) {
	delta, available := 1, len(cursor.rows)-cursor.position
	if !forward {
		if !cursor.IsScrollable {
			return nil, 0, errScanForwardOnly()
		}
		delta, available = -1, cursor.position-1
	}
	if count <= 0 || available < 0 {
		return nil, 0, nil
	}
	var rows []sql.Row
	if returnRows {
		rows = make([]sql.Row, 0, min(int64(available), count))
	}
	moved := 0
	for int64(moved) < count {
		next := cursor.position + delta
		if next < 1 || next > len(cursor.rows) {
			if forward {
				cursor.position = len(cursor.rows) + 1
			} else {
				cursor.position = 0
			}
			break
		}
		cursor.position = next
		moved++
		if returnRows {
			rows = append(rows, cursor.rows[next-1])
		}
	}
	return rows, moved, nil
}

// isAfterLastRow returns whether the cursor has moved past its last row.
func (cursor *Cursor) isAfterLastRow() bool {
	return cursor.position > len(cursor.rows)
}

// rewind moves the cursor to before its first row.
func (cursor *Cursor) rewind() error {
	if cursor.position == 0 {
		return nil
	}
	if !cursor.IsScrollable {
		return errScanForwardOnly()
	}
	cursor.position = 0
	return nil
}

// errScanForwardOnly returns the error for moving backward through a cursor that is not scrollable.
func errScanForwardOnly() error {
	return errors.WithHint(pgerror.New(pgcode.ObjectNotInPrerequisiteState, "cursor can only scan forward"),
		"Declare it with SCROLL option to enable backward scan.")
}
