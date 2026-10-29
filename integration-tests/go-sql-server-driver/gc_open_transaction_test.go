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
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	driver "github.com/dolthub/doltgresql/integration-tests/go-sql-server-driver/driver"
)

func TestGCPreservesOpenTransaction(t *testing.T) {
	t.Parallel()
	for _, mode := range []string{"simple_protocol", "cache_statement"} {
		for _, discard := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/discard=%v", mode, discard), func(t *testing.T) {
				t.Parallel()
				runGCPreservesOpenTransactionTest(t, mode, discard)
			})
		}
	}
}

// runGCPreservesOpenTransactionTest runs one variant of `TestGCPreservesOpenTransaction`, inserting a value large
// enough to be stored out of line, so that only the open transaction references its chunks.
func runGCPreservesOpenTransactionTest(t *testing.T, mode string, discard bool) {
	ctx := context.Background()
	u, err := driver.NewDoltUser()
	require.NoError(t, err)
	t.Cleanup(func() { u.Cleanup() })
	rs, err := u.MakeRepoStore()
	require.NoError(t, err)
	_, err = rs.MakeRepo("gc_open_transaction")
	require.NoError(t, err)
	server := StartServer(t, rs, "gc_open_transaction", &driver.Server{
		DynamicPort: "server",
	}, newPorts(t))
	db, err := server.DB(driver.Connection{DriverParams: map[string]string{"default_query_exec_mode": mode}})
	require.NoError(t, err)
	t.Cleanup(func() { db.Close() })

	_, err = db.ExecContext(ctx, "CREATE TABLE t (id TEXT PRIMARY KEY, body TEXT NOT NULL);")
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, "INSERT INTO t VALUES ('seed', 'seed');")
	require.NoError(t, err)
	_, err = db.ExecContext(ctx, "SELECT dolt_commit('-Am', 'seed');")
	require.NoError(t, err)

	writer, err := db.Conn(ctx)
	require.NoError(t, err)
	defer writer.Close()
	other, err := db.Conn(ctx)
	require.NoError(t, err)
	defer other.Close()
	if discard {
		_, err = writer.ExecContext(ctx, "DISCARD ALL;")
		require.NoError(t, err)
	}
	tx, err := writer.BeginTx(ctx, nil)
	require.NoError(t, err)
	defer tx.Rollback()
	body := strings.Repeat("0123456789abcdef", 1500)
	_, err = tx.ExecContext(ctx, "INSERT INTO t VALUES ('writer', $1);", body)
	require.NoError(t, err)

	_, err = other.ExecContext(ctx, "INSERT INTO t VALUES ('other', 'other');")
	require.NoError(t, err)
	_, err = other.ExecContext(ctx, "SELECT dolt_commit('-Am', 'other');")
	require.NoError(t, err)
	_, err = other.ExecContext(ctx, "SELECT dolt_gc();")
	require.NoError(t, err)

	var actual string
	require.NoError(t, tx.QueryRowContext(ctx, "SELECT body FROM t WHERE id = 'writer';").Scan(&actual))
	require.Equal(t, body, actual)
	require.NoError(t, tx.Commit())
	var count int
	require.NoError(t, other.QueryRowContext(ctx, "SELECT count(*) FROM t;").Scan(&count))
	require.Equal(t, 3, count)
	require.NoError(t, other.QueryRowContext(ctx, "SELECT body FROM t WHERE id = 'writer';").Scan(&actual))
	require.Equal(t, body, actual)
}
