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
)

// Cursor is a cursor created by DECLARE, which reads the rows of its query as they are fetched.
type Cursor struct {
	Name         string
	Statement    string
	Schema       sql.Schema
	IsHoldable   bool
	IsScrollable bool
	CreationTime time.Time
	// Iter returns the rows that the cursor has not read yet, and is nil once every row has been read.
	Iter sql.RowIter
	// Rows holds every row read from Iter, which only scrollable cursors keep so that they can move backward.
	Rows []sql.Row
	// RowsRead is the number of rows read from Iter.
	RowsRead int
	// Position is 0 before the first row, the 1-based index of the current row, or one past the last row.
	Position int
	// InTransaction is true until the transaction that declared the cursor ends.
	InTransaction bool
}

// GetCursors returns the session's open cursors, keyed by name.
func GetCursors(ctx *sql.Context) (map[string]*Cursor, error) {
	cv, err := getContextValues(ctx)
	if err != nil {
		return nil, err
	}
	if cv.cursors == nil {
		cv.cursors = make(map[string]*Cursor)
	}
	return cv.cursors, nil
}

// MaterializeHoldableCursors reads the remaining rows of every WITH HOLD cursor declared in the current transaction,
// so that the cursors can still be read once the transaction commits.
func MaterializeHoldableCursors(ctx *sql.Context) error {
	cursors, err := GetCursors(ctx)
	if err != nil {
		return err
	}
	for _, cursor := range cursors {
		if !cursor.IsHoldable || !cursor.InTransaction || cursor.Iter == nil {
			continue
		}
		rows, err := sql.RowIterToRows(ctx, cursor.Iter)
		if err != nil {
			cursor.Iter = nil
			return err
		}
		cursor.Iter = sql.RowsToRowIter(rows...)
	}
	return nil
}

// EndCursorTransaction closes the cursors that end with the current transaction. Cursors declared without WITH HOLD
// always close, while those declared WITH HOLD only close when the transaction that declared them rolls back.
func EndCursorTransaction(ctx *sql.Context, committed bool) error {
	cursors, err := GetCursors(ctx)
	if err != nil {
		return err
	}
	var closeErr error
	for name, cursor := range cursors {
		if !cursor.IsHoldable || (cursor.InTransaction && !committed) {
			closeErr = errors.CombineErrors(closeErr, cursor.Close(ctx))
			delete(cursors, name)
		} else {
			cursor.InTransaction = false
		}
	}
	return closeErr
}

// Close closes the iterator that the cursor reads its rows from.
func (cursor *Cursor) Close(ctx *sql.Context) error {
	if cursor.Iter == nil {
		return nil
	}
	err := cursor.Iter.Close(ctx)
	cursor.Iter = nil
	return err
}
