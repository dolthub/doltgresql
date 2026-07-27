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

package utils

import (
	"testing"

	"github.com/stretchr/testify/require"
)

// TestWireRWReadPastEndPanicsWithDiagnostic asserts that reading more bytes than are available
// panics with a message naming the requested length, offset, and available bytes, rather than a
// generic Go slice-bounds error.
func TestWireRWReadPastEndPanicsWithDiagnostic(t *testing.T) {
	t.Run("ReadUint32 past end", func(t *testing.T) {
		rw := NewWireReader([]byte{1, 2, 3})
		require.PanicsWithValue(t,
			"wire reader: attempted to read 4 bytes at offset 0, but only 3 bytes are available",
			func() { rw.ReadUint32() })
	})
	t.Run("ReadUint64 past end", func(t *testing.T) {
		rw := NewWireReader([]byte{1, 2, 3, 4})
		require.PanicsWithValue(t,
			"wire reader: attempted to read 8 bytes at offset 0, but only 4 bytes are available",
			func() { rw.ReadUint64() })
	})
	t.Run("ReadBytes exactly at end succeeds", func(t *testing.T) {
		rw := NewWireReader([]byte{1, 2, 3, 4})
		require.Equal(t, []byte{1, 2, 3, 4}, rw.ReadBytes(4))
	})
}

// TestWireRWReadBytesWrappedLengthDoesNotOverflow is a regression test for a near-maximum
// declared length (0xffffffff) wrapping the uint32 readIdx+n bounds check around to a small
// value and not correctly detecting the invalid data.
func TestWireRWReadBytesWrappedLengthDoesNotOverflow(t *testing.T) {
	rw := NewWireReader([]byte{1, 2, 3, 4, 5})
	// Consume one valid prefix byte so readIdx is nonzero
	require.Equal(t, uint8(1), rw.ReadUint8())

	require.PanicsWithValue(t,
		"wire reader: attempted to read 4294967295 bytes at offset 1, but only 4 bytes are available",
		func() { rw.ReadBytes(4294967295) })
}
