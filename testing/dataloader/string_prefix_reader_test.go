// Copyright 2024 Dolthub, Inc.
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

package _dataloader

import (
	"bytes"
	"io"
	"testing"

	"github.com/dolthub/doltgresql/core/dataloader"

	"github.com/stretchr/testify/require"
)

func TestStringPrefixReader(t *testing.T) {
	t.Run("Read prefix and all data in single call", func(t *testing.T) {
		prefix := "prefix"
		reader := bytes.NewReader([]byte("0123456789"))
		prefixReader := dataloader.NewStringPrefixReader(prefix, reader)

		data := make([]byte, 100)
		bytesRead, err := prefixReader.Read(data)
		require.NoError(t, err)
		require.Equal(t, 16, bytesRead)
		require.Equal(t, "prefix0123456789", string(data[:bytesRead]))

		bytesRead, err = prefixReader.Read(data)
		require.Equal(t, io.EOF, err)
		require.Equal(t, 0, bytesRead)
	})

	t.Run("Read part of prefix", func(t *testing.T) {
		prefix := "prefix"
		reader := bytes.NewReader([]byte("0123456789"))
		prefixReader := dataloader.NewStringPrefixReader(prefix, reader)

		data := make([]byte, 5)
		bytesRead, err := prefixReader.Read(data)
		require.NoError(t, err)
		require.Equal(t, 5, bytesRead)
		require.Equal(t, "prefi", string(data[:bytesRead]))

		// Read the next 5 bytes
		bytesRead, err = prefixReader.Read(data)
		require.NoError(t, err)
		require.Equal(t, 5, bytesRead)
		require.Equal(t, "x0123", string(data[:bytesRead]))

		// Read the next 5 bytes
		bytesRead, err = prefixReader.Read(data)
		require.NoError(t, err)
		require.Equal(t, 5, bytesRead)
		require.Equal(t, "45678", string(data[:bytesRead]))

		// Read the last byte
		bytesRead, err = prefixReader.Read(data)
		require.NoError(t, err)
		require.Equal(t, 1, bytesRead)
		require.Equal(t, "9", string(data[:bytesRead]))

		// Read EOF
		bytesRead, err = prefixReader.Read(data)
		require.Equal(t, io.EOF, err)
		require.Equal(t, 0, bytesRead)
	})

	t.Run("Read to prefix boundary", func(t *testing.T) {
		prefix := "prefix"
		reader := bytes.NewReader([]byte("0123456789"))
		prefixReader := dataloader.NewStringPrefixReader(prefix, reader)

		data := make([]byte, 6)
		bytesRead, err := prefixReader.Read(data)
		require.NoError(t, err)
		require.Equal(t, 6, bytesRead)
		require.Equal(t, "prefix", string(data[:bytesRead]))

		// Read the next 6 bytes
		bytesRead, err = prefixReader.Read(data)
		require.NoError(t, err)
		require.Equal(t, 6, bytesRead)
		require.Equal(t, "012345", string(data[:bytesRead]))

		// Read the next 6 bytes
		bytesRead, err = prefixReader.Read(data)
		require.NoError(t, err)
		require.Equal(t, 4, bytesRead)
		require.Equal(t, "6789", string(data[:bytesRead]))

		// Read EOF
		bytesRead, err = prefixReader.Read(data)
		require.Equal(t, io.EOF, err)
		require.Equal(t, 0, bytesRead)
	})
}
