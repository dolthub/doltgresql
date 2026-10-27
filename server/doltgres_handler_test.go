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

package server

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/types"
	"github.com/dolthub/vitess/go/sqltypes"
	"github.com/jackc/pgx/v5/pgproto3"
	"github.com/stretchr/testify/require"
)

// TestResultForMax1RowIter verifies that encoding can still use the query context,
// and that every exit closes the iterator without hiding an earlier error.
func TestResultForMax1RowIter(t *testing.T) {
	readErr := errors.New("read failed")
	encodeErr := errors.New("encoding failed")
	closeErr := errors.New("close failed")
	value := strings.Repeat("out-of-band text ", 256)
	fields := []pgproto3.FieldDescription{{Name: []byte("value")}}

	for _, tt := range []struct {
		name      string
		rows      []sql.Row
		nextErrAt int
		nextErr   error
		encodeErr error
		closeErr  error
		wantErr   error
		wantText  string
	}{
		{name: "lazy value", rows: []sql.Row{{value}}},
		{name: "no rows"},
		{name: "null value", rows: []sql.Row{{nil}}},
		{name: "first read error", nextErrAt: 1, nextErr: readErr, wantErr: readErr},
		{name: "second read error", rows: []sql.Row{{value}}, nextErrAt: 2, nextErr: readErr, wantText: "result max1Row iterator returned more than one row"},
		{name: "extra row", rows: []sql.Row{{value}, {value}}, wantText: "result max1Row iterator returned more than one row"},
		{name: "encoding error", rows: []sql.Row{{value}}, encodeErr: encodeErr, wantErr: encodeErr},
		{name: "close error", rows: []sql.Row{{value}}, closeErr: closeErr, wantErr: closeErr},
		{name: "empty close error", closeErr: closeErr, wantErr: closeErr},
		{name: "read error before close error", nextErrAt: 1, nextErr: readErr, closeErr: closeErr, wantErr: readErr},
		{name: "extra row before close error", rows: []sql.Row{{value}, {value}}, closeErr: closeErr, wantText: "result max1Row iterator returned more than one row"},
		{name: "encoding error before close error", rows: []sql.Row{{value}}, encodeErr: encodeErr, closeErr: closeErr, wantErr: encodeErr},
	} {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			sqlCtx := sql.NewContext(ctx)
			iter := &max1RowTestIter{
				RowIter: sql.RowsToRowIter(tt.rows...), cancel: cancel,
				nextErrAt: tt.nextErrAt, nextErr: tt.nextErr, closeErr: tt.closeErr,
			}
			schema := sql.Schema{{Name: "value", Type: max1RowTestType{max1RowBaseType: types.LongText, encodeErr: tt.encodeErr}}}

			result, err := resultForMax1RowIter(sqlCtx, schema, iter, fields, []int16{0})
			require.Equal(t, 1, iter.closeCalls)
			require.ErrorIs(t, ctx.Err(), context.Canceled)
			if tt.wantErr != nil {
				require.ErrorIs(t, err, tt.wantErr)
				require.Nil(t, result)
			} else if tt.wantText != "" {
				require.EqualError(t, err, tt.wantText)
				require.Nil(t, result)
			} else {
				require.NoError(t, err)
				require.Equal(t, fields, result.Fields)
				require.Equal(t, uint64(len(tt.rows)), result.RowsAffected)
				require.Len(t, result.Rows, len(tt.rows))
				if len(tt.rows) == 1 {
					var expected []byte
					if tt.rows[0][0] != nil {
						expected = []byte(value)
					}
					require.Equal(t, []Row{{[][]byte{expected}}}, result.Rows)
				}
			}
		})
	}
}

// max1RowTestType models a lazy value's need for a live context during encoding.
// Unlike a storage-backed fixture, a warm node cache cannot mask premature Close.
type max1RowTestType struct {
	max1RowBaseType
	encodeErr error
}

// max1RowBaseType avoids shadowing sql.Type's Type method when embedded.
type max1RowBaseType = sql.Type

// SQL implements sql.Type.
func (t max1RowTestType) SQL(ctx *sql.Context, dest []byte, value interface{}) (sqltypes.Value, error) {
	if err := ctx.Err(); err != nil {
		return sqltypes.Value{}, err
	}
	if t.encodeErr != nil {
		return sqltypes.Value{}, t.encodeErr
	}
	return t.max1RowBaseType.SQL(ctx, dest, value)
}

// max1RowTestIter models TrackedRowIter.Close ending the query context.
type max1RowTestIter struct {
	sql.RowIter
	cancel     context.CancelFunc
	nextCalls  int
	nextErrAt  int
	nextErr    error
	closeCalls int
	closeErr   error
}

// Next implements sql.RowIter.
func (i *max1RowTestIter) Next(ctx *sql.Context) (sql.Row, error) {
	i.nextCalls++
	if i.nextCalls == i.nextErrAt {
		return nil, i.nextErr
	}
	return i.RowIter.Next(ctx)
}

// Close implements sql.RowIter.
func (i *max1RowTestIter) Close(ctx *sql.Context) error {
	i.closeCalls++
	i.cancel()
	if err := i.RowIter.Close(ctx); err != nil {
		return err
	}
	return i.closeErr
}
