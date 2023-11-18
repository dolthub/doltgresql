// Copyright 2023 Dolthub, Inc.
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

package connection_test

import (
	"bytes"
	"net"
	"testing"
	"time"

	"github.com/dolthub/doltgresql/postgres/connection"
	"github.com/dolthub/doltgresql/postgres/messages"
	"github.com/jackc/pgx/v5/pgproto3"
	"github.com/stretchr/testify/require"
)

func TestReceive(t *testing.T) {
	// For these tests, we use an artificially small buffer size
	oldBufferSize := connection.BufferSize
	connection.BufferSize = 32
	defer func() {
		connection.BufferSize = oldBufferSize
	}()

	t.Run("Receive Query", func(t *testing.T) {
		mockBuffer := bytes.NewBuffer([]byte{})
		mockConn := &MockConn{buffer: mockBuffer}

		message := &pgproto3.Query{
			String: "SELECT * FROM example",
		}
		encodedMessage := message.Encode(nil)

		// Write the encoded message to the mock connection's buffer
		_, err := mockConn.Write(encodedMessage)
		require.NoError(t, err)

		receivedMessage, err := connection.Receive(mockConn)
		require.NoError(t, err)

		receivedQuery, ok := receivedMessage.(messages.Query)
		require.True(t, ok, "Received message is not a Query type")

		require.Equal(t, "SELECT * FROM example", receivedQuery.String)
	})

	t.Run("Receive Query larger than buffer", func(t *testing.T) {
		mockBuffer := bytes.NewBuffer([]byte{})
		mockConn := &MockConn{buffer: mockBuffer}
		
		message := &pgproto3.Query{
			String: "SELECT abc, def, ghi, jkl, mno, pqr, stuv, wxyz, abc, def, ghi, jkl, mno, pqr, stuv, wxyz FROM example",
		}
		encodedMessage := message.Encode(nil)
		_, err := mockConn.Write(encodedMessage)
		require.NoError(t, err)

		receivedMessage, err := connection.Receive(mockConn)
		require.NoError(t, err)

		receivedQuery, ok := receivedMessage.(messages.Query)
		require.True(t, ok, "Received message is not a Query type")

		require.Equal(t, "SELECT abc, def, ghi, jkl, mno, pqr, stuv, wxyz, abc, def, ghi, jkl, mno, pqr, stuv, wxyz FROM example", receivedQuery.String)
	})
}

// MockConn is a simple in-memory implementation of net.Conn for testing purposes.
type MockConn struct {
	buffer *bytes.Buffer
}

func (m *MockConn) Read(b []byte) (n int, err error) {
	return m.buffer.Read(b)
}

func (m *MockConn) Write(b []byte) (n int, err error) {
	return m.buffer.Write(b)
}

func (m *MockConn) Close() error {
	return nil
}

func (m *MockConn) LocalAddr() net.Addr {
	return nil
}

func (m *MockConn) RemoteAddr() net.Addr {
	return nil
}

func (m *MockConn) SetDeadline(t time.Time) error {
	return nil
}

func (m *MockConn) SetReadDeadline(t time.Time) error {
	return nil
}

func (m *MockConn) SetWriteDeadline(t time.Time) error {
	return nil
}