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
	"bytes"
	"context"
	"encoding/hex"
	"strings"
	"sync"
	"testing"

	"github.com/dolthub/go-mysql-server/sql"
	"github.com/jackc/pgx/v5/pgproto3"
	"github.com/stretchr/testify/require"

	"github.com/dolthub/doltgresql/server/functions"
	"github.com/dolthub/doltgresql/server/functions/framework"
	pgtypes "github.com/dolthub/doltgresql/server/types"
)

var max1RowLazyFunctionsOnce sync.Once

// TestResultForMax1RowIterLazyValues exercises native PostgreSQL text and binary
// encoders with values that must load their contents before Close cancels the query.
func TestResultForMax1RowIterLazyValues(t *testing.T) {
	max1RowLazyFunctionsOnce.Do(func() {
		functions.Init()
		framework.Initialize(nil)
	})

	textValue := strings.Repeat("out-of-band text ", 256)
	byteaValue := bytes.Repeat([]byte{0, 1, 127, 128, 255}, 1024)
	schema := sql.Schema{
		{Name: "text_value", Type: pgtypes.Text},
		{Name: "bytea_value", Type: pgtypes.Bytea},
	}

	for _, tt := range []struct {
		name      string
		format    int16
		wantBytea []byte
	}{
		{name: "text format", format: 0, wantBytea: []byte(`\x` + hex.EncodeToString(byteaValue))},
		{name: "binary format", format: 1, wantBytea: byteaValue},
	} {
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			sqlCtx := sql.NewContext(ctx)
			textWrapper := &max1RowLazyValue[string]{value: textValue}
			byteaWrapper := &max1RowLazyValue[[]byte]{value: byteaValue}
			iter := &max1RowTestIter{
				RowIter: sql.RowsToRowIter(sql.Row{textWrapper, byteaWrapper}),
				cancel:  cancel,
			}
			fields := []pgproto3.FieldDescription{
				{Name: []byte("text_value"), Format: tt.format},
				{Name: []byte("bytea_value"), Format: tt.format},
			}

			result, err := resultForMax1RowIter(sqlCtx, schema, iter, fields, []int16{tt.format})
			require.NoError(t, err)
			require.Equal(t, fields, result.Fields)
			require.Equal(t, uint64(1), result.RowsAffected)
			require.Equal(t, []Row{{[][]byte{[]byte(textValue), tt.wantBytea}}}, result.Rows)
			require.Positive(t, textWrapper.unwrapCalls)
			require.Positive(t, byteaWrapper.unwrapCalls)
			require.Equal(t, 1, iter.closeCalls)
			require.ErrorIs(t, ctx.Err(), context.Canceled)
		})
	}
}

// max1RowLazyValue requires a live context to load a value, making premature Close
// fail deterministically even when a storage-backed test would hit a warm cache.
type max1RowLazyValue[T string | []byte] struct {
	value       T
	unwrapCalls int
}

var _ sql.StringWrapper = (*max1RowLazyValue[string])(nil)
var _ sql.BytesWrapper = (*max1RowLazyValue[[]byte])(nil)

// Unwrap implements sql.Wrapper.
func (v *max1RowLazyValue[T]) Unwrap(ctx context.Context) (T, error) {
	v.unwrapCalls++
	return v.value, ctx.Err()
}

// UnwrapAny implements sql.AnyWrapper.
func (v *max1RowLazyValue[T]) UnwrapAny(ctx context.Context) (any, error) {
	return v.Unwrap(ctx)
}

// IsExactLength implements sql.AnyWrapper.
func (v *max1RowLazyValue[T]) IsExactLength() bool {
	return true
}

// MaxByteLength implements sql.AnyWrapper.
func (v *max1RowLazyValue[T]) MaxByteLength() int64 {
	return int64(len(v.value))
}

// Compare implements sql.AnyWrapper.
func (v *max1RowLazyValue[T]) Compare(context.Context, any) (int, bool, error) {
	return 0, false, nil
}

// Hash implements sql.AnyWrapper.
func (v *max1RowLazyValue[T]) Hash() any {
	return v
}
