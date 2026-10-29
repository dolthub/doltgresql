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

package main

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/stdlib"
	"github.com/stretchr/testify/require"

	driver "github.com/dolthub/doltgresql/integration-tests/go-sql-server-driver/driver"
)

// TestGCProtocolPaths runs a GC while a connection is between the protocol
// messages of a multi-message operation: a COPY FROM STDIN, an open cursor,
// and a session replaced by DISCARD ALL. Each operation's state must survive.
func TestGCProtocolPaths(t *testing.T) {
	t.Parallel()
	ports := newPorts(t)
	u, err := driver.NewDoltUser()
	require.NoError(t, err)
	t.Cleanup(func() { u.Cleanup() })
	rs, err := u.MakeRepoStore()
	require.NoError(t, err)
	const dbName = "gc_protocol_paths"
	_, err = rs.MakeRepo(dbName)
	require.NoError(t, err)
	require.NoError(t, (driver.WithFile{
		Name:     "server.yaml",
		Contents: "behavior:\n  auto_gc_behavior:\n    enable: false\n",
	}).WriteAtDir(rs.Dir))
	server := StartServer(t, rs, dbName, &driver.Server{
		Args:          []string{"--config", "server.yaml"},
		DynamicPort:   "server_port",
		Envs:          []string{"DOLT_GC_SAFEPOINT_CONTROLLER_CHOICE=session_aware"},
		LogNotMatches: []string{"recovered panic", "caught panic"},
	}, ports)
	db, err := server.DB(driver.Connection{})
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })
	ctx, cancel := context.WithTimeout(t.Context(), 60*time.Second)
	defer cancel()

	_, err = db.ExecContext(ctx, "CREATE TABLE vals (id INT PRIMARY KEY, body TEXT NOT NULL)")
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, "SELECT dolt_commit('-Am', 'create vals')")
	require.NoError(t, err)

	// Commits on another connection and collects, so that only the roots held by
	// the connection under test keep its uncommitted state alive.
	checkpointAndGC := func(id int) {
		_, err := db.ExecContext(ctx, "INSERT INTO vals VALUES ($1::int4, 'peer')", id)
		require.NoError(t, err)
		_, err = db.ExecContext(ctx, "SELECT dolt_commit('-am', 'peer checkpoint')")
		require.NoError(t, err)
		_, err = db.ExecContext(ctx, "SELECT dolt_gc()")
		require.NoError(t, err)
	}
	// Out-of-line values make collection of private chunks observable.
	body := func() string {
		bs := make([]byte, 12000)
		_, err := rand.Read(bs)
		require.NoError(t, err)
		return hex.EncodeToString(bs)
	}

	t.Run("CopyFromStdin", func(t *testing.T) {
		conn, err := db.Conn(ctx)
		require.NoError(t, err)
		defer conn.Close()
		_, err = conn.ExecContext(ctx, "BEGIN")
		require.NoError(t, err)

		pr, pw := io.Pipe()
		copyDone := make(chan error, 1)
		go func() {
			copyDone <- conn.Raw(func(dc any) error {
				_, err := dc.(*stdlib.Conn).Conn().PgConn().CopyFrom(ctx, pr, "COPY vals (id, body) FROM STDIN")
				return err
			})
		}()
		var want []string
		writeRows := func(from, to int) {
			var sb strings.Builder
			for i := from; i < to; i++ {
				b := body()
				want = append(want, b)
				fmt.Fprintf(&sb, "%d\t%s\n", i, b)
			}
			_, err := pw.Write([]byte(sb.String()))
			require.NoError(t, err)
		}
		writeRows(1000, 1004)
		// Give the server time to load the first CopyData message.
		time.Sleep(500 * time.Millisecond)
		checkpointAndGC(1)
		writeRows(1004, 1008)
		require.NoError(t, pw.Close())
		require.NoError(t, <-copyDone)
		_, err = conn.ExecContext(ctx, "COMMIT")
		require.NoError(t, err)

		rows, err := db.QueryContext(ctx, "SELECT body FROM vals WHERE id >= 1000 AND id < 1008 ORDER BY id")
		require.NoError(t, err)
		var got []string
		for rows.Next() {
			var b string
			require.NoError(t, rows.Scan(&b))
			got = append(got, b)
		}
		require.NoError(t, rows.Err())
		require.Equal(t, want, got)
	})

	t.Run("Cursor", func(t *testing.T) {
		var want []string
		for i := 2000; i < 2008; i++ {
			b := body()
			want = append(want, b)
			_, err := db.ExecContext(ctx, "INSERT INTO vals VALUES ($1::int4, $2)", i, b)
			require.NoError(t, err)
		}
		conn, err := db.Conn(ctx)
		require.NoError(t, err)
		defer conn.Close()
		_, err = conn.ExecContext(ctx, "BEGIN")
		require.NoError(t, err)
		// The cursor reads rows which are in neither the peer's commits nor the
		// branch working set once they are deleted below.
		_, err = conn.ExecContext(ctx, "DECLARE c CURSOR FOR SELECT body FROM vals WHERE id >= 2000 AND id < 2008 ORDER BY id")
		require.NoError(t, err)
		fetch := func() []string {
			rows, err := conn.QueryContext(ctx, "FETCH 4 FROM c")
			require.NoError(t, err)
			var got []string
			for rows.Next() {
				var b string
				require.NoError(t, rows.Scan(&b))
				got = append(got, b)
			}
			require.NoError(t, rows.Err())
			return got
		}
		got := fetch()
		_, err = db.ExecContext(ctx, "DELETE FROM vals WHERE id >= 2000 AND id < 2008")
		require.NoError(t, err)
		checkpointAndGC(2)
		got = append(got, fetch()...)
		_, err = conn.ExecContext(ctx, "COMMIT")
		require.NoError(t, err)
		require.Equal(t, want, got)
	})

	t.Run("DiscardAll", func(t *testing.T) {
		conn, err := db.Conn(ctx)
		require.NoError(t, err)
		defer conn.Close()
		_, err = conn.ExecContext(ctx, "DISCARD ALL")
		require.NoError(t, err)
		_, err = conn.ExecContext(ctx, "BEGIN")
		require.NoError(t, err)
		b := body()
		_, err = conn.ExecContext(ctx, "INSERT INTO vals VALUES (3000, $1)", b)
		require.NoError(t, err)
		checkpointAndGC(3)
		var got string
		require.NoError(t, conn.QueryRowContext(ctx, "SELECT body FROM vals WHERE id = 3000").Scan(&got))
		require.Equal(t, b, got)
		_, err = conn.ExecContext(ctx, "COMMIT")
		require.NoError(t, err)
	})

	var count int
	require.NoError(t, db.QueryRowContext(ctx, "SELECT COUNT(*) FROM vals").Scan(&count))
	require.Equal(t, 8+3+1, count)
}
