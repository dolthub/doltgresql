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
	"encoding/base64"
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	driver "github.com/dolthub/doltgresql/integration-tests/go-sql-server-driver/driver"
)

// TestGCPreservesOpenTransaction keeps a dirty transaction idle while another
// connection checkpoints the same branch. Its private roots must survive GC even
// though they are not reachable from that checkpoint or the branch working set.
// This is a small, ordered reproducer; no stress workload or timing race is needed.
func TestGCPreservesOpenTransaction(t *testing.T) {
	for _, mode := range []string{"simple_protocol", "cache_statement"} {
		t.Run(mode, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(t.Context(), 20*time.Second)
			defer cancel()
			user, err := driver.NewDoltUser()
			require.NoError(t, err)
			t.Cleanup(func() { require.NoError(t, user.Cleanup()) })
			store, err := user.MakeRepoStore()
			require.NoError(t, err)
			const database = "gc_open_transaction"
			_, err = store.MakeRepo(database)
			require.NoError(t, err)
			require.NoError(t, (driver.WithFile{
				Name:     "server.yaml",
				Contents: "behavior:\n  auto_gc_behavior:\n    enable: false\n",
			}).WriteAtDir(store.Dir))
			server := StartServer(t, store, database, &driver.Server{
				Args: []string{"--config", "server.yaml"}, DynamicPort: "server_port",
			}, newPorts(t))
			db, err := server.DB(driver.Connection{DriverParams: map[string]string{
				"default_query_exec_mode": mode,
			}})
			require.NoError(t, err)
			t.Cleanup(func() { require.NoError(t, db.Close()) })
			_, err = db.ExecContext(ctx, "CREATE TABLE gc_probe (id TEXT PRIMARY KEY, body TEXT NOT NULL)")
			require.NoError(t, err)
			_, err = db.ExecContext(ctx, "INSERT INTO gc_probe VALUES ('seed', 'baseline')")
			require.NoError(t, err)
			_, err = db.ExecContext(ctx, "SELECT dolt_commit('-Am', 'baseline')")
			require.NoError(t, err)

			writer, err := db.Conn(ctx)
			require.NoError(t, err)
			defer writer.Close()
			peer, err := db.Conn(ctx)
			require.NoError(t, err)
			defer peer.Close()
			tx, err := writer.BeginTx(ctx, nil)
			require.NoError(t, err)
			defer tx.Rollback()
			// An out-of-line value makes collection of private transaction chunks
			// observable when the transaction resumes. It need not be large.
			bytes := make([]byte, 24000)
			_, err = rand.Read(bytes)
			require.NoError(t, err)
			body := base64.StdEncoding.EncodeToString(bytes)
			_, err = tx.ExecContext(ctx, "INSERT INTO gc_probe VALUES ($1, $2)", "writer", body)
			require.NoError(t, err)

			_, err = peer.ExecContext(ctx, "INSERT INTO gc_probe VALUES ('peer', 'committed independently')")
			require.NoError(t, err)
			_, err = peer.ExecContext(ctx, "SELECT dolt_add('-A')")
			require.NoError(t, err)
			_, err = peer.ExecContext(ctx, "SELECT dolt_commit('-m', 'peer checkpoint')")
			require.NoError(t, err)
			var privateRows int
			require.NoError(t, peer.QueryRowContext(ctx, "SELECT COUNT(*) FROM gc_probe WHERE id = 'writer'").Scan(&privateRows))
			require.Zero(t, privateRows, "the private transaction must not be included in the peer checkpoint")
			// Neither connection is running a command. The collector must visit
			// the writer's registered session and retain its transaction roots.
			_, err = db.ExecContext(ctx, "SELECT dolt_gc()")
			require.NoError(t, err)
			var actual string
			err = tx.QueryRowContext(ctx, "SELECT body FROM gc_probe WHERE id = 'writer'").Scan(&actual)
			require.NoError(t, err)
			require.Equal(t, body, actual)
			require.NoError(t, tx.Commit())

			// Check both the newly committed private value and the independent
			// peer change through a third connection after transaction merging.
			err = db.QueryRowContext(ctx, "SELECT body FROM gc_probe WHERE id = 'writer'").Scan(&actual)
			require.NoError(t, err)
			require.Equal(t, body, actual)
			var count int
			require.NoError(t, db.QueryRowContext(ctx, "SELECT COUNT(*) FROM gc_probe").Scan(&count))
			require.Equal(t, 3, count)
		})
	}
}
