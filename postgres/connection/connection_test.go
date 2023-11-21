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
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgproto3"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/dolthub/doltgresql/postgres/connection"
	"github.com/dolthub/doltgresql/postgres/messages"
)

func TestReceive(t *testing.T) {
	// For these tests, we use an artificially small buffer size
	oldBufferSize := connection.BufferSize
	connection.BufferSize = 32
	defer func() {
		connection.BufferSize = oldBufferSize
	}()

	t.Run("Receive Query", func(t *testing.T) {
		serverConn, clientConn := getLocalHostConnection(t)
		defer clientConn.Close()
		defer serverConn.Close()

		message := &pgproto3.Query{
			String: "SELECT * FROM example",
		}
		encodedMessage := message.Encode(nil)

		// Write the encoded message to the mock connection's buffer
		_, err := clientConn.Write(encodedMessage)
		require.NoError(t, err)

		receivedMessage, err := connection.Receive(serverConn)
		require.NoError(t, err)

		receivedQuery, ok := receivedMessage.(messages.Query)
		require.True(t, ok, "Received message is not a Query type")

		require.Equal(t, "SELECT * FROM example", receivedQuery.String)
	})

	t.Run("Receive Query in multiple packets", func(t *testing.T) {
		serverConn, clientConn := getLocalHostConnection(t)
		defer clientConn.Close()
		defer serverConn.Close()

		message := &pgproto3.Query{
			String: "SELECT abc, def, ghi, jkl, mno, pqr, stuv, wxyz, abc, def, ghi, jkl, mno, pqr, stuv, wxyz FROM example",
		}
		encodedMessage := message.Encode(nil)
		_, err := clientConn.Write(encodedMessage[:len(encodedMessage)/2])
		require.NoError(t, err)

		wg := &sync.WaitGroup{}
		wg.Add(1)
		var receivedMessage connection.Message
		go func() {
			defer wg.Done()
			receivedMessage, err = connection.Receive(serverConn)
			require.NoError(t, err)
		}()

		// sleep a bit to make sure the goroutine is receiving (can't sync on it because it's blocked on the receive)
		time.Sleep(100 * time.Millisecond)
		_, err = clientConn.Write(encodedMessage[len(encodedMessage)/2:])
		require.NoError(t, err)

		wg.Wait()

		receivedQuery, ok := receivedMessage.(messages.Query)
		require.True(t, ok, "Received message is not a Query type")

		require.Equal(t, "SELECT abc, def, ghi, jkl, mno, pqr, stuv, wxyz, abc, def, ghi, jkl, mno, pqr, stuv, wxyz FROM example", receivedQuery.String)
	})

	t.Run("Receive multiple messages in one packet", func(t *testing.T) {
		serverConn, clientConn := getLocalHostConnection(t)
		defer clientConn.Close()
		defer serverConn.Close()

		queries := []*pgproto3.Query{
			{
				String: "SELECT abc123 FROM example",
			},
			{
				String: "SELECT def456 FROM example",
			},
			{
				String: "INSERT INTO example VALUES (1, 2, 3)",
			},
		}

		b := bytes.Buffer{}
		for _, message := range queries {
			encodedMessage := message.Encode(nil)
			b.Write(encodedMessage)
		}

		_, err := clientConn.Write(b.Bytes())
		require.NoError(t, err)

		messageCount := 0
		for _, message := range queries {
			receivedMessage, err := connection.Receive(serverConn)
			require.NoError(t, err)
			messageCount++

			receivedQuery, ok := receivedMessage.(messages.Query)
			require.True(t, ok, "Received message is not a Query type")
			assert.Equal(t, message.String, receivedQuery.String)
		}

		assert.Equal(t, len(queries), messageCount)
	})

}

func getLocalHostConnection(t *testing.T) (net.Conn, net.Conn) {
	ln, err := net.Listen("tcp", "localhost:0") // 0 for a randomly available port
	require.NoError(t, err)
	defer ln.Close()

	wg := &sync.WaitGroup{}
	wg.Add(2)

	var serverConn net.Conn
	go func() {
		defer wg.Done()
		serverConn, err = ln.Accept()
		require.NoError(t, err)
	}()

	var clientConn net.Conn
	go func() {
		defer wg.Done()
		clientConn, err = net.Dial("tcp", ln.Addr().String())
		require.NoError(t, err)
	}()

	wg.Wait()
	return serverConn, clientConn
}
